//go:build linux

package sys

import (
	"syscall"
	"unsafe"
)

// Setup creates a new io_uring instance.
// Returns the ring file descriptor on success, or an error.
func Setup(entries uint32, params *Params) (int, error) {
	fd, _, errno := syscall.Syscall(
		SYS_IO_URING_SETUP,
		uintptr(entries),
		uintptr(unsafe.Pointer(params)),
		0,
	)
	if errno != 0 {
		return 0, errno
	}
	return int(fd), nil
}

// Enter submits SQEs and/or waits for CQEs.
// toSubmit: number of SQEs to submit
// minComplete: minimum CQEs to wait for (if flags includes IORING_ENTER_GETEVENTS)
// flags: IORING_ENTER_* flags
// sig: optional signal mask (can be nil, pass unsafe.Pointer to sigset_t)
//
// Uses Syscall6 (not RawSyscall) to properly integrate with Go scheduler.
func Enter(fd int, toSubmit, minComplete, flags uint32, sig unsafe.Pointer) (int, error) {
	var sigPtr uintptr
	var sigSz uintptr
	if sig != nil {
		sigPtr = uintptr(sig)
		sigSz = 8 // sizeof(sigset_t) on Linux x86_64 is 128 bytes / 8 = 16 uint64s, but we pass size in bytes
	}

	n, _, errno := syscall.Syscall6(
		SYS_IO_URING_ENTER,
		uintptr(fd),
		uintptr(toSubmit),
		uintptr(minComplete),
		uintptr(flags),
		sigPtr,
		sigSz,
	)
	if errno != 0 {
		return 0, errno
	}
	return int(n), nil
}

// EnterExt uses the extended enter argument (IORING_ENTER_EXT_ARG).
func EnterExt(fd int, toSubmit, minComplete, flags uint32, arg *GetEventsArg) (int, error) {
	n, _, errno := syscall.Syscall6(
		SYS_IO_URING_ENTER,
		uintptr(fd),
		uintptr(toSubmit),
		uintptr(minComplete),
		uintptr(flags|IORING_ENTER_EXT_ARG),
		uintptr(unsafe.Pointer(arg)),
		unsafe.Sizeof(*arg),
	)
	if errno != 0 {
		return 0, errno
	}
	return int(n), nil
}

// Register performs ring registration operations.
// opcode: IORING_REGISTER_* or IORING_UNREGISTER_*
// arg: operation-specific argument (can be nil)
// nrArgs: number of arguments
func Register(fd int, opcode uint32, arg unsafe.Pointer, nrArgs uint32) error {
	_, _, errno := syscall.Syscall6(
		SYS_IO_URING_REGISTER,
		uintptr(fd),
		uintptr(opcode),
		uintptr(arg),
		uintptr(nrArgs),
		0,
		0,
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// RegisterBuffers registers fixed buffers for I/O.
func RegisterBuffers(fd int, iovecs []syscall.Iovec) error {
	if len(iovecs) == 0 {
		return syscall.EINVAL
	}
	return Register(fd, IORING_REGISTER_BUFFERS,
		unsafe.Pointer(&iovecs[0]), uint32(len(iovecs)))
}

// UnregisterBuffers removes registered buffers.
func UnregisterBuffers(fd int) error {
	return Register(fd, IORING_UNREGISTER_BUFFERS, nil, 0)
}

// RegisterFiles registers fixed file descriptors.
func RegisterFiles(fd int, fds []int32) error {
	if len(fds) == 0 {
		return syscall.EINVAL
	}
	return Register(fd, IORING_REGISTER_FILES,
		unsafe.Pointer(&fds[0]), uint32(len(fds)))
}

// UnregisterFiles removes registered files.
func UnregisterFiles(fd int) error {
	return Register(fd, IORING_UNREGISTER_FILES, nil, 0)
}

// RegisterEventfd registers an eventfd for completion notification.
func RegisterEventfd(fd int, eventfd int) error {
	efd := int32(eventfd)
	return Register(fd, IORING_REGISTER_EVENTFD, unsafe.Pointer(&efd), 1)
}

// UnregisterEventfd removes the registered eventfd.
func UnregisterEventfd(fd int) error {
	return Register(fd, IORING_UNREGISTER_EVENTFD, nil, 0)
}

// RegisterEventfdAsync registers eventfd for async completion only.
func RegisterEventfdAsync(fd int, eventfd int) error {
	efd := int32(eventfd)
	return Register(fd, IORING_REGISTER_EVENTFD_ASYNC, unsafe.Pointer(&efd), 1)
}

// RegisterProbe queries supported operations.
func RegisterProbe(fd int, probe *Probe) error {
	return Register(fd, IORING_REGISTER_PROBE,
		unsafe.Pointer(probe), uint32(IORING_OP_LAST))
}

// Mmap wraps the mmap syscall for mapping ring buffers.
func Mmap(fd int, offset uint64, length int, prot, flags int) ([]byte, error) {
	data, err := syscall.Mmap(fd, int64(offset), length, prot, flags)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Munmap unmaps a previously mapped region.
func Munmap(data []byte) error {
	return syscall.Munmap(data)
}
