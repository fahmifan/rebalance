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

	sp := proxy.NewServiceProxy()

	signalCh := make(chan os.Signal, 1)
	defer close(signalCh)

	signal.Notify(signalCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go sp.Start()
	go sp.RunHealthCheck()

	<-signalCh
	log.Println("exiting...")
}
