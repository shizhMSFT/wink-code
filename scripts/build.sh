#!/bin/bash
# Cross-platform build script for wink CLI
# Builds binaries for Linux, macOS, and Windows

set -e

# Version info
VERSION=${VERSION:-"dev"}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}'"

# Output directory
OUT_DIR="dist"
mkdir -p "$OUT_DIR"

echo "Building wink CLI v${VERSION}..."
echo "Build time: ${BUILD_TIME}"
echo ""

# Build for Linux (amd64)
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o "${OUT_DIR}/wink-linux-amd64" ./cmd/wink
echo "✓ ${OUT_DIR}/wink-linux-amd64"

# Build for Linux (arm64)
echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o "${OUT_DIR}/wink-linux-arm64" ./cmd/wink
echo "✓ ${OUT_DIR}/wink-linux-arm64"

# Build for macOS (amd64 - Intel)
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o "${OUT_DIR}/wink-darwin-amd64" ./cmd/wink
echo "✓ ${OUT_DIR}/wink-darwin-amd64"

# Build for macOS (arm64 - Apple Silicon)
echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o "${OUT_DIR}/wink-darwin-arm64" ./cmd/wink
echo "✓ ${OUT_DIR}/wink-darwin-arm64"

# Build for Windows (amd64)
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o "${OUT_DIR}/wink-windows-amd64.exe" ./cmd/wink
echo "✓ ${OUT_DIR}/wink-windows-amd64.exe"

echo ""
echo "Build complete! Binaries in ${OUT_DIR}/"
ls -lh "${OUT_DIR}/"
