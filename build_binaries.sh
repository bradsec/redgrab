#!/bin/sh

set -e

appName="redgrab"
version="1_0"
outputDir="dist"

# Check if 'go' command is available
command -v go >/dev/null 2>&1 || { echo >&2 "Error: 'go' command not found. Make sure Go is installed."; exit 1; }

# Create output directory if it doesn't exist
if [ ! -d "$outputDir" ]; then
  echo "Creating output directory: $outputDir"
  mkdir "$outputDir"
fi

# Linux 64-bit
GOOS=linux GOARCH=amd64 go build
echo "Building Linux 64-bit binary and compress/tar as ${appName}_linux_amd64_v${version}.tar.xz..."
tar -cJf "${outputDir}/${appName}_linux_amd64_v${version}.tar.xz" "${appName}"
rm "${appName}"

# Windows 64-bit
GOOS=windows GOARCH=amd64 go build
echo "Building Windows 64-bit binary and compress/zip as ${appName}_windows64_v${version}.zip..."
zip -r "${outputDir}/${appName}_windows64_v${version}.zip" "${appName}.exe"
rm "${appName}.exe"

# macOS Intel 64-bit
GOOS=darwin GOARCH=amd64 go build
echo "Building macOS Intel 64-bit binary compress/tar as ${appName}_macos_intel_v${version}.tar.xz..."
tar -cJf "${outputDir}/${appName}_macos_intel_v${version}.tar.xz" "${appName}"
rm "${appName}"

# macOS ARM 64-bit
GOOS=darwin GOARCH=arm64 go build
echo "Building macOS ARM 64-bit binary compress/tar as ${appName}_macos_arm64_v${version}.tar.xz..."
tar -cJf "${outputDir}/${appName}_macos_arm64_v${version}.tar.xz" "${appName}"
rm "${appName}"
