#!/usr/bin/env bash
# Copyright (c) 2018-Present Lea Anthony
# SPDX-License-Identifier: MIT

# Fail script on any error
set -euxo pipefail

# Define variables
APP_DIR="${APP_NAME}.AppDir"
ARCH="x86_64"
if [[ $(uname -m) == *aarch64* ]]; then
    ARCH="aarch64"
fi

# Create AppDir structure
mkdir -p "${APP_DIR}/usr/bin"
cp -r "${APP_BINARY}" "${APP_DIR}/usr/bin/"
cp "${ICON_PATH}" "${APP_DIR}/"
cp "${DESKTOP_FILE}" "${APP_DIR}/"

if [[ $(uname -m) == *x86_64* ]]; then
    # Download linuxdeploy and make it executable
    wget -q -4 -N https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
    chmod +x linuxdeploy-x86_64.AppImage

    # Run linuxdeploy to bundle the application
    ./linuxdeploy-x86_64.AppImage --appdir "${APP_DIR}" --output appimage
else
    # Download linuxdeploy and make it executable (arm64)
    wget -q -4 -N https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-aarch64.AppImage
    chmod +x linuxdeploy-aarch64.AppImage

    # Run linuxdeploy to bundle the application (arm64)
    ./linuxdeploy-aarch64.AppImage --appdir "${APP_DIR}" --output appimage
fi

# Ensure an AppImage was produced and provide a lowercase, arch-suffixed filename
shopt -s nullglob
appimages=(*.AppImage)
shopt -u nullglob

if [[ ${#appimages[@]} -eq 0 ]]; then
    echo "No AppImage was generated" >&2
    exit 1
fi

# Pick the first AppImage produced (linuxdeploy typically outputs one)
generated="${appimages[0]}"

# Normalise name to match Wails expectation: lowercase and arch suffix
lower_app_name=$(echo "${APP_NAME}" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
expected="${lower_app_name}-${ARCH}.AppImage"

# Move into build directory (wails expects it there when moving artifacts)
target_dir="${BUILD_DIR:-build}"
mkdir -p "${target_dir}"

if [[ "${generated}" != "${target_dir}/${expected}" ]]; then
    mv "${generated}" "${target_dir}/${expected}"
else
    # If already named/located as expected, ensure path exists
    mkdir -p "$(dirname "${generated}")"
fi

echo "Placed AppImage at ${target_dir}/${expected}"
