//go:build windows

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/updater"
)

func NewUpdateHelperCmd() *cobra.Command {
	var target string
	var staged string
	var result string
	var pid int
	cmd := &cobra.Command{
		Use:    "update-helper",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return updater.ApplyWindowsUpdateHelper(target, staged, result, pid)
		},
	}
	cmd.Flags().StringVar(&target, "target", "", "")
	cmd.Flags().StringVar(&staged, "staged", "", "")
	cmd.Flags().StringVar(&result, "result", "", "")
	cmd.Flags().IntVar(&pid, "pid", 0, "")
	return cmd
}
