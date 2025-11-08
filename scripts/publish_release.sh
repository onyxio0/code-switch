#!/usr/bin/env bash
set -euo pipefail

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required. Install from https://cli.github.com/" >&2
  exit 1
fi

if [ $# -lt 2 ]; then
  cat >&2 <<'USAGE'
Usage: scripts/publish_release.sh <tag> <notes-file>

- <tag>        : git tag to create/release (e.g. v0.1.0)
- <notes-file> : markdown file containing release notes

Environment variables:
  BUILD_OPTS      Optional extra flags passed to `wails3 task build`
  WINDOWS_BUILD   Set to 1 to cross-compile Windows binaries
USAGE
  exit 1
fi

TAG="$1"
NOTES="$2"

if [ ! -f "$NOTES" ]; then
  echo "Release notes file '$NOTES' not found" >&2
  exit 1
fi

wails3 task common:update:build-assets
wails3 task build ${BUILD_OPTS:-}

if [ "${WINDOWS_BUILD:-0}" = "1" ]; then
  env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
    CC=x86_64-w64-mingw32-gcc \
    CXX=x86_64-w64-mingw32-g++ \
    wails3 task windows:build ${BUILD_OPTS:-}
fi

ASSETS=(bin/codeswitch.app)
if [ "${WINDOWS_BUILD:-0}" = "1" ] && [ -f bin/codeswitch.exe ]; then
  ASSETS+=(bin/codeswitch.exe)
fi

for asset in "${ASSETS[@]}"; do
  [ -e "$asset" ] || { echo "Missing asset: $asset" >&2; exit 1; }
  echo "  asset: $asset"
done

gh release create "$TAG" "${ASSETS[@]}" \
  --title "$TAG" \
  --notes-file "$NOTES"
