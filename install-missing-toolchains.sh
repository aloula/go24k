#!/usr/bin/env bash
set -euo pipefail

echo "==> Installing cross-toolchains for go24k GUI builds"

echo "==> Step 1/4: Base toolchains"
sudo apt-get update
sudo apt-get install -y \
  gcc-aarch64-linux-gnu \
  mingw-w64 \
  clang lld cmake ninja-build pkg-config git make autoconf automake libtool patch unzip xz-utils curl

echo "==> Step 2/4: Configure apt sources for arm64 packages via ubuntu-ports"
sudo cp /etc/apt/sources.list.d/ubuntu.sources /etc/apt/sources.list.d/ubuntu.sources.pre-arm64-fix.bak

sudo tee /etc/apt/sources.list.d/ubuntu.sources >/dev/null <<'EOF'
Types: deb
URIs: http://archive.ubuntu.com/ubuntu/
Suites: noble noble-updates noble-backports
Components: main universe restricted multiverse
Architectures: amd64
Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg

Types: deb
URIs: http://security.ubuntu.com/ubuntu/
Suites: noble-security
Components: main universe restricted multiverse
Architectures: amd64
Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg
EOF

sudo tee /etc/apt/sources.list.d/ubuntu-arm64.sources >/dev/null <<'EOF'
Types: deb
URIs: http://ports.ubuntu.com/ubuntu-ports/
Suites: noble noble-updates noble-backports
Components: main universe restricted multiverse
Architectures: arm64
Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg

Types: deb
URIs: http://ports.ubuntu.com/ubuntu-ports/
Suites: noble-security
Components: main universe restricted multiverse
Architectures: arm64
Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg
EOF

# Keep third-party x86_64 repo amd64-only to avoid arm64 404s.
if [ -f /etc/apt/sources.list.d/cuda-ubuntu2204-x86_64.list ]; then
  sudo sed -i 's/^deb \[/deb [arch=amd64 /' /etc/apt/sources.list.d/cuda-ubuntu2204-x86_64.list
fi

sudo apt-get update

echo "==> Step 3/4: Linux arm64 GUI link dependencies"
sudo apt-get install -y \
  libgl1-mesa-dev:arm64 \
  libx11-dev:arm64 \
  libxrandr-dev:arm64 \
  libxxf86vm-dev:arm64 \
  libxi-dev:arm64 \
  libxcursor-dev:arm64 \
  libxinerama-dev:arm64

echo "==> Step 4/4: Optional - install llvm-mingw for Windows ARM64 (user-local)"
mkdir -p "$HOME/.local/toolchains" "$HOME/.local/bin"

LLVM_MINGW_URL=$(curl -s https://api.github.com/repos/mstorsjo/llvm-mingw/releases/latest \
  | grep browser_download_url \
  | cut -d '"' -f 4 \
  | grep 'ubuntu-20.04-x86_64' \
  | grep 'ucrt' \
  | head -n 1 || true)

if [ -n "${LLVM_MINGW_URL}" ]; then
  LLVM_ARCHIVE="$HOME/.local/toolchains/$(basename "$LLVM_MINGW_URL")"
  curl -L "$LLVM_MINGW_URL" -o "$LLVM_ARCHIVE"
  tar -xf "$LLVM_ARCHIVE" -C "$HOME/.local/toolchains"
  LLVM_DIR=$(find "$HOME/.local/toolchains" -maxdepth 1 -type d -name 'llvm-mingw-*' | head -n 1)

  cat > "$HOME/.local/bin/aarch64-w64-mingw32-gcc" <<EOF
#!/usr/bin/env bash
exec "$LLVM_DIR/bin/aarch64-w64-mingw32-clang" "\$@"
EOF

  cat > "$HOME/.local/bin/aarch64-w64-mingw32-g++" <<EOF
#!/usr/bin/env bash
exec "$LLVM_DIR/bin/aarch64-w64-mingw32-clang++" "\$@"
EOF

  chmod +x "$HOME/.local/bin/aarch64-w64-mingw32-gcc" "$HOME/.local/bin/aarch64-w64-mingw32-g++"
  echo "Installed Windows ARM64 compiler wrappers in ~/.local/bin"
else
  echo "Could not auto-resolve llvm-mingw release URL."
  echo "Install llvm-mingw manually from: https://github.com/mstorsjo/llvm-mingw/releases"
fi

echo ""
echo "Done. You can now re-run: ./build.sh"
echo ""
echo "Note: macOS cross-GUI builds still require osxcross (o64-clang/oa64-clang) plus an Apple SDK."
echo "That part cannot be fully automated without providing a local Mac SDK tarball."
