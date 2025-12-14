# Go io_uring API Design

## Design Philosophy

### Core Principles

1. **Zero CGO**: Pure Go implementation using syscall/x/sys/unix
2. **Minimal Allocations**: Pre-allocate SQEs, reuse CQE memory
3. **Type Safety**: Strong typing for operations, flags, and results
4. **Idiomatic Go**: Context support, error wrapping, interface compatibility
5. **Progressive Disclosure**: Simple API for common cases, full control available

### Lessons from Existing Libraries

| Library | Strength | Weakness |
|---------|----------|----------|
| iceber/iouring-go | Channel-based completion API | Allocations per request, limited ops |
| godzie44/go-uring | Three-layer architecture, reactor | Complexity, some flaky tests |
| dshulyak/uring | No CGO, clean code | 750ns overhead, no batching, no SQPOLL |

## Package Structure

```
github.com/yourorg/iouring/
├── ring.go          # Core Ring type and setup
├── sqe.go           # SQE preparation helpers
├── cqe.go           # CQE handling
├── ops.go           # Operation constants and types
├── flags.go         # All flag definitions
├── register.go      # Buffer/file registration
├── probe.go         # Feature detection
├── errors.go        # Error types and handling
├── internal/
│   ├── sys/         # Low-level syscall wrappers
│   └── atomic/      # Memory barriers
└── examples/
    ├── cat/         # Simple file read
    ├── echo/        # TCP echo server
    └── proxy/       # High-performance proxy
```

## Type Definitions

### Core Types

```go
package iouring

import (
    "context"
    "sync/atomic"
    "unsafe"
)

// Ring represents an io_uring instance
type Ring struct {
    fd        int
    params    Params
    sq        submissionQueue
    cq        completionQueue
    features  uint32
    
    // For goroutine-safe access
    sqLock    sync.Mutex
    
    // User data management
    nextID    atomic.Uint64
    pending   sync.Map // map[uint64]*Request
}

// Params mirrors io_uring_params
type Params struct {
    SQEntries    uint32
    CQEntries    uint32
    Flags        uint32
    SQThreadCPU  uint32
    SQThreadIdle uint32
    Features     uint32
    WQFd         uint32
    // ... rest of params
}

// SQE is a submission queue entry (64 or 128 bytes)
type SQE struct {
    Opcode      uint8
    Flags       uint8
    Ioprio      uint16
    Fd          int32
    Off         uint64
    Addr        uint64
    Len         uint32
    OpFlags     uint32
    UserData    uint64
    BufIndex    uint16
    Personality uint16
    SpliceFdIn  int32
    Addr3       uint64
    _           [16]byte // padding or cmd space
}

// CQE is a completion queue entry (16 or 32 bytes)
type CQE struct {
    UserData uint64
    Res      int32
    Flags    uint32
    // If CQE32: additional 16 bytes
}

// Request represents an in-flight operation
type Request struct {
    id       uint64
    op       Op
    done     chan struct{}
    result   int32
    flags    uint32
    err      error
    userData interface{}
}
```

### Operation Types

```go
// Op represents an io_uring operation type
type Op uint8

const (
    OpNop Op = iota
    OpReadv
    OpWritev
    OpFsync
    OpReadFixed
    OpWriteFixed
    OpPollAdd
    OpPollRemove
    OpSyncFileRange
    OpSendmsg
    OpRecvmsg
    OpTimeout
    OpTimeoutRemove
    OpAccept
    OpAsyncCancel
    OpLinkTimeout
    OpConnect
    OpFallocate
    OpOpenat
    OpClose
    OpFilesUpdate
    OpStatx
    OpRead
    OpWrite
    OpFadvise
    OpMadvise
    OpSend
    OpRecv
    OpOpenat2
    OpEpollCtl
    OpSplice
    OpProvideBuffers
    OpRemoveBuffers
    OpTee
    OpShutdown
    OpRenameat
    OpUnlinkat
    OpMkdirat
    OpSymlinkat
    OpLinkat
    OpMsgRing
    OpFsetxattr
    OpSetxattr
    OpFgetxattr
    OpGetxattr
    OpSocket
    OpUringCmd
    OpSendZC
    OpSendmsgZC
    OpWaitid
    OpFutexWait
    OpFutexWake
    OpFutexWaitv
    OpFixedFdInstall
    OpFtruncate
    OpBind
    OpListen
    OpReadMultishot
)

// String returns the operation name
func (o Op) String() string
```

### Flag Types

