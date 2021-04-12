test:
	go test -v -covermode=atomic ./...

run-proxy:
	@go run cmd/rebd/* proxy

build:
	go build -ldflags="-w -s" -o output/rebd ./cmd/rebd/...
	go build -ldflags="-w -s" -o output/rebc ./cmd/rebc/...

bench:
	@go test -benchmem -run=^$ -bench '^(BenchmarkProxy)$' github.com/fahmifan173/rebalance/proxy

changelog:
ifdef version
	@git-chglog --next-tag $(version) -o CHANGELOG.md
else
	@git-chglog -o CHANGELOG.md
endif
