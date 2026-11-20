//go:build !windows

package cmd

import "github.com/spf13/cobra"

func NewUpdateHelperCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "update-helper",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
