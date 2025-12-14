# LLM Research Prompt: go-iouring

Use this document to systematically research the io_uring problem space for building a definitive Go implementation.

---

## Research Tasks

### 1. Kernel Header Analysis

**Objective:** Extract and understand the complete io_uring API surface.

**Instructions:**
1. Fetch the latest `include/uapi/linux/io_uring.h` from the Linux kernel repository
2. Extract ALL of the following:
   - `enum io_uring_op` - All operation codes with their numeric values
   - `IORING_SETUP_*` flags - Ring setup options
   - `IORING_ENTER_*` flags - Submission/wait flags
   - `IORING_REGISTER_*` operations - Registration commands
   - `IOSQE_*` flags - SQE flags
   - `IORING_CQE_F_*` flags - CQE flags
   - `struct io_uring_sqe` - Complete field layout
   - `struct io_uring_cqe` - Complete field layout
   - `struct io_uring_params` - Setup parameters
3. Note which kernel version introduced each operation (check git blame or commit history)
4. Create a mapping of operation → minimum kernel version

**Output Format:**
```
Operation: IORING_OP_READ
Value: 22
Min Kernel: 5.6
SQE Fields Used: fd, buf, len, offset
CQE Result: bytes read or negative errno
```

### 2. Kernel Version Comparison

**Objective:** Understand what changed between kernel 6.8 and 6.11.

**Instructions:**
1. Compare io_uring.h between v6.8 and v6.11 tags
2. List new operations added
3. List new flags added
4. List any structural changes
5. Note any deprecated features

**Questions to Answer:**
- What operations exist in 6.11 but not 6.8?
- Are there any breaking changes?
- What IORING_FEAT_* flags indicate version-specific features?

### 3. liburing API Analysis

**Objective:** Understand the reference C implementation's design decisions.

**Instructions:**
1. Examine liburing's source at https://github.com/axboe/liburing
2. Document the helper functions provided for each operation
3. Note any patterns or abstractions used
4. Identify which parts are convenience vs. necessity
5. Look at how they handle multi-kernel support

**Key Files to Examine:**
- `src/include/liburing.h` - Main API
- `src/include/liburing/io_uring.h` - Kernel header copy
- `src/setup.c` - Ring initialization
- `src/queue.c` - Submission and completion handling

### 4. Existing Go Library Deep Dive

**Objective:** Understand what each existing library got right and wrong.

**Libraries to Analyze:**

#### iceber/iouring-go
- Repository: https://github.com/Iceber/iouring-go
- Questions:
  - What operations are implemented?
  - What operations are missing?
  - How does their channel-based API work?
  - What's the allocation profile?
  - How do they handle errors?

#### godzie44/go-uring
- Repository: https://github.com/godzie44/go-uring
- Questions:
  - What's their three-layer architecture?
  - How does the reactor pattern work?
  - What operations are supported?
  - How do they handle SQPOLL?

#### dshulyak/uring
- Repository: https://github.com/dshulyak/uring
- Questions:
  - How did they implement without CGO?
  - Why couldn't they batch submissions?
  - Why couldn't they use IOPOLL/SQPOLL?
  - What caused the 750ns overhead?

### 5. Performance Characteristics

**Objective:** Understand io_uring performance profiles.

**Research Questions:**
1. What's the syscall overhead of io_uring_enter() vs. traditional syscalls?
2. How does SQPOLL affect CPU usage vs. latency?
3. What's the optimal SQ/CQ ring size for various workloads?
4. How do registered buffers improve performance?
5. What's the overhead of linked operations?
6. How does multishot compare to repeated submissions?

**Benchmarks to Find:**
- io_uring vs. epoll for network I/O
- io_uring vs. pread/pwrite for file I/O
- io_uring vs. aio for async file I/O
- Impact of IOPOLL on NVMe workloads

### 6. Go Runtime Interaction

**Objective:** Understand how io_uring interacts with Go's runtime.

**Research Questions:**
1. How do blocking syscalls interact with GOMAXPROCS?
2. Does io_uring_enter() with IORING_ENTER_GETEVENTS block the OS thread?
3. How should we integrate with Go's netpoller (if at all)?
4. What are the implications of mmap'd memory for GC?
5. How do other Go async I/O libraries handle goroutine parking?

**Potential Issues:**
- Goroutine scheduling during long waits
- Memory pinning for registered buffers
- Thread-local storage with SQPOLL

### 7. Use Case Analysis

**Objective:** Understand what the Go community actually needs.

**Research Questions:**
1. What are the top io_uring use cases in production?
2. Which operations are actually used vs. theoretical?
3. What pain points exist with current Go I/O?
4. What would make developers switch from epoll-based solutions?

**Domains to Research:**
- Database engines (storage I/O)
- Proxy servers (network I/O)
- Message queues (both)
- File processing utilities

### 8. Security Considerations

**Objective:** Understand io_uring security implications.

**Research Questions:**
1. What security issues has io_uring had historically?
2. What restrictions exist in container environments?
3. How do seccomp filters affect io_uring?
4. What are the implications of SQPOLL (kernel thread)?

---

## Output Deliverables

After completing this research, produce:

1. **operations.json** - Complete list of operations with metadata:
   ```json
   {
     "IORING_OP_READ": {
       "value": 22,
       "min_kernel": "5.6",
       "sqe_fields": ["fd", "buf", "len", "offset"],
       "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC"],
       "complexity": "low",
       "priority": "critical"
     }
   }
   ```

2. **kernel_matrix.md** - Feature availability by kernel version

3. **api_design.md** - Recommended Go API design with:
   - Type definitions
   - Function signatures
   - Usage examples
   - Performance considerations

4. **implementation_order.md** - Prioritized implementation plan:
   - Phase 1: Core (ring setup, basic ops)
   - Phase 2: Network (accept, connect, send, recv)
   - Phase 3: Advanced (SQPOLL, registered buffers)
   - Phase 4: Complete (all remaining operations)

5. **test_plan.md** - Testing strategy including:
   - Unit test structure
   - Integration test scenarios
   - Benchmark suite design
   - CI/CD kernel matrix

---

## Key URLs

- Linux kernel io_uring.h: https://github.com/torvalds/linux/blob/master/include/uapi/linux/io_uring.h
- liburing: https://github.com/axboe/liburing
- io_uring manpage: https://man7.org/linux/man-pages/man7/io_uring.7.html
- Lord of the io_uring: https://unixism.net/loti/
- Efficient IO with io_uring (PDF): https://kernel.dk/io_uring.pdf
- LWN io_uring articles: https://lwn.net/Kernel/Index/#io_uring

---

## Questions for Human Review

After research, flag these decisions for human input:

1. Should we support kernels older than 6.8? (increases complexity significantly)
2. Should we integrate with Go's netpoller or stay independent?
3. Should registered buffer management be automatic or manual?
4. How should we handle operations not supported on the running kernel?
5. Should reference implementations be in-repo or separate?
