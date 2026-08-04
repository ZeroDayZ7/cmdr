package annotate_region

import "context"

type Config struct {
	DryRun  bool
	Verbose bool
	Context context.Context
}
