#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$ROOT"

unset CC

ARCH="${1:-amd64}"
case "$ARCH" in
  amd64|arm64|all) ;;
  *) echo "Usage: $0 [amd64|arm64|all]" >&2; exit 2 ;;
esac

if ! command -v go >/dev/null 2>&1; then
  echo "Error: Go is not available in PATH." >&2
  exit 1
fi

if [[ "$(uname -s | tr '[:upper:]' '[:lower:]')" == mingw* || "$(uname -s | tr '[:upper:]' '[:lower:]')" == msys* || "$(uname -s | tr '[:upper:]' '[:lower:]')" == cygwin* ]]; then
  echo "Error: run build-linux.sh in a native Linux environment or WSL/container. Windows GCC cannot build Linux targets." >&2
  exit 2
fi

ARCHES=("$ARCH")
if [[ "$ARCH" == "all" ]]; then
  ARCHES=(amd64 arm64)
fi

for TARGET_ARCH in "${ARCHES[@]}"; do
  OUT_DIR="$ROOT/bin/linux/$TARGET_ARCH"
  OUT_FILE="$OUT_DIR/cmdr"
  mkdir -p "$OUT_DIR"

  if [[ "$TARGET_ARCH" == "amd64" ]]; then
    if ! command -v gcc >/dev/null 2>&1; then
      echo "Error: CGO is required by the WebP dependency; gcc is missing." >&2
      exit 1
    fi
  else
    if ! command -v aarch64-linux-gnu-gcc >/dev/null 2>&1; then
      echo "Error: Linux arm64 cross-build requires aarch64-linux-gnu-gcc." >&2
      exit 1
    fi
    export CC=aarch64-linux-gnu-gcc
  fi

  echo "Building cmdr for linux/$TARGET_ARCH..."
  GOOS=linux GOARCH="$TARGET_ARCH" CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o "$OUT_FILE" .

  if [[ ! -f "$OUT_FILE" || ! -s "$OUT_FILE" ]]; then
    echo "Error: build did not produce a valid executable: $OUT_FILE" >&2
    exit 1
  fi
  echo "SUCCESS: $OUT_FILE ($(stat -c '%s' "$OUT_FILE") bytes)"
done
