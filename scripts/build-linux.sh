#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$ROOT"

unset CC

ARCH="${1:-amd64}"
VERSION_OVERRIDE="${2:-}"
case "$ARCH" in
  amd64) ;;
  *) echo "Usage: $0 amd64 [full-version]" >&2; exit 2 ;;
esac

if ! command -v go >/dev/null 2>&1; then
  echo "Error: Go is not available in PATH." >&2
  exit 1
fi

if [[ "$(uname -s | tr '[:upper:]' '[:lower:]')" == mingw* || "$(uname -s | tr '[:upper:]' '[:lower:]')" == msys* || "$(uname -s | tr '[:upper:]' '[:lower:]')" == cygwin* ]]; then
  echo "Error: run build-linux.sh in a native Linux environment or WSL/container. Windows GCC cannot build Linux targets." >&2
  exit 2
fi

ARCHES=(amd64)

for TARGET_ARCH in "${ARCHES[@]}"; do
  OUT_DIR="$ROOT/bin/linux/$TARGET_ARCH"
  OUT_FILE="$OUT_DIR/cmdr"
  mkdir -p "$OUT_DIR"

  if [[ "$TARGET_ARCH" == "amd64" ]]; then
    if ! command -v gcc >/dev/null 2>&1; then
      echo "Error: CGO is required by the WebP dependency; gcc is missing." >&2
      exit 1
    fi
  fi

  VERSION="${VERSION_OVERRIDE:-$(sed -n 's/^[[:space:]]*Version[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' config/metadata.go | head -1)}"
  COMMIT="$(git rev-parse HEAD 2>/dev/null || echo unknown)"
  BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  if [[ -z "$VERSION" ]]; then
    echo "Error: could not read config.Version from config/metadata.go" >&2
    exit 1
  fi

  echo "Building cmdr $VERSION for linux/$TARGET_ARCH..."
  GOOS=linux GOARCH="$TARGET_ARCH" CGO_ENABLED=1 go build -trimpath -ldflags "-s -w -X github.com/zerodayz7/cmdr/config.Version=$VERSION -X github.com/zerodayz7/cmdr/config.GitCommit=$COMMIT -X github.com/zerodayz7/cmdr/config.BuildDate=$BUILD_DATE" -o "$OUT_FILE" .

  if [[ ! -f "$OUT_FILE" || ! -s "$OUT_FILE" ]]; then
    echo "Error: build did not produce a valid executable: $OUT_FILE" >&2
    exit 1
  fi
  echo "SUCCESS: $OUT_FILE ($(stat -c '%s' "$OUT_FILE") bytes)"
done
