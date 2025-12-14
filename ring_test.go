//go:build linux

package iouring

import (
	"net"
	"os"
	"syscall"
	"testing"
	"unsafe"
)

func skipIfNoIOURing(t *testing.T) {
	t.Helper()
	ring, err := New(4)
	if err != nil {
		if err == syscall.ENOSYS {
			t.Skip("io_uring not supported on this kernel")
		}
		if err == syscall.EPERM {
			t.Skip("io_uring blocked by seccomp or permissions")
		}
		t.Skipf("io_uring unavailable: %v", err)
	}
	ring.Close()
}

func TestNewRing(t *testing.T) {
	skipIfNoIOURing(t)

	tests := []struct {
		name    string
		entries uint32
		opts    []Option
		wantErr bool
	}{
		{"default_64", 64, nil, false},
		{"default_128", 128, nil, false},
		{"default_256", 256, nil, false},
		{"non_power_of_two", 100, nil, false}, // Kernel rounds up
		{"zero_entries", 0, nil, true},
		{"with_cqsize", 64, []Option{WithCQSize(256)}, false},
		{"with_single_issuer", 64, []Option{WithSingleIssuer()}, false},
		{"with_coop_taskrun", 64, []Option{WithCoopTaskrun()}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ring, err := New(tt.entries, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if ring != nil {
				// Verify ring was set up correctly
				if ring.Fd() < 0 {
					t.Error("ring fd should be valid")
				}
				if ring.SQEntries() == 0 {
					t.Error("SQ entries should be non-zero")
				}
				if ring.CQEntries() == 0 {
					t.Error("CQ entries should be non-zero")
				}
				ring.Close()
			}
		})
	}
}

