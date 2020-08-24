package console

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	sp := proxy.NewServiceProxy()

	signalCh := make(chan os.Signal, 1)
	defer close(signalCh)

	signal.Notify(signalCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	localJoinProxy(sp)
	go sp.Start()
	go sp.RunHealthCheck()

	fmt.Println("starting loadbalancer at :9000")

	<-signalCh
	log.Println("exiting...")
}

func localJoinProxy(sp *proxy.ServiceProxy, urls ...string) {
	_, err := os.Stat("config.json")
	if os.IsNotExist(err) {
		log.Info("config.json not found. Skipping local join")
		return
	}

	bt, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	st := &struct {
		Hosts []string `json:"hosts"`
	}{}

	err = json.Unmarshal(bt, st)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("run local join")
	for _, host := range st.Hosts {
		err := sp.AddServer(host)
		if err != nil {
			log.Error(err)
			continue
		}

		log.Infof("succes join %s", host)
	}
}
