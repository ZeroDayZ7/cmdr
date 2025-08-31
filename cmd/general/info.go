package general

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewInfoCmd tworzy podkomendę info
func NewInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show recommended Go project structure and common use cases",
		Long: `Displays an example folder structure for a professional Go application 
and a list of common use cases for the Go programming language.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(`
### Folder structure for a large Go application

Example of a commonly used layout in professional projects:

.
├── cmd/              # Entry points for the app (main.go for each service)
│   ├── api/
│   │   └── main.go
│   └── worker/
│       └── main.go
├── internal/         # Private application code (business logic)
│   ├── service/
│   ├── repository/
│   └── config/
├── pkg/              # Public libraries reusable by other projects
├── api/              # API definitions (OpenAPI/Swagger, gRPC proto files)
├── configs/          # Configuration files (YAML, JSON, TOML)
├── deployments/      # Deployment files (Docker, Kubernetes)
├── scripts/          # Helper scripts (e.g. database migrations)
├── test/             # Integration/e2e tests
├── migrations/       # Database migrations (SQL or tools like golang-migrate)
├── web/              # Frontend if bundled
├── build/            # Build outputs, CI/CD scripts
├── docs/             # Technical documentation
├── tools/            # Developer tools
├── assets/           # Static assets
└── vendor/           # Vendored dependencies (optional)

### Key principles:
* cmd/ contains only startup code.
* internal/ holds core business logic.
* pkg/ is public and reusable.
* configs/ and deployments/ are separated for clarity.
* Tests live either in test/ or next to code (*_test.go).

---

### What is Go commonly used for?

1. API and microservices
2. Network services (HTTP, TCP, gRPC, WebSocket)
3. Queue workers and asynchronous processing (RabbitMQ, Kafka, NATS)
4. CLI tools (e.g., Docker, Hugo, Terraform)
5. Distributed systems (Kubernetes, Etcd, Consul)
6. System integrations and middleware
7. Monitoring and observability tools (Prometheus exporters, Loki)
8. Rapid backend prototyping
9. Stream processing and IoT
10. DevOps automation and CI/CD tools
11. Database engines and management tools
12. Cloud-native apps (Kubernetes-ready)
13. Security tools (network scanners, threat detection)
14. Game backend services
15. IoT device management
16. Compilers and DSL interpreters
17. Automated testing frameworks
18. Blockchain nodes and wallets
19. Machine Learning deployment tools
20. Lightweight desktop applications (Fyne, Wails)

Go is valued for its simplicity, performance, and reliability — making it a strong choice for modern cloud and distributed environments.
			`)
		},
	}
}
