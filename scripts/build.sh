#!/bin/bash

# DB Backup Pro - Cross-platform Build Script
# Usage: ./scripts/build.sh [platform]
# Platforms: macos, windows, linux, all

set -e

VERSION="1.0.0"
APP_NAME="dbbackup"
BUILD_DIR="build/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}  DB Backup Pro - Build Script v${VERSION}${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""

# Check if wails is installed
if ! command -v wails &> /dev/null; then
    echo -e "${RED}Error: wails is not installed or not in PATH${NC}"
    echo "Please install wails: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    exit 1
fi

# Get current OS and architecture
CURRENT_OS=$(go env GOOS)
CURRENT_ARCH=$(go env GOARCH)

build_macos() {
    echo -e "${YELLOW}Building for macOS (arm64 + amd64)...${NC}"

    # Build for ARM64 (Apple Silicon)
    echo "  Building macOS ARM64..."
    wails build -platform darwin/arm64 -o "${APP_NAME}-macos-arm64"

    # Build for AMD64 (Intel)
    echo "  Building macOS AMD64..."
    wails build -platform darwin/amd64 -o "${APP_NAME}-macos-amd64"

    echo -e "${GREEN}macOS builds complete!${NC}"
}

build_windows() {
    echo -e "${YELLOW}Building for Windows (amd64)...${NC}"

    # Note: Cross-compiling to Windows from macOS requires some setup
    # You may need to install mingw-w64: brew install mingw-w64

    wails build -platform windows/amd64 -o "${APP_NAME}.exe"

    echo -e "${GREEN}Windows build complete!${NC}"
}

build_linux() {
    echo -e "${YELLOW}Building for Linux (amd64)...${NC}"

    # Note: Cross-compiling to Linux from macOS may require additional setup

    wails build -platform linux/amd64 -o "${APP_NAME}-linux-amd64"

    echo -e "${GREEN}Linux build complete!${NC}"
}

build_current() {
    echo -e "${YELLOW}Building for current platform (${CURRENT_OS}/${CURRENT_ARCH})...${NC}"
    wails build
    echo -e "${GREEN}Build complete!${NC}"
}

# Parse arguments
PLATFORM=${1:-"current"}

case $PLATFORM in
    "macos")
        build_macos
        ;;
    "windows")
        build_windows
        ;;
    "linux")
        build_linux
        ;;
    "all")
        build_macos
        build_windows
        build_linux
        ;;
    "current")
        build_current
        ;;
    *)
        echo "Usage: $0 [platform]"
        echo "Platforms: macos, windows, linux, all, current (default)"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}  Build Complete!${NC}"
echo -e "${GREEN}  Output directory: ${BUILD_DIR}${NC}"
echo -e "${GREEN}================================================${NC}"

# List built files
echo ""
echo "Built files:"
ls -la ${BUILD_DIR}/ 2>/dev/null || echo "No files in build directory yet"
