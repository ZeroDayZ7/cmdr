package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/cmd/crypto"
	"github.com/zerodayz7/cmdr/cmd/general"
	"github.com/zerodayz7/cmdr/cmd/info"
	"github.com/zerodayz7/cmdr/cmd/system"
	"github.com/zerodayz7/cmdr/cmd/tools"
	"github.com/zerodayz7/cmdr/internal/logger"
)

var l = &logger.ConsoleLogger{
	IsVerbose: false,
}

var rootCmd = &cobra.Command{
	Use:   "cmdr",
	Short: "cmdr is a developer helper tool",
	Long:  `cmdr helps you run tasks like migrate, dev server, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		l.Info("Welcome to cmdr! Use -h to see available commands.")
	},
}

// #region Execute
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		l.Error("Critical failure: %v", err)
		os.Exit(1)
	}
}

// #region init
func init() {
	rootCmd.AddCommand(
		crypto.NewCmd(),               // crypto
		general.NewHelloCmd(l),        // hello
		info.NewInfoCmd(),             // info
		general.NewVersionCmd(l),      // version
		general.NewTreeCmd(),          // tree
		general.NewAskCmd(),           // ask
		tools.NewFilesCombineCmd(),    // files-combine
		tools.NewCleanCmd(),           // clean
		tools.NewCreateServiceCmd(),   // create-service
		tools.NewRemoveCommentsCmd(),  // remove-comments
		tools.NewAnnotateRegionsCmd(), // annotate-regions / reg
		tools.NewAnnotateCmd(),        // adnotacje
		system.NewChecksumCmd(l),      // checksum
		system.NewKillPortCmd(),       // killport
		system.NewKillProcessCmd(),    // killprocess

	)
}
