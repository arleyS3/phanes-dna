#!/usr/bin/env bash
set -e

# Phanes DNA Installer Script for macOS and Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/arleyS3/phanes-dna/main/scripts/install.sh | bash

INSTALL_DIR="$HOME/.phanes-dna/bin"
BINARY_NAME="phanes-dna"

echo "🧬 Installing Phanes DNA..."

mkdir -p "$INSTALL_DIR"

# Auto-detect Go in common system/user locations
if ! command -v go >/dev/null 2>&1; then
    for gopath in "/tmp/opencode/go-install/go/bin" "/usr/local/go/bin" "$HOME/go/bin" "$HOME/.go/bin"; do
        if [ -x "$gopath/go" ]; then
            export PATH="$gopath:$PATH"
            break
        fi
    done
fi

if command -v go >/dev/null 2>&1; then
    echo "🔨 Building phanes-dna from Go toolchain..."
    TMP_DIR=$(mktemp -d)
    git clone https://github.com/arleyS3/phanes-dna.git "$TMP_DIR"
    cd "$TMP_DIR"
    go build -o "$INSTALL_DIR/$BINARY_NAME" ./cmd/phanes-dna
    rm -rf "$TMP_DIR"
else
    echo "⚠️ Go toolchain not found. Attempting to download latest release asset..."
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"
    if [ "$ARCH" = "x86_64" ]; then ARCH="amd64"; fi
    if [ "$ARCH" = "aarch64" ]; then ARCH="arm64"; fi
    
    DOWNLOAD_URL="https://github.com/arleyS3/phanes-dna/releases/latest/download/phanes-dna-${OS}-${ARCH}.tar.gz"
    if ! curl -fsSL "$DOWNLOAD_URL" | tar -xz -C "$INSTALL_DIR" 2>/dev/null; then
        echo "❌ Binary release asset not found on GitHub Releases."
        echo "👉 Please install Go 1.25+ or build from source:"
        echo "   git clone https://github.com/arleyS3/phanes-dna.git && cd phanes-dna && go build -o ~/.phanes-dna/bin/phanes-dna ./cmd/phanes-dna"
        exit 1
    fi
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
