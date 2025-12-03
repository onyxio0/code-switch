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

# Ensure an AppImage was produced
shopt -s nullglob
appimages=(*.AppImage)
shopt -u nullglob

# Filter out linuxdeploy itself
for img in "${appimages[@]}"; do
    if [[ "$img" != linuxdeploy* ]]; then
        generated="$img"
        break
    fi
done

if [[ -z "${generated:-}" ]]; then
    echo "No AppImage was generated" >&2
    exit 1
fi

# Move to output directory with lowercase name and arch suffix
lower_app_name=$(echo "${APP_NAME}" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
target_name="${lower_app_name}-${ARCH}.AppImage"
mkdir -p "${OUTPUT_DIR}"
mv -f "${generated}" "${OUTPUT_DIR}/${target_name}"

echo "AppImage ready at ${OUTPUT_DIR}/${target_name}"
