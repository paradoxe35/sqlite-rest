#!/bin/bash

# Exit on error, undefined variable, or pipe failure
set -euo pipefail

# Get version from git tag
VERSION=$(git describe --tags)
PREFIX=sqlite-rest_${VERSION}_

echo "Building assets for the release ${VERSION}..."

# Cleanup release dir
rm -rf ./release
mkdir -p ./release

# Define platforms to build for
PLATFORMS=(
  "windows:amd64:.exe"
  "windows:386:.exe"
  "linux:amd64:"
  "linux:386:"
  "linux:arm:"
  "linux:arm64:"
  "darwin:amd64:"
  "darwin:arm64:"
)

# Build flags for optimization and smaller binaries
BUILD_FLAGS="-trimpath -ldflags \"-s -w -X main.VERSION=${VERSION}\" -tags netgo"

# Create directories and build binaries for each platform
for platform in "${PLATFORMS[@]}"; do
  IFS=: read -r OS ARCH EXT <<< "${platform}"
  DIR="./release/${PREFIX}${OS}-${ARCH}"
  BIN="${DIR}/sqlite-rest${EXT}"

  echo "Building for ${OS}/${ARCH}..."

  # Create directory
  mkdir -p "${DIR}"

  # Copy assets
  cp -R ./LICENSE ./README.md ./CHANGELOG.md "${DIR}/"

  # Build binary
  GOOS=${OS} GOARCH=${ARCH} CGO_ENABLED=0 go build ${BUILD_FLAGS} -o "${BIN}" ./cmd/sqlite-rest.go

  # Create archive
  pushd ./release > /dev/null
  if [[ "${OS}" == "windows" ]]; then
    zip -r "./${PREFIX}${OS}-${ARCH}.zip" "./${PREFIX}${OS}-${ARCH}/" > /dev/null
  else
    tar -czf "./${PREFIX}${OS}-${ARCH}.tar.gz" "./${PREFIX}${OS}-${ARCH}/" > /dev/null
  fi
  popd > /dev/null

  # Calculate checksum
  if [[ "${OS}" == "windows" ]]; then
    pushd ./release > /dev/null
    sha256sum "./${PREFIX}${OS}-${ARCH}.zip" >> checksums.txt
    popd > /dev/null
  else
    pushd ./release > /dev/null
    sha256sum "./${PREFIX}${OS}-${ARCH}.tar.gz" >> checksums.txt
    popd > /dev/null
  fi

  # Clean up directory
  rm -rf "${DIR}"
done

echo "Build complete! Assets are in the ./release directory."
echo "Generated checksums:"
cat ./release/checksums.txt