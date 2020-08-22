package console

import (
	"github.com/spf13/cobra"
)

var rootCMD = &cobra.Command{}

// Run ..
func Run() {
	rootCMD.Execute()
}
