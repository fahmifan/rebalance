package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/miun173/rebalance/proxy"
)

func main() {
	fmt.Println("starting loadbalancer at :9000")

	rr := proxy.NewServiceProxy()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go rr.Start()
	go rr.RunHealthCheck()

	<-signalCh
	log.Println("exiting...")
}
