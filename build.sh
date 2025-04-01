#!/bin/bash
echo "Installing dependencies..."
npm install

echo "Building executable..."
npm run build

echo "Build complete! Executables are in the dist folder." 