// cmdr: cmdr/internal/annotate_region/config.go

package annotate_region

import "context"

type Config struct {
	DryRun  bool
	Verbose bool
	Context context.Context
}
