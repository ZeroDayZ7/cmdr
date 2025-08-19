package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "Say hello",
	Long:  `This command is used to test write out a greeting in the CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from CLI! 🚀")
	},
}

func init() {
	rootCmd.AddCommand(helloCmd)
}
