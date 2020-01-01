bench_cmd := go test -run=^$ github.com/miun173/rebalance/${package} -bench=.
run_cmd := go run cmd/*

test:
	go test -v -covermode=atomic ./...

run-proxy:
	@$(run_cmd) proxy

run-sidecar:
	@$(run_cmd) sidecar join --url $(url) --service-ports $(service-ports)

build-proxy:
	@go build -o output/rebalance-proxy ./cmd/proxy/

bench:
	@cd proxy && $(bench_cmd)
