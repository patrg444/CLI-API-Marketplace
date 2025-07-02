# API Direct CLI Installation Guide

## Quick Install

Choose your preferred installation method:

### macOS

#### Homebrew (Recommended)
```bash
brew tap api-direct/tap
brew install apidirect
```

#### Direct Download
```bash
curl -sSL https://raw.githubusercontent.com/api-direct/cli/main/install.sh | bash
```

### Windows

#### Chocolatey (Recommended)
```powershell
choco install apidirect
```

#### Scoop
```powershell
scoop bucket add api-direct https://github.com/api-direct/scoop-bucket
scoop install apidirect
```

#### Direct Download
1. Download the latest release from [GitHub Releases](https://github.com/api-direct/cli/releases)
2. Extract the ZIP file
3. Add the directory to your PATH

### Linux

#### Debian/Ubuntu (apt)
```bash
# Add API Direct repository
curl -sSL https://apt.apidirect.io/key.gpg | sudo apt-key add -
echo "deb https://apt.apidirect.io stable main" | sudo tee /etc/apt/sources.list.d/apidirect.list

# Install
sudo apt update
sudo apt install apidirect
```

#### Red Hat/CentOS (yum)
```bash
# Add API Direct repository
sudo tee /etc/yum.repos.d/apidirect.repo <<EOF
[apidirect]
name=API Direct CLI
baseurl=https://yum.apidirect.io/stable
enabled=1
gpgcheck=1
gpgkey=https://yum.apidirect.io/RPM-GPG-KEY-apidirect
EOF

# Install
sudo yum install apidirect
```

#### Arch Linux (AUR)
```bash
yay -S apidirect
# or
paru -S apidirect
```

#### Universal Script
```bash
curl -sSL https://raw.githubusercontent.com/api-direct/cli/main/install.sh | bash
```

### Docker

```bash
# Run directly
docker run -it apidirect/cli:latest --help

# Or use as base image
FROM apidirect/cli:latest
```

### From Source

Requirements:
- Go 1.21 or later
- Git

```bash
git clone https://github.com/api-direct/cli
cd cli
go install
```

## Verify Installation

```bash
apidirect --version
```

Expected output:
```
API Direct CLI v1.0.0 (build: abc123, date: 2024-01-01)
```

## Initial Setup

### 1. Authenticate

```bash
apidirect auth login
```

This will open your browser for OAuth authentication.

### 2. Configure Environment (Optional)

```bash
# Set default region
apidirect config set region us-east-1

# Set default output format
apidirect config set output json

# View current configuration
apidirect config list
```

## Shell Completions

### Bash
```bash
# Add to ~/.bashrc
eval "$(apidirect completion bash)"
```

### Zsh
```bash
# Add to ~/.zshrc
eval "$(apidirect completion zsh)"
```

### Fish
```bash
# Add to ~/.config/fish/config.fish
apidirect completion fish | source
```

### PowerShell
```powershell
# Add to $PROFILE
apidirect completion powershell | Out-String | Invoke-Expression
```

## Environment Variables

The CLI respects these environment variables:

- `APIDIRECT_API_ENDPOINT`: API endpoint URL
- `APIDIRECT_AUTH_TOKEN`: Authentication token (for CI/CD)
- `APIDIRECT_CONFIG_DIR`: Configuration directory (default: `~/.apidirect`)
- `APIDIRECT_NO_COLOR`: Disable colored output
- `APIDIRECT_DEBUG`: Enable debug logging

## Updating

### macOS (Homebrew)
```bash
brew update && brew upgrade apidirect
```

### Windows (Chocolatey)
```powershell
choco upgrade apidirect
```

### Linux (apt)
```bash
sudo apt update && sudo apt upgrade apidirect
```

### Universal
```bash
apidirect self-update
```

## Uninstalling

### macOS (Homebrew)
```bash
brew uninstall apidirect
```

### Windows (Chocolatey)
```powershell
choco uninstall apidirect
```

### Linux (apt)
```bash
sudo apt remove apidirect
```

### Manual Cleanup
```bash
# Remove binary
sudo rm /usr/local/bin/apidirect

# Remove configuration
rm -rf ~/.apidirect
```

## Troubleshooting

### Command not found

Add the installation directory to your PATH:

**Bash/Zsh:**
```bash
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**Windows:**
1. Open System Properties → Advanced → Environment Variables
2. Add the installation directory to PATH

### Permission denied

On Unix systems, ensure the binary has execute permissions:
```bash
chmod +x /usr/local/bin/apidirect
```

### SSL/TLS errors

If you encounter SSL errors, update your certificates:

**macOS:**
```bash
brew install ca-certificates
```

**Linux:**
```bash
sudo apt-get install ca-certificates
# or
sudo yum install ca-certificates
```

### Behind a proxy

Configure proxy settings:
```bash
export HTTP_PROXY=http://proxy.company.com:8080
export HTTPS_PROXY=http://proxy.company.com:8080
export NO_PROXY=localhost,127.0.0.1
```

## Support

- Documentation: https://docs.apidirect.io
- Issues: https://github.com/api-direct/cli/issues
- Community: https://community.apidirect.io