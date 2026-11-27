# CMDR developer, validation, build, packaging and release interface.
# GNU Make is required. On Windows, run this through Git Bash/MSYS or cmd/pwsh
# after GNU Make and Go are available in PATH.

ROOT := $(CURDIR)
GO ?= go
GIT ?= git

ifeq ($(OS),Windows_NT)
HOST_OS := windows
else
UNAME_S := $(shell uname -s 2>/dev/null | tr '[:upper:]' '[:lower:]')
ifneq (,$(findstring mingw,$(UNAME_S)))
HOST_OS := windows
else
ifneq (,$(findstring msys,$(UNAME_S)))
HOST_OS := windows
else
ifneq (,$(findstring cygwin,$(UNAME_S)))
HOST_OS := windows
else
ifeq ($(UNAME_S),darwin)
HOST_OS := darwin
else
HOST_OS := linux
endif
endif
endif
endif
endif

HOST_ARCH := $(shell uname -m 2>/dev/null | tr '[:upper:]' '[:lower:]')
ifeq ($(HOST_ARCH),x86_64)
HOST_ARCH := amd64
else
ifeq ($(HOST_ARCH),amd64)
HOST_ARCH := amd64
else
ifeq ($(HOST_ARCH),aarch64)
HOST_ARCH := arm64
else
ifeq ($(HOST_ARCH),arm64)
HOST_ARCH := arm64
else
HOST_ARCH := $(shell go env GOARCH 2>/dev/null)
endif
endif
endif
endif

VERSION ?= $(shell sed -n 's/^[[:space:]]*Version[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' config/metadata.go 2>/dev/null | head -n 1)
COMMIT ?= $(shell $(GIT) rev-parse HEAD 2>/dev/null || echo unknown)
ifeq ($(OS),Windows_NT)
BUILD_DATE ?= $(shell powershell -NoProfile -Command "(Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')" 2>/dev/null)
else
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null)
endif

TAG ?= v$(VERSION)
LOCAL_OUT_DIR ?= bin/local/$(HOST_OS)/$(HOST_ARCH)
ifeq ($(HOST_OS),windows)
LOCAL_BINARY := $(LOCAL_OUT_DIR)/cmdr.exe
else
LOCAL_BINARY := $(LOCAL_OUT_DIR)/cmdr
endif
DIST_DIR ?= dist

LDFLAGS := -s -w -X github.com/zerodayz7/cmdr/config.Version=$(VERSION) -X github.com/zerodayz7/cmdr/config.GitCommit=$(COMMIT) -X github.com/zerodayz7/cmdr/config.BuildDate=$(BUILD_DATE)

.PHONY: help version fmt fmt-check vet lint test test-race check build build-local install release-build checksums package release-check release clean

help:
	@printf '%s\n' \
	  'CMDR Makefile' \
	  '' \
	  'Project and help:' \
	  '  make help          Show this target list' \
	  '  make version       Show project version and build metadata' \
	  '' \
	  'Development and validation:' \
	  '  make fmt           Format Go source files' \
	  '  make fmt-check     Verify Go formatting without changing files' \
	  '  make vet           Run go vet ./...' \
	  '  make lint          Run golangci-lint ./...' \
	  '  make test          Run the full Go test suite' \
	  '  make test-race     Run the full Go test suite with race detection' \
	  '  make check         Run fmt-check, vet, test, and lint' \
	  '' \
	  'Local builds:' \
	  '  make build         Build CMDR for the current host into bin/local/<os>/<arch>/' \
	  '  make build-local   Alias for make build' \
	  '  make install       Install CMDR with injected build metadata' \
	  '' \
	  'Release preparation:' \
	  '  make release-build Build supported native release binaries' \
	  '  make checksums     Generate SHA-256 checksums for dist/*.zip and *.tar.gz' \
	  '  make package       Package available native release binaries' \
	  '  make release-check Validate the configured tag and repository state' \
	  '  make release CONFIRM=yes  Create and push the version tag (GitHub Actions publishes)' \
	  '' \
	  'Cleanup:' \
	  '  make clean         Remove CMDR build, package, and release artifacts only' \
	  '' \
	  'Examples:' \
	  '  make build OUT_DIR=bin/local/windows/amd64' \
	  '  make release-build VERSION=2.1.7-rc.3' \
	  '  make release-check TAG=v2.1.7-rc.3' \
	  '  make release TAG=v2.1.7-rc.3 CONFIRM=yes'

version:
	@printf 'Version: %s\nCommit: %s\nBuild date: %s\nHost: %s/%s\n' "$(VERSION)" "$(COMMIT)" "$(BUILD_DATE)" "$(HOST_OS)" "$(HOST_ARCH)"

