package cli

import (
	"github.com/spf13/cobra"
)

var rootCMD = &cobra.Command{}

// Run ..
func Run() {
	rootCMD.AddCommand(sideCarCMD())
	rootCMD.Execute()
}
