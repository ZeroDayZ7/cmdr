package info_go

import (
	"fmt"
)

func runTree() {
	fmt.Println(`
.
├── cmd/              # Entry points for the app (main.go for each service)
│   ├── api/          # HTTP API service
│   │   └── main.go
│   └── worker/       # Background worker service
│       └── main.go
├── internal/         # Private application code (business logic)
│   ├── service/      # Service layer
│   ├── repository/   # Database access / repositories
│   └── config/       # App configuration files
├── pkg/              # Public reusable libraries
├── api/              # API definitions (OpenAPI/Swagger, gRPC proto)
├── configs/          # Configuration files (YAML, JSON, TOML)
├── deployments/      # Deployment files (Docker, Kubernetes)
├── scripts/          # Helper scripts (DB migrations, setup, CI)
├── test/             # Integration / e2e tests
├── migrations/       # Database migration scripts
├── web/              # Frontend code (if bundled)
├── build/            # Build outputs, CI/CD scripts
├── docs/             # Technical documentation
├── tools/            # Developer tools
├── assets/           # Static assets (images, fonts, etc.)
└── vendor/           # Vendored dependencies (optional)
	`)
}
