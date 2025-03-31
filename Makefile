BIN := bin

clean: clean-bin

$(BIN):
	mkdir -p $(BIN)

TOOL_INSTALL := env GOBIN=$(PWD)/$(BIN) go install

$(BIN)/golangci-lint: Makefile | $(BIN)
	wget -qO- https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b bin v1.64.8

$(BIN)/gofumpt: Makefile | $(BIN)
	$(TOOL_INSTALL) mvdan.cc/gofumpt@v0.7.0

$(BIN)/godoc: | $(BIN)
	$(TOOL_INSTALL) golang.org/x/tools/cmd/godoc@latest

$(BIN)/gotestsum: Makefile | $(BIN)
	$(TOOL_INSTALL) gotest.tools/gotestsum@v1.12.1

$(BIN)/benchstat: $(BIN)
	$(TOOL_INSTALL) golang.org/x/perf/cmd/benchstat@latest

tools: $(BIN)/golangci-lint $(BIN)/gofumpt $(BIN)/godoc $(BIN)/gotestsum $(BIN)/benchstat

fmt: fmt-go

fmt-go: $(BIN)/gofumpt
	$< -l -w .

lint: $(BIN)/golangci-lint
	$< run \
	  --show-stats --no-config \
	  -E gofumpt,govet,noctx,perfsprint,unconvert \
	  -D errcheck

docs: $(BIN)/godoc
	$< -http=127.0.0.1:6060

test: $(BIN)/gotestsum
	GOEXPERIMENT=synctest $< --format testname -- -v ./...

test-race: $(BIN)/gotestsum
	GOEXPERIMENT=synctest $< --format testname -- -v -race ./...

test-watch: $(BIN)/gotestsum
	GOEXPERIMENT=synctest $< --watch --format testname -- -v ./...

update-benchmarks:
	go -C benchmarks/ test -v -bench=. -count=10 | tee compare.txt
	go run golang.org/x/perf/cmd/benchstat@latest -col /mod compare.txt | tee benchmarks.txt
	echo >> benchmarks.txt
	go test -v . -run=xxx -bench=. | tee -a benchmarks.txt
	rm compare.txt

.PHONY: clean clean-bin tools \
        fmt fmt-go \
        lint \
        update-benchmarks \
        docs test test-watch test-race
