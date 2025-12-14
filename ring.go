//go:build linux

// Package iouring provides a high-performance, zero-allocation io_uring
// interface for Go.
package iouring

import (
	"errors"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/behrlich/go-iouring/internal/sys"
)

// Common errors
var (
	ErrRingClosed   = errors.New("iouring: ring closed")
	ErrSQFull       = errors.New("iouring: submission queue full")
	ErrCQOverflow   = errors.New("iouring: completion queue overflow")
	ErrNotSupported = errors.New("iouring: operation not supported on this kernel")
)

// Timespec is a time specification for timeout operations.
type Timespec = sys.Timespec

// Ring represents an io_uring instance.
type Ring struct {
	fd       int
	params   sys.Params
	features uint32

	// Submission queue
	sqRing    []byte       // mmap'd SQ ring
	sqEntries uint32       // Number of SQ entries
	sqMask    uint32       // SQ ring mask
	sqHead    *uint32      // Pointer into mmap'd region
	sqTail    *uint32      // Pointer into mmap'd region
	sqFlags   *uint32      // Pointer into mmap'd region
	sqDropped *uint32      // Pointer into mmap'd region
	sqArray   []uint32     // SQ index array (into sqes)
	sqes      []sys.SQE    // SQE array
	sqesMmap  []byte       // mmap'd SQE region

	// Completion queue
	cqRing    []byte       // mmap'd CQ ring (may share with sqRing)
	cqEntries uint32       // Number of CQ entries
	cqMask    uint32       // CQ ring mask
	cqHead    *uint32      // Pointer into mmap'd region
	cqTail    *uint32      // Pointer into mmap'd region
	cqFlags   *uint32      // Pointer into mmap'd region
	cqOverflow *uint32     // Pointer into mmap'd region
	cqes      []sys.CQE    // CQE array (view into mmap)

	// Internal state
	sqLock    sync.Mutex   // Protects SQ access for concurrent use
	sqPending uint32       // Number of SQEs pending submission
	closed    atomic.Bool
}

// Option configures ring setup.
type Option func(*sys.Params)

// WithSQPoll enables kernel-side SQ polling.
// This eliminates syscalls for submission but requires CAP_SYS_NICE
// or a recent kernel with io_uring permissions.
func WithSQPoll() Option {
	return func(p *sys.Params) {
		p.Flags |= sys.IORING_SETUP_SQPOLL
	}
}

// WithSQPollCPU pins the SQPOLL kernel thread to a specific CPU.
// Must be used with WithSQPoll.
func WithSQPollCPU(cpu uint32) Option {
	return func(p *sys.Params) {
		p.Flags |= sys.IORING_SETUP_SQ_AFF
		p.SQThreadCPU = cpu
	}
}

// WithSQPollIdle sets the idle timeout (milliseconds) for SQPOLL thread.
func WithSQPollIdle(ms uint32) Option {
	return func(p *sys.Params) {
		p.SQThreadIdle = ms
	}
}

// WithIOPoll enables I/O polling for completions.
// Only works with file descriptors that support polling (e.g., NVMe).
func WithIOPoll() Option {
	return func(p *sys.Params) {
		p.Flags |= sys.IORING_SETUP_IOPOLL
	}
}

// WithCQSize sets a custom completion queue size.
// By default CQ size is 2x SQ size.
func WithCQSize(size uint32) Option {
	return func(p *sys.Params) {
		p.Flags |= sys.IORING_SETUP_CQSIZE
		p.CQEntries = size
	}
}

// WithSingleIssuer indicates only one task will submit to this ring.
// Enables optimizations in the kernel.
func WithSingleIssuer() Option {
	return func(p *sys.Params) {
		p.Flags |= sys.IORING_SETUP_SINGLE_ISSUER
	}
}

// WithDeferTaskrun defers task work until the next io_uring_enter call.
// Useful for batching completions. Requires SINGLE_ISSUER.
func WithDeferTaskrun() Option {
	return func(p *sys.Params) {
		p.Flags |= sys.IORING_SETUP_DEFER_TASKRUN | sys.IORING_SETUP_SINGLE_ISSUER
	}
}

// WithCoopTaskrun enables cooperative task running.
func WithCoopTaskrun() Option {
	return func(p *sys.Params) {
		p.Flags |= sys.IORING_SETUP_COOP_TASKRUN
	}
}

// WithFlags sets arbitrary setup flags.
func WithFlags(flags uint32) Option {
	return func(p *sys.Params) {
		p.Flags |= flags
	}
}

// New creates a new io_uring instance.
// entries specifies the minimum number of submission queue entries
// (will be rounded up to a power of 2 by the kernel).
func New(entries uint32, opts ...Option) (*Ring, error) {
	if entries == 0 {
		return nil, syscall.EINVAL
	}

	params := sys.Params{}
	for _, opt := range opts {
		opt(&params)
	}

	fd, err := sys.Setup(entries, &params)
	if err != nil {
		return nil, err
	}

	r := &Ring{
		fd:       fd,
		params:   params,
		features: params.Features,
	}

	if err := r.mapRings(); err != nil {
		syscall.Close(fd)
		return nil, err
	}

	return r, nil
}

