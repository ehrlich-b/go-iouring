# Go io_uring Test Plan

## Overview

This document outlines a comprehensive testing strategy covering unit tests, integration tests, benchmarks, and CI/CD configuration for multi-kernel testing.

## Test Categories

### 1. Unit Tests

Unit tests verify individual components in isolation.

#### Ring Setup Tests

```go
// ring_test.go

func TestNewRing(t *testing.T) {
    tests := []struct {
        name    string
        entries uint32
        opts    []Option
        wantErr bool
    }{
        {"default", 64, nil, false},
        {"power_of_two", 128, nil, false},
        {"non_power_of_two", 100, nil, false}, // Should clamp
        {"zero_entries", 0, nil, true},
        {"too_large", 1 << 20, nil, true},
        {"sqpoll", 64, []Option{WithSQPoll()}, false}, // May fail without CAP_SYS_NICE
        {"cqsize", 64, []Option{WithCQSize(256)}, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ring, err := New(tt.entries, tt.opts...)
            if (err != nil) != tt.wantErr {
                t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if ring != nil {
                ring.Close()
            }
        })
    }
}

func TestRingClose(t *testing.T) {
    ring, err := New(64)
    require.NoError(t, err)
    
    err = ring.Close()
    assert.NoError(t, err)
    
    // Double close should not panic
    err = ring.Close()
    assert.NoError(t, err)
}

func TestRingFeatures(t *testing.T) {
    ring, err := New(64)
    require.NoError(t, err)
    defer ring.Close()
    
    features := ring.Features()
    
    // Check expected features for modern kernels
    if !ring.HasFeature(IORING_FEAT_SINGLE_MMAP) {
        t.Log("SINGLE_MMAP not supported")
    }
    if !ring.HasFeature(IORING_FEAT_NODROP) {
        t.Log("NODROP not supported")
    }
    
    t.Logf("Ring features: 0x%x", features)
}
```

#### SQE Tests

```go
// sqe_test.go

func TestSQEPool(t *testing.T) {
    ring, err := New(4) // Small ring
    require.NoError(t, err)
    defer ring.Close()
    
    // Should be able to get 4 SQEs
    sqes := make([]*SQE, 4)
    for i := 0; i < 4; i++ {
        sqe := ring.getSQE()
        require.NotNil(t, sqe, "getSQE %d should succeed", i)
        sqes[i] = sqe
    }
    
    // 5th should fail (queue full)
    sqe := ring.getSQE()
    assert.Nil(t, sqe, "getSQE should return nil when full")
    
    // Submit to free up space
    n, err := ring.Submit()
    require.NoError(t, err)
    assert.Equal(t, 4, n)
}

func TestSQEFields(t *testing.T) {
    sqe := &SQE{}
    
    // Verify field offsets match kernel struct
    assert.Equal(t, 0, int(unsafe.Offsetof(sqe.Opcode)))
    assert.Equal(t, 1, int(unsafe.Offsetof(sqe.Flags)))
    assert.Equal(t, 4, int(unsafe.Offsetof(sqe.Fd)))
    assert.Equal(t, 8, int(unsafe.Offsetof(sqe.Off)))
    assert.Equal(t, 16, int(unsafe.Offsetof(sqe.Addr)))
    assert.Equal(t, 24, int(unsafe.Offsetof(sqe.Len)))
    assert.Equal(t, 32, int(unsafe.Offsetof(sqe.UserData)))
}
```

#### CQE Tests

```go
// cqe_test.go

func TestCQEHandling(t *testing.T) {
    ring, err := New(64)
    require.NoError(t, err)
    defer ring.Close()
    
    // Submit NOPs
    for i := 0; i < 10; i++ {
        ring.PrepNop().WithUserData(uint64(i)).Submit()
    }
    
    // Wait and verify all complete
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    
    seen := make(map[uint64]bool)
    for i := 0; i < 10; i++ {
        cqe, err := ring.WaitCQE(ctx)
        require.NoError(t, err)
        assert.Equal(t, int32(0), cqe.Res)
        seen[cqe.UserData] = true
        ring.SeenCQE(cqe)
    }
    
    assert.Len(t, seen, 10)
}

func TestCQEOverflow(t *testing.T) {
    // Small CQ to test overflow
    ring, err := New(4, WithCQSize(4))
    require.NoError(t, err)
    defer ring.Close()
    
    if !ring.HasFeature(IORING_FEAT_NODROP) {
        t.Skip("NODROP not supported, overflow behavior undefined")
    }
    
    // Submit more than CQ can hold
    for i := 0; i < 8; i++ {
        ring.PrepNop().Submit()
    }
    
    // Should still be able to retrieve all
    count := 0
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    
    for count < 8 {
        cqe, err := ring.WaitCQE(ctx)
        if err != nil {
            break
        }
        ring.SeenCQE(cqe)
        count++
    }
    
    assert.Equal(t, 8, count)
}
```

