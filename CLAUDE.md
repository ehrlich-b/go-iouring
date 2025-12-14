# go-iouring Project Guide

## Project Overview

A definitive, pure Go implementation of Linux io_uring targeting comprehensive operation coverage, high performance, and multi-kernel support.

**Goals:**
- Complete coverage of all 57+ io_uring operations
- Zero-allocation hot paths (no channels, no interface{})
- Support for kernel 6.8+ initially, with 5.15 LTS as fallback baseline
- Systematic synchronization with kernel headers via code generation
- Reference implementations (cp, cat, echo server, etc.)

## Build & Test Commands

**All commands must be run via make.** See `make help` for full list.

```bash
# Build all packages
make build

# Test (requires Linux with io_uring support)
make test

# Test with verbose output
make test-v

# Run specific test
make test-run TEST=TestRingSetup

# Benchmarks
make bench
make bench-count COUNT=5

# Coverage
make cover          # Generate coverage report
make cover-html     # Open coverage in browser

# Code quality
make fmt            # Format code
make vet            # Run go vet
make lint           # Run staticcheck (install with: make tools)
make test-race      # Run tests with race detector

# Generate code from kernel headers (when implemented)
make generate

# Check kernel io_uring support
make check-iouring

# Module management
make mod-tidy       # Tidy dependencies
make mod-verify     # Verify dependencies

# Clean
make clean
```

## Project Structure

```
go-iouring/
├── ring.go              # Core Ring type and setup/teardown
├── sqe.go               # Submission Queue Entry building
├── cqe.go               # Completion Queue Entry handling
├── ops.go               # Operation constants (IORING_OP_*)
├── flags.go             # All flag definitions
├── register.go          # Buffer/file registration
├── probe.go             # Feature detection (IORING_REGISTER_PROBE)
├── errors.go            # Error types
├── internal/
│   ├── sys/             # Raw syscall wrappers
│   │   ├── syscall.go   # io_uring_setup, io_uring_enter, io_uring_register
│   │   ├── types.go     # Kernel struct definitions (SQE, CQE, params)
│   │   └── consts.go    # Auto-generated constants
│   └── atomic/          # Memory barrier helpers
├── gen/                 # Code generation from kernel headers
├── examples/
│   ├── cp/              # cp using io_uring
│   ├── cat/             # cat using io_uring
│   └── echoserver/      # High-perf TCP echo server
└── testutil/            # Test helpers
```

## Critical Architecture Decisions

### No CGO
Pure Go syscalls only. Avoids CGO overhead and simplifies cross-compilation.

### No Per-Operation Allocations
This is the core differentiator from existing libraries:
- Pre-allocated SQE/CQE arrays via mmap (kernel requirement)
- User data passed as `uint64`, NOT `interface{}` or channels
- Completion notification via callback or polling, NEVER channels
- Use sync.Pool for any unavoidable allocations

### User Data Strategy (The "Ticket System")
```go
// User maintains their own context mapping
// UserData is a uint64 index, not a pointer
func (r *Ring) SubmitRead(fd int, buf []byte, offset uint64, userData uint64) error

// No allocation on peek
func (r *Ring) PeekCQE() (userData uint64, res int32, flags uint32, ok bool)
```

### Memory Barriers
Go's sync/atomic provides acquire/release semantics:
```go
// SQ tail update (producer) - use StoreUint32 with release semantics
atomic.StoreUint32(sq.tail, newTail)

// CQ head read (consumer) - use LoadUint32 with acquire semantics
head := atomic.LoadUint32(cq.head)
```

### Go Runtime Interaction Risks

1. **Blocking io_uring_enter**: Calling with IORING_ENTER_GETEVENTS blocks the OS thread. Use `syscall.Syscall6` (not RawSyscall) to properly notify the Go scheduler.

2. **Memory Pinning**: Buffers passed to the kernel MUST be heap-allocated. Stack variables can move during GC. Pre-allocate buffer arenas for safety.

3. **SQPOLL Thread Contention**: If Go schedules goroutines onto the SQPOLL CPU, performance degrades. Consider CPU affinity.

### Eventfd Integration
For goroutine parking without burning CPU:
```go
// Register eventfd with the ring
fd, _ := syscall.Eventfd(0, syscall.EFD_NONBLOCK|syscall.EFD_CLOEXEC)
ring.RegisterEventfd(fd)

// Use Go's runtime poller or select on eventfd for wakeup
```

