# All targets are phony (no file named "build", "test", etc. should shadow these).
.PHONY: bench bench-full bench-profile-block bench-profile-cpu bench-profile-mem bench-profile-trace \
	build build-release build-docker codequality-baseline codequality-gate codequality-review codequality-scorecard format help lint lint-ci \
	test test-coverage test-integration test-process-integration test-race test-regression test-unit

# Isolated Redpanda per benchmark: wall time grows with container churn; allow a generous cap.
BENCH_PATTERN ?= BenchmarkBridgeRelay256B|BenchmarkKafkaRoundTrip256B|BenchmarkBridgeRelayBurst256B
BENCH_TIME ?= 2s
BENCH_TIMEOUT ?= 30m
# Set to empty or a higher value if you want more OS threads during a benchmark (default 1 = sequential CPU).
BENCH_GOMAXPROCS ?= 1

# Optional local revision/time for cmd/bifrost/version (matches CI-style ldflags).
REV := $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

help: ## Show available make targets
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "%-24s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

lint: ## Run go vet, module verify, govulncheck, gosec, golangci-lint
	go vet ./...
	go mod verify
	go tool govulncheck ./cmd/... ./pkg/...
	go tool gosec -fmt text -stdout -quiet ./cmd/... ./pkg/...
	golangci-lint run ./...

lint-ci: ## Run CI lint gate for cmd/pkg/unit tests
	go vet ./cmd/... ./pkg/... ./test/unit/...
	go mod verify
	golangci-lint run ./cmd/... ./pkg/... ./test/unit/...
	go run github.com/kisielk/errcheck@latest ./cmd/... ./pkg/... ./test/unit/...
	go run github.com/mgechev/revive@latest -config revive.toml -formatter stylish ./cmd/... ./pkg/... ./test/unit/...

codequality-scorecard: ## Build codequality scorecard JSON/Markdown
	python3 scripts/codequality_pipeline.py --output-prefix scorecard

codequality-baseline: ## Capture baseline codequality snapshot for regression gating
	python3 scripts/codequality_pipeline.py --output-prefix baseline-scorecard --write-baseline

codequality-gate: ## Run codequality gate checks (absolute + regression)
	python3 scripts/codequality_pipeline.py --output-prefix gate-scorecard --enforce

codequality-review: ## Generate scorecard and print top hotspots for weekly triage
	@echo "Generating weekly scorecard..."
	@python3 scripts/codequality_pipeline.py --output-prefix weekly-scorecard
	@echo "Top hotspots (churn x complexity):"
	@python3 scripts/print_codequality_hotspots.py

format: ## Run go fmt and gofmt
	go fmt ./...
	gofmt -w .

build: ## Build bifrost binary from ./cmd/bifrost
	go build -trimpath -ldflags "-s -w -X github.com/lolocompany/bifrost/cmd/bifrost/version.Revision=$(REV) -X github.com/lolocompany/bifrost/cmd/bifrost/version.BuildTime=$(BUILD_TIME)" -o bifrost ./cmd/bifrost

# Env vars match goreleaser/.goreleaser.yaml (CI sets these in the release workflow).
# Requires https://github.com/anchore/syft on PATH for SBOM generation.
build-release: ## Build GoReleaser snapshot to dist/ (no publish)
	@command -v syft >/dev/null 2>&1 || { echo >&2 "syft not on PATH (install: https://github.com/anchore/syft#installation)"; exit 1; }
	@command -v goreleaser >/dev/null 2>&1 || { echo >&2 "goreleaser not on PATH (install: https://goreleaser.com/install/)"; exit 1; }
	BIFROST_BUILD_TIME=$(BUILD_TIME) RELEASE_NAME=local-snapshot RELEASE_BODY='Local snapshot (not a production release).' goreleaser release --snapshot --clean --skip=publish,validate --config goreleaser/.goreleaser.yaml

test: ## Alias for test-unit
	$(MAKE) test-unit

test-unit: ## Run unit tests (./test/unit/...)
	go test -shuffle=on -timeout 120s ./test/unit/...

test-race: ## Run unit tests with race detector
	go test -race -shuffle=on -timeout 180s ./test/unit/...

