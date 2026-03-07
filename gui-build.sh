#!/bin/bash
# Simple GUI Build Script
# Builds production version of the GUI

set -e

echo "🔨 Building GUI for production..."
cd gui && ~/go/bin/wails build

echo ""
echo "✅ Build complete!"
echo "📦 Binary location: gui/build/bin/"
ls -lh gui/build/bin/ | grep -v "^d" | tail -1
