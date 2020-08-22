package console

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/miun173/rebalance/proxy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	proxyCMD = &cobra.Command{
		Use:   "proxy",
		Short: "a reverse proxy",
		Run:   runProxy,
	}
)

func init() {
	rootCMD.AddCommand(proxyCMD)
}

func runProxy(cmd *cobra.Command, args []string) {
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
