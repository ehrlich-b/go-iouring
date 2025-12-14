// Package sys provides low-level io_uring syscall wrappers and types.
package sys

// Syscall numbers for io_uring (x86_64)
const (
	SYS_IO_URING_SETUP    = 425
	SYS_IO_URING_ENTER    = 426
	SYS_IO_URING_REGISTER = 427
)

// io_uring_op - Operation codes for SQE
type Op uint8

const (
	IORING_OP_NOP Op = iota
	IORING_OP_READV
	IORING_OP_WRITEV
	IORING_OP_FSYNC
	IORING_OP_READ_FIXED
	IORING_OP_WRITE_FIXED
	IORING_OP_POLL_ADD
	IORING_OP_POLL_REMOVE
	IORING_OP_SYNC_FILE_RANGE
	IORING_OP_SENDMSG
	IORING_OP_RECVMSG
	IORING_OP_TIMEOUT
	IORING_OP_TIMEOUT_REMOVE
	IORING_OP_ACCEPT
	IORING_OP_ASYNC_CANCEL
	IORING_OP_LINK_TIMEOUT
	IORING_OP_CONNECT
	IORING_OP_FALLOCATE
	IORING_OP_OPENAT
	IORING_OP_CLOSE
	IORING_OP_FILES_UPDATE
	IORING_OP_STATX
	IORING_OP_READ
	IORING_OP_WRITE
	IORING_OP_FADVISE
	IORING_OP_MADVISE
	IORING_OP_SEND
	IORING_OP_RECV
	IORING_OP_OPENAT2
	IORING_OP_EPOLL_CTL
	IORING_OP_SPLICE
	IORING_OP_PROVIDE_BUFFERS
	IORING_OP_REMOVE_BUFFERS
	IORING_OP_TEE
	IORING_OP_SHUTDOWN
	IORING_OP_RENAMEAT
	IORING_OP_UNLINKAT
	IORING_OP_MKDIRAT
	IORING_OP_SYMLINKAT
	IORING_OP_LINKAT
	IORING_OP_MSG_RING
	IORING_OP_FSETXATTR
	IORING_OP_SETXATTR
	IORING_OP_FGETXATTR
	IORING_OP_GETXATTR
	IORING_OP_SOCKET
	IORING_OP_URING_CMD
	IORING_OP_SEND_ZC
	IORING_OP_SENDMSG_ZC
	IORING_OP_READ_MULTISHOT
	IORING_OP_WAITID
	IORING_OP_FUTEX_WAIT
	IORING_OP_FUTEX_WAKE
	IORING_OP_FUTEX_WAITV
	IORING_OP_FIXED_FD_INSTALL
	IORING_OP_FTRUNCATE
	IORING_OP_BIND
	IORING_OP_LISTEN

	IORING_OP_LAST // Sentinel for bounds checking
)

// SQE flags (IOSQE_*)
const (
	IOSQE_FIXED_FILE       uint8 = 1 << 0 // fd is index into registered files
	IOSQE_IO_DRAIN         uint8 = 1 << 1 // Issue after all previous SQEs complete
	IOSQE_IO_LINK          uint8 = 1 << 2 // Link to next SQE
	IOSQE_IO_HARDLINK      uint8 = 1 << 3 // Hard link - chain continues on error
	IOSQE_ASYNC            uint8 = 1 << 4 // Always use async execution
	IOSQE_BUFFER_SELECT    uint8 = 1 << 5 // Select buffer from buf_group
	IOSQE_CQE_SKIP_SUCCESS uint8 = 1 << 6 // Don't generate CQE if successful
)

