package sys

// SQE is the Submission Queue Entry (64 bytes).
// This matches struct io_uring_sqe from the kernel.
// The struct uses unions extensively; we represent the full 64 bytes
// and provide accessor methods for different interpretations.
type SQE struct {
	Opcode      uint8  // Operation code (IORING_OP_*)
	Flags       uint8  // IOSQE_* flags
	Ioprio      uint16 // Request priority or op-specific flags
	Fd          int32  // File descriptor
	Off         uint64 // Offset or addr2 (union)
	Addr        uint64 // Buffer address or splice_off_in (union)
	Len         uint32 // Buffer length or number of iovecs
	OpFlags     uint32 // Op-specific flags (rw_flags, fsync_flags, etc.)
	UserData    uint64 // User data - passed back in CQE
	BufIndex    uint16 // Buffer index or buffer group (union)
	Personality uint16 // Personality for credentials
	SpliceFdIn  int32  // Splice input fd or file_index (union)
	Addr3       uint64 // Additional address field
	_pad2       [1]uint64
}

// CQE is the Completion Queue Entry (16 bytes).
// This matches struct io_uring_cqe from the kernel.
type CQE struct {
	UserData uint64 // User data from the SQE
	Res      int32  // Result (bytes transferred or negative errno)
	Flags    uint32 // IORING_CQE_F_* flags
}

// CQE32 is the extended 32-byte CQE (when IORING_SETUP_CQE32 is used).
type CQE32 struct {
	CQE
	BigCQE [16]byte // Extra 16 bytes for extended data
}

// Params is passed to io_uring_setup and returned with ring parameters.
// This matches struct io_uring_params from the kernel.
type Params struct {
	SQEntries    uint32
	CQEntries    uint32
	Flags        uint32
	SQThreadCPU  uint32
	SQThreadIdle uint32
	Features     uint32
	WQFd         uint32
	Resv         [3]uint32
	SQOff        SQRingOffsets
	CQOff        CQRingOffsets
}

// SQRingOffsets contains offsets into the SQ ring mmap region.
type SQRingOffsets struct {
	Head        uint32
	Tail        uint32
	RingMask    uint32
	RingEntries uint32
	Flags       uint32
	Dropped     uint32
	Array       uint32
	Resv1       uint32
	UserAddr    uint64
}

// CQRingOffsets contains offsets into the CQ ring mmap region.
type CQRingOffsets struct {
	Head        uint32
	Tail        uint32
	RingMask    uint32
	RingEntries uint32
	Overflow    uint32
	CQEs        uint32
	Flags       uint32
	Resv1       uint32
	UserAddr    uint64
}

// ProbeOp describes support for a single operation.
type ProbeOp struct {
	Op    uint8
	Resv  uint8
	Flags uint16
	Resv2 uint32
}

// Probe is the result of IORING_REGISTER_PROBE.
type Probe struct {
	LastOp uint8
	OpsLen uint8
	Resv   uint16
	Resv2  [3]uint32
	Ops    [IORING_OP_LAST]ProbeOp
}

// IO_URING_OP_SUPPORTED indicates an op is supported in ProbeOp.Flags
const IO_URING_OP_SUPPORTED uint16 = 1 << 0

// Timespec matches struct __kernel_timespec.
type Timespec struct {
	Sec  int64
	Nsec int64
}

// FilesUpdate is used with IORING_REGISTER_FILES_UPDATE.
type FilesUpdate struct {
	Offset uint32
	Resv   uint32
	Fds    uint64 // Pointer to fd array
}

// RsrcRegister is used with IORING_REGISTER_BUFFERS2/FILES2.
type RsrcRegister struct {
	Nr    uint32
	Flags uint32
	Resv2 uint64
	Data  uint64 // Pointer to data
	Tags  uint64 // Pointer to tags
}

// RsrcUpdate is used with IORING_REGISTER_BUFFERS_UPDATE/FILES_UPDATE2.
type RsrcUpdate struct {
	Offset uint32
	Resv   uint32
	Data   uint64 // Pointer to data
	Tags   uint64 // Pointer to tags
	Nr     uint32
	Resv2  uint32
}

// GetEventsArg is used with IORING_ENTER_EXT_ARG.
type GetEventsArg struct {
	Sigmask   uint64
	SigmaskSz uint32
	Pad       uint32
	Ts        uint64
}

// BufRingSetup is used with IORING_REGISTER_PBUF_RING.
type BufRingSetup struct {
	BGid    uint16
	Nentries uint16
	Flags   uint32
	Resv    [3]uint64
	RingAddr uint64
}

// Buf describes a provided buffer.
type Buf struct {
	Addr uint64
	Len  uint32
	Bid  uint16
	Resv uint16
}

// BufRing is the header for a provided buffer ring.
type BufRing struct {
	Resv1 uint64
	Resv2 uint32
	Resv3 uint16
	Tail  uint16
	// Followed by Buf entries
}

// SQE accessor methods for union fields

// SetAddr2 sets the addr2 field (alias for Off).
func (s *SQE) SetAddr2(addr2 uint64) {
	s.Off = addr2
}

// SetSpliceOffIn sets the splice_off_in field (alias for Addr).
func (s *SQE) SetSpliceOffIn(off uint64) {
	s.Addr = off
}

// SetBufGroup sets the buf_group field (alias for BufIndex).
func (s *SQE) SetBufGroup(group uint16) {
	s.BufIndex = group
}

// SetFileIndex sets the file_index field (alias for SpliceFdIn).
func (s *SQE) SetFileIndex(index int32) {
	s.SpliceFdIn = index
}

// Reset clears the SQE to zero values.
func (s *SQE) Reset() {
	*s = SQE{}
}

// CQE accessor methods

// GetBufID extracts the buffer ID from flags when IORING_CQE_F_BUFFER is set.
func (c *CQE) GetBufID() uint16 {
	return uint16(c.Flags >> 16)
}

// HasMore returns true if more CQEs are coming (multishot).
func (c *CQE) HasMore() bool {
	return c.Flags&IORING_CQE_F_MORE != 0
}

// IsNotification returns true if this is a notification CQE (zero-copy).
func (c *CQE) IsNotification() bool {
	return c.Flags&IORING_CQE_F_NOTIF != 0
}
