package main

import (
	"log"
	"os/exec"

	"github.com/zerodayz7/cmdr/cmd"
)

// #region init
func init() {
	if err := exec.Command("chcp", "65001").Run(); err != nil {
		log.Printf("Warning: failed to set UTF-8 code page: %v", err)
	}
}

// #region main
func main() {
	cmd.Execute()
}
