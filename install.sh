#!/bin/sh
set -e

REPO="tc6-01/YangDuck"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="yduck"

get_arch() {
    arch=$(uname -m)
    case "$arch" in
        arm64|aarch64) echo "arm64" ;;
        x86_64|amd64)  echo "amd64" ;;
        *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
    esac
}

get_os() {
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$os" in
        darwin) echo "darwin" ;;
        *) echo "Unsupported OS: $os (only macOS is supported)" >&2; exit 1 ;;
    esac
}

main() {
    OS=$(get_os)
    ARCH=$(get_arch)
    
    DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}-${OS}-${ARCH}"
    
    echo "🐤 Installing YangDuck (yduck)..."
    echo "   OS: ${OS}, Arch: ${ARCH}"
    echo "   From: ${DOWNLOAD_URL}"
    echo ""

    if [ ! -d "$INSTALL_DIR" ]; then
        echo "Creating ${INSTALL_DIR}..."
        sudo mkdir -p "$INSTALL_DIR"
    fi

    TMP_FILE=$(mktemp)
    if ! curl -fsSL -o "$TMP_FILE" "$DOWNLOAD_URL"; then
        echo "❌ Download failed. Please check:" >&2
        echo "   - Is there a release at https://github.com/${REPO}/releases ?" >&2
        rm -f "$TMP_FILE"
        exit 1
    fi

    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    echo "✅ yduck installed to ${INSTALL_DIR}/${BINARY_NAME}"
    echo ""
    echo "Run 'yduck' to get started!"
}

main
