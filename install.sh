#!/bin/sh
set -eu

REPO="Fenomen-Alex/go-chat"
VERSION="${VERSION:-v0.1.1}"
BIN_DIR="${BIN_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  linux)   OS="linux" ;;
  darwin)  OS="darwin" ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *)       echo "unsupported os: $OS"; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)           echo "unsupported arch: $ARCH"; exit 1 ;;
esac

EXT=""
if [ "$OS" = "windows" ]; then
  EXT=".exe"
fi

BINARY="chat-${OS}-${ARCH}${EXT}"
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
URL="${BASE_URL}/${BINARY}"

echo "Downloading $BINARY $VERSION..."

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

if command -v curl >/dev/null 2>&1; then
  curl -sfL "$URL" -o "$TMP_DIR/$BINARY"
elif command -v wget >/dev/null 2>&1; then
  wget -q "$URL" -O "$TMP_DIR/$BINARY"
else
  echo "need curl or wget"
  exit 1
fi

# verify checksum
CHECK_FILE="$TMP_DIR/checksums.txt"
if command -v curl >/dev/null 2>&1; then
  curl -sfL "${BASE_URL}/checksums.txt" -o "$CHECK_FILE" 2>/dev/null || true
else
  wget -q "${BASE_URL}/checksums.txt" -O "$CHECK_FILE" 2>/dev/null || true
fi

if [ ! -f "$CHECK_FILE" ]; then
  echo "error: could not fetch checksums — aborting" >&2
  exit 1
fi
EXPECTED=$(grep "$BINARY" "$CHECK_FILE" | awk '{print $1}')
if [ -z "$EXPECTED" ]; then
  echo "error: binary not found in checksums" >&2
  exit 1
fi
ACTUAL=$(sha256sum "$TMP_DIR/$BINARY" 2>/dev/null || shasum -a 256 "$TMP_DIR/$BINARY" 2>/dev/null | awk '{print $1}')
if [ "$EXPECTED" != "$ACTUAL" ]; then
  echo "checksum mismatch" >&2
  exit 1
fi
echo "checksum verified"

chmod +x "$TMP_DIR/$BINARY"

if [ "$OS" = "windows" ]; then
  mv "$TMP_DIR/$BINARY" "./$BINARY"
  echo "installed: ./$BINARY"
else
  if [ -w "$BIN_DIR" ]; then
    mv "$TMP_DIR/$BINARY" "$BIN_DIR/chat"
  else
    sudo mv "$TMP_DIR/$BINARY" "$BIN_DIR/chat"
  fi
  echo "installed: $BIN_DIR/chat"
  echo "run: chat"
fi
