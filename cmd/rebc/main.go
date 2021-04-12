package main

import (
	"errors"
	"strings"

	"github.com/fahmifan173/rebalance/sidecar"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Short: "rebc is client for rebalance",
	}
	cmd.AddCommand(joinCMD())
	cmd.Execute()
}

func joinCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "join",
		Short:   "join a service into proxy",
		Example: "join --url 'http://127.0.0.1:9000'",
	}
	cmd.Flags().String("url", "url", "proxy service url")
	cmd.Flags().String("remote-url", "remote-url", "remote service url")
	cmd.Flags().String("service-ports", "80", "services ports that will be proxied 80,8080,9000")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		url := cmd.Flag("url").Value.String()
		servicePorts := cmd.Flag("service-ports").Value.String()

		if url == "" {
			log.Fatal(errors.New("url should be in form 'http://127.0.0.1:9000'"))
		}

		if url != "" {
			ports := strings.Split(servicePorts, ",")
			sc := sidecar.NewSideCar(url)
			if err := sc.Join(ports...); err != nil {
				log.Fatal(err)
			}
		}

	}

	return cmd
}
