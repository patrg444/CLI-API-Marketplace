#!/bin/bash
# Build Debian package for API Direct CLI

set -e

VERSION=${1:-"1.0.0"}
ARCH=$(dpkg --print-architecture)

echo "Building Debian package for API Direct CLI v${VERSION} (${ARCH})..."

# Create package directory structure
PACKAGE_DIR="apidirect_${VERSION}_${ARCH}"
mkdir -p "${PACKAGE_DIR}/DEBIAN"
mkdir -p "${PACKAGE_DIR}/usr/bin"
mkdir -p "${PACKAGE_DIR}/usr/share/doc/apidirect"
mkdir -p "${PACKAGE_DIR}/usr/share/man/man1"

# Copy binary
cp "../../cli/dist/apidirect_${VERSION}_linux_${ARCH}/apidirect" "${PACKAGE_DIR}/usr/bin/"
chmod 755 "${PACKAGE_DIR}/usr/bin/apidirect"

# Create control file
cat > "${PACKAGE_DIR}/DEBIAN/control" << EOF
Package: apidirect
Version: ${VERSION}
Architecture: ${ARCH}
Maintainer: API Direct Team <support@apidirect.io>
Depends: libc6
Section: devel
Priority: optional
Homepage: https://github.com/api-direct/cli
Description: API Direct CLI - Rapid API deployment tool
 API Direct CLI transforms your code into deployed APIs in minutes.
 Features include auto-detection, containerization, marketplace integration,
 and comprehensive monitoring capabilities.
EOF

# Create postinst script
cat > "${PACKAGE_DIR}/DEBIAN/postinst" << 'EOF'
#!/bin/sh
set -e

case "$1" in
    configure)
        # Generate shell completions
        if [ -x /usr/bin/apidirect ]; then
            echo "Generating shell completions..."
            /usr/bin/apidirect completion bash > /etc/bash_completion.d/apidirect 2>/dev/null || true
            /usr/bin/apidirect completion zsh > /usr/share/zsh/vendor-completions/_apidirect 2>/dev/null || true
        fi
        ;;
esac

exit 0
EOF
chmod 755 "${PACKAGE_DIR}/DEBIAN/postinst"

# Create postrm script
cat > "${PACKAGE_DIR}/DEBIAN/postrm" << 'EOF'
#!/bin/sh
set -e

case "$1" in
    remove|purge)
        # Remove shell completions
        rm -f /etc/bash_completion.d/apidirect
        rm -f /usr/share/zsh/vendor-completions/_apidirect
        ;;
esac

exit 0
EOF
chmod 755 "${PACKAGE_DIR}/DEBIAN/postrm"

# Add documentation
cp ../../README.md "${PACKAGE_DIR}/usr/share/doc/apidirect/"
cp ../../LICENSE "${PACKAGE_DIR}/usr/share/doc/apidirect/"

# Create man page
cat > "${PACKAGE_DIR}/usr/share/man/man1/apidirect.1" << 'EOF'
.TH APIDIRECT 1 "2024" "API Direct CLI" "User Commands"
.SH NAME
apidirect \- rapid API deployment and marketplace management tool
.SH SYNOPSIS
.B apidirect
[\fICOMMAND\fR] [\fIOPTIONS\fR]
.SH DESCRIPTION
API Direct CLI transforms your code into deployed APIs in minutes. It provides
auto-detection of frameworks, container-based deployment, and built-in
marketplace features.
.SH COMMANDS
.TP
.B import
Import an existing API project
.TP
.B validate
Validate API configuration
.TP
.B run
Run API locally for development
.TP
.B env
Manage environment variables
.TP
.B logs
View API logs
.TP
.B scale
Scale API instances
.TP
.B status
View API deployment status
.TP
.B publish
Publish API to marketplace
.TP
.B analytics
View API analytics
.TP
.B earnings
Track API revenue
.TP
.B subscriptions
Manage API subscriptions
.SH OPTIONS
.TP
.B \-h, \-\-help
Show help message
.TP
.B \-v, \-\-version
Show version information
.SH EXAMPLES
.TP
Import an API project:
.B apidirect import /path/to/api
.TP
Run API locally:
.B apidirect run my-api
.TP
View logs:
.B apidirect logs my-api --follow
.SH SEE ALSO
Full documentation at: https://docs.apidirect.io
.SH BUGS
Report bugs at: https://github.com/api-direct/cli/issues
EOF
gzip -9 "${PACKAGE_DIR}/usr/share/man/man1/apidirect.1"

# Build the package
dpkg-deb --build "${PACKAGE_DIR}"

echo "✅ Debian package built: ${PACKAGE_DIR}.deb"