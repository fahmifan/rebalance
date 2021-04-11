bench_cmd := go test -run=^$ github.com/miun173/rebalance/${package} -bench=.
run_cmd := go run cmd/*

test:
	go test -v -covermode=atomic ./...

run-proxy:
	@$(run_cmd) proxy

run-sidecar:
	@$(run_cmd) sidecar join --url $(url) --service-ports $(service-ports)

build:
	go build -ldflags="-w -s" -o output/rebalance ./cmd/rebalance/...
	go build -ldflags="-w -s" -o output/client ./cmd/client/...

bench:
	@cd proxy && $(bench_cmd)

changelog:
ifdef version
	@git-chglog --next-tag $(version) -o CHANGELOG.md
else
	@git-chglog -o CHANGELOG.md
endif
