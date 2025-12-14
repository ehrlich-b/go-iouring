# io_uring Kernel Feature Matrix

## Overview

This document maps io_uring feature availability to Linux kernel versions. Use this to determine minimum kernel requirements for your application.

## Kernel Version Summary

| Version | Release Date | Key Additions |
|---------|-------------|---------------|
| 5.1 | May 2019 | Initial io_uring: readv, writev, fsync, poll, NOP |
| 5.3 | Sep 2019 | sendmsg, recvmsg |
| 5.4 | Nov 2019 | timeout, SINGLE_MMAP |
| 5.5 | Jan 2020 | accept, connect, link_timeout, async_cancel |
| 5.6 | Mar 2020 | read, write, send, recv, openat, close, statx |
| 5.7 | Jun 2020 | splice, provided buffers, FAST_POLL |
| 5.10 | Dec 2020 | R_DISABLED flag |
| 5.11 | Feb 2021 | shutdown, renameat, unlinkat, EXT_ARG |
| 5.13 | Jun 2021 | Multishot poll |
| 5.15 | Oct 2021 | mkdirat, symlinkat, linkat, direct file descriptors |
| 5.18 | May 2022 | MSG_RING |
| 5.19 | Jul 2022 | socket, uring_cmd, multishot accept, xattr ops |
| 6.0 | Oct 2022 | SEND_ZC, multishot recv, SINGLE_ISSUER |
| 6.1 | Dec 2022 | SENDMSG_ZC, DEFER_TASKRUN |
| 6.5 | Aug 2023 | waitid, NO_MMAP |
| 6.7 | Jan 2024 | futex, read_multishot |
| 6.8 | Mar 2024 | FIXED_FD_INSTALL |
| 6.9 | May 2024 | ftruncate |
| 6.11 | Sep 2024 | bind, listen |
| 6.12 | Nov 2024 | Incremental buffer consumption, absolute timeouts |

## Minimum Kernel by Use Case

### Storage I/O (Recommended: 5.6+)
- 5.1: Basic vectored I/O (readv/writev)
- 5.6: Simple read/write, fixed buffers
- 5.7: Provided buffers for automatic buffer selection
- 6.7+: Multishot read for continuous reads

### Network I/O (Recommended: 6.0+)
- 5.3: sendmsg/recvmsg
- 5.5: accept, connect (critical for servers)
- 5.6: send/recv (simpler non-vectored API)
- 5.19: socket(), multishot accept
- 6.0: Zero-copy send, multishot recv
- 6.11: bind(), listen() (required for direct socket workflow)

### High-Performance Servers (Recommended: 6.1+)
- 5.1: SQPOLL for kernel-side submission polling
- 5.11: SQPOLL with non-fixed files
- 6.0: SINGLE_ISSUER optimization
- 6.1: DEFER_TASKRUN for cooperative scheduling

### File Operations (Recommended: 5.11+)
- 5.6: openat, close, statx, fallocate
- 5.11: rename, unlink
- 5.15: mkdir, symlink, hardlink
- 5.19: Extended attributes
- 6.9: ftruncate

## Operations by Kernel Version

### Kernel 5.1-5.5 (Foundation)
```
IORING_OP_NOP           (5.1)
IORING_OP_READV         (5.1)
IORING_OP_WRITEV        (5.1)
IORING_OP_FSYNC         (5.1)
IORING_OP_READ_FIXED    (5.1)
IORING_OP_WRITE_FIXED   (5.1)
IORING_OP_POLL_ADD      (5.1)
IORING_OP_POLL_REMOVE   (5.1)
IORING_OP_SYNC_FILE_RANGE (5.2)
IORING_OP_SENDMSG       (5.3)
IORING_OP_RECVMSG       (5.3)
IORING_OP_TIMEOUT       (5.4)
IORING_OP_TIMEOUT_REMOVE (5.5)
IORING_OP_ACCEPT        (5.5)
IORING_OP_ASYNC_CANCEL  (5.5)
IORING_OP_LINK_TIMEOUT  (5.5)
IORING_OP_CONNECT       (5.5)
```

