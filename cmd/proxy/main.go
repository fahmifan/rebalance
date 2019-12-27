package main

import (
	"fmt"

	"github.com/miun173/rebalance/proxy"
)

var urls = []string{"http://localhost:8080", "http://localhost:8081"}

func main() {
	fmt.Println("starting loadbalancer at :9000")

	rr := proxy.NewRoundRobin()
	_ = rr.AddServer(urls[0])
	_ = rr.AddServer(urls[1])

	rr.Start()
}
