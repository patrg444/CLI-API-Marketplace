#!/bin/bash
# Test the installation process

set -e

echo "🧪 Testing API Direct CLI Installation"
echo "======================================"

# Test if the install script exists and is executable
if [[ -f "install.sh" && -x "install.sh" ]]; then
    echo "✅ Install script found and executable"
else
    echo "❌ Install script not found or not executable"
    exit 1
fi

# Test if the CLI binary exists
if [[ -f "apidirect" ]]; then
    echo "✅ CLI binary found"
    
    # Test basic CLI functionality
    echo "🔍 Testing CLI commands..."
    
    if ./apidirect version > /dev/null 2>&1; then
        echo "✅ Version command works"
        ./apidirect version
    else
        echo "❌ Version command failed"
        exit 1
    fi
    
    if ./apidirect help > /dev/null 2>&1; then
        echo "✅ Help command works"
    else
        echo "❌ Help command failed"
        exit 1
    fi
    
    # Test a few key commands
    for cmd in "marketplace --help" "analytics --help" "search --help"; do
        if ./apidirect $cmd > /dev/null 2>&1; then
            echo "✅ Command '$cmd' works"
        else
            echo "❌ Command '$cmd' failed"
            exit 1
        fi
    done
    
else
    echo "❌ CLI binary not found"
    exit 1
fi

# Test documentation files
docs=("QUICK_START.md" "BUILD_SUCCESS.md" "install.sh")
for doc in "${docs[@]}"; do
    if [[ -f "$doc" ]]; then
        echo "✅ Documentation file $doc exists"
    else
        echo "❌ Documentation file $doc missing"
        exit 1
    fi
done

echo ""
echo "🎉 All tests passed! CLI is ready for distribution."
echo ""
echo "📦 Installation command for users:"
echo "curl -fsSL https://raw.githubusercontent.com/patrg444/CLI-API-Marketplace/main/cli/install.sh | bash"
echo ""
echo "🔗 GitHub Repository:"
echo "https://github.com/patrg444/CLI-API-Marketplace"
echo ""
echo "🚀 Release created: v1.0.0"