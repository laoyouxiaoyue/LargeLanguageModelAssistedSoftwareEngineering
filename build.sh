#!/bin/bash
echo "Building Watermark Application..."

# Install dependencies
echo "Installing dependencies..."
go mod tidy

# Build for macOS
echo "Building for macOS..."
go build -o watermark-app -ldflags "-s -w"

# Build for Windows (if on macOS with cross-compilation)
# GOOS=windows GOARCH=amd64 go build -o watermark-app.exe -ldflags "-s -w"

echo "Build complete!"
echo "Run with: ./watermark-app"
