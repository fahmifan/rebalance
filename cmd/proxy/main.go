package main

import (
	"fmt"

	"github.com/miun173/rebalance/proxy"
)

func main() {
	fmt.Println("starting loadbalancer at :9000")

	rr := proxy.NewRoundRobin()
	rr.Start()
}
