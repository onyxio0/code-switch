# Code Switch

AI relay manager for Claude & Codex providers.  
Builds with [Wails 3](https://v3.wails.io).

## Prerequisites

- Go 1.24+
- Node.js 18+
- npm / pnpm / yarn (project uses npm scripts)
- Wails 3 CLI (`go install github.com/wailsapp/wails/v3/cmd/wails3@latest`)

## Development

```bash
wails3 task dev
```

This installs frontend deps, runs the Vite dev server and Go backend in watch mode.

## Build

Before building, ensure the desktop bundle metadata (company, product name, etc.) is synchronized:

```bash
# Update build assets (Info.plist, icons, etc.) after editing build/config.yml
wails3 task common:update:build-assets

# Produce binaries + .app bundle
wails3 task build
```

The macOS app bundle is generated at `./bin/codeswitch.app`.

### Cross-compile (macOS ➜ Windows)

1. Install mingw-w64:
   ```bash
   brew install mingw-w64
   ```
2. Update build assets (if you changed `build/config.yml`):
   ```bash
   wails3 task common:update:build-assets
   ```
3. Build Windows binaries from macOS using the Windows task:
   ```bash
   env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
     CC=x86_64-w64-mingw32-gcc \
     CXX=x86_64-w64-mingw32-g++ \
     wails3 task windows:build
   ```
   - Output: `./bin/codeswitch.exe` + supporting files.
   - To produce the NSIS installer, run the `windows:package` task with the same environment variables.

### Publish a Release

Use the helper script to build and upload assets to GitHub Releases (requires the `gh` CLI):

```bash
# Build macOS bundle + Windows binary, then create release v0.1.0
WINDOWS_BUILD=1 scripts/publish_release.sh v0.1.0 RELEASE_NOTES.md
```

The script:
- runs `wails3 task common:update:build-assets`
- builds macOS (`bin/codeswitch.app`) and optional Windows (`bin/codeswitch.exe`) artifacts
- calls `gh release create` with the supplied tag and notes file

## Packaging Notes

If `codeswitch.app` fails to open because the executable is “missing”, it usually means the Info.plist was out of sync. Re-run the `common:update:build-assets` task, then rebuild as shown above.
