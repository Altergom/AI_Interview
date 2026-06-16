#!/usr/bin/env bash
# 轻量启动：出题走 LLM wiki（Skill middleware 读本地 SKILL.md），不启动 Milvus/ES。
# 只拉起必需基础设施：Postgres + Redis + MinIO + RabbitMQ。
set -euo pipefail
cd "$(dirname "$0")"

if [ ! -f .env ]; then
  echo "[start-wiki] .env 不存在，请先 cp .env.example .env 并填好 QWEN_API_KEY 等" >&2
  exit 1
fi

echo "[start-wiki] 启动基础设施（postgres redis minio rabbitmq）..."
docker compose up -d --wait postgres redis minio minio-init rabbitmq

# 强制关闭 RAG（export 优先于 .env，godotenv 不覆盖已存在的环境变量）
export RAG_ENABLED=false

echo "[start-wiki] RAG_ENABLED=false，启动应用（出题走 wiki）..."
go run ./cmd
