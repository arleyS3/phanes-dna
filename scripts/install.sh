#!/usr/bin/env bash
set -e

# Phanes DNA Installer Script for macOS and Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/arley/phanes-dna/main/scripts/install.sh | bash

INSTALL_DIR="$HOME/.phanes-dna/bin"
BINARY_NAME="phanes-dna"

echo "🧬 Installing Phanes DNA..."

mkdir -p "$INSTALL_DIR"

if command -v go >/dev/null 2>&1; then
    echo "🔨 Building phanes-dna from Go toolchain..."
    go install github.com/arley/phanes-dna/cmd/phanes-dna@latest || (
        TMP_DIR=$(mktemp -d)
        git clone https://github.com/arley/phanes-dna.git "$TMP_DIR"
        cd "$TMP_DIR"
        go build -o "$INSTALL_DIR/$BINARY_NAME" ./cmd/phanes-dna
        rm -rf "$TMP_DIR"
    )
    if [ -f "$HOME/go/bin/phanes-dna" ]; then
        cp "$HOME/go/bin/phanes-dna" "$INSTALL_DIR/$BINARY_NAME"
    fi
else
    echo "⚠️ Go toolchain not found. Downloading latest binary release..."
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"
    if [ "$ARCH" = "x86_64" ]; then ARCH="amd64"; fi
    if [ "$ARCH" = "aarch64" ]; then ARCH="arm64"; fi
    
    DOWNLOAD_URL="https://github.com/arley/phanes-dna/releases/latest/download/phanes-dna-${OS}-${ARCH}.tar.gz"
    curl -fsSL "$DOWNLOAD_URL" | tar -xz -C "$INSTALL_DIR"
fi

chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Check PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "📝 Adding $INSTALL_DIR to PATH..."
    SHELL_PROFILE="$HOME/.bashrc"
    if [ -n "$ZSH_VERSION" ] || [ -f "$HOME/.zshrc" ]; then
        SHELL_PROFILE="$HOME/.zshrc"
    fi
    echo "" >> "$SHELL_PROFILE"
    echo 'export PATH="$HOME/.phanes-dna/bin:$PATH"' >> "$SHELL_PROFILE"
fi

echo "✅ Phanes DNA installed successfully to $INSTALL_DIR/$BINARY_NAME!"
echo "🚀 Run 'phanes-dna' to launch the interactive terminal UI, or 'phanes-dna doctor' for health checks."
