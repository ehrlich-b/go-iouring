//go:build linux

package iouring

import (
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/behrlich/go-iouring/internal/sys"
)

// getSQE returns the next available SQE, or nil if the queue is full.
// The returned SQE is zeroed and ready for use.
// NOT thread-safe; caller must hold sqLock.
func (r *Ring) getSQE() *sys.SQE {
	head := atomic.LoadUint32(r.sqHead)
	tail := atomic.LoadUint32(r.sqTail) + r.sqPending

	// Check if queue is full
	if tail-head >= r.sqEntries {
		return nil
	}

	idx := tail & r.sqMask
	sqe := &r.sqes[idx]
	sqe.Reset()

	// Update the SQ array to point to this SQE
	r.sqArray[idx] = uint32(idx)
	r.sqPending++

	return sqe
}

// GetSQE returns the next available SQE, or nil if the queue is full.
// Thread-safe.
func (r *Ring) GetSQE() *sys.SQE {
	r.sqLock.Lock()
	sqe := r.getSQE()
	r.sqLock.Unlock()
	return sqe
}

// PrepNop prepares a NOP operation.
// Useful for testing and waking SQPOLL.
func (r *Ring) PrepNop(userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}
	sqe.Opcode = uint8(sys.IORING_OP_NOP)
	sqe.UserData = userData
	r.sqLock.Unlock()
	return nil
}

