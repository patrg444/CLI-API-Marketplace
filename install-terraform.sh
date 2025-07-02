#!/bin/bash
# Install Terraform on macOS

echo "Installing Terraform..."

# Check if Homebrew is installed
if command -v brew &> /dev/null; then
    echo "Installing Terraform via Homebrew..."
    brew tap hashicorp/tap
    brew install hashicorp/tap/terraform
else
    echo "Installing Terraform manually..."
    
    # Detect architecture
    ARCH=$(uname -m)
    if [[ "$ARCH" == "arm64" ]]; then
        TERRAFORM_ARCH="arm64"
    else
        TERRAFORM_ARCH="amd64"
    fi
    
    # Download Terraform
    TERRAFORM_VERSION="1.6.6"
    wget "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_darwin_${TERRAFORM_ARCH}.zip"
    
    # Unzip and install
    unzip terraform_${TERRAFORM_VERSION}_darwin_${TERRAFORM_ARCH}.zip
    sudo mv terraform /usr/local/bin/
    rm terraform_${TERRAFORM_VERSION}_darwin_${TERRAFORM_ARCH}.zip
fi

# Verify installation
terraform --version