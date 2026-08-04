package profiles

const DefaultProfilesConfig = `{
 "global": {
    "ignore": [".git", "node_modules", "dist", "build", ".next", "vendor", "bin"],
    "ignored_extensions": [
      ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".ico",
      ".exe", ".dll", ".so", ".bin",
      ".zip", ".rar", ".7z", ".tar", ".gz",
      ".pdf", ".mp3", ".mp4", ".mov", ".lock"
    ]
  },

  "comment_styles": {
    ".go": "// %s",
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
      "region_patterns": [
        {
          "regex": "^(?:abstract\\s+)?class\\s+([a-zA-Z0-9_]+)",
          "style": "region"
        },
        {
          "regex": "^\\s+(?:async\\s+)?(?:[a-zA-Z0-9_<>]+\\s+)?([a-zA-Z0-9_]+)\\s*\\([^\\)]*\\)\\s*(?:async\\s*)?{",
          "style": "region"
        },
        {
          "regex": "^(?:mixin|extension)\\s+([a-zA-Z0-9_]+)",
          "style": "region"
        }
      ],
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
        ".dart_tool",
        "build",
        "ios",
        "android"
      ],
      "generated": [
        ".g.dart",
        ".freezed.dart"
      ]
    },
    {
      "name": "typescript",
      "priority": 80,
      "region_patterns": [
        {
          "regex": "^(?:export\\s+)?(?:async\\s+)?function\\s+([a-zA-Z0-9_]+)\\s*\\(",
          "style": "region"
        },
        {
          "regex": "^(?:export\\s+)?(?:async\\s+)?const\\s+([a-zA-Z0-9_]+)\\s*=\\s*(?:async\\s*)?\\([^\\)]*\\)\\s*=>",
          "style": "region"
        },
        {
          "regex": "^(?:export\\s+)?(?:abstract\\s+)?class\\s+([a-zA-Z0-9_]+)",
          "style": "region"
        },
        {
          "regex": "^(?:export\\s+)?interface\\s+([a-zA-Z0-9_]+)",
          "style": "region"
        }
      ],
      "detect": {
        "files": ["tsconfig.json", "package.json"],
        "folders": ["src", "app", "pages"],
        "extensions": [".ts", ".tsx"]
      },
      "extensions": [
        ".ts",
        ".tsx",
        ".js",
        ".jsx",
        ".json"
      ],
      "ignore": [
        "node_modules",
        ".next",
        "dist"
      ],
      "generated": [
        "*.d.ts",
        ".next"
      ]
    },
    {
      "name": "go",
      "region_patterns": [
        {
          "regex": "^func\\s+(?:\\([^\\)]+\\)\\s+)?([a-zA-Z0-9_]+)\\s*\\(",
          "style": "region"
        }
      ],
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
