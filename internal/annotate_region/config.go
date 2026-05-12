package annotate

type Config struct {
	DryRun  bool
	Verbose bool
}

// IgnoredExtensions - pliki, których nigdy nie parsujemy w poszukiwaniu funkcji
var IgnoredExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".webp": true, ".svg": true, ".ico": true,
	".exe": true, ".dll": true, ".so": true, ".bin": true,
	".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true,
	".pdf": true, ".mp3": true, ".mp4": true, ".mov": true,
	".lock": true,
}

// DefaultIgnoredDirs - standardowe katalogi do pominięcia
var DefaultIgnoredDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"dist":         true,
	"build":        true,
	".next":        true,
	"vendor":       true,
	"bin":          true,
}
