package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/username/cli/internal/utils"
)

var (
	processName string
	forceKill   bool
)

var killProcessCmd = &cobra.Command{
	Use:     "killprocess",
	Aliases: []string{"kproc"},
	Short:   "Find and optionally kill processes by name",
	Long: `Example usage:
  cmdr killprocess -n WPF
  Will display all processes matching 'WPF' and ask to kill each.`,
	Run: func(cmd *cobra.Command, args []string) {
		if processName == "" {
			fmt.Print("Enter process name: ")
			fmt.Scanln(&processName)
		}

		if runtime.GOOS == "windows" {
			handleWindowsProcess(processName)
		} else {
			handleUnixProcess(processName)
		}
	},
}

func init() {
	killProcessCmd.Flags().StringVarP(&processName, "name", "n", "", "Process name to find")
	killProcessCmd.Flags().BoolVarP(&forceKill, "force", "f", false, "Kill process without confirmation")
	rootCmd.AddCommand(killProcessCmd)
}

// ==========================
// Windows
// ==========================
func handleWindowsProcess(name string) {
	cmd := exec.Command("tasklist", "/FO", "LIST")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing tasklist:", err)
		return
	}

	lines := strings.Split(string(out), "\n")
	var pids []string
	var infos []string

	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "Image Name:") && strings.Contains(strings.ToLower(lines[i]), strings.ToLower(name)) {
			image := strings.TrimSpace(strings.TrimPrefix(lines[i], "Image Name:"))
			if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "PID:") {
				pidLine := strings.TrimSpace(strings.TrimPrefix(lines[i+1], "PID:"))
				pids = append(pids, pidLine)
				infos = append(infos, fmt.Sprintf("%s (PID=%s)", image, pidLine))
			}
		}
	}

	if len(pids) == 0 {
		fmt.Printf("No process found matching '%s'\n", name)
		return
	}

	for i, pid := range pids {
		fmt.Println("Found process:", infos[i])
		utils.ConfirmAndKill(pid, "windows", forceKill)
	}
}

// ==========================
// Unix / Linux / Mac
// ==========================
func handleUnixProcess(name string) {
	cmd := exec.Command("pgrep", "-fl", name)
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		fmt.Printf("No process found matching '%s'\n", name)
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid := fields[0]
		processInfo := strings.Join(fields[1:], " ")
		fmt.Printf("Found process: %s (PID=%s)\n", processInfo, pid)
		utils.ConfirmAndKill(pid, "unix", forceKill)
	}
}
