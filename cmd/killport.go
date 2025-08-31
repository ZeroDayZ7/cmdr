package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/utils"
)

var port int

var killPortCmd = &cobra.Command{
	Use:     "killport",
	Short:   "Find and optionally kill the process using a port",
	Aliases: []string{"kport"},
	Long: `Example usage:
  cmdr killport -p 3000
  Will display the process using port 3000 and allow you to kill it.`,
	Run: func(cmd *cobra.Command, args []string) {
		if port == 0 {
			fmt.Print("Enter port: ")
			fmt.Scanln(&port)
		}

		if runtime.GOOS == "windows" {
			handleWindows(port)
		} else {
			handleUnix(port)
		}
	},
}

var force bool

func init() {
	killPortCmd.Flags().IntVarP(&port, "port", "p", 0, "Port number to check")
	killPortCmd.Flags().BoolVarP(&force, "force", "f", false, "Kill process without confirmation")
	rootCmd.AddCommand(killPortCmd)
}

// ==========================
// Unix (Linux/Mac)
// ==========================
func handleUnix(port int) {
	fmt.Printf("Checking port %d...\n", port)
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-t")
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		fmt.Printf("No process is using port %d\n", port)
		return
	}
	pid := strings.TrimSpace(string(out))
	fmt.Printf("Process using port %d: PID=%s\n", port, pid)

	// show process info
	psCmd := exec.Command("ps", "-p", pid, "-o", "pid,ppid,user,%cpu,%mem,etime,cmd")
	psCmd.Stdout = os.Stdout
	psCmd.Run()

	utils.ConfirmAndKill(pid, "unix", force)
}

// ==========================
// Windows
// ==========================
func handleWindows(port int) {
	fmt.Printf("Checking port %d...\n", port)
	cmd := exec.Command("netstat", "-ano")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing netstat:", err)
		return
	}
	lines := strings.Split(string(out), "\n")
	var pid string
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf(":%d", port)) {
			fields := strings.Fields(line)
			if len(fields) > 4 {
				pid = fields[len(fields)-1]
				break
			}
		}
	}

	if pid == "" {
		fmt.Printf("No process is using port %d\n", port)
		return
	}

	fmt.Printf("Process using port %d: PID=%s\n", port, pid)

	// show basic info
	tasklist := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %s", pid), "/FO", "LIST")
	tasklist.Stdout = os.Stdout
	tasklist.Run()

	utils.ConfirmAndKill(pid, "windows", force)
}
