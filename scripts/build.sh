#!/bin/bash
# CS2 Better Auto Director - Linux Build Script

set -e

echo "============================================"
echo "CS2 Better Auto Director - Build Script"
echo "Version 0.2.2 - Linux Edition"
echo "============================================"
echo ""

# Check if Go is installed
echo "Checking Go installation..."
if ! command -v go &> /dev/null; then
    echo ""
    echo "ERROR: Go is not installed!"
    echo ""
    echo "Please install Go 1.21 or higher from: https://go.dev/dl/"
    echo "Or use your package manager:"
    echo "  Ubuntu/Debian: sudo apt install golang"
    echo "  Fedora: sudo dnf install golang"
    echo "  Arch: sudo pacman -S go"
    echo ""
    exit 1
fi

GO_VER=$(go version)
echo "Go found: $GO_VER"
echo ""

# Check if Node.js is installed
echo "Checking Node.js installation..."
if ! command -v node &> /dev/null; then
    echo ""
    echo "ERROR: Node.js is not installed!"
    echo ""
    echo "Wails requires Node.js for frontend build."
    echo ""
    echo "Install with:"
    echo "  Ubuntu/Debian: sudo apt install nodejs npm"
    echo "  Fedora: sudo dnf install nodejs npm"
    echo "  Arch: sudo pacman -S nodejs npm"
    echo ""
    exit 1
fi

NODE_VER=$(node --version)
NPM_VER=$(npm --version)
echo "Node.js found: $NODE_VER"
echo "npm found: $NPM_VER"
echo ""

# Check if Wails CLI is installed
echo "Checking Wails CLI installation..."
if ! command -v wails &> /dev/null; then
    echo ""
    echo "ERROR: Wails CLI is not installed!"
    echo ""
    echo "Installing Wails CLI..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    
    # Add Go bin to PATH if not already there
    export PATH="$PATH:$(go env GOPATH)/bin"
    
    if ! command -v wails &> /dev/null; then
        echo ""
        echo "Wails CLI installed successfully!"
        echo ""
        echo "IMPORTANT: Please add Go's bin directory to your PATH:"
        echo "  export PATH=\"\$PATH:\$(go env GOPATH)/bin\""
        echo ""
        echo "Add this line to your ~/.bashrc or ~/.zshrc and restart your terminal."
        echo "Then run this script again."
        echo ""
        exit 0
    fi
fi

WAILS_VER=$(wails version)
echo "Wails found:"
echo "$WAILS_VER"
echo ""

# Check Linux build dependencies
echo "Checking Linux build dependencies..."
if ! pkg-config --exists gtk+-3.0 webkit2gtk-4.1; then
    echo ""
    echo "ERROR: Required GTK/WebKit2GTK libraries not found!"
    echo ""
    echo "Please install build dependencies:"
    echo "  Ubuntu/Debian: sudo apt install build-essential libgtk-3-dev libwebkit2gtk-4.1-dev"
    echo "  Fedora: sudo dnf install gtk3-devel webkit2gtk3-devel"
    echo "  Arch: sudo pacman -S gtk3 webkit2gtk"
    echo ""
    exit 1
fi
echo "GTK and WebKit2GTK found!"
echo ""

# Verify environment
echo "Verifying Wails environment..."
wails doctor
echo ""

# Navigate to Wails project directory
echo "Navigating to Wails project..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
WAILS_DIR="$ROOT_DIR/gui"

if [ ! -d "$WAILS_DIR" ]; then
    echo ""
    echo "ERROR: Wails project directory not found!"
    echo "Expected: $WAILS_DIR"
    echo ""
    exit 1
fi

cd "$WAILS_DIR"
echo "Working directory: $WAILS_DIR"
echo ""

# Build with Wails
echo "Building with Wails..."
echo "This may take a few minutes..."
echo ""

wails build -skipbindings -tags webkit2gtk_4_1

echo ""
echo "============================================"
echo "SUCCESS! Build complete"
echo "============================================"
echo ""

# Check if executable was created
BUILT_BIN="$WAILS_DIR/build/bin/cs2-better-autodirector"

if [ -f "$BUILT_BIN" ]; then
    FILE_SIZE=$(stat -c%s "$BUILT_BIN" 2>/dev/null || stat -f%z "$BUILT_BIN" 2>/dev/null)
    FILE_SIZE_MB=$(awk "BEGIN {printf \"%.2f\", $FILE_SIZE/1048576}")
    
    echo "Created: cs2-better-autodirector"
    echo "Location: $BUILT_BIN"
    echo "File size: ${FILE_SIZE_MB} MB"
    echo ""
    echo "You can now run the program with:"
    echo "    ./gui/build/bin/cs2-better-autodirector          (GUI mode - default)"
    echo "    ./gui/build/bin/cs2-better-autodirector -nogui   (CLI mode)"
    echo ""
    
    # Make executable
    chmod +x "$BUILT_BIN"
    echo "Executable permissions set."
    echo ""
else
    echo "WARNING: Executable not found at expected location"
    echo "Expected: $BUILT_BIN"
fi

echo "Build completed successfully!"
