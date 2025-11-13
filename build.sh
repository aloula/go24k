#!/bin/bash

# Build script for Go24K with version information
VERSION=$(grep 'Version.*=' version.go | sed 's/.*"\(.*\)".*/\1/')
BUILD_DATE=$(date +%Y-%m-%d)
BUILD_TIME=$(date +%H:%M:%S)

echo "🏗️  Building Go24K v$VERSION ($BUILD_DATE $BUILD_TIME)"
echo "================================================"

# Update build date in version.go
sed -i.bak "s/BuildDate.*=.*/BuildDate   = \"$BUILD_DATE\"/" version.go

# Create the builds directory if it doesn't exist
mkdir -p builds/{linux/{amd64,arm64},windows/{amd64,arm64},macos/{intel,arm}}

# Build flags for optimization and version info
BUILD_FLAGS="-ldflags=-s"

# Build for Linux (amd64)
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build $BUILD_FLAGS -o builds/linux/amd64/go24k .
if [ $? -ne 0 ]; then
    echo "❌ Failed to build for Linux (amd64)"
    exit 1
fi
echo "✅ Linux AMD64 build completed"

# Build for Linux (arm64)
echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build $BUILD_FLAGS -o builds/linux/arm64/go24k .
if [ $? -ne 0 ]; then
    echo "❌ Failed to build for Linux (arm64)"
    exit 1
fi
echo "✅ Linux ARM64 build completed"

# Build for Windows (amd64)
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build $BUILD_FLAGS -o builds/windows/amd64/go24k.exe .
if [ $? -ne 0 ]; then
    echo "❌ Failed to build for Windows (amd64)"
    exit 1
fi
echo "✅ Windows AMD64 build completed"

# Build for Windows (arm64)
echo "Building for Windows (arm64)..."
GOOS=windows GOARCH=arm64 go build $BUILD_FLAGS -o builds/windows/arm64/go24k.exe .
if [ $? -ne 0 ]; then
    echo "❌ Failed to build for Windows (arm64)"
    exit 1
fi
echo "✅ Windows ARM64 build completed"

# Build for macOS (Intel)
echo "Building for macOS (Intel)..."
GOOS=darwin GOARCH=amd64 go build $BUILD_FLAGS -o builds/macos/intel/go24k .
if [ $? -ne 0 ]; then
    echo "❌ Failed to build for macOS (Intel)"
    exit 1
fi
echo "✅ macOS Intel build completed"

# Build for macOS (ARM)
echo "Building for macOS (ARM)..."
GOOS=darwin GOARCH=arm64 go build $BUILD_FLAGS -o builds/macos/arm/go24k .
if [ $? -ne 0 ]; then
    echo "❌ Failed to build for macOS (ARM)"
    exit 1
fi
echo "✅ macOS ARM (Apple Silicon) build completed"

echo ""
echo "🎉 All builds completed successfully!"
echo "📦 Generated executables:"
echo "   • Linux: AMD64, ARM64"
echo "   • Windows: AMD64, ARM64" 
echo "   • macOS: Intel, Apple Silicon"
echo ""
echo "📁 Output directory: builds/"
echo "🏷️  Version: v$VERSION"
echo "📅 Build date: $BUILD_DATE $BUILD_TIME"

# Restore backup of version.go
if [ -f version.go.bak ]; then
    mv version.go.bak version.go
fi