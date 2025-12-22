# go-iouring Implementation Roadmap

## Current Status: Phase 3 In Progress

Phases 0-3 mostly complete. Core I/O, Network I/O, SQPOLL, zero-copy send, and linked operations working.
All 26 tests passing (TestBindListen skips on kernels < 6.11, TestSendZC requires 6.0+).

### Benchmark Results (i7-8700K)
| Operation | ns/op | allocs |
|-----------|-------|--------|
| NOP single | 841 | 0 |
| NOP batched | 92 | 0 |
| Read syscall | 885 | 0 |
| Read io_uring | 1071 | 0 |
| Read batched | 373 | 0 |

## Decision Points (Resolved)

- [x] **Minimum kernel**: 5.15 LTS as floor, 6.8+ recommended
- [x] **Buffer management**: Auto for simple, manual for perf-critical
- [x] **Unsupported ops**: Return ErrNotSupported

---

## Phase 0: Foundation ✅

### Syscall Layer
- [x] `internal/sys/syscall.go` - io_uring_setup, io_uring_enter, io_uring_register
- [x] `internal/sys/types.go` - io_uring_params, io_uring_sqe (64 bytes), io_uring_cqe
- [x] `internal/sys/consts.go` - All IORING_OP_*, IOSQE_*, IORING_SETUP_* constants

### Core Ring
- [x] `ring.go` - Ring struct, New(), Close()
- [x] `sqe.go` - getSQE(), SQE pool management
- [x] `cqe.go` - PeekCQE(), SeenCQE(), ForEachCQE()
- [x] Basic Submit() without waiting

### Validation
- [x] Ring creation/destruction works
- [x] NOP submit and receive CQE
- [x] Memory properly unmapped on close

---

## Phase 1: Core I/O Operations ✅

### File I/O
- [x] PrepRead / PrepWrite
- [x] PrepReadv / PrepWritev (vectored)
- [x] PrepReadFixed / PrepWriteFixed (registered buffers)
- [x] PrepFsync

### Timeouts & Cancellation
- [x] PrepTimeout (with clock selection)
- [x] PrepTimeoutRemove
- [x] PrepLinkTimeout
- [x] PrepAsyncCancel (PrepCancel)

### Registration
- [x] RegisterBuffers / UnregisterBuffers
- [x] RegisterFiles / UnregisterFiles
- [x] RegisterEventfd

### Feature Detection
- [x] `probe.go` - Probe(), SupportsOp(), HasFeature()

### Validation
- [x] File read/write benchmarks vs syscall
- [x] Timeout actually expires
- [x] Cancellation works
- [x] Feature probe returns accurate results

---

## Phase 2: Network I/O ✅

### Socket Operations
- [x] PrepAccept (with multishot support flag)
- [x] PrepAcceptMultishot
- [x] PrepConnect
- [x] PrepSend / PrepRecv
- [x] PrepRecvMultishot
- [x] PrepSendmsg / PrepRecvmsg
- [x] PrepShutdown

### New Socket Lifecycle (5.19+/6.11+)
- [x] PrepSocket (create socket async)
- [x] PrepBind (6.11+)
- [x] PrepListen (6.11+)

### Polling
- [x] PrepPollAdd (with multishot)
- [x] PrepPollRemove

### Validation
- [x] Accept/connect tests pass
- [x] Send/recv tests pass
- [x] Poll tests pass
- [ ] Echo server handles 100k+ conn/sec
- [ ] Multishot accept works without leaks
- [x] Connect with timeout works

---

## Phase 3: Advanced Features

### SQPOLL Mode
- [x] Setup with IORING_SETUP_SQPOLL (WithSQPoll option)
- [x] SQ_AFF CPU pinning (WithSQPollCPU option)
- [x] Wakeup handling (IORING_SQ_NEED_WAKEUP)
- [x] Idle timeout configuration (WithSQPollIdle option)

