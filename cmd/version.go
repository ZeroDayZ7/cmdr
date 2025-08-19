package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cmdr v0.1.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
