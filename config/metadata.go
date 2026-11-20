package config

import "runtime"

var (
	Name      = "cmdr"
	Version   = "2.1.4"
	GitCommit = "unknown"
	BuildDate = "unknown"
	Contact   = "ZeroDayZ7@proton.me"
)

type BuildInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	Platform  string `json:"platform"`
	GoVersion string `json:"go_version"`
}

func GetBuildInfo() BuildInfo {
	commit := unknownIfEmpty(GitCommit)
	if len(commit) > 12 {
		commit = commit[:12]
	}
	return BuildInfo{
		Name:      Name,
		Version:   unknownIfEmpty(Version),
		GitCommit: commit,
		BuildDate: unknownIfEmpty(BuildDate),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
		GoVersion: runtime.Version(),
	}
}

func unknownIfEmpty(value string) string {
	if value == "" || value == "unknown" {
		return "unknown"
	}
	return value
}
