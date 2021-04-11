package cli

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/miun173/rebalance/sidecar"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func sideCarCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sidecar",
		Short: "a sidecar",
	}

	joinCMD := &cobra.Command{
		Use:     "join",
		Short:   "join a service into proxy",
		Example: "join --url 'http://127.0.0.1:9000'",
	}
	joinCMD.Flags().String("url", "url", "proxy service url")
	joinCMD.Flags().String("remote-url", "remote-url", "remote service url")
	joinCMD.Flags().String("service-ports", "80", "services ports that will be proxied 80,8080,9000")
	joinCMD.Run = func(cmd *cobra.Command, args []string) {
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

	remoteCMD := &cobra.Command{
		Use:   "remote",
		Short: "join a remote service",
	}
	remoteCMD.Flags().String("host", "host", "service host")
	remoteCMD.Flags().String("lb", "lb", "loadbalancer host")
	remoteCMD.Run = func(cmd *cobra.Command, args []string) {
		host := cmd.Flag("host").Value.String()
		lb := cmd.Flag("lb").Value.String()

		reqURL := lb + "/rebalance/local-join?host=" + host
		resp, err := http.Get(reqURL)
		if err != nil {
			log.Fatal(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatal(string(body))
		}

		log.Println(string(body))
	}

	joinCMD.AddCommand(remoteCMD)
	cmd.AddCommand(joinCMD)
	return cmd
}
