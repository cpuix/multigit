#!/bin/bash

# Simple release script for multigit

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

# Check if GitHub token is set
if [ -z "$GITHUB_TOKEN" ]; then
  echo "Error: GITHUB_TOKEN environment variable is not set."
  echo "Please create a GitHub personal access token with 'repo' scope and run:"
  echo "  export GITHUB_TOKEN=your_token_here"
  echo "Then run this script again."
  exit 1
fi

# Create a release branch
echo "Creating release branch for ${NEW_VERSION}..."
git checkout -b "release-${NEW_VERSION}"

# Create tag
echo "Creating tag ${NEW_VERSION}..."
git tag -a "${NEW_VERSION}" -m "Release ${NEW_VERSION}"

# Run GoReleaser
echo "Running GoReleaser..."
rm -rf dist/
GITHUB_TOKEN=$GITHUB_TOKEN goreleaser release

# Push changes
echo "Pushing changes to remote..."
git push origin "${NEW_VERSION}"
git push origin "release-${NEW_VERSION}"

echo "\n🎉 Release ${NEW_VERSION} is ready!"
echo "- Release notes: https://github.com/cpuix/multigit/releases/tag/${NEW_VERSION}"
echo "- Docker image: ghcr.io/cpuix/multigit:${NEW_VERSION}"

echo "\n📝 Don't forget to merge the release branch into main/master when ready:"
echo "  git checkout main"
echo "  git merge release-${NEW_VERSION}"
echo "  git push origin main"
