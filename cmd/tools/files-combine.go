package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/profiles"
	"github.com/zerodayz7/cmdr/internal/utils"
)

var (
	outputFile  string
	extensions  string
	excludeDirs string
)

// #region NewFilesCombineCmd
func NewFilesCombineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "files-combine",
		Aliases: []string{"fc", "combine-files"},
		Short:   "Combine contents of files based on detected project profile",
		RunE:    runFilesCombine,
	}

	cmd.Flags().StringVarP(&extensions, "extensions", "e", "", "Override: Comma-separated extensions (e.g. go,proto)")
	cmd.Flags().StringVarP(&outputFile, "name", "n", "combined.txt", "Name of the output file")
	cmd.Flags().StringVarP(&excludeDirs, "exclude", "x", "", "Override: Additional directories to exclude")

	return cmd
}

// #region runFilesCombine
func runFilesCombine(cmd *cobra.Command, args []string) error {

	cfg, err := profiles.LoadConfig()
	if err != nil {
		return err
	}

	currDir, _ := os.Getwd()
	activeProfile := profiles.DetectProjectProfile(currDir, cfg)

	finalExtensions := prepareExtensions(activeProfile)
	ignoredItems := prepareIgnored(cfg, activeProfile)

	if activeProfile != nil {
		fmt.Printf("Detected profile: %s\n", activeProfile.Name)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		name := info.Name()

		if info.IsDir() {
			if isIgnored(name, ignoredItems) {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !isAllowed(ext, name, finalExtensions, activeProfile) {
			return nil
		}

		return writeFileContent(writer, path, name)
	})

	if err == nil {
		fmt.Printf("✅ Finished! Combined into: %s\n", outputFile)
	}
	return err
}

// #region prepareExtensions
func prepareExtensions(p *profiles.Profile) []string {
	if extensions != "" {
		return utils.ParseCommaSeparated(extensions)
	}
	if p != nil {
		return p.Extensions
	}
	return []string{".ts", ".tsx", ".jsx", ".json"}
}

// #region prepareIgnored
func prepareIgnored(cfg *profiles.Config, p *profiles.Profile) []string {
	list := cfg.Global.Ignore
	if p != nil {
		list = append(list, p.Ignore...)
	}
	if excludeDirs != "" {
		list = append(list, utils.ParseCommaSeparated(excludeDirs)...)
	}
	return list
}

// #region isIgnored
func isIgnored(name string, ignoredList []string) bool {
	return slices.Contains(ignoredList, name)
}

// #region isAllowed
func isAllowed(ext, name string, allowedExts []string, p *profiles.Profile) bool {

	found := false
	for _, e := range allowedExts {
		if strings.TrimPrefix(e, ".") == strings.TrimPrefix(ext, ".") {
			found = true
			break
		}
	}
	if !found {
		return false
	}

	if p != nil && len(p.Generated) > 0 {
		for _, gen := range p.Generated {
			if strings.HasSuffix(name, gen) {
				return false
			}
		}
	}

	return true
}

// #region writeFileContent
func writeFileContent(w *bufio.Writer, path, name string) error {
	content, err := os.ReadFile(path)
	if err != nil {

		fmt.Fprintf(os.Stderr, "Warning: could not read file %s: %v\n", path, err)
		return nil
	}

	if _, err := fmt.Fprintf(w, "───────────── %s ─────────────\n", name); err != nil {
		return fmt.Errorf("failed to write header for %s: %w", name, err)
	}

	if _, err := w.Write(content); err != nil {
		return fmt.Errorf("failed to write content of %s: %w", name, err)
	}

	if _, err := w.WriteString("\n\n"); err != nil {
		return fmt.Errorf("failed to write separator after %s: %w", name, err)
	}

	return nil
}