### Provided Buffers
- [x] PrepProvideBuffers / PrepRemoveBuffers
- [ ] Buffer ring setup (IORING_REGISTER_PBUF_RING, 5.19+)
- [x] Automatic buffer selection in recv (PrepRecvMultishot with buf_group)

### Zero-Copy Networking
- [x] PrepSendZC (6.0+)
- [x] PrepSendmsgZC (6.1+)
- [x] Handle notification CQE (IORING_CQE_F_NOTIF)
- [ ] Buffer lifetime management (user responsibility)

### Linked Operations
- [x] IOSQE_IO_LINK flag support (SetSQELink)
- [x] SetSQEFlags / SetSQEAsync / SetSQEDrain helpers
- [x] IOSQE_IO_HARDLINK flag support (SetSQEHardlink)
- [x] Chain error propagation (handled by kernel, exposed via CQE)

### Validation
- [ ] SQPOLL reduces syscalls >90%
- [ ] Zero-copy doesn't corrupt buffers
- [ ] Linked operations chain correctly

---

## Phase 4: Filesystem & Completion (Partial)

### File Management
- [x] PrepOpenat
- [ ] PrepOpenat2
- [x] PrepClose
- [x] PrepStatx
- [ ] PrepFallocate
- [ ] PrepFtruncate (6.9+)

### Directory Operations
- [ ] PrepRenameat
- [ ] PrepUnlinkat
- [ ] PrepMkdirat
- [ ] PrepSymlinkat
- [ ] PrepLinkat

### Data Movement
- [x] PrepSplice
- [ ] PrepTee

### Extended Attributes
- [ ] PrepFsetxattr / PrepSetxattr
- [ ] PrepFgetxattr / PrepGetxattr

---

## Phase 5: Polish & Production

### Performance
- [ ] Profile hot paths
- [ ] Eliminate remaining allocations
- [ ] Optimize memory barriers
- [ ] Benchmark vs epoll, libaio, other Go libs

### Error Handling
- [ ] CQ overflow handling
- [ ] SQPOLL thread death recovery
- [ ] Graceful seccomp EPERM handling

### Testing
- [ ] Full kernel matrix CI (5.15, 6.1, 6.6, 6.8, 6.11)
- [ ] Stress tests
- [ ] Fuzzing
- [ ] Memory leak detection

### Documentation
- [ ] API documentation (godoc)
- [ ] Performance guide
- [ ] Migration guide from other libs

### Examples
- [ ] `examples/cp/` - Async file copy
- [ ] `examples/cat/` - Simple file read
- [ ] `examples/echoserver/` - TCP echo server
- [ ] `examples/proxy/` - High-performance proxy

---

## Low Priority / Future

- [ ] MSG_RING (inter-ring messaging)
- [ ] URING_CMD (driver passthrough)
- [ ] Futex operations (6.7+)
- [ ] WAITID
- [ ] SQE128 / CQE32 modes
- [ ] Go netpoller integration (complex, may not be worth it)

---

## Research Artifacts

Stored in repo for reference:
- `response_gemini.md` - Gemini analysis
- `response_grok.md` - Grok analysis
- `response_opus/` - Opus detailed deliverables
  - `operations.json` - Complete operation catalog
  - `kernel_matrix.md` - Feature availability by kernel
  - `api_design.md` - Recommended Go API design
  - `implementation_order.md` - Phased implementation plan
  - `test_plan.md` - Testing strategy

---

## Effort Estimate

| Phase | Scope | Effort |
|-------|-------|--------|
| 0 | Foundation | 1-2 weeks |
| 1 | Core I/O | 2-3 weeks |
| 2 | Network I/O | 2-3 weeks |
| 3 | Advanced | 3-4 weeks |
| 4 | Filesystem | 2-3 weeks |
| 5 | Polish | 2-3 weeks |
| **Total** | **Full coverage** | **12-18 weeks** |

For 90% use case (Phases 0-2): **5-8 weeks**
