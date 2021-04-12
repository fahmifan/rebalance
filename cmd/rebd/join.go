package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func joinCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join",
		Short: "join a remote service",
	}
	cmd.Flags().String("hosts", "http://localhost:8001", `--hosts "http://localhost:8001,http://localhost:8002"`)
	cmd.Flags().String("lbport", "9000", "loadbalancer port")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		hosts := cmd.Flag("hosts").Value.String()
		lbport := cmd.Flag("lbport").Value.String()

		for _, host := range strings.Split(hosts, ",") {
			reqURL := "http://localhost:" + lbport + "/rebalance/local-join?host=" + host
			resp, err := http.Get(reqURL)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Println(string(body))
				os.Exit(1)
			}

			log.Println(string(body))
		}
	}

	return cmd
}
