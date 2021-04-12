# Rebalance ![goreport](https://goreportcard.com/badge/github.com/fahmifan173/rebalance)

Experimentation on server load balancing using Round Robin algorithm with self join.

## TODO
- [x] proxy request to several ip
    - [x] proxy request to an ip
- [ ] add self join & discovery
    - [x] can join from proxied services
- [x] handle concurrent proxy requests
    - [reff](https://kasvith.github.io/posts/lets-create-a-simple-lb-go)

## Usages

There are two binaries `rebc` and `rebd`.
- `rebc`
```
rebc is client for rebalance

Usage:
   [command]

Available Commands:
  help        Help about any command
  join        join a service into proxy

Flags:
  -h, --help   help for this command
```

- `rebd`
```
rebd is server for rebalance

Usage:
   [command]

Available Commands:
  help        Help about any command
  join        join a remote service
  proxy       run reverse proxy

Flags:
  -h, --help   help for this command
```

## Build
```
make build
```

## Benchmarks
`$ make bench package=proxy`
```
goos: linux
goarch: amd64
pkg: github.com/fahmifan173/rebalance/proxy
cpu: Intel(R) Core(TM) i5-5300U CPU @ 2.30GHz
BenchmarkProxy/1_upstream-4                 4453            246196 ns/op           42159 B/op        109 allocs/op
BenchmarkProxy/2_upstream-4                 4616            240572 ns/op           42157 B/op        109 allocs/op
BenchmarkProxy/3_upstream-4                 4918            244510 ns/op           42172 B/op        109 allocs/op
BenchmarkProxy/4_upstream-4                 4933            245854 ns/op           42196 B/op        109 allocs/op
BenchmarkProxy/5_upstream-4                 4905            243645 ns/op           42231 B/op        109 allocs/op
BenchmarkProxy/6_upstream-4                 4760            246456 ns/op           42252 B/op        109 allocs/op
BenchmarkProxy/7_upstream-4                 4820            257239 ns/op           42230 B/op        109 allocs/op
BenchmarkProxy/8_upstream-4                 4958            233508 ns/op           42200 B/op        109 allocs/op
PASS
ok      github.com/fahmifan173/rebalance/proxy  14.629s
```
