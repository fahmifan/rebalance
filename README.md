# Rebalance

Experimentation on server load balancing using Round Robin algorithm with self join.

## TODO
- [x] proxy request to several ip
    - [x] proxy request to an ip
- [ ] add self join & discovery
    - [x] can join from proxied services
- [x] handle concurrent proxy requests
    - [reff](https://kasvith.github.io/posts/lets-create-a-simple-lb-go)

# Benchmarks
`$ make bench package=proxy`
```
goos: linux
goarch: amd64
pkg: github.com/miun173/rebalance/proxy
Benchmark4Upstream/200_microsecond/response/1000_req-4              1000000000          0.87 ns/op
Benchmark4Upstream/200_microsecond/response/10000_req-4             1                   8659798397 ns/op
Benchmark4Upstream/20_microsecond/response/1000_req-4               1000000000          0.87 ns/op
Benchmark4Upstream/20_microsecond/response/10000_req-4              1                   8673090255 ns/op
PASS
ok      github.com/miun173/rebalance/proxy      79.571s
```