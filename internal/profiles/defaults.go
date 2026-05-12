package profiles

const DefaultProfilesConfig = `{
  "global": {
    "ignore": [
      ".git",
      "node_modules",
      "dist",
      "build",
      ".next",
      "vendor",
      "bin"
    ]
  },

  "comment_styles": {
    ".go": "// cmdr: %s",
    ".js": "// cmdr: %s",
    ".jsx": "// cmdr: %s",
    ".ts": "// cmdr: %s",
    ".tsx": "// cmdr: %s",
    ".py": "# cmdr: %s",
    ".yaml": "# cmdr: %s",
    ".yml": "# cmdr: %s",
    ".sh": "# cmdr: %s",
    ".sql": "-- cmdr: %s",
    ".css": "/* cmdr: %s */",
    ".scss": "/* cmdr: %s */",
    ".html": "<!-- cmdr: %s -->",
    ".md": "<!-- cmdr: %s -->"
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

      "extensions": [
        ".dart",
        ".yaml"
      ],

      "ignore": [
        ".dart_tool"
      ],

      "generated": [
        ".g.dart",
        ".freezed.dart"
      ]
    },

    {
      "name": "go",
      "priority": 90,

      "detect": {
        "files": ["go.mod"],
        "folders": ["cmd", "internal"],
        "extensions": [".go"]
      },

      "extensions": [
        ".go",
        ".mod",
        ".sum"
      ],

      "ignore": [
        "vendor"
      ],

      "generated": []
    }
  ]
}`