```go
// SQEFlags are per-submission flags
type SQEFlags uint8

const (
    SQEFixedFile       SQEFlags = 1 << 0
    SQEIODrain         SQEFlags = 1 << 1
    SQEIOLink          SQEFlags = 1 << 2
    SQEIOHardlink      SQEFlags = 1 << 3
    SQEAsync           SQEFlags = 1 << 4
    SQEBufferSelect    SQEFlags = 1 << 5
    SQECQESkipSuccess  SQEFlags = 1 << 6
)

// SetupFlags are ring setup options
type SetupFlags uint32

const (
    SetupIOPoll        SetupFlags = 1 << 0
    SetupSQPoll        SetupFlags = 1 << 1
    SetupSQAff         SetupFlags = 1 << 2
    SetupCQSize        SetupFlags = 1 << 3
    SetupClamp         SetupFlags = 1 << 4
    SetupAttachWQ      SetupFlags = 1 << 5
    SetupRDisabled     SetupFlags = 1 << 6
    SetupSubmitAll     SetupFlags = 1 << 7
    SetupCoopTaskrun   SetupFlags = 1 << 8
    SetupTaskrunFlag   SetupFlags = 1 << 9
    SetupSQE128        SetupFlags = 1 << 10
    SetupCQE32         SetupFlags = 1 << 11
    SetupSingleIssuer  SetupFlags = 1 << 12
    SetupDeferTaskrun  SetupFlags = 1 << 13
    SetupNoMmap        SetupFlags = 1 << 14
    SetupRegisteredFdOnly SetupFlags = 1 << 15
    SetupNoSQArray     SetupFlags = 1 << 16
)
```

## Ring API

### Construction and Setup

```go
// New creates a new io_uring instance
func New(entries uint32, opts ...Option) (*Ring, error)

// Option configures ring setup
type Option func(*Params)

// Common options
func WithSQPoll() Option
func WithSQPollCPU(cpu uint32) Option
func WithSQPollIdle(idle time.Duration) Option
func WithIOPoll() Option
func WithCQSize(size uint32) Option
func WithSingleIssuer() Option
func WithDeferTaskrun() Option
func WithFlags(flags SetupFlags) Option

// Example usage
ring, err := iouring.New(256,
    iouring.WithSQPoll(),
    iouring.WithSQPollCPU(0),
)
if err != nil {
    return fmt.Errorf("creating ring: %w", err)
}
defer ring.Close()
```

### Submission API (Builder Pattern)

```go
// PrepRead prepares a read operation
func (r *Ring) PrepRead(fd int, buf []byte, offset uint64) *PreparedOp

// PrepWrite prepares a write operation
func (r *Ring) PrepWrite(fd int, buf []byte, offset uint64) *PreparedOp

// PrepAccept prepares an accept operation
func (r *Ring) PrepAccept(fd int, flags int) *PreparedOp

// PrepConnect prepares a connect operation
func (r *Ring) PrepConnect(fd int, addr syscall.Sockaddr) *PreparedOp

// PrepSend prepares a send operation
func (r *Ring) PrepSend(fd int, buf []byte, flags int) *PreparedOp

// PrepRecv prepares a recv operation
func (r *Ring) PrepRecv(fd int, buf []byte, flags int) *PreparedOp

// PrepTimeout prepares a timeout operation
func (r *Ring) PrepTimeout(ts time.Duration, count uint64) *PreparedOp

// PrepCancel prepares a cancel operation
func (r *Ring) PrepCancel(userData uint64) *PreparedOp

// PrepSocket prepares a socket creation
func (r *Ring) PrepSocket(domain, typ, protocol int) *PreparedOp

// PrepBind prepares a bind operation (6.11+)
func (r *Ring) PrepBind(fd int, addr syscall.Sockaddr) *PreparedOp

// PrepListen prepares a listen operation (6.11+)
func (r *Ring) PrepListen(fd int, backlog int) *PreparedOp

// PreparedOp represents a prepared operation with fluent modifiers
type PreparedOp struct {
    ring *Ring
    sqe  *SQE
}

func (p *PreparedOp) WithFlags(flags SQEFlags) *PreparedOp
func (p *PreparedOp) WithLink() *PreparedOp
func (p *PreparedOp) WithAsync() *PreparedOp
func (p *PreparedOp) WithDrain() *PreparedOp
func (p *PreparedOp) WithUserData(data interface{}) *PreparedOp
func (p *PreparedOp) WithFixedFile(index uint16) *PreparedOp
func (p *PreparedOp) WithBufferSelect(group uint16) *PreparedOp
func (p *PreparedOp) Submit() (*Request, error)
func (p *PreparedOp) SubmitAndWait(ctx context.Context) (int32, error)
```

### Completion API

