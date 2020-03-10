package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/miun173/rebalance/proxy"
	"github.com/miun173/rebalance/sidecar"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// used for flags
	url          string
	hosts        string
	servicePorts string

	rootCMD  = &cobra.Command{}
	proxyCMD = &cobra.Command{
		Use:   "proxy",
		Short: "a reverse proxy",
		Run:   runProxy,
	}

	sidecarCMD = &cobra.Command{
		Use:   "sidecar",
		Short: "a sidecar proxy",
	}

	joinCMD = &cobra.Command{
		Use:     "join",
		Short:   "join a service into proxy",
		Example: "join --url 'http://127.0.0.1:9000'",
		Run:     runJoinProxy,
	}

	joinFromConfigCMD = &cobra.Command{
		Use:     "join-config",
		Short:   "join-config a service into proxy",
		Example: "join-config --url 'http://127.0.0.1:9000' --service-hosts '192.168.1.0:8000'",
		Run:     runJoinProxyFromConfig,
	}
)

func init() {
	joinCMD.PersistentFlags().StringVar(&url, "url", "", "proxy service url")
	joinCMD.PersistentFlags().StringVar(&servicePorts, "service-ports", "80", "services ports that will be proxied 80,8080,9000")

	joinFromConfigCMD.PersistentFlags().StringVar(&hosts, "service-hosts", "192.168.0.1", "service host to be joined")
	joinFromConfigCMD.PersistentFlags().StringVar(&url, "url", "", "proxy service url")

	sidecarCMD.AddCommand(joinCMD)
	sidecarCMD.AddCommand(joinFromConfigCMD)
	rootCMD.AddCommand(sidecarCMD)
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

func runJoinProxy(cmd *cobra.Command, args []string) {
	if url == "" {
		log.Fatal(errors.New("url should be in form 'http://127.0.0.1:9000'"))
	}

	ports := strings.Split(servicePorts, ",")
	sc := sidecar.NewSideCar(url)
	if err := sc.Join(ports...); err != nil {
		log.Fatal(err)
	}
}

func runJoinProxyFromConfig(cmd *cobra.Command, args []string) {
	if url == "" {
		log.Fatal(errors.New("url should be in form 'http://127.0.0.1:9000'"))
	}

	serviceHosts := strings.Split(hosts, ",")
	sc := sidecar.NewSideCar(url)
	if err := sc.JoinFromConfig(serviceHosts...); err != nil {
		log.Fatal(err)
	}
}