// mapRings maps the SQ, CQ, and SQE arrays into memory.
func (r *Ring) mapRings() error {
	p := &r.params

	// Calculate sizes
	sqRingSize := p.SQOff.Array + p.SQEntries*4
	cqRingSize := p.CQOff.CQEs + p.CQEntries*uint32(unsafe.Sizeof(sys.CQE{}))

	// If SINGLE_MMAP is supported, SQ and CQ share memory
	singleMmap := p.Features&sys.IORING_FEAT_SINGLE_MMAP != 0
	if singleMmap {
		if cqRingSize > sqRingSize {
			sqRingSize = cqRingSize
		}
	}

	// Map SQ ring
	var err error
	r.sqRing, err = sys.Mmap(r.fd, sys.IORING_OFF_SQ_RING, int(sqRingSize),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_POPULATE)
	if err != nil {
		return err
	}

	// Map CQ ring (may be same as SQ ring)
	if singleMmap {
		r.cqRing = r.sqRing
	} else {
		r.cqRing, err = sys.Mmap(r.fd, sys.IORING_OFF_CQ_RING, int(cqRingSize),
			syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_POPULATE)
		if err != nil {
			sys.Munmap(r.sqRing)
			return err
		}
	}

	// Map SQE array
	sqeSize := p.SQEntries * uint32(unsafe.Sizeof(sys.SQE{}))
	r.sqesMmap, err = sys.Mmap(r.fd, sys.IORING_OFF_SQES, int(sqeSize),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_POPULATE)
	if err != nil {
		if !singleMmap {
			sys.Munmap(r.cqRing)
		}
		sys.Munmap(r.sqRing)
		return err
	}

	// Set up SQ pointers
	r.sqEntries = *(*uint32)(unsafe.Pointer(&r.sqRing[p.SQOff.RingEntries]))
	r.sqMask = *(*uint32)(unsafe.Pointer(&r.sqRing[p.SQOff.RingMask]))
	r.sqHead = (*uint32)(unsafe.Pointer(&r.sqRing[p.SQOff.Head]))
	r.sqTail = (*uint32)(unsafe.Pointer(&r.sqRing[p.SQOff.Tail]))
	r.sqFlags = (*uint32)(unsafe.Pointer(&r.sqRing[p.SQOff.Flags]))
	r.sqDropped = (*uint32)(unsafe.Pointer(&r.sqRing[p.SQOff.Dropped]))

	// SQ array is uint32 indices into the SQE array
	sqArrayPtr := unsafe.Pointer(&r.sqRing[p.SQOff.Array])
	r.sqArray = unsafe.Slice((*uint32)(sqArrayPtr), r.sqEntries)

	// SQE array
	sqesPtr := unsafe.Pointer(&r.sqesMmap[0])
	r.sqes = unsafe.Slice((*sys.SQE)(sqesPtr), p.SQEntries)

	// Set up CQ pointers
	r.cqEntries = *(*uint32)(unsafe.Pointer(&r.cqRing[p.CQOff.RingEntries]))
	r.cqMask = *(*uint32)(unsafe.Pointer(&r.cqRing[p.CQOff.RingMask]))
	r.cqHead = (*uint32)(unsafe.Pointer(&r.cqRing[p.CQOff.Head]))
	r.cqTail = (*uint32)(unsafe.Pointer(&r.cqRing[p.CQOff.Tail]))
	r.cqFlags = (*uint32)(unsafe.Pointer(&r.cqRing[p.CQOff.Flags]))
	r.cqOverflow = (*uint32)(unsafe.Pointer(&r.cqRing[p.CQOff.Overflow]))

	// CQE array
	cqesPtr := unsafe.Pointer(&r.cqRing[p.CQOff.CQEs])
	r.cqes = unsafe.Slice((*sys.CQE)(cqesPtr), r.cqEntries)

	return nil
}

// Close closes the ring and releases all resources.
func (r *Ring) Close() error {
	if r.closed.Swap(true) {
		return nil // Already closed
	}

	// Unmap CQ if separate from SQ
	if r.params.Features&sys.IORING_FEAT_SINGLE_MMAP == 0 && r.cqRing != nil {
		sys.Munmap(r.cqRing)
	}

	// Unmap SQ and SQEs
	if r.sqRing != nil {
		sys.Munmap(r.sqRing)
	}
	if r.sqesMmap != nil {
		sys.Munmap(r.sqesMmap)
	}

	return syscall.Close(r.fd)
}

// Fd returns the ring file descriptor.
func (r *Ring) Fd() int {
	return r.fd
}

