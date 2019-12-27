package main

import (
	"errors"
	"flag"
	"log"

	"github.com/miun173/rebalance/sidecar"
)

func main() {
	url := flag.String("url", "", "load balancer url")
	flag.Parse()

	if url == nil || *url == "" {
		log.Fatal(errors.New("url cannot empty"))
	}

	sc := sidecar.NewSideCar(*url)
	if err := sc.Join(); err != nil {
		log.Fatal(err)
	}
}
