package console

import (
	"errors"
	"strings"

	"github.com/miun173/rebalance/sidecar"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	sidecarCMD = &cobra.Command{
		Use:   "sidecar",
		Short: "a sidecar proxy",
	}

	sideCarJoinCMD = &cobra.Command{
		Use:     "join",
		Short:   "join a service into proxy",
		Example: "join --url 'http://127.0.0.1:9000'",
		Run:     runSideCarJoinProxy,
	}
)

func init() {
	sideCarJoinCMD.Flags().String("url", "url", "proxy service url")
	sideCarJoinCMD.Flags().String("service-ports", "80", "services ports that will be proxied 80,8080,9000")

	sidecarCMD.AddCommand(sideCarJoinCMD)
	rootCMD.AddCommand(sidecarCMD)
}

func runSideCarJoinProxy(cmd *cobra.Command, args []string) {
	url := cmd.Flag("url").Value.String()
	servicePorts := cmd.Flag("service-ports").Value.String()

	if url == "" {
		log.Fatal(errors.New("url should be in form 'http://127.0.0.1:9000'"))
	}

	ports := strings.Split(servicePorts, ",")
	sc := sidecar.NewSideCar(url)
	if err := sc.Join(ports...); err != nil {
		log.Fatal(err)
	}
}
