#!/bin/bash

set -e

BIN_NAME="cfdk"
BIN_URL="https://raw.githubusercontent.com/nithinkjoy-tech/cfdk/main/cfdk" # change this to correct URL if needed
LOCAL_INSTALL_DIR="$HOME/.local/bin"
SYSTEM_INSTALL_DIR="/usr/local/bin"

echo "📦 Installing $BIN_NAME..."

# Create temporary download directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# Download the prebuilt binary
curl -fsSL "$BIN_URL" -o "$BIN_NAME"
chmod +x "$BIN_NAME"

# Check if we can install to /usr/local/bin
if [ -w "$SYSTEM_INSTALL_DIR" ]; then
  echo "✅ Admin access detected. Installing to $SYSTEM_INSTALL_DIR..."
  mv "$BIN_NAME" "$SYSTEM_INSTALL_DIR/"
else
  echo "⚠️  No admin rights. Installing to $LOCAL_INSTALL_DIR instead..."
  mkdir -p "$LOCAL_INSTALL_DIR"
  mv "$BIN_NAME" "$LOCAL_INSTALL_DIR/"

  # Ensure ~/.local/bin is in PATH
  if [[ ":$PATH:" != *":$LOCAL_INSTALL_DIR:"* ]]; then
    SHELL_CONFIG="${HOME}/.zshrc"
    [[ $SHELL == *bash* ]] && SHELL_CONFIG="${HOME}/.bash_profile"

    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$SHELL_CONFIG"
    echo "✅ Added $LOCAL_INSTALL_DIR to your PATH in $SHELL_CONFIG. Run: source $SHELL_CONFIG"
  fi
fi

echo "🎉 Installed successfully! You can now run: $BIN_NAME"
