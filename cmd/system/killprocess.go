package system

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/utils"
)

func NewKillProcessCmd() *cobra.Command {
	var name string
	var force bool

	cmd := &cobra.Command{
		Use:     "killprocess",
		Aliases: []string{"kproc"},
		Short:   "Find and optionally kill processes by name",
		Long: `Example usage:
  cmdr killprocess -n chrome
  Will display all processes matching 'chrome' and allow you to kill them.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				fmt.Print("Enter process name: ")
				if _, err := fmt.Scanln(&name); err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}
			}

			processes, err := utils.GetProcessesByName(name)
			if err != nil {
				return fmt.Errorf("failed to get processes: %w", err)
			}

			if len(processes) == 0 {
				fmt.Printf("No process found matching '%s'\n", name)
				return nil
			}

			for _, p := range processes {
				fmt.Printf("Found process: %s (PID=%s)\n", p.Name, p.PID)
				if force || utils.AskConfirmation("Do you want to terminate this process?") {
					utils.TerminateProcess(p.PID, runtime.GOOS)
					fmt.Printf("Process %s terminated.\n", p.PID)
				} else {
					fmt.Println("Process left running.")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Process name to find")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Kill process without confirmation")

	return cmd
}
