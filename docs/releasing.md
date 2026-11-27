# Releasing cmdr

This repository uses the root `Makefile` as the primary interface for development, validation, local builds, release preparation and publishing.

The existing automated release pipeline in `.github/workflows/release.yml` remains the release authority. Pushing a version tag to GitHub triggers the workflow, which validates the tag, tests the project, builds Windows and Linux binaries, packages archives, generates checksums and creates the GitHub Release.

## Required tools

Required for core Go targets:

- Go 1.27.0 or compatible
- GNU Make
- Git

Additional tools:

- GCC and CGO support: required because `github.com/chai2010/webp` uses CGO.
- PowerShell: required for Windows release builds and Windows artifact/checksum commands.
- Bash: required for the Linux release script.
- GCC on Linux: required for Linux `amd64` release builds.
- `zip`: required only when packaging Windows archives from Linux.
- `tar` and `sha256sum`: used for Linux packaging/checksums.
- `golangci-lint`: required only by `make lint` and `make check`.
- `goversioninfo`: installed automatically by `scripts/build-windows.ps1` when missing.

## Windows setup

Run GNU Make from Git Bash/MSYS or a shell where `make`, `go`, `git`, PowerShell and GCC are available. The Makefile is intended to work with GNU Make on Windows and Unix-like systems.

PowerShell functions for Windows release packaging and checksum generation are provided in:

- `scripts/package-release.ps1`
- `scripts/checksums.ps1`

## Local development setup

Clone the repository and verify the Makefile:

```bash
make help
make version
```

## Testing and validation

```bash
make fmt
make fmt-check
make vet
make lint
make test
make test-race
make check
```

- `make fmt` modifies Go source files with `gofmt`.
- `make fmt-check` verifies formatting without modifying files.
- `make vet` runs `go vet ./...`.
- `make lint` runs `golangci-lint run ./...` when `golangci-lint` is installed.
- `make test` runs `go test ./...`.
- `make test-race` runs `go test -race ./...`.
- `make check` runs `fmt-check`, `vet`, `test`, and `lint`.

## Local builds

Build CMDR for the detected current host:

```bash
make build
```

The alias `make build-local` performs the same build. The default output is:

- Windows: `bin/local/windows/amd64/cmdr.exe`
- Linux: `bin/local/linux/amd64/cmdr`

A custom local output directory can be selected:

```bash
make build LOCAL_OUT_DIR=bin/local/windows/amd64
```

Install to the Go bin directory:

```bash
make install
```

Local builds are separate from `bin/windows/amd64` and `bin/linux/amd64` release artifacts.

CGO is enabled intentionally because the WebP dependency requires it. Windows and Linux builds require GCC. macOS local builds are not included in the existing cross-platform release matrix, but the host detection supports a local macOS build when the host toolchain can compile the project with CGO enabled.

## Cross-platform release builds

The Makefile executes the repository's existing native release scripts:

```bash
make release-build
```

Supported release matrix:

| Operating system | Architecture | Binary location |
|---|---|---|
| Windows | amd64 | `bin/windows/amd64/cmdr.exe` |
| Linux | amd64 | `bin/linux/amd64/cmdr` |

This matrix matches the current GitHub Actions workflow. macOS is not a supported release target of the existing workflow.

Native Windows builds require PowerShell, MinGW/GCC and the Windows build resources in `build/windows/`. Linux builds must run in native Linux or WSL, not through Windows GCC/MSYS.

A release version can be passed explicitly:

```bash
make release-build VERSION=2.1.7-rc.3
```

If omitted, `VERSION` is read from `config/metadata.go`.

## Version metadata

The build scripts inject values using `-ldflags`:

- `config.Version`
- `config.GitCommit`
- `config.BuildDate`

The default version comes from `config/metadata.go`. `make version` prints the effective version, commit SHA, build date and host platform.

Stable tags use `vMAJOR.MINOR.PATCH`. Pre-release tags use `vMAJOR.MINOR.PATCH-PRERELEASE`, for example `v2.1.7-rc.3`.

## Release preparation

Build the supported binaries:

```bash
make release-build
```

Create archives:

```bash
make package
```

Generate SHA-256 checksums:

```bash
make checksums
```

`make package` reads native release binaries from:

- `bin/windows/amd64/cmdr.exe`
- `bin/linux/amd64/cmdr`

Generated archives are written to `dist/`:

- `dist/cmdr_<version>_windows_amd64.zip`
- `dist/cmdr_<version>_linux_amd64.tar.gz`
- `dist/checksums.txt`

The Linux archive contains `cmdr`, `README.md` and `LICENSE`. The Windows ZIP contains `cmdr.exe`, `README.md` and `LICENSE`.

Verify a checksum from the repository root:

```bash
cd dist
sha256sum -c checksums.txt
```

On Windows PowerShell:

```powershell
Get-Content .\dist\checksums.txt
Get-FileHash -Algorithm SHA256 .\dist\cmdr_2.1.7-rc.2_windows_amd64.zip
```

## Release check and GitHub publishing

Publishing is handled exclusively by pushing a `v*` tag. The repository's GitHub Actions `release.yml` workflow performs validation, testing, builds, packaging and `gh release create` publication.

Before tagging, update `config/metadata.go` to the release version, commit and push the source changes:

```bash
git add config/metadata.go
git commit -m "Prepare release v2.1.7"
git push origin main
```

For a prerelease:

```bash
git add config/metadata.go
git commit -m "Prepare release v2.1.7-rc.3"
git push origin main
```

Validate the release without publishing:

```bash
make release-check TAG=v2.1.7-rc.3
```

`release-check` verifies:

- the tag format;
- that the tag matches `config.Version`;
- that the working tree and index are clean;
- that the local tag does not already exist;
- that the remote tag does not already exist.

Create and push the tag with confirmation:

```bash
make release TAG=v2.1.7-rc.3 CONFIRM=yes
```

`make release` never publishes directly through GitHub CLI. It creates one local tag and pushes that tag. The push triggers `.github/workflows/release.yml`, which then creates the GitHub Release.

Do not run `make release` without explicit authorization. The target does not force-push, move existing tags or overwrite releases.

To verify without changing a remote repository:

```bash
make release-check TAG=v2.1.7-rc.3
```

## Cleanup

```bash
make clean
```

Cleanup removes only generated project artifacts:

- `dist/`
- `bin/local/`
- `bin/linux/`
- `bin/windows/`
- `bin/cmdr.exe`
- temporary `resource_windows_*.syso` files

It does not remove source files or user configuration.

## Complete target reference

| Target | Purpose |
|---|---|
| `make help` | Display available targets and examples. |
| `make version` | Print version, commit, build date and host platform. |
| `make fmt` | Format all Go source files. |
| `make fmt-check` | Verify Go formatting without modifications. |
| `make vet` | Run `go vet ./...`. |
| `make lint` | Run `golangci-lint run ./...`. |
| `make test` | Run `go test ./...`. |
| `make test-race` | Run `go test -race ./...`. |
| `make check` | Run formatting, vet, tests and lint. |
| `make build` | Build for the detected local host. |
| `make build-local` | Alias for `make build`. |
| `make install` | Run `go install` with CMDR build metadata. |
| `make release-build` | Build native Windows/Linux amd64 release binaries. |
| `make package` | Create Windows ZIP and/or Linux tarball archives. |
| `make checksums` | Generate `dist/checksums.txt`. |
| `make release-check` | Validate tag and repository release state. |
| `make release` | Create and push the tag after confirmation. |
| `make clean` | Remove generated CMDR build and release artifacts. |

## Troubleshooting

- **`go` or `gofmt` missing:** install Go and ensure Go's `bin` directory is in `PATH`.
- **`gcc` missing:** install MinGW/GCC on Windows or GCC on Linux; CGO is required.
- **Linux build says Windows GCC is invalid:** run the Linux build in native Linux or WSL.
- **`golangci-lint` missing:** install it before `make lint` or `make check`, or run `make fmt-check vet test` individually.
- **Windows resource build fails:** verify `build/windows/icon.ico`, `build/windows/manifest.xml` and `goversioninfo` availability.
- **Tag mismatch:** update `config/metadata.go`, commit the change and ensure `TAG` equals `v<config.Version>`.
- **Tag already exists:** never force-move it; increment to a new version or prerelease.
- **No package output:** run `make release-build` before `make package`.
- **No archive for checksum:** run `make package` before `make checksums`.