fmt:
	@command -v $(GO) >/dev/null 2>&1 || { echo 'Error: Go is not available in PATH.' >&2; exit 1; }
	gofmt -w .

fmt-check:
	@command -v gofmt >/dev/null 2>&1 || { echo 'Error: gofmt is not available in PATH.' >&2; exit 1; }
	@files="$$(gofmt -l .)"; if [ -n "$$files" ]; then printf 'Unformatted Go files:\n%s\n' "$$files" >&2; exit 1; fi
	@echo 'Go formatting OK'

vet:
	@command -v $(GO) >/dev/null 2>&1 || { echo 'Error: Go is not available in PATH.' >&2; exit 1; }
	$(GO) vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo 'Error: golangci-lint is not available in PATH. Install it only if linting is required.' >&2; exit 1; }
	golangci-lint run ./...

test:
	@command -v $(GO) >/dev/null 2>&1 || { echo 'Error: Go is not available in PATH.' >&2; exit 1; }
	$(GO) test ./...

test-race:
	@command -v $(GO) >/dev/null 2>&1 || { echo 'Error: Go is not available in PATH.' >&2; exit 1; }
	$(GO) test -race ./...

check: fmt-check vet test lint

build build-local:
	@command -v $(GO) >/dev/null 2>&1 || { echo 'Error: Go is not available in PATH.' >&2; exit 1; }
	@if [ "$(HOST_OS)" = "windows" ] || [ "$(HOST_OS)" = "linux" ] || [ "$(HOST_OS)" = "darwin" ]; then :; else echo 'Error: unsupported local host OS: $(HOST_OS).' >&2; exit 1; fi
	@if [ "$(HOST_ARCH)" != "amd64" ] && [ "$(HOST_ARCH)" != "arm64" ]; then echo 'Error: unsupported local host architecture: $(HOST_ARCH).' >&2; exit 1; fi
	@if [ "$(HOST_OS)" = "windows" ] || [ "$(HOST_OS)" = "linux" ]; then command -v gcc >/dev/null 2>&1 || { echo 'Error: CGO is required and gcc is not available in PATH.' >&2; exit 1; }; fi
	@mkdir -p "$(LOCAL_OUT_DIR)"
	CGO_ENABLED=1 GOOS=$(HOST_OS) GOARCH=$(HOST_ARCH) $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o "$(LOCAL_BINARY)" .
	@test -s "$(LOCAL_BINARY)"
	@echo "SUCCESS: $(LOCAL_BINARY)"

install:
	@command -v $(GO) >/dev/null 2>&1 || { echo 'Error: Go is not available in PATH.' >&2; exit 1; }
	$(GO) install -trimpath -ldflags "$(LDFLAGS)" .

release-build:
	@command -v $(GIT) >/dev/null 2>&1 || { echo 'Error: Git is not available in PATH.' >&2; exit 1; }
	@command -v $(GO) >/dev/null 2>&1 || { echo 'Error: Go is not available in PATH.' >&2; exit 1; }
	@if [ -z "$(VERSION)" ]; then echo 'Error: could not read config.Version from config/metadata.go; set VERSION=... .' >&2; exit 1; fi
	@if [ "$(HOST_OS)" = "windows" ]; then \
	  command -v powershell >/dev/null 2>&1 || { echo 'Error: PowerShell is required for Windows release builds.' >&2; exit 1; }; \
	  powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/build-windows.ps1 -Arch amd64 -Version "$(VERSION)"; \
	elif [ "$(HOST_OS)" = "linux" ]; then \
	  command -v bash >/dev/null 2>&1 || { echo 'Error: Bash is required for Linux release builds.' >&2; exit 1; }; \
	  bash ./scripts/build-linux.sh amd64 "$(VERSION)"; \
	else \
	  echo 'Error: native release builds are supported only on Windows amd64 and Linux amd64; macOS is not in the existing release matrix.' >&2; exit 1; \
	fi

