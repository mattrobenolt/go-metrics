TEST_FLAGS := "--rerun-fails=5 --packages=./... --format testname"

default:
    @just --list

# Format Go code
fmt: fmt-go

# Format Go code with golangci-lint
fmt-go:
    golangci-lint fmt

# Run linter
lint:
    golangci-lint run --show-stats

# Start documentation server
docs:
    go doc -http

# Run tests
test:
    GOEXPERIMENT=synctest gotestsum {{ TEST_FLAGS }}

# Run tests with race detector
test-race:
    GOEXPERIMENT=synctest gotestsum {{ TEST_FLAGS }} -- -race

# Run tests in watch mode
test-watch:
    GOEXPERIMENT=synctest gotestsum --watch {{ TEST_FLAGS }}

# Update benchmark results
update-benchmarks:
    go -C benchmarks/ test -v -bench=. -count=10 | tee compare.txt
    go run golang.org/x/perf/cmd/benchstat@latest -col /mod compare.txt | tee benchmarks.txt
    echo >> benchmarks.txt
    go test -v . -run=xxx -bench=. | tee -a benchmarks.txt
    rm compare.txt
