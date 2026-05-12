package general

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/logger"
	"github.com/zerodayz7/cmdr/internal/profiles"
	"github.com/zerodayz7/cmdr/internal/tree"
	"github.com/zerodayz7/cmdr/internal/utils"
)

var (
	output          string
	exclude         []string
	format          string
	copyToClipboard bool
	log             = &logger.ConsoleLogger{}
)

func NewTreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tree [path]",
		Short: "Display a tree view of folders and files",
		RunE: func(cmd *cobra.Command, args []string) error {

			genIgnore, _ := cmd.Flags().GetBool("generate-ignore")
			if genIgnore {
				if err := tree.GenerateIgnoreFile(); err != nil {
					log.Error("Failed to generate ignore file: %v", err)
					return err
				}
				configDir, _ := profiles.GetConfigDir()
				log.Success("Generated %s in %s", tree.IgnoreFileName, configDir)
				return nil
			}

			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			defaultIgnore, _ := tree.ReadIgnoreFile()

			excludeList := append(defaultIgnore, utils.ParseCommaSeparated(strings.Join(exclude, ","))...)

			opts := tree.TreeOptions{
				Path:        path,
				ExcludeList: excludeList,
				Format:      format,
			}

			result, err := tree.Generate(opts)
			if err != nil {
				log.Error("Failed to generate tree: %v", err)
				return err
			}

			if output != "" {
				if err := os.WriteFile(output, []byte(result), 0644); err != nil {
					log.Error("Failed to save file: %v", err)
					return err
				}
				log.Success("Tree saved to %s", output)
			} else if !copyToClipboard {

				os.Stdout.WriteString(result + "\n")
			}

			if copyToClipboard {
				if err := utils.CopyToClipboard(result); err != nil {
					log.Error("Clipboard error: %v", err)
					return err
				}
				log.Info("Content copied to clipboard")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file to save the tree")
	cmd.Flags().StringSliceVarP(&exclude, "exclude", "x", []string{}, "Exclude patterns")
	cmd.Flags().BoolP("generate-ignore", "g", false, "Generate .cmdrignore in your config folder")
	cmd.Flags().StringVarP(&format, "format", "f", "ascii", "ascii|json|csv|md")
	cmd.Flags().BoolVarP(&copyToClipboard, "copy", "c", false, "Copy to clipboard")

	return cmd
}