## API Layers

1. **Low-level (internal/sys)**: Direct syscall wrappers, mirrors kernel interface exactly
2. **Core (ring.go, sqe.go, cqe.go)**: Ring management, SQE preparation, CQE consumption
3. **Operations**: Type-safe Prep* functions for each operation
4. **High-level (optional)**: Convenience wrappers like SubmitAndWait with context

## Critical Operations (90% Use Case)

### Must Implement First
| Operation | Kernel | Priority | Notes |
|-----------|--------|----------|-------|
| READ/WRITE | 5.6 | Critical | Basic file I/O |
| READV/WRITEV | 5.1 | Critical | Vectored I/O |
| ACCEPT | 5.5 | Critical | TCP servers |
| CONNECT | 5.5 | Critical | TCP clients |
| SEND/RECV | 5.6 | Critical | Socket I/O |
| SENDMSG/RECVMSG | 5.3 | Critical | Vectored socket I/O |
| TIMEOUT | 5.4 | High | Essential for timeouts |
| ASYNC_CANCEL | 5.5 | High | Cancel in-flight ops |
| OPENAT/CLOSE | 5.6 | High | File management |
| FSYNC | 5.1 | High | Durability |

### Advanced (Phase 2+)
- SQPOLL mode (5.1+) - zero syscall submissions
- Registered buffers/files (5.1+) - zero-copy
- Multishot accept (5.19+) - single submit, multiple accepts
- Multishot recv (6.0+) - single submit, multiple receives
- Zero-copy send (6.0+) - SEND_ZC, SENDMSG_ZC
- SOCKET/BIND/LISTEN (5.19/6.11) - full async socket lifecycle

## Kernel Version Strategy

### Minimum: 5.15 LTS
- Has all essential operations
- Widely deployed in production
- Docker/container support

### Recommended: 6.8+
- bind/listen ops for pure io_uring servers
- Multishot read
- Latest performance optimizations

### Runtime Detection
```go
// Use Probe() to detect operation support
probe, _ := ring.Probe()
if probe.SupportsOp(IORING_OP_BIND) {
    // Use async bind
} else {
    // Fall back to syscall.Bind
}
```

## Security Considerations

**WARNING**: io_uring has had significant security issues:
- Google disabled io_uring on ChromeOS, Android apps, and production servers (2023)
- Docker 25+ blocks io_uring by default via seccomp
- 60% of Google's 2022 bug bounty submissions exploited io_uring

**Mitigation:**
1. Check if io_uring is blocked: `make check-iouring`
2. Handle EPERM gracefully (seccomp blocked)
3. Validate all fd inputs and buffer bounds
4. Stay current on kernel security patches

## Testing Strategy

1. **Unit tests**: Each operation in isolation
2. **Integration tests**: Multi-operation workflows (echo server, file copy)
3. **Kernel matrix**: CI tests on 5.15, 6.1, 6.6, 6.8, 6.11
4. **Benchmarks**: Compare against syscall baseline and other libraries
5. **Fuzz testing**: Random operation sequences

## Existing Go Libraries (Deficiencies)

| Library | Status | Issues |
|---------|--------|--------|
| iceber/iouring-go | Active | Channel per response (allocations!), incomplete ops |
| godzie44/go-uring | Active | Heap-allocated closures for callbacks |
| dshulyak/uring | Archived 2022 | No batching, can't use SQPOLL, +750ns overhead |
| hodgesds/iouring-go | Stale | POC only, author says needs runtime integration |

**Our differentiators**: Complete coverage, zero allocations, multi-kernel, systematic header sync, proper Go scheduler integration.

## Performance Targets

| Metric | Target |
|--------|--------|
| NOP submit+complete | <150ns |
| Batched ops overhead | <10ns per op |
| Memory allocations | 0 in hot path |
| SQPOLL syscall reduction | >90% |

## References

- [Kernel io_uring.h](https://github.com/torvalds/linux/blob/master/include/uapi/linux/io_uring.h)
- [liburing](https://github.com/axboe/liburing)
- [Lord of the io_uring](https://unixism.net/loti/)
- [Efficient IO with io_uring (PDF)](https://kernel.dk/io_uring.pdf)
- [io_uring manpage](https://man7.org/linux/man-pages/man7/io_uring.7.html)
