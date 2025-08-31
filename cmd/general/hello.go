package general

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewHelloCmd tworzy podkomendę hello
func NewHelloCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hello",
		Short: "Say hello",
		Long:  `This command is used to test write out a greeting in the CLI.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello from CLI! 🚀")
		},
	}
}