package:
	@mkdir -p "$(DIST_DIR)"
	@if [ "$(HOST_OS)" = "windows" ]; then \
	  if [ -f bin/windows/amd64/cmdr.exe ]; then \
	    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/package-release.ps1 -DistDir "$(DIST_DIR)" -Version "$(VERSION)" -Root "$(ROOT)"; \
	  fi; \
	elif [ "$(HOST_OS)" = "linux" ]; then \
	  if [ -f bin/windows/amd64/cmdr.exe ]; then \
	    mkdir -p "$(DIST_DIR)/.package/windows"; \
	    cp bin/windows/amd64/cmdr.exe "$(DIST_DIR)/.package/windows/cmdr.exe"; \
	    cp README.md LICENSE "$(DIST_DIR)/.package/windows/"; \
	    if command -v zip >/dev/null 2>&1; then \
	      (cd "$(DIST_DIR)/.package/windows" && zip -q -r "$(ROOT)/$(DIST_DIR)/cmdr_$(VERSION)_windows_amd64.zip" cmdr.exe README.md LICENSE); \
	    else \
	      echo 'Error: zip is required to package Windows artifacts on Linux.' >&2; exit 1; \
	    fi; \
	  fi; \
	fi
	@rm -rf "$(DIST_DIR)/.package"
	@if [ "$(HOST_OS)" = "linux" ] && [ -f bin/linux/amd64/cmdr ]; then \
	  mkdir -p "$(DIST_DIR)/.package/linux"; \
	  cp bin/linux/amd64/cmdr "$(DIST_DIR)/.package/linux/cmdr"; \
	  cp README.md LICENSE "$(DIST_DIR)/.package/linux/"; \
	  chmod 0755 "$(DIST_DIR)/.package/linux/cmdr"; \
	  tar -czf "$(DIST_DIR)/cmdr_$(VERSION)_linux_amd64.tar.gz" -C "$(DIST_DIR)/.package/linux" cmdr README.md LICENSE; \
	fi
	@rm -rf "$(DIST_DIR)/.package"
	@if [ ! -f "$(DIST_DIR)/cmdr_$(VERSION)_windows_amd64.zip" ] && [ ! -f "$(DIST_DIR)/cmdr_$(VERSION)_linux_amd64.tar.gz" ]; then echo 'Error: no release artifact found in bin/windows/amd64 or bin/linux/amd64.' >&2; exit 1; fi
	@echo "Packages created in $(DIST_DIR)/"

checksums:
	@mkdir -p "$(DIST_DIR)"
	@if [ "$(HOST_OS)" = "windows" ]; then \
	  powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/checksums.ps1 -DistDir "$(DIST_DIR)"; \
	else \
	  cd "$(DIST_DIR)" && files="$$(find . -maxdepth 1 -type f \( -name '*.zip' -o -name '*.tar.gz' \) -printf '%f\n' | sort)"; if [ -z "$$files" ]; then echo 'Error: no release archives found in $(DIST_DIR).' >&2; exit 1; fi; cd "$(DIST_DIR)" && sha256sum $$files > checksums.txt; \
	fi
	@echo "Checksums: $(DIST_DIR)/checksums.txt"

release-check:
	@command -v $(GIT) >/dev/null 2>&1 || { echo 'Error: Git is not available in PATH.' >&2; exit 1; }
	@tag="$(TAG)"; case "$$tag" in v*) ;; *) echo 'Error: TAG must start with v, for example v2.1.7-rc.3.' >&2; exit 1;; esac
	@tag="$(TAG)"; version=$${tag#v}; project="$(VERSION)"; if [ -z "$$project" ]; then echo 'Error: could not read config.Version.' >&2; exit 1; fi; if [ "$$version" != "$$project" ]; then echo "Error: TAG $$tag does not match config.Version $$project." >&2; exit 1; fi
	@tag="$(TAG)"; if echo "$$tag" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)($|-([0-9A-Za-z-]+)(\.[0-9A-Za-z-]+)*$)'; then :; else echo "Error: invalid release tag: $$tag" >&2; exit 1; fi
	@if ! $(GIT) diff --quiet || ! $(GIT) diff --cached --quiet; then echo 'Error: release-check requires a clean working tree.' >&2; exit 1; fi
	@if $(GIT) rev-parse -q --verify "refs/tags/$(TAG)" >/dev/null 2>&1; then echo 'Error: local tag $(TAG) already exists.' >&2; exit 1; fi
	@existing="$$( $(GIT) ls-remote --exit-code --tags origin "refs/tags/$(TAG)" 2>/dev/null || true )"; if [ -n "$$existing" ]; then echo 'Error: remote tag $(TAG) already exists.' >&2; exit 1; fi
	@echo 'Release check OK: $(TAG)'

release: release-check
	@if [ "$(CONFIRM)" != "yes" ]; then echo 'Error: refusing to publish. Re-run: make release TAG=$(TAG) CONFIRM=yes' >&2; exit 1; fi
	@$(GIT) tag "$(TAG)"
	@$(GIT) push origin "$(TAG)"
	@echo 'Tag $(TAG) pushed. GitHub Actions release.yml will build and publish the GitHub Release.'

clean:
	@rm -rf "$(DIST_DIR)" bin/local bin/linux bin/windows
	@rm -f bin/cmdr.exe
	@find . -maxdepth 1 -type f -name 'resource_windows_*.syso' -delete 2>/dev/null || true
	@echo 'Generated CMDR build and release artifacts removed'