```go
// Submit submits pending SQEs and returns count
func (r *Ring) Submit() (int, error)

// SubmitAndWait submits and waits for at least n completions
func (r *Ring) SubmitAndWait(n uint32) (int, error)

// WaitCQE waits for at least one CQE
func (r *Ring) WaitCQE(ctx context.Context) (*CQE, error)

// WaitCQEs waits for n CQEs (batch retrieval)
func (r *Ring) WaitCQEs(ctx context.Context, n uint32) ([]*CQE, error)

// PeekCQE returns a CQE if available without waiting
func (r *Ring) PeekCQE() (*CQE, bool)

// SeenCQE marks a CQE as consumed
func (r *Ring) SeenCQE(cqe *CQE)

// ForEachCQE processes all available CQEs
func (r *Ring) ForEachCQE(fn func(*CQE) bool)
```

### Registration API

```go
// RegisterFiles registers file descriptors for fixed file ops
func (r *Ring) RegisterFiles(fds []int) error

// UpdateFiles updates registered files at offset
func (r *Ring) UpdateFiles(fds []int, offset uint32) error

// UnregisterFiles removes all registered files
func (r *Ring) UnregisterFiles() error

// RegisterBuffers registers buffers for fixed buffer ops
func (r *Ring) RegisterBuffers(bufs [][]byte) error

// UnregisterBuffers removes registered buffers
func (r *Ring) UnregisterBuffers() error

// SetupBufferRing sets up a provided buffer ring (5.19+)
func (r *Ring) SetupBufferRing(entries uint32, groupID uint16) (*BufferRing, error)
```

### Feature Detection

```go
// Probe queries supported operations
func (r *Ring) Probe() (*Probe, error)

// SupportsOp checks if an operation is supported
func (r *Ring) SupportsOp(op Op) bool

// Features returns the feature flags
func (r *Ring) Features() uint32

// HasFeature checks for a specific feature
func (r *Ring) HasFeature(feat uint32) bool
```

## Usage Examples

### Simple File Read

```go
func readFile(path string) ([]byte, error) {
    ring, err := iouring.New(8)
    if err != nil {
        return nil, err
    }
    defer ring.Close()
    
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    
    info, _ := f.Stat()
    buf := make([]byte, info.Size())
    
    n, err := ring.PrepRead(int(f.Fd()), buf, 0).SubmitAndWait(context.Background())
    if err != nil {
        return nil, err
    }
    
    return buf[:n], nil
}
```

### TCP Echo Server

```go
func echoServer(ctx context.Context, port int) error {
    ring, err := iouring.New(256, iouring.WithSQPoll())
    if err != nil {
        return err
    }
    defer ring.Close()
    
    // Create listening socket via io_uring
    req, err := ring.PrepSocket(syscall.AF_INET, syscall.SOCK_STREAM, 0).Submit()
    if err != nil {
        return err
    }
    <-req.Done()
    listenFd := int(req.Result())
    
    addr := &syscall.SockaddrInet4{Port: port}
    if _, err := ring.PrepBind(listenFd, addr).SubmitAndWait(ctx); err != nil {
        return err
    }
    if _, err := ring.PrepListen(listenFd, 128).SubmitAndWait(ctx); err != nil {
        return err
    }
    
    // Accept loop
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        // Multishot accept
        req, err := ring.PrepAccept(listenFd, 0).
            WithFlags(iouring.SQEMultishot).
            Submit()
        if err != nil {
            return err
        }
        
        // Handle connections...
    }
}
```

### Linked Operations

```go
// Read then write as linked chain
func copyWithLink(ring *iouring.Ring, srcFd, dstFd int, buf []byte) error {
    // Read with link flag - write won't start until read completes
    ring.PrepRead(srcFd, buf, 0).WithLink()
    
    // Write is chained to read
    _, err := ring.PrepWrite(dstFd, buf, 0).SubmitAndWait(context.Background())
    return err
}

// Read with timeout
func readWithTimeout(ring *iouring.Ring, fd int, buf []byte, timeout time.Duration) (int, error) {
    ring.PrepRead(fd, buf, 0).WithLink()
    ring.PrepLinkTimeout(timeout)
    
    // Submit both
    if _, err := ring.Submit(); err != nil {
        return 0, err
    }
    
    // Wait for read completion (or timeout)
    cqe, err := ring.WaitCQE(context.Background())
    if err != nil {
        return 0, err
    }
    
    if cqe.Res < 0 {
        return 0, syscall.Errno(-cqe.Res)
    }
    return int(cqe.Res), nil
}
```

## Performance Considerations

### Memory Barriers

