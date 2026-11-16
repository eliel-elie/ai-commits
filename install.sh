#!/usr/bin/env sh
set -eu

REPO="eliel-elie/ai-commits"
BINARY_NAME="aic"
VERSION="${AIC_VERSION:-latest}"

detect_os() {
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    linux) echo "linux" ;;
    darwin) echo "darwin" ;;
    mingw*|msys*|cygwin*) echo "windows" ;;
    *)
      echo "Unsupported OS: $os" >&2
      exit 1
      ;;
  esac
}

detect_arch() {
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *)
      echo "Unsupported architecture: $arch" >&2
      exit 1
      ;;
  esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"
EXT=""
if [ "$OS" = "windows" ]; then
  EXT=".exe"
fi
ASSET="aic-${OS}-${ARCH}${EXT}"
TARGET_NAME="${BINARY_NAME}${EXT}"

if [ "$VERSION" = "latest" ]; then
  URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"
else
  URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"
fi

if [ "$OS" = "windows" ]; then
  DEFAULT_DIR="${LOCALAPPDATA:-$HOME}/Programs/aic"
else
  DEFAULT_DIR="/usr/local/bin"
fi
INSTALL_DIR="${AIC_INSTALL_DIR:-$DEFAULT_DIR}"
TMP_FILE="$(mktemp)"
trap 'rm -f "$TMP_FILE"' EXIT

echo "Downloading ${ASSET} (${VERSION})..."
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$URL" -o "$TMP_FILE"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$TMP_FILE" "$URL"
else
  echo "Error: curl or wget is required." >&2
  exit 1
fi

chmod +x "$TMP_FILE"

if [ ! -d "$INSTALL_DIR" ]; then
  mkdir -p "$INSTALL_DIR" 2>/dev/null || sudo mkdir -p "$INSTALL_DIR"
fi

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_FILE" "${INSTALL_DIR}/${TARGET_NAME}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "$TMP_FILE" "${INSTALL_DIR}/${TARGET_NAME}"
fi

echo "Installed to ${INSTALL_DIR}/${TARGET_NAME}"
"${INSTALL_DIR}/${TARGET_NAME}" --version
