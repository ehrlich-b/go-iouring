# Go io_uring Implementation Order

## Overview

This document outlines a phased implementation plan prioritizing the most commonly needed operations while building a solid foundation.

## Phase 0: Foundation (Week 1-2)

### Objectives
- Establish syscall interface without CGO
- Implement ring setup and teardown
- Basic memory mapping

### Deliverables

```
/internal/sys/
├── syscall.go       # io_uring_setup, io_uring_enter, io_uring_register
├── types.go         # io_uring_params, io_uring_sqe, io_uring_cqe
├── consts.go        # All constants (ops, flags, etc.)
└── mmap.go          # Ring memory mapping

/ring.go             # Ring struct and New()
/sqe.go              # SQE pool and basic operations
/cqe.go              # CQE reading
```

### Tasks

1. **Syscall Wrappers** (2 days)
   ```go
   func io_uring_setup(entries uint32, params *Params) (int, error)
   func io_uring_enter(fd int, toSubmit, minComplete uint32, flags uint32, sig *unix.Sigset_t) (int, error)
   func io_uring_register(fd int, opcode uint32, arg unsafe.Pointer, nrArgs uint32) error
   ```

2. **Memory Mapping** (2 days)
   - Map SQ and CQ rings
   - Handle SINGLE_MMAP feature
   - Map SQE array

3. **Ring Operations** (3 days)
   ```go
   func New(entries uint32, opts ...Option) (*Ring, error)
   func (r *Ring) Close() error
   func (r *Ring) getSQE() *SQE       // Get next available SQE
   func (r *Ring) Submit() (int, error)
   func (r *Ring) WaitCQE(ctx context.Context) (*CQE, error)
   ```

4. **Basic Operations** (3 days)
   - IORING_OP_NOP (testing)
   - IORING_OP_READ
   - IORING_OP_WRITE

### Validation
- [ ] Ring creation and destruction works
- [ ] Can submit NOP and receive CQE
- [ ] Simple file read/write works
- [ ] Memory is properly unmapped on close

---

## Phase 1: Core I/O (Week 3-4)

### Objectives
- Complete file I/O operations
- Implement timeout support
- Add feature detection

### Operations to Implement

| Operation | Priority | Notes |
|-----------|----------|-------|
| IORING_OP_READV | Critical | Vectored read |
| IORING_OP_WRITEV | Critical | Vectored write |
| IORING_OP_FSYNC | High | File sync |
| IORING_OP_READ_FIXED | High | Pre-registered buffers |
| IORING_OP_WRITE_FIXED | High | Pre-registered buffers |
| IORING_OP_TIMEOUT | High | Wait with timeout |
| IORING_OP_TIMEOUT_REMOVE | Medium | Cancel timeout |
| IORING_OP_ASYNC_CANCEL | High | Cancel any operation |
| IORING_OP_LINK_TIMEOUT | Medium | Linked timeout |

### Deliverables

```
/prep.go             # PrepRead, PrepWrite, PrepReadv, etc.
/timeout.go          # Timeout operations
/cancel.go           # Cancellation support
/register.go         # File and buffer registration
/probe.go            # Feature detection
```

### Tasks

1. **Buffer Registration** (2 days)
   ```go
   func (r *Ring) RegisterBuffers(bufs [][]byte) error
   func (r *Ring) UnregisterBuffers() error
   ```

2. **File Registration** (2 days)
   ```go
   func (r *Ring) RegisterFiles(fds []int) error
   func (r *Ring) UpdateFiles(fds []int, offset uint32) error
   func (r *Ring) UnregisterFiles() error
   ```

3. **Vectored I/O** (2 days)
   - Handle iovec conversion from [][]byte
   - Pin memory during operation

4. **Timeout Support** (2 days)
   - TIMEOUT, TIMEOUT_REMOVE, LINK_TIMEOUT
   - Clock selection (MONOTONIC, BOOTTIME, REALTIME)

5. **Feature Probing** (1 day)
   ```go
   func (r *Ring) Probe() (*Probe, error)
   func (r *Ring) SupportsOp(op Op) bool
   ```

### Validation
- [ ] Vectored I/O works correctly
- [ ] Fixed buffers improve performance
- [ ] Timeout operations work
- [ ] Linked timeout cancels parent
- [ ] Feature probe returns accurate results

---