// PrepRead prepares a read operation.
// Reads up to len(buf) bytes from fd at offset into buf.
func (r *Ring) PrepRead(fd int, buf []byte, offset uint64, userData uint64) error {
	if len(buf) == 0 {
		return nil
	}

	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_READ)
	sqe.Fd = int32(fd)
	sqe.Addr = uint64(uintptr(unsafe.Pointer(&buf[0])))
	sqe.Len = uint32(len(buf))
	sqe.Off = offset
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepWrite prepares a write operation.
// Writes len(buf) bytes from buf to fd at offset.
func (r *Ring) PrepWrite(fd int, buf []byte, offset uint64, userData uint64) error {
	if len(buf) == 0 {
		return nil
	}

	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_WRITE)
	sqe.Fd = int32(fd)
	sqe.Addr = uint64(uintptr(unsafe.Pointer(&buf[0])))
	sqe.Len = uint32(len(buf))
	sqe.Off = offset
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepReadFixed prepares a read using a pre-registered buffer.
// bufIndex is the index into the registered buffer array.
func (r *Ring) PrepReadFixed(fd int, buf []byte, offset uint64, bufIndex uint16, userData uint64) error {
	if len(buf) == 0 {
		return nil
	}

	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_READ_FIXED)
	sqe.Fd = int32(fd)
	sqe.Addr = uint64(uintptr(unsafe.Pointer(&buf[0])))
	sqe.Len = uint32(len(buf))
	sqe.Off = offset
	sqe.BufIndex = bufIndex
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepWriteFixed prepares a write using a pre-registered buffer.
// bufIndex is the index into the registered buffer array.
func (r *Ring) PrepWriteFixed(fd int, buf []byte, offset uint64, bufIndex uint16, userData uint64) error {
	if len(buf) == 0 {
		return nil
	}

	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_WRITE_FIXED)
	sqe.Fd = int32(fd)
	sqe.Addr = uint64(uintptr(unsafe.Pointer(&buf[0])))
	sqe.Len = uint32(len(buf))
	sqe.Off = offset
	sqe.BufIndex = bufIndex
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepReadv prepares a vectored read operation.
// iovecs must remain valid until the operation completes.
func (r *Ring) PrepReadv(fd int, iovecs []syscall.Iovec, offset uint64, userData uint64) error {
	if len(iovecs) == 0 {
		return nil
	}

	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_READV)
	sqe.Fd = int32(fd)
	sqe.Addr = uint64(uintptr(unsafe.Pointer(&iovecs[0])))
	sqe.Len = uint32(len(iovecs))
	sqe.Off = offset
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepWritev prepares a vectored write operation.
// iovecs must remain valid until the operation completes.
func (r *Ring) PrepWritev(fd int, iovecs []syscall.Iovec, offset uint64, userData uint64) error {
	if len(iovecs) == 0 {
		return nil
	}

	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_WRITEV)
	sqe.Fd = int32(fd)
	sqe.Addr = uint64(uintptr(unsafe.Pointer(&iovecs[0])))
	sqe.Len = uint32(len(iovecs))
	sqe.Off = offset
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepFsync prepares an fsync operation.
// flags can be 0 or IORING_FSYNC_DATASYNC.
func (r *Ring) PrepFsync(fd int, flags uint32, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_FSYNC)
	sqe.Fd = int32(fd)
	sqe.OpFlags = flags
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepTimeout prepares a timeout operation.
// ts specifies the timeout duration.
// count specifies the number of completions to wait for (0 = just timeout).
// flags can include IORING_TIMEOUT_ABS, IORING_TIMEOUT_BOOTTIME, etc.
func (r *Ring) PrepTimeout(ts *sys.Timespec, count uint64, flags uint32, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_TIMEOUT)
	sqe.Fd = -1
	sqe.Addr = uint64(uintptr(unsafe.Pointer(ts)))
	sqe.Len = 1
	sqe.Off = count
	sqe.OpFlags = flags
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepTimeoutRemove prepares a timeout removal operation.
// targetUserData is the userData of the timeout to remove.
func (r *Ring) PrepTimeoutRemove(targetUserData uint64, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_TIMEOUT_REMOVE)
	sqe.Fd = -1
	sqe.Addr = targetUserData
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepLinkTimeout prepares a linked timeout operation.
// Must be used after PrepXxx + SetSQELink to timeout the linked operation.
// ts specifies the timeout duration.
// flags can include IORING_TIMEOUT_ABS, IORING_TIMEOUT_BOOTTIME, etc.
func (r *Ring) PrepLinkTimeout(ts *sys.Timespec, flags uint32, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_LINK_TIMEOUT)
	sqe.Fd = -1
	sqe.Addr = uint64(uintptr(unsafe.Pointer(ts)))
	sqe.Len = 1
	sqe.OpFlags = flags
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepCancel prepares an async cancel operation.
// targetUserData is the userData of the operation to cancel.
// flags can include IORING_ASYNC_CANCEL_*.
func (r *Ring) PrepCancel(targetUserData uint64, flags uint32, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_ASYNC_CANCEL)
	sqe.Fd = -1
	sqe.Addr = targetUserData
	sqe.OpFlags = flags
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepAccept prepares an accept operation.
// addr and addrLen can be nil if peer address isn't needed.
// flags are accept4 flags (e.g., syscall.SOCK_NONBLOCK).
func (r *Ring) PrepAccept(fd int, addr unsafe.Pointer, addrLen *uint32, flags uint32, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_ACCEPT)
	sqe.Fd = int32(fd)
	sqe.Addr = uint64(uintptr(addr))
	sqe.Off = uint64(uintptr(unsafe.Pointer(addrLen)))
	sqe.OpFlags = flags
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepAcceptMultishot prepares a multishot accept operation.
// Each accept generates a CQE with IORING_CQE_F_MORE flag.
func (r *Ring) PrepAcceptMultishot(fd int, addr unsafe.Pointer, addrLen *uint32, flags uint32, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_ACCEPT)
	sqe.Fd = int32(fd)
	sqe.Addr = uint64(uintptr(addr))
	sqe.Off = uint64(uintptr(unsafe.Pointer(addrLen)))
	sqe.OpFlags = flags
	sqe.Ioprio = uint16(sys.IORING_ACCEPT_MULTISHOT)
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepConnect prepares a connect operation.
func (r *Ring) PrepConnect(fd int, addr unsafe.Pointer, addrLen uint32, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_CONNECT)
	sqe.Fd = int32(fd)
	sqe.Addr = uint64(uintptr(addr))
	sqe.Off = uint64(addrLen)
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepSend prepares a send operation.
func (r *Ring) PrepSend(fd int, buf []byte, flags int, userData uint64) error {
	if len(buf) == 0 {
		return nil
	}

	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_SEND)
	sqe.Fd = int32(fd)
	sqe.Addr = uint64(uintptr(unsafe.Pointer(&buf[0])))
	sqe.Len = uint32(len(buf))
	sqe.OpFlags = uint32(flags)
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepRecv prepares a recv operation.
func (r *Ring) PrepRecv(fd int, buf []byte, flags int, userData uint64) error {
	if len(buf) == 0 {
		return nil
	}

	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_RECV)
	sqe.Fd = int32(fd)
	sqe.Addr = uint64(uintptr(unsafe.Pointer(&buf[0])))
	sqe.Len = uint32(len(buf))
	sqe.OpFlags = uint32(flags)
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepRecvMultishot prepares a multishot recv operation.
// Requires buffer group selection (bufGroup).
func (r *Ring) PrepRecvMultishot(fd int, bufGroup uint16, flags int, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_RECV)
	sqe.Fd = int32(fd)
	sqe.Flags = sys.IOSQE_BUFFER_SELECT
	sqe.Ioprio = sys.IORING_RECV_MULTISHOT
	sqe.SetBufGroup(bufGroup)
	sqe.OpFlags = uint32(flags)
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepClose prepares a close operation.
func (r *Ring) PrepClose(fd int, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_CLOSE)
	sqe.Fd = int32(fd)
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepShutdown prepares a shutdown operation.
// how is SHUT_RD, SHUT_WR, or SHUT_RDWR.
func (r *Ring) PrepShutdown(fd int, how int, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_SHUTDOWN)
	sqe.Fd = int32(fd)
	sqe.Len = uint32(how)
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepSendmsg prepares a sendmsg operation.
// msg must remain valid until the operation completes.
func (r *Ring) PrepSendmsg(fd int, msg *syscall.Msghdr, flags int, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_SENDMSG)
	sqe.Fd = int32(fd)
	sqe.Addr = uint64(uintptr(unsafe.Pointer(msg)))
	sqe.Len = 1
	sqe.OpFlags = uint32(flags)
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepRecvmsg prepares a recvmsg operation.
// msg must remain valid until the operation completes.
func (r *Ring) PrepRecvmsg(fd int, msg *syscall.Msghdr, flags int, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_RECVMSG)
	sqe.Fd = int32(fd)
	sqe.Addr = uint64(uintptr(unsafe.Pointer(msg)))
	sqe.Len = 1
	sqe.OpFlags = uint32(flags)
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepSocket prepares an async socket creation operation (5.19+).
// Returns the new socket fd in the CQE result.
func (r *Ring) PrepSocket(domain, typ, protocol int, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_SOCKET)
	sqe.Fd = int32(domain)
	sqe.Off = uint64(typ)
	sqe.Len = uint32(protocol)
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepPollAdd prepares a poll add operation.
// pollMask is POLLIN, POLLOUT, etc.
func (r *Ring) PrepPollAdd(fd int, pollMask uint32, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_POLL_ADD)
	sqe.Fd = int32(fd)
	sqe.OpFlags = pollMask
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepPollAddMultishot prepares a multishot poll operation.
// Generates multiple CQEs until explicitly removed.
func (r *Ring) PrepPollAddMultishot(fd int, pollMask uint32, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_POLL_ADD)
	sqe.Fd = int32(fd)
	sqe.OpFlags = pollMask
	sqe.Len = uint32(sys.IORING_POLL_ADD_MULTI)
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepPollRemove prepares a poll remove operation.
// targetUserData is the userData of the poll to remove.
func (r *Ring) PrepPollRemove(targetUserData uint64, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_POLL_REMOVE)
	sqe.Fd = -1
	sqe.Addr = targetUserData
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepOpenat prepares an openat operation.
// path must be a null-terminated string that remains valid until completion.
func (r *Ring) PrepOpenat(dirfd int, path *byte, flags int, mode uint32, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_OPENAT)
	sqe.Fd = int32(dirfd)
	sqe.Addr = uint64(uintptr(unsafe.Pointer(path)))
	sqe.Len = uint32(mode)
	sqe.OpFlags = uint32(flags)
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepStatx prepares a statx operation.
// path and statxbuf must remain valid until completion.
func (r *Ring) PrepStatx(dirfd int, path *byte, flags, mask int, statxbuf unsafe.Pointer, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_STATX)
	sqe.Fd = int32(dirfd)
	sqe.Addr = uint64(uintptr(unsafe.Pointer(path)))
	sqe.Len = uint32(mask)
	sqe.OpFlags = uint32(flags)
	sqe.Off = uint64(uintptr(statxbuf))
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// PrepSplice prepares a splice operation.
func (r *Ring) PrepSplice(fdIn int, offIn int64, fdOut int, offOut int64, nbytes uint32, flags uint32, userData uint64) error {
	r.sqLock.Lock()
	sqe := r.getSQE()
	if sqe == nil {
		r.sqLock.Unlock()
		return ErrSQFull
	}

	sqe.Opcode = uint8(sys.IORING_OP_SPLICE)
	sqe.Fd = int32(fdOut)
	sqe.SpliceFdIn = int32(fdIn)
	sqe.Len = nbytes
	sqe.Off = uint64(offOut)
	sqe.SetSpliceOffIn(uint64(offIn))
	sqe.OpFlags = flags
	sqe.UserData = userData

	r.sqLock.Unlock()
	return nil
}

// SetSQEFlags sets flags on the most recently prepared SQE.
// Must be called immediately after a Prep* function.
// NOT thread-safe with other Prep calls.
func (r *Ring) SetSQEFlags(flags uint8) {
	r.sqLock.Lock()
	if r.sqPending > 0 {
		tail := atomic.LoadUint32(r.sqTail) + r.sqPending - 1
		idx := tail & r.sqMask
		r.sqes[idx].Flags |= flags
	}
	r.sqLock.Unlock()
}

// SetSQELink links the most recently prepared SQE to the next one.
func (r *Ring) SetSQELink() {
	r.SetSQEFlags(sys.IOSQE_IO_LINK)
}

// SetSQEAsync forces async execution for the most recently prepared SQE.
func (r *Ring) SetSQEAsync() {
	r.SetSQEFlags(sys.IOSQE_ASYNC)
}
