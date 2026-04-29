package main

import (
	"os/exec"

	"github.com/zerodayz7/cmdr/cmd"
)

func init() {
	exec.Command("chcp", "65001").Run()
}

func main() {
	cmd.Execute()
}