## Phase 2: Network I/O (Week 5-7)

### Objectives
- Full TCP/UDP socket support
- Multishot operations
- Basic performance optimization

### Operations to Implement

| Operation | Priority | Notes |
|-----------|----------|-------|
| IORING_OP_ACCEPT | Critical | Accept connections |
| IORING_OP_CONNECT | Critical | Initiate connections |
| IORING_OP_SEND | Critical | Send data |
| IORING_OP_RECV | Critical | Receive data |
| IORING_OP_SENDMSG | High | Vectored/ancillary send |
| IORING_OP_RECVMSG | High | Vectored/ancillary recv |
| IORING_OP_SHUTDOWN | High | Socket shutdown |
| IORING_OP_SOCKET | High | Create socket (5.19+) |
| IORING_OP_BIND | High | Bind socket (6.11+) |
| IORING_OP_LISTEN | High | Listen (6.11+) |
| IORING_OP_POLL_ADD | Medium | Poll for events |
| IORING_OP_POLL_REMOVE | Medium | Cancel poll |

### Deliverables

```
/net.go              # Network operations
/socket.go           # Socket creation helpers
/multishot.go        # Multishot accept/recv handling
/poll.go             # Poll operations
```

### Tasks

1. **Basic Socket Ops** (3 days)
   - Accept, Connect, Send, Recv
   - Proper sockaddr handling

2. **Advanced Socket Ops** (3 days)
   - Sendmsg, Recvmsg with control messages
   - Shutdown

3. **Socket Creation** (2 days)
   - Socket op (5.19+)
   - Bind, Listen (6.11+)
   - Fallback for older kernels

4. **Multishot** (3 days)
   - Multishot accept (5.19+)
   - Multishot recv (6.0+)
   - CQE_F_MORE handling

5. **Polling** (2 days)
   - POLL_ADD/POLL_REMOVE
   - Multishot poll (5.13+)

### Example: TCP Server

```go
// After Phase 2, this should work
ring, _ := iouring.New(256)

// Create socket entirely through io_uring (6.11+)
sockReq := ring.PrepSocket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
sockReq.Submit()
cqe, _ := ring.WaitCQE(ctx)
listenFd := int(cqe.Res)

ring.PrepBind(listenFd, addr).SubmitAndWait(ctx)
ring.PrepListen(listenFd, 128).SubmitAndWait(ctx)

// Multishot accept
ring.PrepAccept(listenFd, 0).WithMultishot().Submit()
```

### Validation
- [ ] Echo server works
- [ ] Multishot accept doesn't leak memory
- [ ] Connect with timeout works
- [ ] Can handle 10k+ connections

---

## Phase 3: Advanced Features (Week 8-10)

### Objectives
- SQPOLL support
- Zero-copy networking
- Provided buffer rings
- Performance optimization

### Features to Implement

| Feature | Priority | Notes |
|---------|----------|-------|
| SQPOLL | High | Kernel polls SQ |
| Provided Buffers | High | Kernel selects buffers |
| Zero-Copy Send | High | SEND_ZC, SENDMSG_ZC |
| Linked Operations | Medium | Chain SQEs |
| CQE Skip | Medium | No CQE on success |
| Buffer Rings | Medium | Efficient buffer mgmt (5.19+) |

### Deliverables

```
/sqpoll.go           # SQPOLL handling
/zerocopy.go         # Zero-copy operations
/bufring.go          # Provided buffer rings
/link.go             # Linked operation helpers
```

### Tasks

1. **SQPOLL** (3 days)
   - Setup with SQ_AFF
   - Wakeup handling
   - Thread idle timeout

2. **Linked Operations** (2 days)
   - Link flag handling
   - Hard link support
   - Chain error propagation

3. **Provided Buffers** (4 days)
   - PROVIDE_BUFFERS / REMOVE_BUFFERS
   - Buffer rings (io_uring_buf_ring)
   - Buffer selection in recv

4. **Zero-Copy** (4 days)
   - SEND_ZC implementation
   - SENDMSG_ZC implementation
   - Notification handling (CQE_F_NOTIF)
   - Buffer lifetime management

### Validation
- [ ] SQPOLL reduces syscall count
- [ ] Provided buffers work with recv
- [ ] Zero-copy send produces two CQEs
- [ ] Linked operations chain correctly

---

## Phase 4: Filesystem & Completion (Week 11-13)

