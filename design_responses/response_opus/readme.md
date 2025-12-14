Research complete. I've produced all five deliverables:
View operations.json - Complete catalog of 57 io_uring operations with metadata including minimum kernel versions, SQE fields used, supported flags, and implementation priority.
View kernel_matrix.md - Feature availability mapped to kernel versions from 5.1 through 6.12, with recommendations for minimum kernel by use case.
View api_design.md - Recommended Go API design including type definitions, builder pattern for operations, memory management considerations, and integration strategies with Go's runtime.
View implementation_order.md - Phased implementation plan across 16 weeks covering foundation, core I/O, network I/O, advanced features, filesystem operations, and production hardening.
View test_plan.md - Comprehensive testing strategy with unit tests, integration tests, benchmarks, CI/CD configuration for multi-kernel testing, and fuzzing approaches.

Key Findings for Decision Points
1. Kernel support threshold: Recommend targeting 5.15 LTS as minimum, with graceful feature detection for newer operations. Supporting kernels older than 5.6 adds significant complexity for minimal benefit.
2. Go netpoller integration: Stay independent initially. The complexity of integrating with the runtime internals (gopark/goready are internal) isn't worth it. Use eventfd registration + epoll on the ring fd for goroutine parking, similar to how godzie44/go-uring handles it.
3. Registered buffer management: Provide both automatic and manual APIs. Automatic for simple cases (buffer pools for recv), manual for performance-critical paths where the user knows their access patterns.
4. Unsupported operations: Use Probe() for runtime detection. Return ErrNotSupported for operations the kernel doesn't support, with clear documentation of minimum kernel versions.
5. Reference implementations: Keep examples in-repo under /examples - echo server, file copy, HTTP server. This keeps them tested alongside the library.

Security Note
io_uring has had significant security issues. Google's security team reported that 60% of their 2022 bug bounty submissions exploited io_uring vulnerabilities, leading them to disable it on Android apps, ChromeOS, and production servers. Docker 25+ blocks io_uring by default via seccomp. Wikipedia Applications should detect and handle seccomp restrictions gracefully, and production deployments should stay current on kernel security patches.
