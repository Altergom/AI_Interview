#!/usr/bin/env bash
# 离线采集：拉取小林 CS-Base 仓库 → LLM ingest → 生成 wiki/questions + index.md。
# 这是构建 wiki 知识库的一次性/增量操作，不是启动服务。需要时手动跑，不要塞进 start-wiki.sh。
#
# 用法：
#   bash collect-wiki.sh                 # 全量 ingest（约 123 篇，100~150 万 token，5~15 分钟）
#   bash collect-wiki.sh --dry-run       # 只解析不写库，先看规模
#   bash collect-wiki.sh --limit 5       # 只导入前 5 篇，验证流程
#   bash collect-wiki.sh --category 网络  # 只导入指定目录
# 其余 collector flag（--repo-url / --repo-branch / --local-path）均可透传。
set -euo pipefail
cd "$(dirname "$0")"

if [ ! -f .env ]; then
  echo "[collect-wiki] .env 不存在，请先 cp .env.example .env 并填好 QWEN_API_KEY 等" >&2
  exit 1
fi

echo "[collect-wiki] 启动采集所需基础设施（postgres redis minio rabbitmq）..."
docker compose up -d --wait postgres redis minio rabbitmq
docker compose up -d minio-init

echo "[collect-wiki] 开始采集（首次会 git clone 小林 CS-Base 到 internal/wiki/raw/）..."
go run ./tools/collector "$@"

echo "[collect-wiki] 完成。产物在 internal/wiki/questions/ 与 internal/wiki/index.md"
