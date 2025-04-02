#!/bin/bash

# Create output directory if it doesn't exist
mkdir -p build

# Build for Windows (64-bit)
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o build/flex-processor.exe main.go

# Build for macOS (64-bit)
echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -o build/flex-processor_mac main.go

echo "Build complete! Executables are in the 'build' directory:"
echo "- Windows: flex-processor.exe"
echo "- macOS: flex-processor_mac" 