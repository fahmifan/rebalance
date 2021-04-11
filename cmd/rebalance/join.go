package main

import (
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func joinCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join",
		Short: "join a remote service",
	}
	cmd.Flags().String("host", "host", "service host")
	cmd.Flags().String("lbport", "lbport", "loadbalancer port")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		host := cmd.Flag("host").Value.String()
		lbport := cmd.Flag("lbport").Value.String()

		reqURL := "http://localhost:" + lbport + "/rebalance/local-join?host=" + host
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

	return cmd
}
