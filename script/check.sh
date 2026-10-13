#!/usr/bin/env bash
# Agent 验收回路：后端三连（可选 WEB=1 附带前端 build）
set -euo pipefail
cd "$(dirname "$0")/.."

if [ -z "${PATH##*"$HOME/.local/go/bin"*}" ] || [ -x "$HOME/.local/go/bin/go" ]; then
  export PATH="$HOME/.local/go/bin:$PATH"
fi

echo "==> go test ./..."
go test ./...

echo "==> go vet ./..."
go vet ./...

echo "==> go build ./..."
go build ./...

if [ "${WEB:-0}" = "1" ]; then
  echo "==> web npm run build"
  (cd web && npm run build)
fi

# 开发文档站（缺 hugo 时 SKIP，不阻塞验收；见 script/build-docs.sh）
echo "==> build docs"
./script/build-docs.sh

echo "check OK"
