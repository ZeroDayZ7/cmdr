package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ConfirmAndKill asks the user and kills the process by PID
func ConfirmAndKill(pid string, osType string, force bool) {
	if force {
		kill(pid, osType)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Do you want to kill this process? (no/yes) [default no]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "" || input == "no" {
			fmt.Println("Process left running.")
			break
		} else if input == "yes" {
			kill(pid, osType)
			break
		} else {
			fmt.Println("Please enter 'yes' or 'no'.")
		}
	}
}

func kill(pid string, osType string) {
	var cmd *exec.Cmd
	if osType == "windows" {
		cmd = exec.Command("taskkill", "/PID", pid, "/F")
	} else {
		cmd = exec.Command("kill", "-9", pid)
	}
	if err := cmd.Run(); err != nil {
		fmt.Println("Error killing process:", err)
		return
	}
	fmt.Printf("Process %s killed.\n", pid)
}
