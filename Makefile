run-proxy:
	@go run cmd/proxy/main.go

run-sidecar:
	@go run cmd/sidecar/main.go $(args)

build-proxy:
	@go build -o output/rebalance-proxy ./cmd/proxy/