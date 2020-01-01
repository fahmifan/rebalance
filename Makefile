bench_cmd := go test -run=^$ github.com/miun173/rebalance/${package} -bench=.

run-proxy:
	@go run cmd/proxy/main.go

run-sidecar:
	@go run cmd/sidecar/main.go $(args)

build-proxy:
	@go build -o output/rebalance-proxy ./cmd/proxy/

bench:
	@cd proxy && $(bench_cmd)
