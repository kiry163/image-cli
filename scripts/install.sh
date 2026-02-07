#!/usr/bin/env bash

set -euo pipefail

REPO="kiry163/image-cli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
CONFIG_PATH="${HOME}/.config/image-cli/config.yaml"

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

fail() {
  echo "Error: $1" >&2
  exit 1
}

need_cmd() {
  if ! command_exists "$1"; then
    fail "缺少依赖: $1"
  fi
}

need_cmd curl

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Darwin)
    OS_NAME="darwin"
    ;;
  Linux)
    OS_NAME="linux"
    ;;
  *)
    fail "不支持的系统: $OS"
    ;;
esac

case "$ARCH" in
  x86_64|amd64)
    ARCH_NAME="amd64"
    ;;
  arm64|aarch64)
    ARCH_NAME="arm64"
    ;;
  *)
    fail "不支持的架构: $ARCH"
    ;;
esac

if [ "$OS_NAME" = "darwin" ] && [ "$ARCH_NAME" = "amd64" ]; then
  fail "当前未提供 darwin-amd64 二进制，建议源码编译"
fi

if [ "$OS_NAME" = "linux" ] && [ "$ARCH_NAME" = "arm64" ]; then
  fail "当前未提供 linux-arm64 二进制，建议源码编译"
fi

BIN_NAME="image-cli-${OS_NAME}-${ARCH_NAME}"

if ! command_exists pkg-config; then
  fail "缺少依赖: pkg-config"
fi

if ! pkg-config --exists vips; then
  echo "缺少 libvips，请先安装" >&2
  if [ "$OS_NAME" = "darwin" ]; then
    echo "  brew install vips" >&2
  elif [ "$OS_NAME" = "linux" ]; then
    echo "  sudo apt-get update && sudo apt-get install -y libvips libvips-dev pkg-config" >&2
  fi
  exit 1
fi

TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')"
if [ -z "$TAG" ]; then
  fail "无法获取最新版本"
fi

URL="https://github.com/${REPO}/releases/download/${TAG}/${BIN_NAME}"
TMP_DIR="$(mktemp -d)"
TMP_BIN="${TMP_DIR}/${BIN_NAME}"

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

echo "下载: ${URL}"
curl -fL -o "$TMP_BIN" "$URL"
chmod +x "$TMP_BIN"

mkdir -p "$INSTALL_DIR"
TARGET_BIN="${INSTALL_DIR}/image-cli"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_BIN" "$TARGET_BIN"
else
  sudo mv "$TMP_BIN" "$TARGET_BIN"
fi

echo "已安装: $TARGET_BIN"

if [ ! -f "$CONFIG_PATH" ]; then
  "$TARGET_BIN" config init >/dev/null 2>&1 || true
fi

echo "完成: ImageCLI ${TAG}"
