package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/config"
	"github.com/zerodayz7/cmdr/internal/updater"
)

func NewUpdateCmd() *cobra.Command {
	var checkOnly bool
	var force bool
	var targetVersion string
	var prerelease bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for or install cmdr updates",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			updater.PrintStage(cmd.OutOrStdout(), "Checking for a newer version")

			goos, goarch := updater.CurrentPlatform()
			if goos != "windows" && goos != "linux" {
				return fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
			}
			if goarch != "amd64" {
				return fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
			}

			var release *updater.Release
			var err error
			if strings.TrimSpace(targetVersion) != "" {
				version := strings.TrimSpace(targetVersion)
				_, prereleasePart, parseErr := updater.ParseVersion(version)
				if parseErr != nil {
					return parseErr
				}
				if prereleasePart != "" && !prerelease {
					return fmt.Errorf("pre-release versions require --prerelease")
				}
				version = strings.TrimPrefix(version, "v")
				tag := "v" + version
				if prerelease {
					release, err = updater.FetchReleaseByTagAny(tag)
				} else {
					release, err = updater.FetchReleaseByTag(tag)
				}
			} else if prerelease {
				release, err = updater.FetchLatestRelease(updater.ReleasesAPIURL, true)
			} else {
				release, err = updater.FetchLatestRelease(updater.LatestReleaseURL, false)
			}
			if err != nil {
				return err
			}

			latestTag := strings.TrimSpace(release.TagName)
			latestVersion, isPrerelease, err := updater.ReleaseVersionTag(latestTag)
			if err != nil {
				return err
			}
			if isPrerelease && !prerelease && strings.TrimSpace(targetVersion) == "" {
				return fmt.Errorf("latest release is a pre-release; use --prerelease")
			}
			if isPrerelease != release.Prerelease {
				return fmt.Errorf("release %s has inconsistent prerelease metadata", latestTag)
			}
			if release.Draft {
				return fmt.Errorf("release %s has inconsistent release metadata", latestTag)
			}
			comparison, err := updater.CompareVersions(latestVersion, config.Version)
			if err != nil {
				return err
			}

			cmd.Printf("Current version: %s\n", config.Version)
			cmd.Printf("Latest version:  %s\n", latestVersion)
			if comparison < 0 && !force {
				if strings.TrimSpace(targetVersion) != "" {
					return fmt.Errorf("downgrade from %s to %s requires --force", config.Version, latestVersion)
				}
				cmd.Println("cmdr is already up to date.")
				return nil
			}
			if comparison == 0 && !force {
				cmd.Println("cmdr is already up to date.")
				return nil
			}
			if checkOnly {
				cmd.Println("Update available! Run 'cmdr update' to install it.")
				return nil
			}

			asset, err := updater.FindAsset(release, goos, goarch, latestVersion)
			if err != nil {
				return err
			}
			var checksumAsset *updater.Asset
			for i := range release.Assets {
				if release.Assets[i].Name == "checksums.txt" {
					checksumAsset = &release.Assets[i]
					break
				}
			}
			if checksumAsset == nil {
				return fmt.Errorf("release %s does not contain checksums.txt", release.TagName)
			}

			working, err := updater.CreateUpdateTempDir()
			if err != nil {
				return err
			}
			defer updater.CleanUpdateTempDir(working)

			archivePath := filepath.Join(working, asset.Name)
			checksumPath := filepath.Join(working, "checksums.txt")
			updater.PrintStage(cmd.OutOrStdout(), "Downloading update")
			progress := updater.NewProgressRenderer(cmd.OutOrStdout())
			started := time.Now()
			var downloaded int64
			if err := updater.DownloadToFileWithProgress(asset.BrowserDownloadURL, archivePath, updater.MaxArchiveBytes, func(downloadedBytes, total int64, elapsed time.Duration) {
				downloaded = downloadedBytes
				progress.Render(updater.DownloadProgress{Downloaded: downloadedBytes, Total: total, Elapsed: elapsed})
			}); err != nil {
				progress.Abort(err.Error())
				return err
			}
			progress.Finish(downloaded, time.Since(started))
			if err := updater.DownloadToFile(checksumAsset.BrowserDownloadURL, checksumPath, 1<<20); err != nil {
				return err
			}

			updater.PrintStage(cmd.OutOrStdout(), "Verifying SHA-256 checksum")
			checksumFile, err := os.Open(checksumPath)
			if err != nil {
				return err
			}
			checksums, err := updater.ParseChecksums(checksumFile)
			checksumFile.Close()
			if err != nil {
				return err
			}
			expected, ok := checksums[asset.Name]
			if !ok {
				return fmt.Errorf("checksums.txt does not contain %s", asset.Name)
			}
			archiveFile, err := os.Open(archivePath)
			if err != nil {
				return err
			}
			if err := updater.VerifySHA256(archiveFile, expected); err != nil {
				archiveFile.Close()
				return err
			}
			archiveFile.Close()

			binaryPath, err := updater.ExtractArchive(archivePath, working, asset.Name, goos)
			if err != nil {
				return err
			}
			currentPath, err := updater.ExecutablePath()
			if err != nil {
				return err
			}
			stagedPath, err := updater.PrepareStagedBinary(binaryPath, filepath.Dir(currentPath))
			if err != nil {
				return err
			}

			updater.PrintStage(cmd.OutOrStdout(), "Installing update")
			if err := updater.ApplyPlatformUpdate(currentPath, stagedPath); err != nil {
				_ = os.Remove(stagedPath)
				return err
			}
			if goos == "windows" {
				cmd.Println("Update helper is replacing cmdr.exe after this process exits.")
				cmd.Println("If the update helper reports an error, run 'cmdr version' and inspect the update result.")
				return nil
			}
			updater.PrintCompleted(cmd.OutOrStdout(), latestVersion)
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "check for updates without installing")
	cmd.Flags().BoolVar(&force, "force", false, "allow reinstall or downgrade")
	cmd.Flags().StringVar(&targetVersion, "version", "", "install a specific stable version")
	cmd.Flags().BoolVar(&prerelease, "prerelease", false, "allow prerelease releases")
	return cmd
}