// Setup flags (IORING_SETUP_*)
const (
	IORING_SETUP_IOPOLL        uint32 = 1 << 0  // Use I/O polling
	IORING_SETUP_SQPOLL        uint32 = 1 << 1  // Kernel polls SQ
	IORING_SETUP_SQ_AFF        uint32 = 1 << 2  // Pin SQPOLL thread to CPU
	IORING_SETUP_CQSIZE        uint32 = 1 << 3  // App provides CQ size
	IORING_SETUP_CLAMP         uint32 = 1 << 4  // Clamp SQ/CQ to max
	IORING_SETUP_ATTACH_WQ     uint32 = 1 << 5  // Share async workers
	IORING_SETUP_R_DISABLED    uint32 = 1 << 6  // Start ring disabled
	IORING_SETUP_SUBMIT_ALL    uint32 = 1 << 7  // Continue on submit error
	IORING_SETUP_COOP_TASKRUN  uint32 = 1 << 8  // Cooperative task run
	IORING_SETUP_TASKRUN_FLAG  uint32 = 1 << 9  // Set taskrun flag
	IORING_SETUP_SQE128        uint32 = 1 << 10 // 128-byte SQEs
	IORING_SETUP_CQE32         uint32 = 1 << 11 // 32-byte CQEs
	IORING_SETUP_SINGLE_ISSUER uint32 = 1 << 12 // Single task submits
	IORING_SETUP_DEFER_TASKRUN uint32 = 1 << 13 // Defer task work to enter
	IORING_SETUP_NO_MMAP       uint32 = 1 << 14 // App provides memory
	IORING_SETUP_REGISTERED_FD_ONLY uint32 = 1 << 15 // Return registered fd
	IORING_SETUP_NO_SQARRAY    uint32 = 1 << 16 // No SQ array indirection
)

// Feature flags (IORING_FEAT_*)
const (
	IORING_FEAT_SINGLE_MMAP     uint32 = 1 << 0  // SQ/CQ share mmap
	IORING_FEAT_NODROP          uint32 = 1 << 1  // No CQE drops
	IORING_FEAT_SUBMIT_STABLE   uint32 = 1 << 2  // SQE data stable after submit
	IORING_FEAT_RW_CUR_POS      uint32 = 1 << 3  // off=-1 uses file position
	IORING_FEAT_CUR_PERSONALITY uint32 = 1 << 4  // Use current personality
	IORING_FEAT_FAST_POLL       uint32 = 1 << 5  // Fast poll supported
	IORING_FEAT_POLL_32BITS     uint32 = 1 << 6  // 32-bit poll flags
	IORING_FEAT_SQPOLL_NONFIXED uint32 = 1 << 7  // SQPOLL non-fixed files
	IORING_FEAT_EXT_ARG         uint32 = 1 << 8  // Extended argument
	IORING_FEAT_NATIVE_WORKERS  uint32 = 1 << 9  // Native IO workers
	IORING_FEAT_RSRC_TAGS       uint32 = 1 << 10 // Resource tagging
	IORING_FEAT_CQE_SKIP        uint32 = 1 << 11 // CQE skip supported
	IORING_FEAT_LINKED_FILE     uint32 = 1 << 12 // File slot linking
	IORING_FEAT_REG_REG_RING    uint32 = 1 << 13 // Can register ring fd
)

// Enter flags (IORING_ENTER_*)
const (
	IORING_ENTER_GETEVENTS       uint32 = 1 << 0 // Wait for events
	IORING_ENTER_SQ_WAKEUP       uint32 = 1 << 1 // Wake SQPOLL thread
	IORING_ENTER_SQ_WAIT         uint32 = 1 << 2 // Wait for SQ space
	IORING_ENTER_EXT_ARG         uint32 = 1 << 3 // Extended argument
	IORING_ENTER_REGISTERED_RING uint32 = 1 << 4 // Ring fd is registered
)