// Features returns the feature flags from io_uring_params.
func (r *Ring) Features() uint32 {
	return r.features
}

// HasFeature checks if a specific feature is supported.
func (r *Ring) HasFeature(feat uint32) bool {
	return r.features&feat != 0
}

// SQEntries returns the number of submission queue entries.
func (r *Ring) SQEntries() uint32 {
	return r.sqEntries
}

// CQEntries returns the number of completion queue entries.
func (r *Ring) CQEntries() uint32 {
	return r.cqEntries
}

// SQReady returns the number of SQEs ready for submission.
func (r *Ring) SQReady() uint32 {
	return r.sqPending
}

// SQSpace returns the available space in the submission queue.
func (r *Ring) SQSpace() uint32 {
	head := atomic.LoadUint32(r.sqHead)
	tail := atomic.LoadUint32(r.sqTail)
	return r.sqEntries - (tail - head)
}

// CQReady returns the number of CQEs ready for consumption.
func (r *Ring) CQReady() uint32 {
	head := atomic.LoadUint32(r.cqHead)
	tail := atomic.LoadUint32(r.cqTail)
	return tail - head
}

// needsWakeup returns true if SQPOLL thread needs waking.
func (r *Ring) needsWakeup() bool {
	if r.params.Flags&sys.IORING_SETUP_SQPOLL == 0 {
		return false
	}
	return atomic.LoadUint32(r.sqFlags)&sys.IORING_SQ_NEED_WAKEUP != 0
}

// Submit submits all pending SQEs to the kernel.
// Returns the number of SQEs submitted.
func (r *Ring) Submit() (int, error) {
	if r.closed.Load() {
		return 0, ErrRingClosed
	}

	r.sqLock.Lock()
	submitted := r.sqPending
	if submitted == 0 {
		r.sqLock.Unlock()
		return 0, nil
	}

	// Update the SQ tail with release semantics
	tail := atomic.LoadUint32(r.sqTail)
	atomic.StoreUint32(r.sqTail, tail+submitted)
	r.sqPending = 0
	r.sqLock.Unlock()

	// Determine if we need a syscall
	var flags uint32
	if r.needsWakeup() {
		flags |= sys.IORING_ENTER_SQ_WAKEUP
	}

	// If SQPOLL and no wakeup needed, no syscall required
	if r.params.Flags&sys.IORING_SETUP_SQPOLL != 0 && flags == 0 {
		return int(submitted), nil
	}

	n, err := sys.Enter(r.fd, submitted, 0, flags, nil)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// SubmitAndWait submits pending SQEs and waits for at least n completions.
func (r *Ring) SubmitAndWait(n uint32) (int, error) {
	if r.closed.Load() {
		return 0, ErrRingClosed
	}

	r.sqLock.Lock()
	submitted := r.sqPending
	if submitted > 0 {
		tail := atomic.LoadUint32(r.sqTail)
		atomic.StoreUint32(r.sqTail, tail+submitted)
		r.sqPending = 0
	}
	r.sqLock.Unlock()

	var flags uint32 = sys.IORING_ENTER_GETEVENTS
	if r.needsWakeup() {
		flags |= sys.IORING_ENTER_SQ_WAKEUP
	}

	result, err := sys.Enter(r.fd, submitted, n, flags, nil)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// RegisterEventfd registers an eventfd for completion notification.
func (r *Ring) RegisterEventfd(eventfd int) error {
	return sys.RegisterEventfd(r.fd, eventfd)
}

// UnregisterEventfd removes the registered eventfd.
func (r *Ring) UnregisterEventfd() error {
	return sys.UnregisterEventfd(r.fd)
}

// RegisterBuffers registers fixed buffers for I/O operations.
func (r *Ring) RegisterBuffers(bufs [][]byte) error {
	if len(bufs) == 0 {
		return syscall.EINVAL
	}

	iovecs := make([]syscall.Iovec, len(bufs))
	for i, buf := range bufs {
		if len(buf) > 0 {
			iovecs[i].Base = &buf[0]
			iovecs[i].Len = uint64(len(buf))
		}
	}

	return sys.RegisterBuffers(r.fd, iovecs)
}

// UnregisterBuffers removes registered buffers.
func (r *Ring) UnregisterBuffers() error {
	return sys.UnregisterBuffers(r.fd)
}

// RegisterFiles registers fixed file descriptors.
func (r *Ring) RegisterFiles(fds []int) error {
	if len(fds) == 0 {
		return syscall.EINVAL
	}

	fds32 := make([]int32, len(fds))
	for i, fd := range fds {
		fds32[i] = int32(fd)
	}

	return sys.RegisterFiles(r.fd, fds32)
}

// UnregisterFiles removes registered files.
func (r *Ring) UnregisterFiles() error {
	return sys.UnregisterFiles(r.fd)
}
