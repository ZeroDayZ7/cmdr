package annotate

import "github.com/zerodayz7/cmdr/internal/profiles"

type Logger interface {
	Success(msg string, args ...any)
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

type Config struct {
	DryRun         bool
	Verbose        bool
	Profile        *profiles.Profile
	ProfilesConfig *profiles.Config
	Log            Logger
}
