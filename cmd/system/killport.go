package system

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/utils"
)

func NewKillPortCmd() *cobra.Command {
	var port int
	var force bool

	cmd := &cobra.Command{
		Use:     "killport",
		Aliases: []string{"kport"},
		Short:   "Find and optionally kill the process using a port",
		Long: `Example usage:
  cmdr killport -p 3000
  Will display the process using port 3000 and allow you to kill it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if port == 0 {
				fmt.Print("Enter port: ")
				fmt.Scanln(&port)
			}

			pid, err := utils.GetPIDByPort(port)
			if err != nil || pid == "" {
				fmt.Printf("No process is using port %d\n", port)
				return nil
			}

			fmt.Printf("Process using port %d: PID=%s\n", port, pid)
			utils.ShowProcessInfo(pid, runtime.GOOS)
			if force || utils.AskConfirmation("Do you want to terminate this process?") {
				utils.TerminateProcess(pid, runtime.GOOS)
			} else {
				fmt.Println("Process left running.")
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 0, "Port number to check")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Kill process without confirmation")

	return cmd
}
