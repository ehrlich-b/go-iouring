//go:build linux

package iouring

import (
	"context"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/behrlich/go-iouring/internal/sys"
)

// PeekCQE returns the next completion queue entry without blocking.
// Returns userData, result, flags, and whether a CQE was available.
// This is the zero-allocation path - use this in hot loops.
func (r *Ring) PeekCQE() (userData uint64, res int32, flags uint32, ok bool) {
	head := atomic.LoadUint32(r.cqHead)
	tail := atomic.LoadUint32(r.cqTail)

	if head == tail {
		return 0, 0, 0, false
	}

	idx := head & r.cqMask
	cqe := &r.cqes[idx]

	return cqe.UserData, cqe.Res, cqe.Flags, true
}

// SeenCQE advances the CQ head, marking the current CQE as consumed.
// Must be called after processing a CQE from PeekCQE.
func (r *Ring) SeenCQE() {
	head := atomic.LoadUint32(r.cqHead)
	atomic.StoreUint32(r.cqHead, head+1)
}

// SeenCQEs advances the CQ head by n entries.
func (r *Ring) SeenCQEs(n uint32) {
	head := atomic.LoadUint32(r.cqHead)
	atomic.StoreUint32(r.cqHead, head+n)
}

// WaitCQE waits for at least one CQE to be available.
// Returns userData, result, flags, or an error.
// Does NOT automatically advance the CQ head - call SeenCQE after processing.
func (r *Ring) WaitCQE() (userData uint64, res int32, flags uint32, err error) {
	if r.closed.Load() {
		return 0, 0, 0, ErrRingClosed
	}

	// Try non-blocking first
	if userData, res, flags, ok := r.PeekCQE(); ok {
		return userData, res, flags, nil
	}

	// Need to wait - submit pending and wait for 1 completion
	_, err = r.SubmitAndWait(1)
	if err != nil {
		return 0, 0, 0, err
	}

	// Should have a CQE now
	if userData, res, flags, ok := r.PeekCQE(); ok {
		return userData, res, flags, nil
	}

	// This shouldn't happen
	return 0, 0, 0, syscall.EAGAIN
}

// WaitCQETimeout waits for a CQE with a timeout.
// Returns userData, result, flags, or an error (syscall.ETIME on timeout).
func (r *Ring) WaitCQETimeout(timeout time.Duration) (userData uint64, res int32, flags uint32, err error) {
	if r.closed.Load() {
		return 0, 0, 0, ErrRingClosed
	}

	// Try non-blocking first
	if userData, res, flags, ok := r.PeekCQE(); ok {
		return userData, res, flags, nil
	}

	// Need to wait with timeout
	if !r.HasFeature(sys.IORING_FEAT_EXT_ARG) {
		// Fallback: poll in a loop (less efficient)
		return r.waitCQETimeoutPoll(timeout)
	}

	ts := sys.Timespec{
		Sec:  int64(timeout / time.Second),
		Nsec: int64(timeout % time.Second),
	}

	arg := sys.GetEventsArg{
		Ts: uint64(uintptr(unsafe.Pointer(&ts))),
	}

	r.sqLock.Lock()
	submitted := r.sqPending
	if submitted > 0 {
		tail := atomic.LoadUint32(r.sqTail)
		atomic.StoreUint32(r.sqTail, tail+submitted)
		r.sqPending = 0
	}
	r.sqLock.Unlock()

	_, err = sys.EnterExt(r.fd, submitted, 1, sys.IORING_ENTER_GETEVENTS, &arg)
	if err != nil {
		return 0, 0, 0, err
	}

	if userData, res, flags, ok := r.PeekCQE(); ok {
		return userData, res, flags, nil
	}

	return 0, 0, 0, syscall.ETIME
}

// waitCQETimeoutPoll is a fallback for kernels without EXT_ARG support.
func (r *Ring) waitCQETimeoutPoll(timeout time.Duration) (userData uint64, res int32, flags uint32, err error) {
	deadline := time.Now().Add(timeout)

	for {
		if userData, res, flags, ok := r.PeekCQE(); ok {
			return userData, res, flags, nil
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			return 0, 0, 0, syscall.ETIME
		}

		// Short sleep to avoid busy-waiting
		sleepTime := remaining
		if sleepTime > 10*time.Millisecond {
			sleepTime = 10 * time.Millisecond
		}

		_, err := r.SubmitAndWait(1)
		if err == nil {
			continue
		}
		if err == syscall.EINTR {
			continue
		}
		return 0, 0, 0, err
	}
}

// WaitCQEContext waits for a CQE with context cancellation support.
func (r *Ring) WaitCQEContext(ctx context.Context) (userData uint64, res int32, flags uint32, err error) {
	if r.closed.Load() {
		return 0, 0, 0, ErrRingClosed
	}

	// Try non-blocking first
	if userData, res, flags, ok := r.PeekCQE(); ok {
		return userData, res, flags, nil
	}

	// Poll in a loop checking context
	for {
		select {
		case <-ctx.Done():
			return 0, 0, 0, ctx.Err()
		default:
		}

		// Try with short timeout
		userData, res, flags, err := r.WaitCQETimeout(100 * time.Millisecond)
		if err == syscall.ETIME {
			continue
		}
		return userData, res, flags, err
	}
}

// ForEachCQE iterates over all available CQEs.
// The callback receives userData, result, and flags for each CQE.
// Returns the number of CQEs processed.
// The CQ head is advanced after all processing is complete.
func (r *Ring) ForEachCQE(fn func(userData uint64, res int32, flags uint32) bool) int {
	head := atomic.LoadUint32(r.cqHead)
	tail := atomic.LoadUint32(r.cqTail)
	count := 0

	for head != tail {
		idx := head & r.cqMask
		cqe := &r.cqes[idx]

		if !fn(cqe.UserData, cqe.Res, cqe.Flags) {
			break
		}

		head++
		count++
	}

	if count > 0 {
		atomic.StoreUint32(r.cqHead, head)
	}

	return count
}

// DrainCQEs processes all available CQEs and advances the head.
// Returns the number of CQEs drained.
func (r *Ring) DrainCQEs() int {
	head := atomic.LoadUint32(r.cqHead)
	tail := atomic.LoadUint32(r.cqTail)
	count := int(tail - head)

	if count > 0 {
		atomic.StoreUint32(r.cqHead, tail)
	}

	return count
}

// CQOverflow returns the number of CQE overflows (dropped completions).
func (r *Ring) CQOverflow() uint32 {
	return atomic.LoadUint32(r.cqOverflow)
}

// ResultError converts a CQE result to an error if negative.
// Returns nil if the result is non-negative.
func ResultError(res int32) error {
	if res >= 0 {
		return nil
	}
	return syscall.Errno(-res)
}
