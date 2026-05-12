package annotate

import "github.com/zerodayz7/cmdr/internal/profiles"

type Config struct {
	DryRun         bool
	Verbose        bool
	Profile        *profiles.Profile
	ProfilesConfig *profiles.Config
}