### Kernel 5.6-5.10 (File & Network Essentials)
```
IORING_OP_FALLOCATE     (5.6)
IORING_OP_OPENAT        (5.6)
IORING_OP_CLOSE         (5.6)
IORING_OP_FILES_UPDATE  (5.6)
IORING_OP_STATX         (5.6)
IORING_OP_READ          (5.6)
IORING_OP_WRITE         (5.6)
IORING_OP_FADVISE       (5.6)
IORING_OP_MADVISE       (5.6)
IORING_OP_SEND          (5.6)
IORING_OP_RECV          (5.6)
IORING_OP_OPENAT2       (5.6)
IORING_OP_EPOLL_CTL     (5.6)
IORING_OP_SPLICE        (5.7)
IORING_OP_PROVIDE_BUFFERS (5.7)
IORING_OP_REMOVE_BUFFERS (5.7)
IORING_OP_TEE           (5.8)
```

### Kernel 5.11-5.17 (Filesystem & Optimization)
```
IORING_OP_SHUTDOWN      (5.11)
IORING_OP_RENAMEAT      (5.11)
IORING_OP_UNLINKAT      (5.11)
IORING_OP_MKDIRAT       (5.15)
IORING_OP_SYMLINKAT     (5.15)
IORING_OP_LINKAT        (5.15)
```

### Kernel 5.18-5.19 (IPC & Advanced)
```
IORING_OP_MSG_RING      (5.18)
IORING_OP_FSETXATTR     (5.19)
IORING_OP_SETXATTR      (5.19)
IORING_OP_FGETXATTR     (5.19)
IORING_OP_GETXATTR      (5.19)
IORING_OP_SOCKET        (5.19)
IORING_OP_URING_CMD     (5.19)
```

### Kernel 6.0+ (Performance & Zero-Copy)
```
IORING_OP_SEND_ZC       (6.0)
IORING_OP_SENDMSG_ZC    (6.1)
IORING_OP_WAITID        (6.5)
IORING_OP_FUTEX_WAIT    (6.7)
IORING_OP_FUTEX_WAKE    (6.7)
IORING_OP_FUTEX_WAITV   (6.7)
IORING_OP_FIXED_FD_INSTALL (6.8)
IORING_OP_FTRUNCATE     (6.9)
IORING_OP_BIND          (6.11)
IORING_OP_LISTEN        (6.11)
IORING_OP_READ_MULTISHOT (6.7)
```

## Changes Between 6.8 and 6.11

### New Operations
- `IORING_OP_FTRUNCATE` (6.9): Async file truncation
- `IORING_OP_BIND` (6.11): Bind socket to address
- `IORING_OP_LISTEN` (6.11): Start listening on socket

### New Features
- MSG_RING efficiency improvements for DEFER_TASKRUN rings (6.11)
- Async discard operations for block devices (6.12)
- Minimum timeout support in io_uring_enter (6.12)
- Absolute timeout support with clock selection (6.12)
- Incremental buffer consumption (`IOU_PBUF_RING_INC`) (6.12)
- Buffer cloning between rings (6.12)
- Huge page optimization for registered buffers (6.12)

### Breaking Changes
- None between 6.8 and 6.11

## Feature Detection

### Runtime Detection with Probe
```go
// Use IORING_REGISTER_PROBE to detect operation support
type IOURingProbe struct {
    LastOp uint8
    OpsLen uint8
    // ...
    Ops    []IOURingProbeOp
}

// Check if operation is supported
func (p *IOURingProbe) IsSupported(op uint8) bool {
    if int(op) >= len(p.Ops) {
        return false
    }
    return p.Ops[op].Flags & IO_URING_OP_SUPPORTED != 0
}
```

### Feature Flags from io_uring_params
```go
// After io_uring_setup, check params.features for:
// IORING_FEAT_* flags indicating kernel capabilities
if params.Features & IORING_FEAT_FAST_POLL != 0 {
    // Fast poll available
}
```

## Recommended Minimum Kernel

| Target Use Case | Minimum | Recommended | Notes |
|-----------------|---------|-------------|-------|
| Basic async I/O | 5.6 | 5.11 | 5.6 has read/write, 5.11 has full fs ops |
| TCP Server | 5.5 | 6.0 | 5.5 has accept/connect, 6.0 has multishot |
| High-perf network | 6.0 | 6.11 | Zero-copy, bind/listen for full async |
| Production workloads | 5.15 LTS | 6.6 LTS | LTS kernels recommended |

## Security Considerations

**Warning**: io_uring has been a source of security vulnerabilities. Consider:

- Container environments: Docker 25+ blocks io_uring by default via seccomp
- Google disabled io_uring on ChromeOS, Android apps, and production servers (2023)
- Some cloud providers restrict io_uring in shared environments

Before deploying, verify:
1. io_uring is not blocked by seccomp filters
2. Kernel version has latest security patches
3. Application properly validates all file descriptors and buffers
