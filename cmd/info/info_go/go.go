package info_go

import (
	"github.com/spf13/cobra"
)

func NewGoCmd() *cobra.Command {
	var showTree bool
	var showDesc bool
	var showTemplate bool
	var projectName string
	var targetDir string

	cmd := &cobra.Command{
		Use:   "go",
		Short: "Show Go project structure and examples",
		Long: `Show Go project structure, use cases, and generate templates.

Flags:
  -t, --tree         Show recommended Go project folder structure
  -d, --description  Show what Go is commonly used for
  -p, --template     Generate Go project template
  -n, --name         Name of the Go project (required for template)
  -D, --dir          Target directory for the project (default is current directory)

Examples:
  # Show folder structure
  cmdr info go -t

  # Show common Go use cases
  cmdr info go -d

	# Generate a Go project interactively
  cmdr info go -p
  Enter project name:

  # Generate a Go project template in current folder
  cmdr info go -p -n myservice

  # Generate a Go project template in a specific directory
  cmdr info go -p -n myservice -D ./projects
`,
		Run: func(cmd *cobra.Command, args []string) {
			if showTree {
				runTree()
				return
			}
			if showDesc {
				runDescription()
				return
			}
			if showTemplate {
				runTemplate(projectName, targetDir)
				return
			}
			cmd.Help()
		},
	}

	cmd.Flags().BoolVarP(&showTree, "tree", "t", false, "Show recommended Go project folder structure")
	cmd.Flags().BoolVarP(&showDesc, "description", "d", false, "Show what Go is commonly used for")
	cmd.Flags().BoolVarP(&showTemplate, "template", "p", false, "Generate Go project template")
	cmd.Flags().StringVarP(&projectName, "name", "n", "", "Name of the Go project (for template)")
	cmd.Flags().StringVarP(&targetDir, "dir", "D", ".", "Target directory for the project (for template)")

	return cmd
}
