# go-iouring Makefile
# All project commands should be run via make

.PHONY: all build test test-v test-run bench clean generate check-iouring lint fmt vet help

# Default target
all: build

# Build all packages
build:
	go build ./...

# Run all tests
test:
	go test ./...

# Run tests with verbose output
test-v:
	go test -v ./...

# Run specific test (usage: make test-run TEST=TestRingSetup)
test-run:
	go test -v -run $(TEST) ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Run benchmarks with count for stability (usage: make bench-count COUNT=5)
bench-count:
	go test -bench=. -benchmem -count=$(COUNT) ./...

# Run code generation from kernel headers
generate:
	go generate ./...

# Check if io_uring is enabled on this system
check-iouring:
	@echo "Checking io_uring status..."
	@if [ -f /proc/sys/kernel/io_uring_disabled ]; then \
		STATUS=$$(cat /proc/sys/kernel/io_uring_disabled); \
		case $$STATUS in \
			0) echo "io_uring: ENABLED (fully available)" ;; \
			1) echo "io_uring: RESTRICTED (unprivileged disabled)" ;; \
			2) echo "io_uring: DISABLED (fully disabled)" ;; \
			*) echo "io_uring: UNKNOWN status ($$STATUS)" ;; \
		esac; \
	else \
		echo "io_uring: status file not found (kernel may not support it)"; \
	fi
	@echo ""
	@echo "Kernel version:"
	@uname -r

# Run go vet
vet:
	go vet ./...

# Format code
fmt:
	go fmt ./...

# Run staticcheck if installed
lint:
	@which staticcheck > /dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"

# Clean build artifacts
clean:
	go clean ./...
	rm -f coverage.out

# Run tests with coverage
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Run tests with coverage and open HTML report
cover-html: cover
	go tool cover -html=coverage.out

# Run tests with race detector
test-race:
	go test -race ./...

# Install development tools
tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest

# Verify module dependencies
mod-verify:
	go mod verify

# Tidy module dependencies
mod-tidy:
	go mod tidy

# Download dependencies
mod-download:
	go mod download

# Show help
help:
	@echo "go-iouring Makefile targets:"
	@echo ""
	@echo "  build         - Build all packages"
	@echo "  test          - Run all tests"
	@echo "  test-v        - Run tests with verbose output"
	@echo "  test-run      - Run specific test (TEST=TestName)"
	@echo "  test-race     - Run tests with race detector"
	@echo "  bench         - Run benchmarks"
	@echo "  bench-count   - Run benchmarks with count (COUNT=N)"
	@echo "  cover         - Run tests with coverage report"
	@echo "  cover-html    - Open coverage report in browser"
	@echo "  generate      - Run code generation"
	@echo "  check-iouring - Check if io_uring is enabled"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Run staticcheck"
	@echo "  clean         - Clean build artifacts"
	@echo "  tools         - Install development tools"
	@echo "  mod-tidy      - Tidy module dependencies"
	@echo "  mod-verify    - Verify module dependencies"
	@echo "  mod-download  - Download dependencies"
	@echo "  help          - Show this help"
