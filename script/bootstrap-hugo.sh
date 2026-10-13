#!/usr/bin/env bash
# 一次性引导：把 pinned 版本 Hugo 放到 script/bin/hugo。
#
# 获取策略（GitHub Release CDN 在部分网络极慢/不可达，故提供两条路径）：
#   1. linux：下载官方 tar.gz（hugo_<ver>_linux-<arch>.tar.gz）
#   2. 其余情况（含 darwin：>0.152.2 起官方只发 .pkg，装包要 sudo）：
#      `go install github.com/gohugoio/hugo@v<ver>` 源码构建（GOBIN=script/bin，
#      模块完整性由 Go checksum DB 校验）。
# 已就绪时直接退出，可重复执行。
set -euo pipefail
cd "$(dirname "$0")/.."

HUGO_VERSION="${HUGO_VERSION:-0.167.0}"
BIN_DIR="$PWD/script/bin"
mkdir -p "$BIN_DIR"

if [ -x "$BIN_DIR/hugo" ] && "$BIN_DIR/hugo" version 2>/dev/null | grep -q "v$HUGO_VERSION"; then
  echo "hugo v$HUGO_VERSION 已就绪：script/bin/hugo"
  exit 0
fi

os="$(uname -s)"
arch="$(uname -m)"

download_tarball() {
  local platform="$1"
  local archive="hugo_${HUGO_VERSION}_${platform}.tar.gz"
  local url="https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/${archive}"
  local tmp
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' RETURN
  echo "==> download $url"
  curl -fsSL --connect-timeout 15 -o "$tmp/$archive" "$url" || return 1
  tar -xzf "$tmp/$archive" -C "$tmp" hugo
  mv "$tmp/hugo" "$BIN_DIR/hugo"
  chmod +x "$BIN_DIR/hugo"
}

build_from_source() {
  command -v go >/dev/null 2>&1 || {
    echo "go 不可用，无法源码构建 hugo。请安装 Go 或手动提供 script/bin/hugo" >&2
    return 1
  }
  echo "==> go install github.com/gohugoio/hugo@v${HUGO_VERSION}（首次需数分钟）"
  GOBIN="$BIN_DIR" go install "github.com/gohugoio/hugo@v${HUGO_VERSION}"
}

acquired=""
if [ "$os" = "Linux" ]; then
  case "$arch" in
    arm64|aarch64) platform="linux-arm64" ;;
    x86_64|amd64)  platform="linux-amd64" ;;
    *) platform="" ;;
  esac
  if [ -n "$platform" ]; then
    download_tarball "$platform" || build_from_source
    acquired=1
  fi
fi

if [ -z "$acquired" ]; then
  build_from_source
fi

"$BIN_DIR/hugo" version
