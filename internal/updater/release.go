package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const (
	LatestReleaseURL = "https://api.github.com/repos/ZeroDayZ7/cmdr/releases/latest"
	ReleasesAPIURL   = "https://api.github.com/repos/ZeroDayZ7/cmdr/releases"
	maxResponseBytes = 4 << 20
)

var httpClient = &http.Client{Timeout: 20 * time.Second}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type Release struct {
	TagName    string  `json:"tag_name"`
	Draft      bool    `json:"draft"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
}

func PlatformAssetName(goos, goarch, version string) (string, error) {
	if goos == "windows" && goarch == "amd64" {
		return fmt.Sprintf("cmdr_%s_windows_amd64.zip", version), nil
	}
	if goos == "linux" && goarch == "amd64" {
		return fmt.Sprintf("cmdr_%s_linux_amd64.tar.gz", version), nil
	}
	return "", fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
}

func FetchLatestRelease(baseURL string, allowPrerelease bool) (*Release, error) {
	url := baseURL
	if allowPrerelease {
		url = ReleasesAPIURL
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "cmdr-updater")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub Releases request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub Releases returned HTTP %d", resp.StatusCode)
	}
	limit := io.LimitReader(resp.Body, maxResponseBytes+1)
	data, err := io.ReadAll(limit)
	if err != nil {
		return nil, fmt.Errorf("read GitHub Releases response: %w", err)
	}
	if len(data) > maxResponseBytes {
		return nil, fmt.Errorf("GitHub Releases response exceeds size limit")
	}

	if allowPrerelease {
		var releases []Release
		if err := json.Unmarshal(data, &releases); err != nil {
			return nil, fmt.Errorf("decode GitHub Releases response: %w", err)
		}
		var selected *Release
		for i := range releases {
			candidate := &releases[i]
			if candidate.Draft {
				continue
			}
			if candidate.TagName == "" {
				continue
			}
			if selected == nil {
				selected = candidate
				continue
			}
			a, err := CompareVersions(strings.TrimPrefix(candidate.TagName, "v"), strings.TrimPrefix(selected.TagName, "v"))
			if err == nil && a > 0 {
				selected = candidate
			}
		}
		if selected == nil {
			return nil, fmt.Errorf("no GitHub releases found")
		}
		return selected, nil
	}

	var release Release
	if err := json.Unmarshal(data, &release); err != nil {
		return nil, fmt.Errorf("decode GitHub release response: %w", err)
	}
	if release.Draft {
		return nil, fmt.Errorf("latest GitHub release is a draft")
	}
	if release.Prerelease {
		return nil, fmt.Errorf("latest GitHub release is a prerelease")
	}
	return &release, nil
}

func FindAsset(release *Release, goos, goarch, version string) (*Asset, error) {
	name, err := PlatformAssetName(goos, goarch, version)
	if err != nil {
		return nil, err
	}
	for i := range release.Assets {
		if release.Assets[i].Name == name {
			return &release.Assets[i], nil
		}
	}
	return nil, fmt.Errorf("release %s does not contain artifact %s", release.TagName, name)
}

func CurrentPlatform() (string, string) {
	return runtime.GOOS, runtime.GOARCH
}

func StableVersionTag(raw string) (string, error) {
	text := strings.TrimSpace(raw)
	if strings.HasPrefix(text, "v.") {
		return "", fmt.Errorf("invalid release tag %q", raw)
	}
	if !strings.HasPrefix(text, "v") {
		return "", fmt.Errorf("invalid release tag %q", raw)
	}
	core := strings.TrimPrefix(text, "v")
	_, prerelease, err := ParseVersion(core)
	if err != nil || prerelease != "" {
		return "", fmt.Errorf("invalid release tag %q", raw)
	}
	return core, nil
}

func FetchReleaseByTag(tag string) (*Release, error) {
	url := strings.TrimRight(ReleasesAPIURL, "/") + "/tags/" + tag
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "cmdr-updater")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub Release request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release %s was not found", tag)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub Release returned HTTP %d", resp.StatusCode)
	}
	limit := io.LimitReader(resp.Body, maxResponseBytes+1)
	data, err := io.ReadAll(limit)
	if err != nil {
		return nil, err
	}
	if len(data) > maxResponseBytes {
		return nil, fmt.Errorf("GitHub Release response exceeds size limit")
	}
	var release Release
	if err := json.Unmarshal(data, &release); err != nil {
		return nil, fmt.Errorf("decode GitHub release response: %w", err)
	}
	if release.Draft {
		return nil, fmt.Errorf("release %s is a draft", tag)
	}
	if release.Prerelease {
		return nil, fmt.Errorf("release %s is a prerelease", tag)
	}
	return &release, nil
}
