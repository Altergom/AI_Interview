#!/usr/bin/env bash
# 全量启动：启用向量/关键词检索，拉起全套基础设施（含 etcd + Milvus + Elasticsearch）。
# 需要 8G+ 内存，首次启动会拉取较大镜像。
set -euo pipefail
cd "$(dirname "$0")"

if [ ! -f .env ]; then
  echo "[start-rag] .env 不存在，请先 cp .env.example .env 并填好 QWEN_API_KEY 等" >&2
  exit 1
fi

echo "[start-rag] 启动全套基础设施（含 milvus/es/etcd，较重，首次拉镜像耗时较长）..."
docker compose --profile rag up -d --wait postgres redis minio rabbitmq etcd milvus elasticsearch
docker compose up -d minio-init

export RAG_ENABLED=true

echo "[start-rag] RAG_ENABLED=true，启动应用（启用 Milvus + ES 检索）..."
go run ./cmd
