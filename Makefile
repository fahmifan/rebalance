bench_cmd := go test -run=^$ github.com/miun173/rebalance/${package} -bench=.
run_cmd := go run cmd/*

test:
	go test -v -covermode=atomic ./...

run-proxy:
	@$(run_cmd) proxy

run-sidecar:
	@$(run_cmd) sidecar join --url $(url) --service-ports $(service-ports)

build:
	@go build -o output/rebalance ./cmd/...

bench:
	@cd proxy && $(bench_cmd)
