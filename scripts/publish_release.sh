#!/usr/bin/env bash
set -euo pipefail

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required. Install from https://cli.github.com/" >&2
  exit 1
fi

if [ $# -lt 1 ]; then
  echo "Usage: scripts/publish_release.sh <tag> [notes-file]" >&2
  exit 1
fi

TAG="$1"
NOTES="${2:-RELEASE_NOTES.md}"

if [ ! -f "$NOTES" ]; then
  echo "Release notes file '$NOTES' not found" >&2
  exit 1
fi

wails3 task common:update:build-assets
wails3 task package ${BUILD_OPTS:-}

env ARCH=amd64 wails3 task windows:package ${BUILD_OPTS:-}

if [ ! -d "bin/codeswitch.app" ]; then
  echo "Missing asset: bin/codeswitch.app" >&2
  exit 1
fi

echo "==> Archiving macOS app bundle"
MAC_ZIP="bin/codeswitch-macos.zip"
rm -f "$MAC_ZIP"
ditto -c -k --sequesterRsrc --keepParent "bin/codeswitch.app" "$MAC_ZIP"
rm -rf "bin/codeswitch.app"

ASSETS=(
  "$MAC_ZIP"
  "bin/codeswitch-amd64-installer.exe"
  "bin/codeswitch.exe"
)

for asset in "${ASSETS[@]}"; do
  [ -e "$asset" ] || { echo "Missing asset: $asset" >&2; exit 1; }
  echo "  asset: $asset"
done

gh release create "$TAG" "${ASSETS[@]}" \
  --title "$TAG" \
  --notes-file "$NOTES"
