package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fahmifan/rebalance/sidecar"
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
		Example: "join --url 'http://127.0.0.1:9000' --service-ports '8001,8002,8003'",
	}
	cmd.Flags().String("url", "", "proxy service url")
	cmd.Flags().String("service-ports", "", "services ports that will be proxied 80,8080,9000")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		url := cmd.Flag("url").Value.String()
		if url == "" {
			fmt.Println(errors.New("url should be in form 'http://127.0.0.1:9000'"))
			os.Exit(1)
		}

		ports := strings.Split(cmd.Flag("service-ports").Value.String(), ",")
		sc := sidecar.NewSideCar(url)
		if err := sc.Join(ports...); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	return cmd
}
