package tools

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var serviceName string

// NewCreateServiceCmd generates a new microservice folder from a predefined template.
func NewCreateServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-service",
		Aliases: []string{"cs"},
		Short:   "Create a new microservice from template",
		Long:    "Generates a new microservice folder based on a predefined template",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serviceName == "" {
				fmt.Print("Enter the name of the microservice: ")
				if _, err := fmt.Scanln(&serviceName); err != nil {
					return fmt.Errorf("failed to read service name: %w", err)
				}
			}

			templateDir := filepath.Join(".", "templates", "microservice-template")
			targetDir := filepath.Join(".", serviceName)

			if _, err := os.Stat(templateDir); os.IsNotExist(err) {
				return fmt.Errorf("template folder not found: %s", templateDir)
			}

			if _, err := os.Stat(targetDir); err == nil {
				return fmt.Errorf("folder already exists: %s", targetDir)
			}

			if err := copyDir(templateDir, targetDir); err != nil {
				return fmt.Errorf("error copying template: %w", err)
			}

			fmt.Printf("Microservice created successfully at: %s\n", targetDir)
			return nil
		},
	}

	cmd.Flags().StringVarP(&serviceName, "name", "n", "", "Name of the microservice")
	return cmd
}

// copyDir recursively copies a directory from src to dest.
func copyDir(src string, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		targetPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath, info.Mode())
	})
}

// copyFile handles the low-level file copying logic.
func copyFile(src, dest string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err = io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	return nil
}