### 2. Operation Tests

Test each operation individually.

```go
// ops_test.go

func TestOpRead(t *testing.T) {
    ring, err := New(64)
    require.NoError(t, err)
    defer ring.Close()
    
    // Create test file
    content := []byte("hello, io_uring!")
    f, err := os.CreateTemp("", "iouring_test")
    require.NoError(t, err)
    defer os.Remove(f.Name())
    
    _, err = f.Write(content)
    require.NoError(t, err)
    
    buf := make([]byte, len(content))
    n, err := ring.PrepRead(int(f.Fd()), buf, 0).SubmitAndWait(context.Background())
    
    require.NoError(t, err)
    assert.Equal(t, int32(len(content)), n)
    assert.Equal(t, content, buf)
}

func TestOpWrite(t *testing.T) {
    ring, err := New(64)
    require.NoError(t, err)
    defer ring.Close()
    
    f, err := os.CreateTemp("", "iouring_test")
    require.NoError(t, err)
    defer os.Remove(f.Name())
    
    content := []byte("hello from io_uring!")
    n, err := ring.PrepWrite(int(f.Fd()), content, 0).SubmitAndWait(context.Background())
    
    require.NoError(t, err)
    assert.Equal(t, int32(len(content)), n)
    
    // Verify
    readBack, _ := os.ReadFile(f.Name())
    assert.Equal(t, content, readBack)
}

func TestOpAccept(t *testing.T) {
    ring, err := New(64)
    require.NoError(t, err)
    defer ring.Close()
    
    // Create listener
    ln, err := net.Listen("tcp", "127.0.0.1:0")
    require.NoError(t, err)
    defer ln.Close()
    
    tcpLn := ln.(*net.TCPListener)
    rawConn, _ := tcpLn.SyscallConn()
    var listenFd int
    rawConn.Control(func(fd uintptr) { listenFd = int(fd) })
    
    // Start accept
    req, err := ring.PrepAccept(listenFd, 0).Submit()
    require.NoError(t, err)
    
    // Connect from another goroutine
    go func() {
        time.Sleep(10 * time.Millisecond)
        conn, _ := net.Dial("tcp", ln.Addr().String())
        if conn != nil {
            conn.Close()
        }
    }()
    
    // Wait for accept
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    
    cqe, err := ring.WaitCQE(ctx)
    require.NoError(t, err)
    
    assert.Greater(t, cqe.Res, int32(0), "should return valid fd")
    syscall.Close(int(cqe.Res))
}

func TestOpTimeout(t *testing.T) {
    ring, err := New(64)
    require.NoError(t, err)
    defer ring.Close()
    
    start := time.Now()
    timeout := 100 * time.Millisecond
    
    _, err = ring.PrepTimeout(timeout, 0).Submit()
    require.NoError(t, err)
    
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    
    cqe, err := ring.WaitCQE(ctx)
    require.NoError(t, err)
    
    elapsed := time.Since(start)
    
    // Should complete with -ETIME
    assert.Equal(t, int32(-int(syscall.ETIME)), cqe.Res)
    assert.GreaterOrEqual(t, elapsed, timeout)
}

func TestOpLinked(t *testing.T) {
    ring, err := New(64)
    require.NoError(t, err)
    defer ring.Close()
    
    // Create test file
    f, err := os.CreateTemp("", "iouring_test")
    require.NoError(t, err)
    defer os.Remove(f.Name())
    
    content := []byte("linked operations test")
    _, _ = f.Write(content)
    
    buf := make([]byte, len(content))
    
    // Link read -> write (copy to stdout)
    ring.PrepRead(int(f.Fd()), buf, 0).WithLink()
    ring.PrepWrite(int(os.Stdout.Fd()), buf, 0)
    
    _, err = ring.Submit()
    require.NoError(t, err)
    
    // Should get 2 completions
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    
    for i := 0; i < 2; i++ {
        cqe, err := ring.WaitCQE(ctx)
        require.NoError(t, err)
        assert.GreaterOrEqual(t, cqe.Res, int32(0))
        ring.SeenCQE(cqe)
    }
}
```

### 3. Integration Tests

Test complete workflows and interaction with system resources.

