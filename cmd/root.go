package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/cmd/crypto"
)

var rootCmd = &cobra.Command{
	Use:   "cmdr",
	Short: "cmdr is a developer helper tool",
	Long:  `cmdr helps you run tasks like migrate, dev server, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to cmdr! Use -h to see available commands.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Rejestracja podkomend
	rootCmd.AddCommand(
		crypto.NewCmd(), // <-- to z Twojego crypto.go
		// możesz dodać kolejne, np. ask.NewCmd(), hashdir.NewCmd(), itp.
	)
}
