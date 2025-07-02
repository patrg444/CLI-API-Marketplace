#!/bin/bash

# Build script for API Direct CLI
# This script builds the CLI for multiple platforms

set -e

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Get version from git tag or commit
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build directory
BUILD_DIR="build"
DIST_DIR="dist"

echo -e "${BLUE}🏗️  Building API Direct CLI v${VERSION}${NC}"
echo "Build date: ${BUILD_DATE}"
echo "Commit: ${COMMIT}"
echo ""

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf ${BUILD_DIR} ${DIST_DIR}
mkdir -p ${BUILD_DIR} ${DIST_DIR}

# Build flags
LDFLAGS="-s -w"
LDFLAGS="${LDFLAGS} -X github.com/api-direct/cli/cmd.Version=${VERSION}"
LDFLAGS="${LDFLAGS} -X github.com/api-direct/cli/cmd.BuildDate=${BUILD_DATE}"
LDFLAGS="${LDFLAGS} -X github.com/api-direct/cli/cmd.GitCommit=${COMMIT}"

# Platforms to build
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "linux/386"
    "windows/amd64"
    "windows/386"
)

# Build for each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    
    OUTPUT_NAME="apidirect"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="apidirect.exe"
    fi
    
    OUTPUT_PATH="${BUILD_DIR}/${GOOS}-${GOARCH}/${OUTPUT_NAME}"
    
    echo -e "${GREEN}Building for ${GOOS}/${GOARCH}...${NC}"
    
    # Create platform directory
    mkdir -p "${BUILD_DIR}/${GOOS}-${GOARCH}"
    
    # Build
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "${LDFLAGS}" \
        -o "${OUTPUT_PATH}" \
        main.go
    
    if [ $? -eq 0 ]; then
        echo -e "  ✅ Built: ${OUTPUT_PATH}"
        
        # Create archive
        ARCHIVE_NAME="apidirect_${VERSION}_${GOOS}_${GOARCH}"
        if [ "$GOOS" = "windows" ]; then
            # Create zip for Windows
            cd "${BUILD_DIR}/${GOOS}-${GOARCH}"
            zip -q "../../${DIST_DIR}/${ARCHIVE_NAME}.zip" "${OUTPUT_NAME}"
            cd - > /dev/null
            echo -e "  📦 Archived: ${DIST_DIR}/${ARCHIVE_NAME}.zip"
        else
            # Create tar.gz for Unix-like systems
            cd "${BUILD_DIR}/${GOOS}-${GOARCH}"
            tar -czf "../../${DIST_DIR}/${ARCHIVE_NAME}.tar.gz" "${OUTPUT_NAME}"
            cd - > /dev/null
            echo -e "  📦 Archived: ${DIST_DIR}/${ARCHIVE_NAME}.tar.gz"
        fi
    else
        echo -e "${RED}  ❌ Failed to build for ${GOOS}/${GOARCH}${NC}"
    fi
    
    echo ""
done

# Create checksums
echo -e "${YELLOW}Generating checksums...${NC}"
cd ${DIST_DIR}
sha256sum * > checksums.txt
cd - > /dev/null
echo -e "✅ Checksums generated: ${DIST_DIR}/checksums.txt"

# Create version file
echo "${VERSION}" > ${DIST_DIR}/VERSION

# Summary
echo ""
echo -e "${GREEN}✅ Build complete!${NC}"
echo -e "Version: ${VERSION}"
echo -e "Artifacts: ${DIST_DIR}/"
ls -lh ${DIST_DIR}/

# Create install script
cat > ${DIST_DIR}/install.sh << 'EOF'
#!/bin/bash
# Install script for API Direct CLI

set -e

VERSION="latest"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="apidirect"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    i386|i686) ARCH="386" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# GitHub release URL
GITHUB_REPO="api-direct/cli"
RELEASE_URL="https://github.com/${GITHUB_REPO}/releases"

if [ "$VERSION" = "latest" ]; then
    DOWNLOAD_URL="${RELEASE_URL}/latest/download/apidirect_${OS}_${ARCH}.tar.gz"
else
    DOWNLOAD_URL="${RELEASE_URL}/download/${VERSION}/apidirect_${VERSION}_${OS}_${ARCH}.tar.gz"
fi

echo "📦 Installing API Direct CLI..."
echo "Platform: ${OS}/${ARCH}"
echo "Download URL: ${DOWNLOAD_URL}"

# Download
TEMP_DIR=$(mktemp -d)
trap "rm -rf ${TEMP_DIR}" EXIT

echo "⬇️  Downloading..."
curl -L -o "${TEMP_DIR}/apidirect.tar.gz" "${DOWNLOAD_URL}"

# Extract
echo "📂 Extracting..."
tar -xzf "${TEMP_DIR}/apidirect.tar.gz" -C "${TEMP_DIR}"

# Install
echo "🔧 Installing to ${INSTALL_DIR}..."
sudo mv "${TEMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/"
sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# Verify
if command -v apidirect &> /dev/null; then
    echo "✅ Installation complete!"
    echo "Version: $(apidirect version)"
else
    echo "❌ Installation failed"
    exit 1
fi

echo ""
echo "🎉 API Direct CLI installed successfully!"
echo "Run 'apidirect --help' to get started"
EOF

chmod +x ${DIST_DIR}/install.sh
echo -e "\n📄 Install script created: ${DIST_DIR}/install.sh"