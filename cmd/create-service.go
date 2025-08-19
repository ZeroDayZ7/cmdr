package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var serviceName string

var createCmd = &cobra.Command{
	Use:     "create-service",
	Aliases: []string{"cs"},
	Short:   "Create a new microservice from template",
	Long:    "Generates a new microservice folder based on a predefined template",
	Run: func(cmd *cobra.Command, args []string) {
		if serviceName == "" {
			fmt.Print("Enter the name of the microservice: ")
			fmt.Scanln(&serviceName)
		}

		templateDir := filepath.Join(".", "templates", "microservice-template")
		targetDir := filepath.Join(".", serviceName)

		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			fmt.Println("Template folder not found:", templateDir)
			return
		}

		if _, err := os.Stat(targetDir); err == nil {
			fmt.Println("Folder already exists:", targetDir)
			return
		}

		err := copyDir(templateDir, targetDir)
		if err != nil {
			fmt.Println("Error copying template:", err)
			return
		}

		fmt.Println("Microservice created at:", targetDir)
	},
}

func init() {
	createCmd.Flags().StringVarP(&serviceName, "name", "n", "", "Name of the microservice")
	rootCmd.AddCommand(createCmd)
}

// ==========================
// Recursive folder copy
// ==========================
func copyDir(src string, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		targetPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		return err
	})
}
