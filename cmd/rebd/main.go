package main

import (
	"github.com/spf13/cobra"
)

func main() {
	rootCMD := &cobra.Command{
		Short: "rebd is server for rebalance",
	}
	rootCMD.AddCommand(proxyCMD(), joinCMD())
	rootCMD.Execute()
}