### Objectives
- Complete filesystem operations
- Remaining misc operations
- Documentation and examples

### Operations to Implement

| Operation | Priority | Notes |
|-----------|----------|-------|
| IORING_OP_OPENAT | High | Open file |
| IORING_OP_OPENAT2 | Medium | Extended open |
| IORING_OP_CLOSE | High | Close fd |
| IORING_OP_STATX | Medium | File stat |
| IORING_OP_RENAMEAT | Medium | Rename file |
| IORING_OP_UNLINKAT | Medium | Delete file |
| IORING_OP_MKDIRAT | Low | Create directory |
| IORING_OP_SYMLINKAT | Low | Create symlink |
| IORING_OP_LINKAT | Low | Create hard link |
| IORING_OP_FALLOCATE | Low | Preallocate space |
| IORING_OP_SPLICE | Low | Move data |
| IORING_OP_TEE | Low | Duplicate pipe data |
| IORING_OP_FTRUNCATE | Low | Truncate file |

### Deliverables

```
/fs.go               # Filesystem operations
/splice.go           # Splice/tee operations
/examples/           # Complete examples
/README.md           # Documentation
```

### Tasks

1. **File Operations** (4 days)
   - Open, Close with direct fd support
   - Statx
   - Rename, Unlink

2. **Directory Operations** (2 days)
   - Mkdir, Symlink, Link

3. **Splice/Tee** (2 days)
   - Splice between fds
   - Tee for pipes

4. **Documentation** (3 days)
   - API documentation
   - Performance guide
   - Example applications

5. **Examples** (3 days)
   - File copy utility
   - HTTP server
   - Database-style I/O patterns

---

## Phase 5: Polish & Production (Week 14-16)

### Objectives
- Performance optimization
- Edge case handling
- Production hardening

### Tasks

1. **Performance Audit** (4 days)
   - Profile hot paths
   - Reduce allocations
   - Optimize memory barriers
   - Benchmark vs epoll

2. **Error Handling** (3 days)
   - Graceful CQ overflow
   - SQPOLL thread death
   - Resource cleanup on error

3. **Security Review** (2 days)
   - Validate all fd inputs
   - Check buffer bounds
   - Review unsafe usage

4. **Testing** (4 days)
   - Stress tests
   - Fuzzing
   - CI matrix for kernel versions

5. **Documentation** (2 days)
   - Final API review
   - Migration guide
   - Troubleshooting guide

---

## Priority Matrix

### Must Have (Phase 0-2)
- Ring setup/teardown
- Read/Write/Readv/Writev
- Accept/Connect/Send/Recv
- Timeout
- Feature probe
- Basic error handling

### Should Have (Phase 3)
- SQPOLL
- Provided buffers
- Zero-copy send
- Linked operations
- Multishot accept/recv

### Nice to Have (Phase 4)
- Full filesystem ops
- Splice/Tee
- MSG_RING
- Extended attributes

### Future Consideration
- URING_CMD (driver-specific)
- Futex operations
- Waitid
- Integration with Go netpoller

---

## Resource Requirements

### Development
- 1-2 developers full-time
- Access to multiple kernel versions (VM or containers)
- Performance testing hardware

### Testing Infrastructure
- CI with kernel matrix (5.15, 6.1, 6.6, 6.11)
- Memory leak detection (valgrind via CGO test mode)
- Benchmarking suite

### Documentation
- API reference (godoc)
- Performance guide
- Example repository

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Kernel version fragmentation | Feature detection, graceful fallbacks |
| Security vulnerabilities | Follow kernel security advisories, validate inputs |
| Go memory model issues | Conservative memory barriers, extensive testing |
| Container restrictions | Document seccomp requirements |
| Performance regression | Continuous benchmarking |

---

## Success Criteria

### Phase 1 Complete
- [ ] Basic file I/O 2x faster than syscall for batched ops
- [ ] All tests pass on kernel 5.15+

### Phase 2 Complete
- [ ] TCP echo server handles 100k conn/sec
- [ ] No goroutine leaks under load

### Phase 3 Complete
- [ ] SQPOLL reduces syscalls by 90%+
- [ ] Zero-copy send works without buffer corruption

### Final Release
- [ ] Production-ready documentation
- [ ] Performance competitive with C liburing
- [ ] Adopted by at least one production system
