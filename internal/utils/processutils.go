package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// #region ProcessInfo
// Represents basic information about a process.
type ProcessInfo struct {
	PID  string
	Name string
}

// #region AskConfirmation
// Prompts user for yes/no confirmation.
func AskConfirmation(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(message + " (no/yes) [default no]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		switch input {
		case "", "no":
			return false
		case "yes":
			return true
		default:
			fmt.Println("Please enter 'yes' or 'no'.")
		}
	}
}

// #region terminateProcess
// Kills the process by PID, using the appropriate command for the OS.
func TerminateProcess(pid string, osType string) {
	var cmdArgs []string
	var name string

	if osType == "windows" {
		name = "taskkill"
		cmdArgs = []string{"/PID", pid, "/F"}
	} else {
		name = "kill"
		cmdArgs = []string{"-9", pid}
	}

	if _, err := runCommand(name, cmdArgs...); err != nil {
		fmt.Printf("Failed to terminate process %s: %v\n", pid, err)
		return
	}
	fmt.Printf("Process %s terminated.\n", pid)
}

// #region GetPIDByPort
// returns the PID of the process using the specified port.
func GetPIDByPort(port int) (string, error) {
	if runtime.GOOS == "windows" {
		out, err := runCommand("netstat", "-ano")
		if err != nil {
			return "", err
		}
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.Contains(line, fmt.Sprintf(":%d", port)) {
				fields := strings.Fields(line)
				if len(fields) > 4 {
					return fields[len(fields)-1], nil
				}
			}
		}
		return "", nil
	}

	out, err := runCommand("lsof", "-i", fmt.Sprintf(":%d", port), "-t")
	if err != nil || len(out) == 0 {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// #region GetProcessesByName
// Returns a list of processes matching the given name.
func GetProcessesByName(name string) ([]ProcessInfo, error) {
	var processes []ProcessInfo

	if runtime.GOOS == "windows" {
		out, err := runCommand("tasklist", "/FO", "LIST")
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(out), "\n")
		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "Image Name:") &&
				strings.Contains(strings.ToLower(line), strings.ToLower(name)) {
				image := strings.TrimSpace(strings.TrimPrefix(line, "Image Name:"))
				if i+1 < len(lines) {
					next := strings.TrimSpace(lines[i+1])
					if strings.HasPrefix(next, "PID:") {
						pid := strings.TrimSpace(strings.TrimPrefix(next, "PID:"))
						processes = append(processes, ProcessInfo{PID: pid, Name: image})
					}
				}
			}
		}
		return processes, nil
	}

	out, err := runCommand("pgrep", "-fl", name)
	if err != nil || len(out) == 0 {
		return nil, nil
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		processes = append(processes, ProcessInfo{
			PID:  fields[0],
			Name: strings.Join(fields[1:], " "),
		})
	}
	return processes, nil
}

// #region ShowProcessInfo
// displays detailed information about a process.
func ShowProcessInfo(pid string, osType string) {
	var cmdArgs []string
	var name string

	if osType == "windows" {
		name = "tasklist"
		cmdArgs = []string{"/FI", fmt.Sprintf("PID eq %s", pid), "/FO", "LIST"}
	} else {
		name = "ps"
		cmdArgs = []string{"-p", pid, "-o", "pid,ppid,user,%cpu,%mem,etime,cmd"}
	}

	_, _ = runCommand(name, cmdArgs...)
}

// #region runCommand
// helper function to execute a command and return output or error.
func runCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}
