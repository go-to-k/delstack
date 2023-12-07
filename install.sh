#!/bin/bash

# This script automates the installation of the delstack tool.
# It checks for the specified version (or fetches the latest one),
# downloads the binary, and installs it on the system.

# Check for required tools: curl and tar.
# These tools are necessary for downloading and extracting the delstack binary.
if ! command -v curl &>/dev/null; then
	echo "curl could not be found"
	exit 1
fi

if ! command -v tar &>/dev/null; then
	echo "tar could not be found"
	exit 1
fi

# Determine the version of delstack to install.
# If no version is specified as a command line argument, fetch the latest version.
if [ -z "$1" ]; then
	VERSION=$(curl -s https://api.github.com/repos/go-to-k/delstack/releases/latest | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
	if [ -z "$VERSION" ]; then
		echo "Failed to fetch the latest version"
		exit 1
	fi
else
	VERSION=$1
fi

# Normalize the version string by removing any leading 'v'.
VERSION=${VERSION#v}

# Detect the architecture of the current system.
# This script supports x86_64, arm64, and i386 architectures.
ARCH=$(uname -m)
case $ARCH in
x86_64 | amd64) ARCH="x86_64" ;;
arm64 | aarch64) ARCH="arm64" ;;
i386 | i686) ARCH="i386" ;;
*)
	echo "Unsupported architecture: $ARCH"
	exit 1
	;;
esac

# Detect the operating system (OS) of the current system.
# This script supports Linux, Darwin (macOS) and Windows operating systems.
OS=$(uname -s)
case $OS in
Linux) OS="Linux" ;;
Darwin) OS="Darwin" ;;
MINGW* | MSYS* | CYGWIN*) OS="Windows" ;;
*)
	echo "Unsupported OS: $OS"
	exit 1
	;;
esac

# Construct the download URL for the delstack binary based on the version, OS, and architecture.
FILE_NAME="delstack_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/go-to-k/delstack/releases/download/v${VERSION}/${FILE_NAME}"

# Download the delstack binary.
echo "Downloading delstack..."
if ! curl -L -o "$FILE_NAME" "$URL"; then
	echo "Failed to download delstack"
	exit 1
fi

# Install delstack.
# This involves extracting the binary and moving it to /usr/local/bin.
echo "Installing delstack..."
if ! tar -xzf "$FILE_NAME"; then
	echo "Failed to extract delstack"
	exit 1
fi
if ! sudo mv delstack /usr/local/bin/delstack; then
	echo "Failed to install delstack"
	exit 1
fi

# Clean up by removing the downloaded tar file.
rm "$FILE_NAME"

echo "delstack installation complete."
echo "Run 'delstack -h' to see how to use delstack."