```go
// integration_test.go

func TestEchoServer(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    ring, err := New(256)
    require.NoError(t, err)
    defer ring.Close()
    
    // Setup server
    ln, err := net.Listen("tcp", "127.0.0.1:0")
    require.NoError(t, err)
    defer ln.Close()
    
    serverDone := make(chan struct{})
    go func() {
        defer close(serverDone)
        runEchoServer(t, ring, ln)
    }()
    
    // Run client test
    conn, err := net.Dial("tcp", ln.Addr().String())
    require.NoError(t, err)
    defer conn.Close()
    
    message := []byte("hello io_uring echo!")
    _, err = conn.Write(message)
    require.NoError(t, err)
    
    buf := make([]byte, len(message))
    n, err := conn.Read(buf)
    require.NoError(t, err)
    
    assert.Equal(t, message, buf[:n])
}

func TestHighConcurrency(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    ring, err := New(4096, WithSQPoll())
    if err != nil {
        ring, err = New(4096) // Fallback without SQPOLL
    }
    require.NoError(t, err)
    defer ring.Close()
    
    f, _ := os.CreateTemp("", "iouring_concurrent")
    defer os.Remove(f.Name())
    f.WriteString(strings.Repeat("x", 4096))
    
    const numOps = 10000
    var wg sync.WaitGroup
    errors := make(chan error, numOps)
    
    for i := 0; i < numOps; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            buf := make([]byte, 64)
            offset := uint64((id * 64) % 4096)
            
            _, err := ring.PrepRead(int(f.Fd()), buf, offset).
                SubmitAndWait(context.Background())
            if err != nil {
                errors <- err
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    for err := range errors {
        t.Errorf("concurrent operation failed: %v", err)
    }
}

func TestMemoryStability(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Run many iterations to check for memory leaks
    for iteration := 0; iteration < 100; iteration++ {
        ring, err := New(64)
        require.NoError(t, err)
        
        for i := 0; i < 1000; i++ {
            ring.PrepNop().Submit()
        }
        
        ring.Close()
    }
    
    // Force GC and check memory
    runtime.GC()
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    t.Logf("Alloc: %d MB", m.Alloc/1024/1024)
}
```

### 4. Benchmarks

```go
// benchmark_test.go

func BenchmarkNopSubmit(b *testing.B) {
    ring, _ := New(1024)
    defer ring.Close()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ring.PrepNop().Submit()
        cqe, _ := ring.WaitCQE(context.Background())
        ring.SeenCQE(cqe)
    }
}

func BenchmarkNopBatch(b *testing.B) {
    ring, _ := New(1024)
    defer ring.Close()
    
    batchSize := 32
    
    b.ResetTimer()
    for i := 0; i < b.N; i += batchSize {
        // Submit batch
        for j := 0; j < batchSize; j++ {
            ring.PrepNop()
        }
        ring.Submit()
        
        // Collect completions
        for j := 0; j < batchSize; j++ {
            cqe, _ := ring.WaitCQE(context.Background())
            ring.SeenCQE(cqe)
        }
    }
    
    b.SetBytes(int64(batchSize))
}

func BenchmarkFileRead(b *testing.B) {
    f, _ := os.CreateTemp("", "bench")
    defer os.Remove(f.Name())
    
    data := make([]byte, 4096)
    f.Write(data)
    
    ring, _ := New(256)
    defer ring.Close()
    
    buf := make([]byte, 4096)
    
    b.ResetTimer()
    b.SetBytes(4096)
    
    for i := 0; i < b.N; i++ {
        ring.PrepRead(int(f.Fd()), buf, 0).Submit()
        cqe, _ := ring.WaitCQE(context.Background())
        ring.SeenCQE(cqe)
    }
}

func BenchmarkFileReadSyscall(b *testing.B) {
    f, _ := os.CreateTemp("", "bench")
    defer os.Remove(f.Name())
    
    data := make([]byte, 4096)
    f.Write(data)
    
    buf := make([]byte, 4096)
    
    b.ResetTimer()
    b.SetBytes(4096)
    
    for i := 0; i < b.N; i++ {
        syscall.Pread(int(f.Fd()), buf, 0)
    }
}

func BenchmarkTCPEcho(b *testing.B) {
    // Setup server
    ring, _ := New(1024, WithSQPoll())
    defer ring.Close()
    
    // ... setup echo server ...
    
    b.ResetTimer()
    
    // Run echo benchmark
}
```

### 5. Kernel Version Matrix Tests

