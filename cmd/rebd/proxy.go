package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fahmifan173/rebalance/proxy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func proxyCMD() *cobra.Command {
	proxyCMD := &cobra.Command{
		Use:   "proxy",
		Short: "run reverse proxy",
		Run: func(cmd *cobra.Command, args []string) {
			sp := proxy.NewProxy()

			signalCh := make(chan os.Signal, 1)
			defer close(signalCh)
			signal.Notify(signalCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

			joinFromConfig(sp)
			go sp.Start()
			go sp.RunHealthCheck(signalCh)

			fmt.Println("starting loadbalancer at :9000")
			<-signalCh

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			log.Println("stopping server ...")
			sp.Stop(ctx)
		},
	}

	return proxyCMD
}

func joinFromConfig(sp *proxy.Proxy, urls ...string) {
	_, err := os.Stat("config.json")
	if os.IsNotExist(err) {
		log.Info("config.json not found. Skipping local join")
		return
	}

	bt, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	st := struct {
		Hosts []string `json:"hosts"`
	}{}

	err = json.Unmarshal(bt, &st)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("run local join")
	for _, host := range st.Hosts {
		err := sp.AddService(host)
		if err != nil {
			log.Error(err)
			continue
		}

		log.Infof("succes join %s", host)
	}
}
