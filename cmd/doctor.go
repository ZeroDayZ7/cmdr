package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/doctor"
)

func NewDoctorCmd() *cobra.Command {
	var jsonOutput bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose cmdr environment and updater readiness",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := doctor.Run(doctor.Options{})
			if jsonOutput {
				if err := doctor.WriteJSON(cmd.OutOrStdout(), report); err != nil {
					return err
				}
			} else {
				if err := doctor.WriteText(cmd.OutOrStdout(), report, verbose); err != nil {
					return err
				}
			}
			if report.Summary.Failed > 0 {
				return errors.New("doctor checks found failures")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print diagnostics as JSON")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "show additional diagnostic messages")
	return cmd
}