```go
// kernel_test.go

func TestKernelFeatures(t *testing.T) {
    ring, err := New(64)
    require.NoError(t, err)
    defer ring.Close()
    
    probe, err := ring.Probe()
    require.NoError(t, err)
    
    // Log supported operations
    t.Logf("Kernel supports %d operations", probe.LastOp)
    
    // Test version-specific features
    tests := []struct {
        op      Op
        minKern string
    }{
        {OpRead, "5.6"},
        {OpSend, "5.6"},
        {OpAccept, "5.5"},
        {OpSocket, "5.19"},
        {OpSendZC, "6.0"},
        {OpBind, "6.11"},
    }
    
    for _, tt := range tests {
        supported := ring.SupportsOp(tt.op)
        t.Logf("Op %v (min %s): %v", tt.op, tt.minKern, supported)
    }
}

func TestOperationFallback(t *testing.T) {
    ring, err := New(64)
    require.NoError(t, err)
    defer ring.Close()
    
    // Test that unsupported ops return proper error
    if !ring.SupportsOp(OpBind) {
        _, err := ring.PrepBind(0, nil).SubmitAndWait(context.Background())
        assert.ErrorIs(t, err, ErrNotSupported)
    }
}
```

## CI/CD Configuration

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        kernel: ['5.15', '6.1', '6.6', '6.11']
        go: ['1.21', '1.22']
    
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}
      
      - name: Install kernel ${{ matrix.kernel }}
        run: |
          # Use specific kernel for testing
          # This requires custom runner or VM
          echo "Testing on kernel ${{ matrix.kernel }}"
      
      - name: Run tests
        run: |
          go test -v -race ./...
      
      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem ./...

  integration:
    runs-on: ubuntu-latest
    needs: test
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Integration tests
        run: |
          go test -v -tags=integration ./...
          
  coverage:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Run coverage
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
```

### Docker Test Matrix

```dockerfile
# Dockerfile.test
ARG KERNEL_VERSION=6.1

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    golang \
    linux-tools-generic

# Install specific kernel headers if needed
RUN apt-get install -y linux-headers-${KERNEL_VERSION}

WORKDIR /app
COPY . .

CMD ["go", "test", "-v", "./..."]
```

```yaml
# docker-compose.test.yml
version: '3.8'

services:
  test-5.15:
    build:
      context: .
      dockerfile: Dockerfile.test
      args:
        KERNEL_VERSION: "5.15"
    privileged: true

  test-6.1:
    build:
      context: .
      dockerfile: Dockerfile.test
      args:
        KERNEL_VERSION: "6.1"
    privileged: true
    
  test-6.6:
    build:
      context: .
      dockerfile: Dockerfile.test
      args:
        KERNEL_VERSION: "6.6"
    privileged: true
```

## Test Utilities

```go
// testutil/testutil.go

package testutil

import (
    "os"
    "testing"
)

// SkipIfNoIOUring skips test if io_uring is unavailable
func SkipIfNoIOUring(t *testing.T) {
    t.Helper()
    
    // Try to create a ring
    ring, err := iouring.New(4)
    if err != nil {
        t.Skipf("io_uring unavailable: %v", err)
    }
    ring.Close()
}

// SkipIfSeccomp skips test if io_uring is blocked by seccomp
func SkipIfSeccomp(t *testing.T) {
    t.Helper()
    
    _, err := iouring.New(4)
    if err != nil && strings.Contains(err.Error(), "EPERM") {
        t.Skip("io_uring blocked by seccomp")
    }
}

// RequireKernel skips test if kernel is too old
func RequireKernel(t *testing.T, major, minor int) {
    t.Helper()
    
    var uname unix.Utsname
    unix.Uname(&uname)
    
    release := string(uname.Release[:])
    // Parse and compare version...
}

// TempFile creates a temporary file with content
func TempFile(t *testing.T, content []byte) *os.File {
    t.Helper()
    
    f, err := os.CreateTemp("", "iouring_test")
    if err != nil {
        t.Fatal(err)
    }
    
    t.Cleanup(func() {
        f.Close()
        os.Remove(f.Name())
    })
    
    if content != nil {
        f.Write(content)
        f.Seek(0, 0)
    }
    
    return f
}
```

## Test Coverage Goals

| Package | Target Coverage |
|---------|----------------|
| ring.go | 90% |
| sqe.go | 85% |
| cqe.go | 85% |
| ops.go | 80% |
| register.go | 80% |
| net.go | 85% |
| internal/sys | 70% |

## Fuzzing

```go
// fuzz_test.go

func FuzzSQEPrepare(f *testing.F) {
    f.Add(uint8(0), int32(0), uint64(0), uint32(0))
    f.Add(uint8(22), int32(3), uint64(0), uint32(4096)) // Read
    
    f.Fuzz(func(t *testing.T, opcode uint8, fd int32, offset uint64, len uint32) {
        ring, err := New(4)
        if err != nil {
            t.Skip()
        }
        defer ring.Close()
        
        sqe := ring.getSQE()
        if sqe == nil {
            return
        }
        
        sqe.Opcode = opcode
        sqe.Fd = fd
        sqe.Off = offset
        sqe.Len = len
        
        // Should not panic
        ring.Submit()
    })
}
```
