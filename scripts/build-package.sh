#!/bin/bash
# CS2 Better Auto Director - Linux Package Builder
# Creates a distributable tar.gz archive

set -e

echo "============================================"
echo "CS2 Better Auto Director - Package Builder"
echo "Version 0.2.2 - Linux Edition"
echo "============================================"
echo ""

# Get directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
WAILS_DIR="$ROOT_DIR/gui"
BUILD_DIR="$ROOT_DIR/build"
BUILT_BIN="$WAILS_DIR/build/bin/cs2-better-autodirector"

# Check if binary exists
if [ ! -f "$BUILT_BIN" ]; then
    echo "ERROR: Binary not found at $BUILT_BIN"
    echo ""
    echo "Please build the application first:"
    echo "    ./scripts/build.sh"
    echo ""
    exit 1
fi

echo "Binary found: $BUILT_BIN"
echo ""

# Create build directory
echo "Creating package directory..."
mkdir -p "$BUILD_DIR"

# Create temporary package directory
TEMP_DIR="$BUILD_DIR/cs2-better-autodirector-linux"
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"

# Copy executable
echo "Copying executable..."
cp "$BUILT_BIN" "$TEMP_DIR/"
chmod +x "$TEMP_DIR/cs2-better-autodirector"

# Copy config files
echo "Copying config files..."
mkdir -p "$TEMP_DIR/config"
cp "$ROOT_DIR/config/gamestate_integration_autoobserver.cfg" "$TEMP_DIR/config/"
cp "$ROOT_DIR/config/spectator_bindings.cfg" "$TEMP_DIR/config/"

# Copy documentation
echo "Copying documentation..."
cp "$ROOT_DIR/README.md" "$TEMP_DIR/"
cp "$ROOT_DIR/CHANGELOG.md" "$TEMP_DIR/"

# Create install script
echo "Creating install script..."
cat > "$TEMP_DIR/install.sh" << 'EOF'
#!/bin/bash
# CS2 Better Auto Director - Installation Script

echo "CS2 Better Auto Director - Linux Installation"
echo "=============================================="
echo ""

# Get install directory
DEFAULT_INSTALL_DIR="$HOME/.local/share/cs2-better-autodirector"
read -p "Install directory [$DEFAULT_INSTALL_DIR]: " INSTALL_DIR
INSTALL_DIR="${INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"

# Create directory
echo ""
echo "Creating installation directory..."
mkdir -p "$INSTALL_DIR"

# Copy files
echo "Copying files..."
cp -r ./* "$INSTALL_DIR/"

# Create desktop entry
DESKTOP_FILE="$HOME/.local/share/applications/cs2-better-autodirector.desktop"
mkdir -p "$(dirname "$DESKTOP_FILE")"

echo "Creating desktop entry..."
cat > "$DESKTOP_FILE" << DESKTOPEOF
[Desktop Entry]
Version=1.0
Type=Application
Name=CS2 Better Auto Director
Comment=Intelligent spectator camera control for CS2
Exec=$INSTALL_DIR/cs2-better-autodirector
Terminal=false
Categories=Game;
DESKTOPEOF

chmod +x "$INSTALL_DIR/cs2-better-autodirector"

echo ""
echo "=============================================="
echo "Installation complete!"
echo "=============================================="
echo ""
echo "Installed to: $INSTALL_DIR"
echo ""
echo "To run: $INSTALL_DIR/cs2-better-autodirector"
echo "Or find it in your application menu."
echo ""
echo "Next steps:"
echo "1. Copy config/gamestate_integration_autoobserver.cfg to:"
echo "   ~/.local/share/Steam/steamapps/common/Counter-Strike Global Offensive/game/csgo/cfg/"
echo ""
echo "2. Copy config/spectator_bindings.cfg to your CS2 cfg folder"
echo ""
echo "3. In CS2 console, execute: exec spectator_bindings"
echo ""
EOF

chmod +x "$TEMP_DIR/install.sh"

# Create uninstall script
echo "Creating uninstall script..."
cat > "$TEMP_DIR/uninstall.sh" << 'EOF'
#!/bin/bash
# CS2 Better Auto Director - Uninstallation Script

echo "CS2 Better Auto Director - Uninstallation"
echo "=========================================="
echo ""

DEFAULT_INSTALL_DIR="$HOME/.local/share/cs2-better-autodirector"
read -p "Installation directory [$DEFAULT_INSTALL_DIR]: " INSTALL_DIR
INSTALL_DIR="${INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"

if [ -d "$INSTALL_DIR" ]; then
    echo ""
    read -p "Remove $INSTALL_DIR? (y/N): " CONFIRM
    if [ "$CONFIRM" = "y" ] || [ "$CONFIRM" = "Y" ]; then
        rm -rf "$INSTALL_DIR"
        echo "Removed installation directory."
    fi
fi

DESKTOP_FILE="$HOME/.local/share/applications/cs2-better-autodirector.desktop"
if [ -f "$DESKTOP_FILE" ]; then
    rm "$DESKTOP_FILE"
    echo "Removed desktop entry."
fi

echo ""
echo "Uninstallation complete!"
EOF

chmod +x "$TEMP_DIR/uninstall.sh"

# Create README for package
cat > "$TEMP_DIR/INSTALL.txt" << 'EOF'
CS2 Better Auto Director - Linux Installation
==============================================

Quick Installation:
1. Extract this archive
2. Run: ./install.sh
3. Follow the prompts

Manual Installation:
1. Copy cs2-better-autodirector to any directory
2. Copy config files to your CS2 installation
3. Run: ./cs2-better-autodirector

Configuration Files:
- config/gamestate_integration_autoobserver.cfg
  → Copy to: ~/.local/share/Steam/steamapps/common/Counter-Strike Global Offensive/game/csgo/cfg/

- config/spectator_bindings.cfg
  → Copy to your CS2 cfg folder
  → Execute in CS2: exec spectator_bindings

Running the Application:
- GUI mode (default): ./cs2-better-autodirector
- CLI mode: ./cs2-better-autodirector -nogui

For more information, see README.md
EOF

# Create tar.gz archive
ARCHIVE_NAME="CS2BetterAutoDirector-Linux-x64.tar.gz"
echo ""
echo "Creating archive: $ARCHIVE_NAME"
cd "$BUILD_DIR"
tar -czf "$ARCHIVE_NAME" "cs2-better-autodirector-linux/"

# Get archive size
ARCHIVE_SIZE=$(stat -c%s "$ARCHIVE_NAME" 2>/dev/null || stat -f%z "$ARCHIVE_NAME" 2>/dev/null)
ARCHIVE_SIZE_MB=$(awk "BEGIN {printf \"%.2f\", $ARCHIVE_SIZE/1048576}")

# Cleanup temp directory
rm -rf "$TEMP_DIR"

echo ""
echo "============================================"
echo "Package created successfully!"
echo "============================================"
echo ""
echo "Archive: $BUILD_DIR/$ARCHIVE_NAME"
echo "Size: ${ARCHIVE_SIZE_MB} MB"
echo ""
echo "To distribute:"
echo "  tar -xzf $ARCHIVE_NAME"
echo "  cd cs2-better-autodirector-linux"
echo "  ./install.sh"
echo ""
