#!/bin/bash

# Test package script for multigit
# This script builds and tests the package locally

set -e

echo "🚀 Starting package test..."

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf dist/

# Build for current platform
echo "🔨 Building for current platform..."
./scripts/build.sh

# Verify the binary works
echo "✅ Verifying the binary works..."
./dist/multigit-*-$(go env GOOS)-$(go env GOARCH) --help

# Test Docker build
echo "🐳 Testing Docker build..."
docker build -t multigit:test .

echo "✅ Package test completed successfully!"
echo "📦 You can find the built packages in the 'dist' directory."
