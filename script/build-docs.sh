#!/usr/bin/env bash
# 构建 docs/hugo → docs/hugo/public。
# hugo 缺失时打印 SKIP 并 exit 0（不阻塞 check.sh）；有 hugo 但构建失败则 exit 1。
set -euo pipefail
cd "$(dirname "$0")/.."

if command -v hugo >/dev/null 2>&1; then
  HUGO="$(command -v hugo)"
elif [ -x "script/bin/hugo" ]; then
  HUGO="$PWD/script/bin/hugo"   # 子 shell 会 cd，转绝对路径
else
  echo "==> build docs SKIP (hugo 未安装，可执行 script/bootstrap-hugo.sh 下载)"
  exit 0
fi

echo "==> build docs ($HUGO)"
( cd docs/hugo && "$HUGO" --cleanDestinationDir )
echo "==> build docs OK → docs/hugo/public"
