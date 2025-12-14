//go:build linux

package iouring

import (
	"github.com/behrlich/go-iouring/internal/sys"
)

// Probe contains information about supported io_uring operations.
type Probe struct {
	probe    sys.Probe
	features uint32
}

// Probe queries the kernel for supported operations.
// Returns a Probe that can be used to check operation support.
func (r *Ring) Probe() (*Probe, error) {
	p := &Probe{
		features: r.features,
	}
	err := sys.RegisterProbe(r.fd, &p.probe)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// SupportsOp returns true if the kernel supports the given operation.
func (p *Probe) SupportsOp(op sys.Op) bool {
	if uint8(op) > p.probe.LastOp {
		return false
	}
	return p.probe.Ops[op].Flags&sys.IO_URING_OP_SUPPORTED != 0
}

// LastOp returns the highest operation code supported by the kernel.
func (p *Probe) LastOp() sys.Op {
	return sys.Op(p.probe.LastOp)
}

// Features returns the feature flags from ring setup.
func (p *Probe) Features() uint32 {
	return p.features
}

// HasFeature returns true if the ring has the given feature.
func (p *Probe) HasFeature(feature uint32) bool {
	return p.features&feature != 0
}

// Ring feature check methods

// HasSingleMmap returns true if SQ and CQ share a single mmap region.
func (r *Ring) HasSingleMmap() bool {
	return r.features&sys.IORING_FEAT_SINGLE_MMAP != 0
}

// HasNoDrop returns true if CQ overflow will block rather than drop.
func (r *Ring) HasNoDrop() bool {
	return r.features&sys.IORING_FEAT_NODROP != 0
}

// HasSubmitStable returns true if buffers don't need to be stable until submit.
func (r *Ring) HasSubmitStable() bool {
	return r.features&sys.IORING_FEAT_SUBMIT_STABLE != 0
}

// HasRWCurPos returns true if read/write can use current file position.
func (r *Ring) HasRWCurPos() bool {
	return r.features&sys.IORING_FEAT_RW_CUR_POS != 0
}

// HasCurPersonality returns true if current personality is used for ops.
func (r *Ring) HasCurPersonality() bool {
	return r.features&sys.IORING_FEAT_CUR_PERSONALITY != 0
}

// HasFastPoll returns true if fast poll is supported.
func (r *Ring) HasFastPoll() bool {
	return r.features&sys.IORING_FEAT_FAST_POLL != 0
}

// HasPoll32Bits returns true if poll uses 32-bit masks.
func (r *Ring) HasPoll32Bits() bool {
	return r.features&sys.IORING_FEAT_POLL_32BITS != 0
}

// HasSQPollNonFixed returns true if SQPOLL works with non-fixed files.
func (r *Ring) HasSQPollNonFixed() bool {
	return r.features&sys.IORING_FEAT_SQPOLL_NONFIXED != 0
}

// HasExtArg returns true if extended enter arguments are supported.
func (r *Ring) HasExtArg() bool {
	return r.features&sys.IORING_FEAT_EXT_ARG != 0
}

// HasNativeWorkers returns true if native workers are used.
func (r *Ring) HasNativeWorkers() bool {
	return r.features&sys.IORING_FEAT_NATIVE_WORKERS != 0
}

// HasRsrcTags returns true if resource tags are supported.
func (r *Ring) HasRsrcTags() bool {
	return r.features&sys.IORING_FEAT_RSRC_TAGS != 0
}

// HasCQESkip returns true if CQE skip is supported.
func (r *Ring) HasCQESkip() bool {
	return r.features&sys.IORING_FEAT_CQE_SKIP != 0
}

// HasLinkedFile returns true if linked operations can use direct file.
func (r *Ring) HasLinkedFile() bool {
	return r.features&sys.IORING_FEAT_LINKED_FILE != 0
}

// HasRegRegRing returns true if registered ring fds are supported.
func (r *Ring) HasRegRegRing() bool {
	return r.features&sys.IORING_FEAT_REG_REG_RING != 0
}
