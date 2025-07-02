#!/bin/bash

# Simple build test for API Direct CLI

echo "🧪 Testing API Direct CLI Build"
echo "================================"

# Set Go path
export PATH=/Users/patrickgloria/CLI-API-Marketplace/cli/go/bin:$PATH

# Clean previous builds
rm -f apidirect-test

# Try to build
echo "Building CLI..."
if go build -o apidirect-test main.go 2>&1 | grep -q "imported and not used"; then
    echo "⚠️  Build has unused import warnings, cleaning up..."
    
    # For now, just try to build ignoring warnings
    go build -a -o apidirect-test main.go 2>/dev/null || true
fi

# Check if binary was created
if [ -f apidirect-test ]; then
    echo "✅ Binary created!"
    
    # Test basic commands
    echo ""
    echo "Testing --version:"
    ./apidirect-test --version || echo "Version command needs fixing"
    
    echo ""
    echo "Testing --help:"
    ./apidirect-test --help | head -10 || echo "Help command needs fixing"
    
    echo ""
    echo "Testing completion command:"
    ./apidirect-test completion bash | head -5 || echo "Completion needs fixing"
else
    echo "❌ Build failed - checking specific errors..."
    go build -o apidirect-test main.go 2>&1 | head -20
fi