#!/bin/bash

# Build script for multigit
# Usage: ./scripts/build.sh [version]

set -e

VERSION=${1:-$(git describe --tags 2>/dev/null || echo "dev")}
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "")
GIT_TREE_STATE=""
if [[ -n $(git status --porcelain 2>/dev/null) ]]; then
  GIT_TREE_STATE="dirty"
else
  GIT_TREE_STATE="clean"
fi

LDFLAGS="-X 'main.version=${VERSION}'
        -X 'main.buildDate=${BUILD_DATE}'
        -X 'main.gitCommit=${GIT_COMMIT}'
        -X 'main.gitTreeState=${GIT_TREE_STATE}'"

PLATFORMS=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
  "windows/386"
)

mkdir -p dist

echo "Building multigit version: ${VERSION}"
echo "Build date: ${BUILD_DATE}"
echo "Git commit: ${GIT_COMMIT}"
echo "Git tree state: ${GIT_TREE_STATE}"

for PLATFORM in "${PLATFORMS[@]}"; do
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}
  BINARY_NAME="multigit"
  OUTPUT="dist/multigit-${VERSION}-${GOOS}-${GOARCH}"
  
  if [ "${GOOS}" = "windows" ]; then
    BINARY_NAME="multigit.exe"
    OUTPUT="${OUTPUT}.exe"
  fi

  echo "Building for ${GOOS}/${GOARCH}..."
  
  CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
    -ldflags "${LDFLAGS}" \
    -o "${OUTPUT}" \
    .
    
  echo "  Built: ${OUTPUT}"
done

echo "Build complete!"
echo "Binaries are available in the 'dist' directory."
