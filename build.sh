#!/bin/bash

# Create the builds directory if it doesn't exist
mkdir -p builds

# Build for Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o builds/linux/go24k main.go
if [ $? -ne 0 ]; then
    echo "Failed to build for Linux"
    exit 1
fi

# Build for Windows (amd64)
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o builds/windows/amd64/go24k.exe main.go
if [ $? -ne 0 ]; then
    echo "Failed to build for Windows (amd64)"
    exit 1
fi

# Build for Windows (arm64)
echo "Building for Windows (arm64)..."
GOOS=windows GOARCH=arm64 go build -o builds/windows/arm64/go24k.exe main.go
if [ $? -ne 0 ]; then
    echo "Failed to build for Windows (arm64)"
    exit 1
fi

# Build for macOS (Intel)
echo "Building for macOS (Intel)..."
GOOS=darwin GOARCH=amd64 go build -o builds/macos/intel/go24k main.go
if [ $? -ne 0 ]; then
    echo "Failed to build for macOS (Intel)"
    exit 1
fi

# Build for macOS (ARM)
echo "Building for macOS (ARM)..."
GOOS=darwin GOARCH=arm64 go build -o builds/macos/arm/go24k main.go
if [ $? -ne 0 ]; then
    echo "Failed to build for macOS (ARM)"
    exit 1
fi

echo "Builds completed successfully!"