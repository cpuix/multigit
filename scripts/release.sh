#!/bin/bash

# Release script for multigit
# Creates a new release with version tag and GitHub release
# Requires: git, gh (GitHub CLI), goreleaser

set -e

# Check if GitHub CLI is installed
if ! command -v gh &> /dev/null; then
  echo "Error: GitHub CLI (gh) is not installed. Please install it first."
  echo "Visit: https://cli.github.com/"
  exit 1
fi

# Check if GoReleaser is installed
if ! command -v goreleaser &> /dev/null; then
  echo "Error: GoReleaser is not installed. Please install it first."
  echo "Run: brew install goreleaser/tap/goreleaser"
  exit 1
fi

# Get current version from git tags
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
echo "Current version: ${CURRENT_VERSION}"

# Ask for new version
read -p "Enter new version (e.g., v1.0.0): " NEW_VERSION

# Validate version format
if ! [[ "$NEW_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
  echo "Error: Invalid version format. Use semantic versioning (e.g., v1.0.0)"
  exit 1
fi

# Update version in files (if needed)
# sed -i '' "s/${CURRENT_VERSION}/${NEW_VERSION}/g" path/to/file

# Create release commit
echo "Creating release commit for ${NEW_VERSION}..."
git checkout -b "release-${NEW_VERSION}" 2>/dev/null || git checkout "release-${NEW_VERSION}"
git add .
git commit -m "chore: release ${NEW_VERSION}"

# Create tag
echo "Creating tag ${NEW_VERSION}..."
git tag -a "${NEW_VERSION}" -m "Release ${NEW_VERSION}"

# Build and release with GoReleaser
echo "Building and releasing with GoReleaser..."
goreleaser release --rm-dist --skip-validate

# Push changes
echo "Pushing changes to remote..."
git push origin "${NEW_VERSION}"
git push origin "release-${NEW_VERSION}"

echo "\n🎉 Release ${NEW_VERSION} is ready!"
echo "- Release notes: https://github.com/cpuix/multigit/releases/tag/${NEW_VERSION}"
echo "- Docker image: ghcr.io/cpuix/multigit:${NEW_VERSION}"
