# Rebalance ![goreport](https://goreportcard.com/badge/github.com/fahmifan/rebalance)

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
pkg: github.com/fahmifan/rebalance/proxy
cpu: Intel(R) Core(TM) i5-5300U CPU @ 2.30GHz
BenchmarkProxy/1_upstream-4                 4628            248533 ns/op           42164 B/op        109 allocs/op
BenchmarkProxy/2_upstream-4                 4952            243464 ns/op           42160 B/op        109 allocs/op
BenchmarkProxy/3_upstream-4                 4960            241166 ns/op           42174 B/op        109 allocs/op
BenchmarkProxy/4_upstream-4                 4137            246276 ns/op           42201 B/op        109 allocs/op
BenchmarkProxy/5_upstream-4                 4810            246440 ns/op           42226 B/op        109 allocs/op
BenchmarkProxy/6_upstream-4                 4791            252848 ns/op           42265 B/op        109 allocs/op
BenchmarkProxy/7_upstream-4                 4732            245577 ns/op           42228 B/op        109 allocs/op
BenchmarkProxy/8_upstream-4                 5017            233513 ns/op           42197 B/op        109 allocs/op
PASS
ok      github.com/fahmifan/rebalance/proxy  13.459s
```