// Register opcodes (IORING_REGISTER_*)
const (
	IORING_REGISTER_BUFFERS           uint32 = 0
	IORING_UNREGISTER_BUFFERS         uint32 = 1
	IORING_REGISTER_FILES             uint32 = 2
	IORING_UNREGISTER_FILES           uint32 = 3
	IORING_REGISTER_EVENTFD           uint32 = 4
	IORING_UNREGISTER_EVENTFD         uint32 = 5
	IORING_REGISTER_FILES_UPDATE      uint32 = 6
	IORING_REGISTER_EVENTFD_ASYNC     uint32 = 7
	IORING_REGISTER_PROBE             uint32 = 8
	IORING_REGISTER_PERSONALITY       uint32 = 9
	IORING_UNREGISTER_PERSONALITY     uint32 = 10
	IORING_REGISTER_RESTRICTIONS      uint32 = 11
	IORING_REGISTER_ENABLE_RINGS      uint32 = 12
	IORING_REGISTER_FILES2            uint32 = 13
	IORING_REGISTER_FILES_UPDATE2     uint32 = 14
	IORING_REGISTER_BUFFERS2          uint32 = 15
	IORING_REGISTER_BUFFERS_UPDATE    uint32 = 16
	IORING_REGISTER_IOWQ_AFF          uint32 = 17
	IORING_UNREGISTER_IOWQ_AFF        uint32 = 18
	IORING_REGISTER_IOWQ_MAX_WORKERS  uint32 = 19
	IORING_REGISTER_RING_FDS          uint32 = 20
	IORING_UNREGISTER_RING_FDS        uint32 = 21
	IORING_REGISTER_PBUF_RING         uint32 = 22
	IORING_UNREGISTER_PBUF_RING       uint32 = 23
	IORING_REGISTER_SYNC_CANCEL       uint32 = 24
	IORING_REGISTER_FILE_ALLOC_RANGE  uint32 = 25
)

// CQE flags (IORING_CQE_F_*)
const (
	IORING_CQE_F_BUFFER        uint32 = 1 << 0 // Buffer ID in upper 16 bits
	IORING_CQE_F_MORE          uint32 = 1 << 1 // More CQEs coming (multishot)
	IORING_CQE_F_SOCK_NONEMPTY uint32 = 1 << 2 // Socket has more data
	IORING_CQE_F_NOTIF         uint32 = 1 << 3 // Notification (zero-copy)
)

// SQ ring flags
const (
	IORING_SQ_NEED_WAKEUP uint32 = 1 << 0 // SQPOLL needs wakeup
	IORING_SQ_CQ_OVERFLOW uint32 = 1 << 1 // CQ overflow
	IORING_SQ_TASKRUN     uint32 = 1 << 2 // Task work pending
)

// Timeout flags
const (
	IORING_TIMEOUT_ABS           uint32 = 1 << 0
	IORING_TIMEOUT_UPDATE        uint32 = 1 << 1
	IORING_TIMEOUT_BOOTTIME      uint32 = 1 << 2
	IORING_TIMEOUT_REALTIME      uint32 = 1 << 3
	IORING_TIMEOUT_ETIME_SUCCESS uint32 = 1 << 5
	IORING_TIMEOUT_MULTISHOT     uint32 = 1 << 6
)

// Fsync flags
const (
	IORING_FSYNC_DATASYNC uint32 = 1 << 0
)

// Poll flags for multishot
const (
	IORING_POLL_ADD_MULTI       uint32 = 1 << 0
	IORING_POLL_UPDATE_EVENTS   uint32 = 1 << 1
	IORING_POLL_UPDATE_USER_DATA uint32 = 1 << 2
	IORING_POLL_ADD_LEVEL       uint32 = 1 << 3
)

// Accept flags
const (
	IORING_ACCEPT_MULTISHOT uint32 = 1 << 0
)

// Recv/Send flags in ioprio
const (
	IORING_RECVSEND_POLL_FIRST uint16 = 1 << 0
	IORING_RECV_MULTISHOT      uint16 = 1 << 1
	IORING_RECVSEND_FIXED_BUF  uint16 = 1 << 2
	IORING_SEND_ZC_REPORT_USAGE uint16 = 1 << 3
)

// Cancel flags
const (
	IORING_ASYNC_CANCEL_ALL uint32 = 1 << 0
	IORING_ASYNC_CANCEL_FD  uint32 = 1 << 1
	IORING_ASYNC_CANCEL_ANY uint32 = 1 << 2
	IORING_ASYNC_CANCEL_FD_FIXED uint32 = 1 << 3
)

// MSG_RING commands
const (
	IORING_MSG_DATA uint32 = 0
	IORING_MSG_SEND_FD uint32 = 1
)

// mmap offsets for the ring buffers
const (
	IORING_OFF_SQ_RING uint64 = 0
	IORING_OFF_CQ_RING uint64 = 0x8000000
	IORING_OFF_SQES    uint64 = 0x10000000
)

// Magic value for file_index to allocate a direct descriptor
const (
	IORING_FILE_INDEX_ALLOC uint32 = 0xffffffff - 1
)
