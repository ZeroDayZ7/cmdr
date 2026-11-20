package doctor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/zerodayz7/cmdr/config"
	"github.com/zerodayz7/cmdr/internal/profiles"
)

type Status string

const (
	OK   Status = "OK"
	WARN Status = "WARN"
	FAIL Status = "FAIL"
)

type Result struct {
	Section string `json:"section"`
	Name    string `json:"name"`
	Status  Status `json:"status"`
	Message string `json:"message"`
}

type Summary struct {
	Checks   int `json:"checks"`
	Passed   int `json:"passed"`
	Warnings int `json:"warnings"`
	Failed   int `json:"failed"`
}

type Report struct {
	Results []Result `json:"results"`
	Summary Summary  `json:"summary"`
}

type Options struct {
	HTTPClient *http.Client
	APIURL     string
}

func Run(options Options) Report {
	if options.HTTPClient == nil {
		options.HTTPClient = &http.Client{Timeout: 5 * time.Second}
	}
	if options.APIURL == "" {
		options.APIURL = "https://api.github.com"
	}

	results := make([]Result, 0, 10)
	results = append(results, environmentChecks()...)
	results = append(results, executableChecks()...)
	results = append(results, configChecks()...)
	results = append(results, networkCheck(options)...)
	results = append(results, updaterChecks()...)

	report := Report{Results: results}
	for _, result := range results {
		report.Summary.Checks++
		switch result.Status {
		case OK:
			report.Summary.Passed++
		case WARN:
			report.Summary.Warnings++
		case FAIL:
			report.Summary.Failed++
		}
	}
	return report
}

func environmentChecks() []Result {
	info := config.GetBuildInfo()
	return []Result{
		{Section: "Environment", Name: "Operating system", Status: OK, Message: info.Platform},
		{Section: "Environment", Name: "Application version", Status: OK, Message: info.Version},
	}
}

func executableChecks() []Result {
	path, err := os.Executable()
	if err != nil {
		return []Result{{Section: "Filesystem", Name: "Executable", Status: FAIL, Message: "Could not determine executable path"}}
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		path = resolved
	}
	results := []Result{{Section: "Filesystem", Name: "Executable path", Status: OK, Message: path}}
	info, err := os.Stat(path)
	if err != nil {
		results = append(results, Result{Section: "Filesystem", Name: "Executable file", Status: FAIL, Message: "Executable file is inaccessible"})
		return results
	}
	if info.IsDir() || !info.Mode().IsRegular() {
		results = append(results, Result{Section: "Filesystem", Name: "Executable file", Status: FAIL, Message: "Executable path is not a regular file"})
		return results
	}
	file, err := os.Open(path)
	if err != nil {
		results = append(results, Result{Section: "Filesystem", Name: "Executable file", Status: FAIL, Message: "Executable file cannot be read"})
		return results
	}
	_ = file.Close()
	results = append(results, Result{Section: "Filesystem", Name: "Executable file", Status: OK, Message: "Accessible"})
	return results
}

func configChecks() []Result {
	dir, err := profiles.GetConfigDir()
	if err != nil {
		return []Result{{Section: "Filesystem", Name: "Configuration directory", Status: FAIL, Message: "Could not determine configuration directory"}}
	}

	info, err := os.Stat(dir)
	if err == nil {
		if !info.IsDir() {
			return []Result{{Section: "Filesystem", Name: "Configuration directory", Status: FAIL, Message: dir + " is not a directory"}}
		}
		if !directoryWritable(dir) {
			return []Result{
				{Section: "Filesystem", Name: "Configuration directory", Status: OK, Message: dir},
				{Section: "Filesystem", Name: "Configuration directory write", Status: WARN, Message: "Directory is not currently writable"},
			}
		}
		return []Result{
			{Section: "Filesystem", Name: "Configuration directory", Status: OK, Message: dir},
			{Section: "Filesystem", Name: "Configuration directory write", Status: OK, Message: "Temporary write test passed"},
		}
	}
	if !os.IsNotExist(err) {
		return []Result{{Section: "Filesystem", Name: "Configuration directory", Status: FAIL, Message: "Configuration directory cannot be inspected"}}
	}
	parent := filepath.Dir(dir)
	if _, err := os.Stat(parent); err == nil {
		if directoryWritable(parent) {
			return []Result{{Section: "Filesystem", Name: "Configuration directory", Status: WARN, Message: dir + " does not exist; it can be created without administrator access"}}
		}
	}
	return []Result{{Section: "Filesystem", Name: "Configuration directory", Status: WARN, Message: dir + " does not exist and could not be created safely"}}
}

func directoryWritable(dir string) bool {
	probe, err := os.CreateTemp(dir, ".cmdr-doctor-*")
	if err != nil {
		return false
	}
	name := probe.Name()
	_ = probe.Close()
	_ = os.Remove(name)
	return true
}

func networkCheck(options Options) []Result {
	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(options.APIURL, "/")+"/", nil)
	if err != nil {
		return []Result{{Section: "Network", Name: "GitHub API", Status: FAIL, Message: "Could not prepare API request"}}
	}
	req.Header.Set("User-Agent", "cmdr-doctor")
	resp, err := options.HTTPClient.Do(req)
	if err != nil {
		status := WARN
		message := "GitHub API is unreachable"
		if os.IsTimeout(err) {
			message = "GitHub API request timed out"
		}
		return []Result{{Section: "Network", Name: "GitHub API", Status: status, Message: message}}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return []Result{{Section: "Network", Name: "GitHub API", Status: OK, Message: "Reachable"}}
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		return []Result{{Section: "Network", Name: "GitHub API", Status: WARN, Message: fmt.Sprintf("GitHub API returned HTTP %d", resp.StatusCode)}}
	}
	if resp.StatusCode >= 500 {
		return []Result{{Section: "Network", Name: "GitHub API", Status: WARN, Message: fmt.Sprintf("GitHub API returned HTTP %d", resp.StatusCode)}}
	}
	return []Result{{Section: "Network", Name: "GitHub API", Status: FAIL, Message: fmt.Sprintf("GitHub API returned HTTP %d", resp.StatusCode)}}
}

func updaterChecks() []Result {
	goos, goarch := runtime.GOOS, runtime.GOARCH
	results := make([]Result, 0, 3)
	if (goos == "windows" || goos == "linux") && goarch == "amd64" {
		results = append(results, Result{Section: "Updater", Name: "Platform support", Status: OK, Message: goos + "/" + goarch})
	} else {
		results = append(results, Result{Section: "Updater", Name: "Platform support", Status: FAIL, Message: "Unsupported platform: " + goos + "/" + goarch})
		return results
	}

	path, err := os.Executable()
	if err == nil {
		resolved, err := filepath.EvalSymlinks(path)
		if err == nil {
			path = resolved
		}
		dir := filepath.Dir(path)
		if directoryWritable(dir) {
			results = append(results, Result{Section: "Updater", Name: "Installation directory", Status: OK, Message: "Writable: " + dir})
		} else {
			results = append(results, Result{Section: "Updater", Name: "Installation directory", Status: WARN, Message: "Updating may require elevated permissions: " + dir})
		}
	} else {
		results = append(results, Result{Section: "Updater", Name: "Installation directory", Status: FAIL, Message: "Could not determine installation directory"})
	}
	return results
}

func WriteJSON(w io.Writer, report Report) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