test-integration: ## Run integration tests (BIFROST_INTEGRATION=1)
	BIFROST_INTEGRATION=1 go test -shuffle=on -timeout 120s ./test/integration/...

test-process-integration: ## Run process-driven integration tests
	BIFROST_INTEGRATION=1 go test -shuffle=on -timeout 180s ./test/integration/...

test-regression: ## Run config regression tests (BIFROST_INTEGRATION=1)
	BIFROST_INTEGRATION=1 go test -shuffle=on -timeout 300s ./test/regression/...

# Tests live under test/..., so default -cover only sees external test packages (no pkg statements) → 0%.
# -coverpkg instruments pkg/ and cmd/ when those packages are exercised from test/ packages.
test-coverage: ## Run tests with coverage output and HTML report
	BIFROST_INTEGRATION=1 go test -coverprofile=coverage.out -coverpkg=./pkg/... ./test/...
	go tool cover -html=coverage.out

# GOMAXPROCS=$(BENCH_GOMAXPROCS) and -p 1: one package worker; default one OS thread for stable CPU.
# Benchmark lines go to stdout; slog and other diagnostics go to stderr. Keep them separate so
# go tool benchstat can parse bench.txt (merged stderr breaks benchmark line continuations).
bench: ## Run benchmark subset (Docker-backed)
	BIFROST_BENCHMARK=1 GOMAXPROCS=$(BENCH_GOMAXPROCS) go test -p 1 -bench='$(BENCH_PATTERN)' -benchmem -benchtime=$(BENCH_TIME) -timeout=$(BENCH_TIMEOUT) ./test/benchmark/... 2>bench.err | tee bench.txt
	go tool benchstat bench.txt

# Run all benchmarks 6 times.
bench-full: ## Run all benchmarks with count=6
	BIFROST_BENCHMARK=1 GOMAXPROCS=$(BENCH_GOMAXPROCS) go test -count=6 -p 1 -bench=. -benchmem -benchtime=$(BENCH_TIME) -timeout=$(BENCH_TIMEOUT) ./test/benchmark/... 2>bench.err | tee bench.txt
	go tool benchstat bench.txt

bench-profile-cpu: ## Run CPU profile benchmark and open pprof UI
	BIFROST_BENCHMARK=1 GOMAXPROCS=$(BENCH_GOMAXPROCS) go test -p 1 -bench=$(BENCH_PATTERN) -benchmem -benchtime=10s -timeout=$(BENCH_TIMEOUT) -cpuprofile=test/benchmark/.prof/cpu.prof 	./test/benchmark/...
	go tool pprof -http=:5432 test/benchmark/.prof/cpu.prof

bench-profile-block: ## Run block profile benchmark and open pprof UI
	BIFROST_BENCHMARK=1 GOMAXPROCS=$(BENCH_GOMAXPROCS) go test -p 1 -bench=$(BENCH_PATTERN) -benchmem -benchtime=10s -timeout=$(BENCH_TIMEOUT) -blockprofile=test/benchmark/.prof/block.prof 	./test/benchmark/...
	go tool pprof -http=:5432 test/benchmark/.prof/block.prof

bench-profile-trace: ## Run trace benchmark and open go tool trace
	BIFROST_BENCHMARK=1 GOMAXPROCS=$(BENCH_GOMAXPROCS) go test -p 1 -bench=$(BENCH_PATTERN) -benchmem -benchtime=10s -timeout=$(BENCH_TIMEOUT) -trace=test/benchmark/.prof/trace.out 	./test/benchmark/...
	go tool trace test/benchmark/.prof/trace.out

bench-profile-mem: ## Run memory profile benchmark and open pprof UI
	BIFROST_BENCHMARK=1 GOMAXPROCS=$(BENCH_GOMAXPROCS) go test -p 1 -bench=$(BENCH_PATTERN) -benchmem -benchtime=10s -timeout=$(BENCH_TIMEOUT) -memprofile=test/benchmark/.prof/mem.prof 	./test/benchmark/...
	go tool pprof -http=:5432 test/benchmark/.prof/mem.prof