func TestRingClose(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// First close should succeed
	err = ring.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Second close should be idempotent (not panic or error)
	err = ring.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestRingFeatures(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	features := ring.Features()
	t.Logf("Ring features: 0x%x", features)

	// Log which features are supported
	featureNames := map[uint32]string{
		0x1:    "SINGLE_MMAP",
		0x2:    "NODROP",
		0x4:    "SUBMIT_STABLE",
		0x8:    "RW_CUR_POS",
		0x10:   "CUR_PERSONALITY",
		0x20:   "FAST_POLL",
		0x40:   "POLL_32BITS",
		0x80:   "SQPOLL_NONFIXED",
		0x100:  "EXT_ARG",
		0x200:  "NATIVE_WORKERS",
		0x400:  "RSRC_TAGS",
		0x800:  "CQE_SKIP",
		0x1000: "LINKED_FILE",
		0x2000: "REG_REG_RING",
	}

	for flag, name := range featureNames {
		if ring.HasFeature(flag) {
			t.Logf("  %s: supported", name)
		}
	}
}

func TestNopOperation(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Submit multiple NOPs
	const numNops = 10
	for i := 0; i < numNops; i++ {
		err := ring.PrepNop(uint64(i + 1))
		if err != nil {
			t.Fatalf("PrepNop(%d) error = %v", i, err)
		}
	}

	if ring.SQReady() != numNops {
		t.Errorf("SQReady() = %d, want %d", ring.SQReady(), numNops)
	}

	// Submit
	n, err := ring.Submit()
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if n != numNops {
		t.Errorf("Submit() = %d, want %d", n, numNops)
	}

	// Wait for completions
	seen := make(map[uint64]bool)
	for i := 0; i < numNops; i++ {
		userData, res, _, err := ring.WaitCQE()
		if err != nil {
			t.Fatalf("WaitCQE() error = %v", err)
		}
		if res != 0 {
			t.Errorf("CQE res = %d, want 0", res)
		}
		seen[userData] = true
		ring.SeenCQE()
	}

	// Verify all NOPs completed
	for i := 1; i <= numNops; i++ {
		if !seen[uint64(i)] {
			t.Errorf("Missing completion for userData %d", i)
		}
	}
}

func TestReadWrite(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Create a temp file
	f, err := os.CreateTemp("", "iouring_test")
	if err != nil {
		t.Fatalf("CreateTemp error = %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// Write some data using io_uring
	writeData := []byte("Hello, io_uring!")
	err = ring.PrepWrite(int(f.Fd()), writeData, 0, 1)
	if err != nil {
		t.Fatalf("PrepWrite error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	userData, res, _, err := ring.WaitCQE()
	if err != nil {
		t.Fatalf("WaitCQE error = %v", err)
	}
	if userData != 1 {
		t.Errorf("userData = %d, want 1", userData)
	}
	if res != int32(len(writeData)) {
		t.Errorf("write res = %d, want %d", res, len(writeData))
	}
	ring.SeenCQE()

	// Read it back
	readBuf := make([]byte, len(writeData))
	err = ring.PrepRead(int(f.Fd()), readBuf, 0, 2)
	if err != nil {
		t.Fatalf("PrepRead error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	userData, res, _, err = ring.WaitCQE()
	if err != nil {
		t.Fatalf("WaitCQE error = %v", err)
	}
	if userData != 2 {
		t.Errorf("userData = %d, want 2", userData)
	}
	if res != int32(len(writeData)) {
		t.Errorf("read res = %d, want %d", res, len(writeData))
	}
	ring.SeenCQE()

	// Verify data
	if string(readBuf) != string(writeData) {
		t.Errorf("read data = %q, want %q", string(readBuf), string(writeData))
	}
}

func TestSQFull(t *testing.T) {
	skipIfNoIOURing(t)

	// Create a small ring
	ring, err := New(4) // Small ring
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Fill the queue
	sqEntries := ring.SQEntries()
	for i := uint32(0); i < sqEntries; i++ {
		err := ring.PrepNop(uint64(i))
		if err != nil {
			t.Fatalf("PrepNop(%d) unexpected error = %v", i, err)
		}
	}

	// Next one should fail
	err = ring.PrepNop(999)
	if err != ErrSQFull {
		t.Errorf("PrepNop on full queue error = %v, want ErrSQFull", err)
	}

	// Submit to clear queue
	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	// Drain completions
	for i := uint32(0); i < sqEntries; i++ {
		_, _, _, err := ring.WaitCQE()
		if err != nil {
			t.Fatalf("WaitCQE error = %v", err)
		}
		ring.SeenCQE()
	}

	// Now should be able to submit again
	err = ring.PrepNop(1000)
	if err != nil {
		t.Errorf("PrepNop after drain error = %v", err)
	}
}

func TestForEachCQE(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Submit multiple NOPs
	const numNops = 5
	for i := 0; i < numNops; i++ {
		ring.PrepNop(uint64(i + 1))
	}
	ring.Submit()

	// Wait a bit for completions
	ring.SubmitAndWait(uint32(numNops))

	// Process all at once
	count := ring.ForEachCQE(func(userData uint64, res int32, flags uint32) bool {
		if res != 0 {
			t.Errorf("CQE res = %d, want 0", res)
		}
		return true // continue iteration
	})

	if count != numNops {
		t.Errorf("ForEachCQE processed %d, want %d", count, numNops)
	}

	// CQEs should be consumed
	if ring.CQReady() != 0 {
		t.Errorf("CQReady() = %d after ForEachCQE, want 0", ring.CQReady())
	}
}

func BenchmarkNopSubmit(b *testing.B) {
	ring, err := New(1024)
	if err != nil {
		b.Skipf("io_uring unavailable: %v", err)
	}
	defer ring.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ring.PrepNop(uint64(i))
		ring.Submit()
		ring.WaitCQE()
		ring.SeenCQE()
	}
}

func BenchmarkNopBatch(b *testing.B) {
	ring, err := New(1024)
	if err != nil {
		b.Skipf("io_uring unavailable: %v", err)
	}
	defer ring.Close()

	const batchSize = 32

	b.ResetTimer()
	for i := 0; i < b.N; i += batchSize {
		// Submit batch
		for j := 0; j < batchSize && i+j < b.N; j++ {
			ring.PrepNop(uint64(i + j))
		}
		ring.Submit()

		// Collect completions
		for j := 0; j < batchSize && i+j < b.N; j++ {
			ring.WaitCQE()
			ring.SeenCQE()
		}
	}
}

func TestProbe(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	probe, err := ring.Probe()
	if err != nil {
		t.Fatalf("Probe() error = %v", err)
	}

	t.Logf("Last operation supported: %d", probe.LastOp())
	t.Logf("Features: 0x%x", probe.Features())

	// NOP should always be supported
	if !probe.SupportsOp(0) { // IORING_OP_NOP
		t.Error("NOP should be supported")
	}

	// READV should be supported (available since 5.1)
	if !probe.SupportsOp(1) { // IORING_OP_READV
		t.Log("READV not supported (unusual)")
	}

	// Check a definitely unsupported op (high number)
	if probe.SupportsOp(255) {
		t.Error("Op 255 should not be supported")
	}
}

func TestTimeout(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Create a short timeout (100ms)
	ts := &Timespec{Sec: 0, Nsec: 100_000_000}
	err = ring.PrepTimeout(ts, 0, 0, 1)
	if err != nil {
		t.Fatalf("PrepTimeout error = %v", err)
	}

	start := nanotime()
	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	_, res, _, err := ring.WaitCQE()
	elapsed := nanotime() - start
	if err != nil {
		t.Fatalf("WaitCQE error = %v", err)
	}
	ring.SeenCQE()

	// Result should be -ETIME (62)
	if res != -62 {
		t.Errorf("timeout res = %d, want -62 (ETIME)", res)
	}

	// Should have taken roughly 100ms
	if elapsed < 50_000_000 {
		t.Errorf("timeout elapsed = %dns, expected >= 50ms", elapsed)
	}
	t.Logf("Timeout elapsed: %dms", elapsed/1_000_000)
}

func TestCancel(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Submit a long timeout
	ts := &Timespec{Sec: 10, Nsec: 0}
	err = ring.PrepTimeout(ts, 0, 0, 100)
	if err != nil {
		t.Fatalf("PrepTimeout error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	// Now cancel it
	err = ring.PrepCancel(100, 0, 200)
	if err != nil {
		t.Fatalf("PrepCancel error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit cancel error = %v", err)
	}

	// Should get two CQEs: the cancelled timeout and the cancel itself
	seenCancel := false
	seenTimeout := false

	for i := 0; i < 2; i++ {
		userData, res, _, err := ring.WaitCQE()
		if err != nil {
			t.Fatalf("WaitCQE error = %v", err)
		}
		ring.SeenCQE()

		switch userData {
		case 100:
			// Cancelled timeout should return -ECANCELED (125)
			if res != -125 {
				t.Errorf("cancelled timeout res = %d, want -125 (ECANCELED)", res)
			}
			seenTimeout = true
		case 200:
			// Cancel op should return 0 on success
			if res != 0 {
				t.Errorf("cancel res = %d, want 0", res)
			}
			seenCancel = true
		default:
			t.Errorf("unexpected userData %d", userData)
		}
	}

	if !seenCancel {
		t.Error("did not see cancel completion")
	}
	if !seenTimeout {
		t.Error("did not see timeout completion")
	}
}

func TestReadvWritev(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Create a temp file
	f, err := os.CreateTemp("", "iouring_test_v")
	if err != nil {
		t.Fatalf("CreateTemp error = %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// Create multiple buffers for vectored I/O
	buf1 := []byte("Hello, ")
	buf2 := []byte("vectored ")
	buf3 := []byte("io_uring!")

	// Build iovec slice
	iovecs := []syscall.Iovec{
		{Base: &buf1[0], Len: uint64(len(buf1))},
		{Base: &buf2[0], Len: uint64(len(buf2))},
		{Base: &buf3[0], Len: uint64(len(buf3))},
	}

	err = ring.PrepWritev(int(f.Fd()), iovecs, 0, 1)
	if err != nil {
		t.Fatalf("PrepWritev error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	userData, res, _, err := ring.WaitCQE()
	if err != nil {
		t.Fatalf("WaitCQE error = %v", err)
	}
	ring.SeenCQE()

	expectedLen := int32(len(buf1) + len(buf2) + len(buf3))
	if userData != 1 || res != expectedLen {
		t.Errorf("writev: userData=%d res=%d, want userData=1 res=%d", userData, res, expectedLen)
	}

	// Read it back with readv
	readBuf := make([]byte, expectedLen)
	readIovecs := []syscall.Iovec{
		{Base: &readBuf[0], Len: uint64(len(readBuf))},
	}

	err = ring.PrepReadv(int(f.Fd()), readIovecs, 0, 2)
	if err != nil {
		t.Fatalf("PrepReadv error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	userData, res, _, err = ring.WaitCQE()
	if err != nil {
		t.Fatalf("WaitCQE error = %v", err)
	}
	ring.SeenCQE()

	if userData != 2 || res != expectedLen {
		t.Errorf("readv: userData=%d res=%d, want userData=2 res=%d", userData, res, expectedLen)
	}

	expected := "Hello, vectored io_uring!"
	if string(readBuf) != expected {
		t.Errorf("readv data = %q, want %q", string(readBuf), expected)
	}
}

func TestRegisterBuffers(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Create temp file
	f, err := os.CreateTemp("", "iouring_test_buf")
	if err != nil {
		t.Fatalf("CreateTemp error = %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// Register buffers
	bufs := [][]byte{
		make([]byte, 4096),
		make([]byte, 4096),
	}
	copy(bufs[0], "Hello from registered buffer!")

	err = ring.RegisterBuffers(bufs)
	if err != nil {
		t.Fatalf("RegisterBuffers error = %v", err)
	}

	// Write using fixed buffer (index 0)
	dataLen := len("Hello from registered buffer!")
	err = ring.PrepWriteFixed(int(f.Fd()), bufs[0][:dataLen], 0, 0, 1)
	if err != nil {
		t.Fatalf("PrepWriteFixed error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	_, res, _, err := ring.WaitCQE()
	if err != nil {
		t.Fatalf("WaitCQE error = %v", err)
	}
	ring.SeenCQE()

	if res != int32(dataLen) {
		t.Errorf("write_fixed res = %d, want %d", res, dataLen)
	}

	// Read back using fixed buffer (index 1)
	err = ring.PrepReadFixed(int(f.Fd()), bufs[1][:dataLen], 0, 1, 2)
	if err != nil {
		t.Fatalf("PrepReadFixed error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	_, res, _, err = ring.WaitCQE()
	if err != nil {
		t.Fatalf("WaitCQE error = %v", err)
	}
	ring.SeenCQE()

	if res != int32(dataLen) {
		t.Errorf("read_fixed res = %d, want %d", res, dataLen)
	}

	if string(bufs[1][:dataLen]) != "Hello from registered buffer!" {
		t.Errorf("read_fixed data = %q, want %q", string(bufs[1][:dataLen]), "Hello from registered buffer!")
	}

	// Unregister buffers
	err = ring.UnregisterBuffers()
	if err != nil {
		t.Errorf("UnregisterBuffers error = %v", err)
	}
}

func TestRegisterFiles(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Create temp files
	f1, err := os.CreateTemp("", "iouring_test_f1")
	if err != nil {
		t.Fatalf("CreateTemp error = %v", err)
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := os.CreateTemp("", "iouring_test_f2")
	if err != nil {
		t.Fatalf("CreateTemp error = %v", err)
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	// Register files
	err = ring.RegisterFiles([]int{int(f1.Fd()), int(f2.Fd())})
	if err != nil {
		t.Fatalf("RegisterFiles error = %v", err)
	}

	// Unregister files
	err = ring.UnregisterFiles()
	if err != nil {
		t.Errorf("UnregisterFiles error = %v", err)
	}
}

func TestLinkTimeout(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Create a temp file
	f, err := os.CreateTemp("", "iouring_test_link")
	if err != nil {
		t.Fatalf("CreateTemp error = %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// Prep a read that will wait (file is empty)
	buf := make([]byte, 100)
	err = ring.PrepRead(int(f.Fd()), buf, 0, 1)
	if err != nil {
		t.Fatalf("PrepRead error = %v", err)
	}
	ring.SetSQELink()

	// Link a timeout to it
	ts := &Timespec{Sec: 0, Nsec: 50_000_000} // 50ms
	err = ring.PrepLinkTimeout(ts, 0, 2)
	if err != nil {
		t.Fatalf("PrepLinkTimeout error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	// Read completes immediately with 0 bytes (empty file)
	// Link timeout may or may not fire depending on timing
	cqeCount := 0
	for cqeCount < 2 {
		userData, res, _, err := ring.WaitCQE()
		if err != nil {
			break
		}
		ring.SeenCQE()
		cqeCount++
		t.Logf("CQE: userData=%d res=%d", userData, res)
	}

	if cqeCount < 1 {
		t.Error("expected at least 1 CQE")
	}
}

func TestFsync(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Create temp file
	f, err := os.CreateTemp("", "iouring_test_fsync")
	if err != nil {
		t.Fatalf("CreateTemp error = %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// Write data
	data := []byte("test data for fsync")
	err = ring.PrepWrite(int(f.Fd()), data, 0, 1)
	if err != nil {
		t.Fatalf("PrepWrite error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	_, _, _, err = ring.WaitCQE()
	if err != nil {
		t.Fatalf("WaitCQE error = %v", err)
	}
	ring.SeenCQE()

	// Fsync the file
	err = ring.PrepFsync(int(f.Fd()), 0, 2)
	if err != nil {
		t.Fatalf("PrepFsync error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	_, res, _, err := ring.WaitCQE()
	if err != nil {
		t.Fatalf("WaitCQE error = %v", err)
	}
	ring.SeenCQE()

	if res != 0 {
		t.Errorf("fsync res = %d, want 0", res)
	}
}

// nanotime returns current time in nanoseconds
func nanotime() int64 {
	var ts syscall.Timespec
	syscall.Syscall(syscall.SYS_CLOCK_GETTIME, 1 /* CLOCK_MONOTONIC */, uintptr(unsafe.Pointer(&ts)), 0)
	return ts.Sec*1e9 + ts.Nsec
}

// Benchmarks comparing io_uring vs syscall

func BenchmarkReadSyscall(b *testing.B) {
	f, err := os.CreateTemp("", "bench_read")
	if err != nil {
		b.Fatalf("CreateTemp error = %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// Write test data
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i)
	}
	f.Write(data)

	buf := make([]byte, 4096)
	fd := int(f.Fd())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		syscall.Pread(fd, buf, 0)
	}
}

func BenchmarkReadIOUring(b *testing.B) {
	ring, err := New(1024)
	if err != nil {
		b.Skipf("io_uring unavailable: %v", err)
	}
	defer ring.Close()

	f, err := os.CreateTemp("", "bench_read")
	if err != nil {
		b.Fatalf("CreateTemp error = %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// Write test data
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i)
	}
	f.Write(data)

	buf := make([]byte, 4096)
	fd := int(f.Fd())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ring.PrepRead(fd, buf, 0, uint64(i))
		ring.Submit()
		ring.WaitCQE()
		ring.SeenCQE()
	}
}

func BenchmarkReadIOUringBatch(b *testing.B) {
	ring, err := New(1024)
	if err != nil {
		b.Skipf("io_uring unavailable: %v", err)
	}
	defer ring.Close()

	f, err := os.CreateTemp("", "bench_read")
	if err != nil {
		b.Fatalf("CreateTemp error = %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// Write test data
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i)
	}
	f.Write(data)

	buf := make([]byte, 4096)
	fd := int(f.Fd())
	const batchSize = 32

	b.ResetTimer()
	for i := 0; i < b.N; i += batchSize {
		count := batchSize
		if i+count > b.N {
			count = b.N - i
		}

		// Submit batch
		for j := 0; j < count; j++ {
			ring.PrepRead(fd, buf, 0, uint64(i+j))
		}
		ring.Submit()

		// Collect completions
		for j := 0; j < count; j++ {
			ring.WaitCQE()
			ring.SeenCQE()
		}
	}
}

func BenchmarkWriteSyscall(b *testing.B) {
	f, err := os.CreateTemp("", "bench_write")
	if err != nil {
		b.Fatalf("CreateTemp error = %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	buf := make([]byte, 4096)
	for i := range buf {
		buf[i] = byte(i)
	}
	fd := int(f.Fd())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		syscall.Pwrite(fd, buf, 0)
	}
}

func BenchmarkWriteIOUring(b *testing.B) {
	ring, err := New(1024)
	if err != nil {
		b.Skipf("io_uring unavailable: %v", err)
	}
	defer ring.Close()

	f, err := os.CreateTemp("", "bench_write")
	if err != nil {
		b.Fatalf("CreateTemp error = %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	buf := make([]byte, 4096)
	for i := range buf {
		buf[i] = byte(i)
	}
	fd := int(f.Fd())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ring.PrepWrite(fd, buf, 0, uint64(i))
		ring.Submit()
		ring.WaitCQE()
		ring.SeenCQE()
	}
}

func BenchmarkWriteIOUringBatch(b *testing.B) {
	ring, err := New(1024)
	if err != nil {
		b.Skipf("io_uring unavailable: %v", err)
	}
	defer ring.Close()

	f, err := os.CreateTemp("", "bench_write")
	if err != nil {
		b.Fatalf("CreateTemp error = %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	buf := make([]byte, 4096)
	for i := range buf {
		buf[i] = byte(i)
	}
	fd := int(f.Fd())
	const batchSize = 32

	b.ResetTimer()
	for i := 0; i < b.N; i += batchSize {
		count := batchSize
		if i+count > b.N {
			count = b.N - i
		}

		// Submit batch
		for j := 0; j < count; j++ {
			ring.PrepWrite(fd, buf, 0, uint64(i+j))
		}
		ring.Submit()

		// Collect completions
		for j := 0; j < count; j++ {
			ring.WaitCQE()
			ring.SeenCQE()
		}
	}
}

func BenchmarkReadFixedBuffer(b *testing.B) {
	ring, err := New(1024)
	if err != nil {
		b.Skipf("io_uring unavailable: %v", err)
	}
	defer ring.Close()

	f, err := os.CreateTemp("", "bench_read_fixed")
	if err != nil {
		b.Fatalf("CreateTemp error = %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// Write test data
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i)
	}
	f.Write(data)

	// Register buffer
	buf := make([]byte, 4096)
	if err := ring.RegisterBuffers([][]byte{buf}); err != nil {
		b.Fatalf("RegisterBuffers error = %v", err)
	}
	defer ring.UnregisterBuffers()

	fd := int(f.Fd())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ring.PrepReadFixed(fd, buf, 0, 0, uint64(i))
		ring.Submit()
		ring.WaitCQE()
		ring.SeenCQE()
	}
}

// Network tests

func TestAcceptConnect(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Create a TCP listener using standard library
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen error = %v", err)
	}
	defer ln.Close()

	addr := ln.Addr().(*net.TCPAddr)

	// Get the listener's file descriptor
	tcpLn := ln.(*net.TCPListener)
	lnFile, err := tcpLn.File()
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}
	defer lnFile.Close()
	lnFd := int(lnFile.Fd())

	// Create client socket
	clientFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM|syscall.SOCK_NONBLOCK, 0)
	if err != nil {
		t.Fatalf("Socket error = %v", err)
	}
	defer syscall.Close(clientFd)

	// Prepare sockaddr
	sa := &syscall.SockaddrInet4{Port: addr.Port}
	copy(sa.Addr[:], addr.IP.To4())

	// Submit accept and connect in parallel
	err = ring.PrepAccept(lnFd, nil, nil, syscall.SOCK_NONBLOCK, 1)
	if err != nil {
		t.Fatalf("PrepAccept error = %v", err)
	}

	// For connect we need raw sockaddr
	rawSa := syscall.RawSockaddrInet4{
		Family: syscall.AF_INET,
		Port:   htons(uint16(addr.Port)),
	}
	copy(rawSa.Addr[:], addr.IP.To4())

	err = ring.PrepConnect(clientFd, unsafe.Pointer(&rawSa), uint32(unsafe.Sizeof(rawSa)), 2)
	if err != nil {
		t.Fatalf("PrepConnect error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	// Wait for both completions
	seenAccept := false
	seenConnect := false
	var acceptedFd int32

	for i := 0; i < 2; i++ {
		userData, res, _, err := ring.WaitCQE()
		if err != nil {
			t.Fatalf("WaitCQE error = %v", err)
		}
		ring.SeenCQE()

		switch userData {
		case 1: // Accept
			if res < 0 {
				t.Errorf("accept failed: %v", syscall.Errno(-res))
			} else {
				acceptedFd = res
				seenAccept = true
			}
		case 2: // Connect
			// Connect might fail with EINPROGRESS initially on non-blocking socket
			if res < 0 && res != -int32(syscall.EINPROGRESS) {
				t.Errorf("connect failed: %v", syscall.Errno(-res))
			} else {
				seenConnect = true
			}
		}
	}

	if !seenAccept {
		t.Error("did not see accept completion")
	}
	if !seenConnect {
		t.Error("did not see connect completion")
	}

	// Close accepted fd
	if acceptedFd > 0 {
		syscall.Close(int(acceptedFd))
	}
}

func TestSendRecv(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Create a socket pair for testing
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf("Socketpair error = %v", err)
	}
	defer syscall.Close(fds[0])
	defer syscall.Close(fds[1])

	// Send data through one end
	sendData := []byte("Hello from io_uring!")
	err = ring.PrepSend(fds[0], sendData, 0, 1)
	if err != nil {
		t.Fatalf("PrepSend error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	userData, res, _, err := ring.WaitCQE()
	if err != nil {
		t.Fatalf("WaitCQE error = %v", err)
	}
	ring.SeenCQE()

	if userData != 1 {
		t.Errorf("send userData = %d, want 1", userData)
	}
	if res != int32(len(sendData)) {
		t.Errorf("send res = %d, want %d", res, len(sendData))
	}

	// Receive data through the other end
	recvBuf := make([]byte, 64)
	err = ring.PrepRecv(fds[1], recvBuf, 0, 2)
	if err != nil {
		t.Fatalf("PrepRecv error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	userData, res, _, err = ring.WaitCQE()
	if err != nil {
		t.Fatalf("WaitCQE error = %v", err)
	}
	ring.SeenCQE()

	if userData != 2 {
		t.Errorf("recv userData = %d, want 2", userData)
	}
	if res != int32(len(sendData)) {
		t.Errorf("recv res = %d, want %d", res, len(sendData))
	}
	if string(recvBuf[:res]) != string(sendData) {
		t.Errorf("recv data = %q, want %q", string(recvBuf[:res]), string(sendData))
	}
}

func TestPollAdd(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Create a socket pair
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf("Socketpair error = %v", err)
	}
	defer syscall.Close(fds[0])
	defer syscall.Close(fds[1])

	// Poll for write readiness (should be immediately ready)
	const POLLOUT = 0x0004
	err = ring.PrepPollAdd(fds[0], POLLOUT, 1)
	if err != nil {
		t.Fatalf("PrepPollAdd error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	userData, res, _, err := ring.WaitCQE()
	if err != nil {
		t.Fatalf("WaitCQE error = %v", err)
	}
	ring.SeenCQE()

	if userData != 1 {
		t.Errorf("poll userData = %d, want 1", userData)
	}
	// Result contains the poll events that occurred
	if res <= 0 {
		t.Errorf("poll res = %d, expected > 0 (poll events)", res)
	}
	t.Logf("Poll events: 0x%x", res)
}

func TestCloseOperation(t *testing.T) {
	skipIfNoIOURing(t)

	ring, err := New(64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer ring.Close()

	// Create a temp file
	f, err := os.CreateTemp("", "iouring_close_test")
	if err != nil {
		t.Fatalf("CreateTemp error = %v", err)
	}
	name := f.Name()
	defer os.Remove(name)

	// Get fd and close the Go file handle without closing the underlying fd
	fd := int(f.Fd())

	// Close using io_uring
	err = ring.PrepClose(fd, 1)
	if err != nil {
		t.Fatalf("PrepClose error = %v", err)
	}

	_, err = ring.Submit()
	if err != nil {
		t.Fatalf("Submit error = %v", err)
	}

	userData, res, _, err := ring.WaitCQE()
	if err != nil {
		t.Fatalf("WaitCQE error = %v", err)
	}
	ring.SeenCQE()

	if userData != 1 {
		t.Errorf("close userData = %d, want 1", userData)
	}
	if res != 0 {
		t.Errorf("close res = %d, want 0", res)
	}
}

// htons converts a uint16 to network byte order
func htons(v uint16) uint16 {
	return (v << 8) | (v >> 8)
}
