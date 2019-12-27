run-proxy:
	@go run cmd/proxy/main.go

run-sidecar:
	@go run cmd/sidecar/main.go $(args)