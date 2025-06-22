#!/bin/bash
# API Direct CLI Installation Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
INSTALL_DIR="/usr/local/bin"
VERSION="latest"
BASE_URL="https://github.com/api-direct/cli/releases"

# Helper functions
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Detect platform
detect_platform() {
    case "$(uname -s)" in
        Darwin*)
            if [[ "$(uname -m)" == "arm64" ]]; then
                PLATFORM="darwin_arm64"
            else
                PLATFORM="darwin_x86_64"
            fi
            ;;
        Linux*)
            if [[ "$(uname -m)" == "x86_64" ]]; then
                PLATFORM="linux_x86_64"
            elif [[ "$(uname -m)" == "aarch64" ]]; then
                PLATFORM="linux_arm64"
            else
                error "Unsupported Linux architecture: $(uname -m)"
            fi
            ;;
        CYGWIN*|MINGW*|MSYS*)
            PLATFORM="windows_x86_64"
            ;;
        *)
            error "Unsupported platform: $(uname -s)"
            ;;
    esac
    log "Detected platform: $PLATFORM"
}

# Get latest version if not specified
get_latest_version() {
    if [[ "$VERSION" == "latest" ]]; then
        log "Fetching latest version..."
        VERSION=$(curl -s "$BASE_URL/latest" | grep -o 'tag/v[^"]*' | head -1 | cut -d'/' -f2)
        if [[ -z "$VERSION" ]]; then
            error "Could not determine latest version"
        fi
        log "Latest version: $VERSION"
    fi
}

# Download and install CLI
install_cli() {
    local filename="apidirect_${VERSION#v}_${PLATFORM}.tar.gz"
    local download_url="$BASE_URL/download/$VERSION/$filename"
    local temp_dir=$(mktemp -d)
    
    log "Downloading $download_url..."
    
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$temp_dir/$filename" "$download_url"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$temp_dir/$filename" "$download_url"
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
    
    log "Extracting to $temp_dir..."
    tar -xzf "$temp_dir/$filename" -C "$temp_dir"
    
    # Find the binary (it might be in a subdirectory)
    local binary_path
    if [[ -f "$temp_dir/apidirect" ]]; then
        binary_path="$temp_dir/apidirect"
    elif [[ -f "$temp_dir/apidirect.exe" ]]; then
        binary_path="$temp_dir/apidirect.exe"
    else
        binary_path=$(find "$temp_dir" -name "apidirect*" -type f -executable | head -1)
        if [[ -z "$binary_path" ]]; then
            error "Could not find apidirect binary in downloaded archive"
        fi
    fi
    
    log "Installing to $INSTALL_DIR..."
    
    # Check if we need sudo
    if [[ ! -w "$INSTALL_DIR" ]]; then
        if command -v sudo >/dev/null 2>&1; then
            sudo cp "$binary_path" "$INSTALL_DIR/apidirect"
            sudo chmod +x "$INSTALL_DIR/apidirect"
        else
            error "No write permission to $INSTALL_DIR and sudo not available"
        fi
    else
        cp "$binary_path" "$INSTALL_DIR/apidirect"
        chmod +x "$INSTALL_DIR/apidirect"
    fi
    
    # Cleanup
    rm -rf "$temp_dir"
    
    success "API Direct CLI installed successfully!"
}

# Setup shell completion
setup_completion() {
    log "Setting up shell completion..."
    
    case "$SHELL" in
        */bash)
            if [[ -d "/usr/local/etc/bash_completion.d" ]]; then
                apidirect completion bash | sudo tee /usr/local/etc/bash_completion.d/apidirect >/dev/null
            elif [[ -d "/etc/bash_completion.d" ]]; then
                apidirect completion bash | sudo tee /etc/bash_completion.d/apidirect >/dev/null
            else
                warn "Could not find bash completion directory"
            fi
            ;;
        */zsh)
            local zsh_comp_dir
            if [[ -d "/usr/local/share/zsh/site-functions" ]]; then
                zsh_comp_dir="/usr/local/share/zsh/site-functions"
            elif [[ -d "/usr/share/zsh/site-functions" ]]; then
                zsh_comp_dir="/usr/share/zsh/site-functions"
            else
                warn "Could not find zsh completion directory"
                return
            fi
            apidirect completion zsh | sudo tee "$zsh_comp_dir/_apidirect" >/dev/null
            ;;
        */fish)
            if [[ -d "$HOME/.config/fish/completions" ]]; then
                apidirect completion fish > "$HOME/.config/fish/completions/apidirect.fish"
            else
                warn "Could not find fish completion directory"
            fi
            ;;
        *)
            log "Shell completion not configured for: $SHELL"
            ;;
    esac
}

# Verify installation
verify_installation() {
    log "Verifying installation..."
    
    if command -v apidirect >/dev/null 2>&1; then
        local version_output=$(apidirect version)
        success "Installation verified!"
        echo "$version_output"
        echo ""
        log "Run 'apidirect help' to get started"
        
        # Offer to set up completion
        echo ""
        read -p "Would you like to set up shell completion? [y/N]: " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            setup_completion
        fi
    else
        error "Installation failed - apidirect command not found in PATH"
    fi
}

# Main execution
main() {
    echo ""
    echo "🚀 API Direct CLI Installer"
    echo "=========================="
    echo ""
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version)
                VERSION="$2"
                shift 2
                ;;
            --install-dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [options]"
                echo ""
                echo "Options:"
                echo "  --version VERSION     Install specific version (default: latest)"
                echo "  --install-dir DIR     Installation directory (default: /usr/local/bin)"
                echo "  --help                Show this help message"
                echo ""
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                ;;
        esac
    done
    
    detect_platform
    get_latest_version
    install_cli
    verify_installation
    
    echo ""
    success "🎉 API Direct CLI installation complete!"
    echo ""
    echo "Next steps:"
    echo "  1. Run 'apidirect --help' to see available commands"
    echo "  2. Run 'apidirect init' to create your first API"
    echo "  3. Check out the docs: https://docs.api-direct.com"
    echo ""
}

# Run main function
main "$@"