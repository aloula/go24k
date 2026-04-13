#!/bin/bash

# Build script for Go24K with version information
VERSION_FILE="utils/version.go"
VERSION=$(grep 'Version.*=' "$VERSION_FILE" | sed 's/.*"\(.*\)".*/\1/')
BUILD_DATE=$(date +%Y-%m-%d)
BUILD_TIME=$(date +%H:%M:%S)

echo "🏗️  Building Go24K v$VERSION ($BUILD_DATE $BUILD_TIME)"
echo "================================================"

# Update build date in version file
sed -i.bak "s/BuildDate.*=.*/BuildDate   = \"$BUILD_DATE\"/" "$VERSION_FILE"

# Create the builds directory if it doesn't exist
mkdir -p builds/{linux/{amd64,arm64},windows/{amd64,arm64},macos/{intel,arm}}

# Build flags for optimization and version info
BUILD_FLAGS="-ldflags=-s"
GUI_BUILD_FLAGS="$BUILD_FLAGS -tags fyne"

GUI_TARGETS=()

build_gui_target() {
    local target_os="$1"
    local target_arch="$2"
    local compiler="$3"
    local output_dir="builds/gui/${target_os}/${target_arch}"
    local binary_name="go24k-gui"
    local gui_ldflags="-s"

    if [ "$target_os" = "windows" ]; then
        binary_name="go24k-gui.exe"
        gui_ldflags="-s -H windowsgui"
    fi

    mkdir -p "$output_dir"
    echo "Building GUI for ${target_os} (${target_arch})..."

    if [ -n "$compiler" ]; then
        GOOS="$target_os" GOARCH="$target_arch" CGO_ENABLED=1 CC="$compiler" go build -tags fyne -ldflags "$gui_ldflags" -o "$output_dir/$binary_name" .
    else
        GOOS="$target_os" GOARCH="$target_arch" CGO_ENABLED=1 go build -tags fyne -ldflags "$gui_ldflags" -o "$output_dir/$binary_name" .
    fi

    if [ $? -ne 0 ]; then
        echo "⚠️  GUI build failed for ${target_os}/${target_arch}"
        return 1
    fi

    GUI_TARGETS+=("${target_os}/${target_arch}")
    echo "✅ GUI build completed for ${target_os}/${target_arch}"
    return 0
}

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

# Build GUI binaries where the required toolchains are available.
NATIVE_GOOS=$(go env GOOS)
NATIVE_GOARCH=$(go env GOARCH)

if [ "$NATIVE_GOOS" = "linux" ] && [ "$NATIVE_GOARCH" = "amd64" ]; then
    build_gui_target linux amd64 ""
fi

if command -v aarch64-linux-gnu-gcc >/dev/null 2>&1; then
    build_gui_target linux arm64 "aarch64-linux-gnu-gcc"
else
    echo "Skipping GUI for linux/arm64 (missing aarch64-linux-gnu-gcc)"
fi

if command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1; then
    build_gui_target windows amd64 "x86_64-w64-mingw32-gcc"
else
    echo "Skipping GUI for windows/amd64 (missing x86_64-w64-mingw32-gcc)"
fi

if command -v aarch64-w64-mingw32-gcc >/dev/null 2>&1; then
    build_gui_target windows arm64 "aarch64-w64-mingw32-gcc"
else
    echo "Skipping GUI for windows/arm64 (missing aarch64-w64-mingw32-gcc)"
fi

if command -v o64-clang >/dev/null 2>&1; then
    build_gui_target darwin amd64 "o64-clang"
else
    echo "Skipping GUI for darwin/amd64 (missing o64-clang)"
fi

if command -v oa64-clang >/dev/null 2>&1; then
    build_gui_target darwin arm64 "oa64-clang"
else
    echo "Skipping GUI for darwin/arm64 (missing oa64-clang)"
fi

echo ""
echo "🎉 All builds completed successfully!"
echo "📦 Generated executables:"
echo "   • Linux: AMD64, ARM64"
echo "   • Windows: AMD64, ARM64" 
echo "   • macOS: Intel, Apple Silicon"
if [ ${#GUI_TARGETS[@]} -gt 0 ]; then
    echo "   • GUI: ${GUI_TARGETS[*]}"
else
    echo "   • GUI: not generated"
fi
echo ""
echo "📁 Output directory: builds/"
echo "🏷️  Version: v$VERSION"
echo "📅 Build date: $BUILD_DATE $BUILD_TIME"

# Restore backup of version file
if [ -f "$VERSION_FILE.bak" ]; then
    mv "$VERSION_FILE.bak" "$VERSION_FILE"
fi