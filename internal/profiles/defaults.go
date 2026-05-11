package profiles

import (
	"os"
	"path/filepath"
)

func GetConfigPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return ".cmdr/profiles.json"
	}
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, ".cmdr_profiles.json")
}

const DefaultProfilesConfig = `{
  "global": {
    "ignore": [".git", "node_modules", "dist", "build", ".next", "vendor", "bin"]
  },
  "profiles": [
    {
      "name": "flutter",
      "priority": 100,
      "detect": {
        "files": ["pubspec.yaml"],
        "folders": ["lib", "android", "ios"],
        "extensions": [".dart"]
      },
      "extensions": [".dart", ".yaml"],
      "ignore": [".dart_tool"],
      "generated": [".g.dart", ".freezed.dart"]
    },
    {
      "name": "go",
      "priority": 90,
      "detect": {
        "files": ["go.mod"],
        "folders": ["cmd", "internal"],
        "extensions": [".go"]
      },
      "extensions": [".go", ".mod", ".sum"],
      "ignore": ["vendor"],
      "generated": []
    }
  ]
}`