Go's sync/atomic provides acquire/release semantics needed for ring buffer operations:

```go
// For SQ tail update (producer)
func (sq *submissionQueue) flush() int {
    tail := atomic.LoadUint32(sq.tail)
    // ... prepare entries ...
    atomic.StoreUint32(sq.tail, newTail) // release semantics
    return int(newTail - tail)
}

// For CQ head read (consumer)
func (cq *completionQueue) peek() *CQE {
    head := atomic.LoadUint32(cq.head) // acquire semantics
    tail := atomic.LoadUint32(cq.tail)
    if head == tail {
        return nil
    }
    return &cq.cqes[head & cq.mask]
}
```

### Avoiding Allocations

```go
// Use sync.Pool for frequently allocated types
var sqePool = sync.Pool{
    New: func() interface{} {
        return &PreparedOp{}
    },
}

// Pre-allocate buffers for multishot operations
type BufferPool struct {
    bufs    [][]byte
    free    chan int
    bufSize int
}
```

### SQPOLL Considerations

```go
// SQPOLL needs wakeup check
func (r *Ring) needWakeup() bool {
    if r.params.Flags&SetupSQPoll == 0 {
        return false
    }
    flags := atomic.LoadUint32(r.sq.flags)
    return flags&IORING_SQ_NEED_WAKEUP != 0
}

func (r *Ring) Submit() (int, error) {
    flushed := r.sq.flush()
    if flushed == 0 {
        return 0, nil
    }
    
    var flags uint32
    if r.needWakeup() {
        flags |= IORING_ENTER_SQ_WAKEUP
    }
    
    if r.params.Flags&SetupSQPoll != 0 && flags == 0 {
        return flushed, nil // No syscall needed
    }
    
    return iouring_enter(r.fd, uint32(flushed), 0, flags, nil)
}
```

## Error Handling

```go
// Error wraps io_uring errors with context
type Error struct {
    Op     string
    Ring   int
    Errno  syscall.Errno
    Detail string
}

func (e *Error) Error() string {
    return fmt.Sprintf("iouring %s: %v", e.Op, e.Errno)
}

func (e *Error) Unwrap() error {
    return e.Errno
}

// Common errors
var (
    ErrRingClosed    = errors.New("ring closed")
    ErrSQFull        = errors.New("submission queue full")
    ErrCQOverflow    = errors.New("completion queue overflow")
    ErrNotSupported  = errors.New("operation not supported on this kernel")
)
```

## Integration with Go Runtime

### Goroutine Parking

```go
// For blocking waits without burning CPU
func (r *Ring) WaitCQE(ctx context.Context) (*CQE, error) {
    // Try non-blocking first
    if cqe, ok := r.PeekCQE(); ok {
        return cqe, nil
    }
    
    // Set up eventfd for notification
    if r.eventfd == 0 {
        fd, err := syscall.Eventfd(0, syscall.EFD_NONBLOCK|syscall.EFD_CLOEXEC)
        if err != nil {
            return nil, err
        }
        r.eventfd = fd
        if err := r.RegisterEventfd(fd); err != nil {
            return nil, err
        }
    }
    
    // Use netpoll or select to wait
    for {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }
        
        // Poll eventfd
        var pfd syscall.PollFd
        pfd.Fd = int32(r.eventfd)
        pfd.Events = syscall.POLLIN
        
        _, err := syscall.Poll([]syscall.PollFd{pfd}, 100)
        if err != nil && err != syscall.EINTR {
            return nil, err
        }
        
        // Check for CQE
        if cqe, ok := r.PeekCQE(); ok {
            return cqe, nil
        }
    }
}
```

### mmap Memory Management

```go
// Ensure mmap'd memory isn't moved by GC
func mapRing(fd int, params *Params) (*mappedRing, error) {
    sqSize := params.SQOff.Array + params.SQEntries*uint32(unsafe.Sizeof(uint32(0)))
    cqSize := params.CQOff.CQEs + params.CQEntries*uint32(unsafe.Sizeof(CQE{}))
    
    // Single mmap if supported
    if params.Features&IORING_FEAT_SINGLE_MMAP != 0 {
        if cqSize > sqSize {
            sqSize = cqSize
        }
    }
    
    sqPtr, err := syscall.Mmap(fd, IORING_OFF_SQ_RING, int(sqSize),
        syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
    if err != nil {
        return nil, err
    }
    
    // ... setup ring pointers from mapped memory ...
    
    return &mappedRing{
        sqRing: sqPtr,
        cqRing: cqPtr, // May be same as sqPtr with SINGLE_MMAP
        sqes:   sqesPtr,
    }, nil
}
```
