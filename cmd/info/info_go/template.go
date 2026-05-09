package info_go

import (
	"fmt"
	"os"
	"path/filepath"
)

func runTemplate(projectName, targetDir string) {
	if projectName == "" {
		fmt.Print("Enter project name: ")

		if _, err := fmt.Scanln(&projectName); err != nil {
			fmt.Printf("❌ Error reading project name: %v\n", err)
			return
		}

		if projectName == "" {
			fmt.Println("❌ Project name cannot be empty")
			return
		}
	}

	if targetDir == "" {
		targetDir = "."
	}

	projectPath := filepath.Join(targetDir, projectName)

	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		fmt.Printf("⚠️ Directory already exists: %s\n", projectPath)
		return
	}

	folders := []string{
		"cmd/api", "cmd/worker", "internal/service", "internal/repository",
		"internal/config", "pkg", "api", "configs", "deployments",
		"scripts", "test", "migrations", "web", "bin", "docs",
		"tools", "assets", "vendor",
	}

	for _, f := range folders {
		path := filepath.Join(projectPath, f)
		if err := os.MkdirAll(path, 0755); err != nil {
			fmt.Printf("❌ Failed to create folder %s: %v\n", path, err)
			return
		}
	}

	mainFiles := map[string]string{
		"cmd/api/main.go":    "package main\n\nfunc main() {\n\tprintln(\"API service starting...\")\n}\n",
		"cmd/worker/main.go": "package main\n\nfunc main() {\n\tprintln(\"Worker service starting...\")\n}\n",
	}

	for file, content := range mainFiles {
		fullPath := filepath.Join(projectPath, file)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			fmt.Printf("❌ Failed to create file %s: %v\n", fullPath, err)
			return
		}
	}

	fmt.Printf("🚀 Go project template created successfully at: %s\n", projectPath)
}
