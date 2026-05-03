# 环境变量配置说明

## 配置文件

- `.env.example` - 配置模板（提交到版本库）
- `.env` - 实际配置（不提交到版本库，包含敏感信息）

## 使用方式

```bash
# 复制模板
cp .env.example .env

# 编辑配置
vim .env
```

---

## 配置项说明

### 基本配置

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `APP_ENV` | 运行环境 | `development` | `development` / `staging` / `production` |
| `LOG_LEVEL` | 日志级别 | `info` | `debug` / `info` / `warn` / `error` |
| `LOG_FORMAT` | 日志格式 | `text` | `text`（本地）/ `json`（生产） |

### HTTP 服务

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `HTTP_ADDR` | HTTP 监听地址 | `:8080` | `:8080` / `0.0.0.0:8080` |

### 数据库

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `POSTGRES_DSN` | PostgreSQL 连接字符串 | - | `postgres://user:pass@host:5432/dbname` |
| `REDIS_ADDR` | Redis 地址 | `127.0.0.1:6379` | `127.0.0.1:6379` |
| `REDIS_PASSWORD` | Redis 密码 | - | `your_password` |
| `REDIS_DB` | Redis 数据库编号 | `0` | `0` |

### 消息队列

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `MQ_BROKER_URL` | 消息队列连接地址 | - | `amqp://guest:guest@localhost:5672/` |
| `MQ_TOPIC_INTERVIEW_FINISHED` | 面试完成事件主题 | `interview_finished` | `interview_finished` |

### 对象存储（S3 兼容）

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `S3_ENDPOINT` | S3 端点 | - | `s3.amazonaws.com` / `localhost:9000` |
| `S3_ACCESS_KEY` | S3 访问密钥 | - | `your_access_key` |
| `S3_SECRET_KEY` | S3 密钥 | - | `your_secret_key` |
| `S3_BUCKET` | S3 存储桶 | - | `ai-interview` |
| `S3_REGION` | S3 区域 | `us-east-1` | `us-east-1` / `cn-north-1` |
| `S3_USE_SSL` | 是否使用 SSL | `true` | `true` / `false` |

### JWT 认证

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `JWT_SECRET` | JWT 密钥 | - | **生产环境必须设置强随机字符串** |
| `JWT_ISSUER` | JWT 签发者 | `ai_interview` | `ai_interview` |
| `JWT_ACCESS_EXP_MIN` | 访问令牌过期时间（分钟） | `60` | `60` |

### LLM 模型提供商

#### OpenAI

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `OPENAI_API_KEY` | OpenAI API Key | - | `sk-...` |
| `OPENAI_BASE_URL` | OpenAI API 地址 | - | `https://api.openai.com/v1` |

#### 豆包（Doubao）

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `DOUBAO_API_KEY` | 豆包 API Key | - | `...` |
| `DOUBAO_BASE_URL` | 豆包 API 地址 | - | `...` |

#### 千问（Qwen）⭐ 推荐

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `QWEN_API_KEY` | 千问 API Key | - | `sk-...` |
| `QWEN_BASE_URL` | 千问 API 地址 | - | `https://dashscope.aliyuncs.com/compatible-mode/v1` |

**说明**：
- 千问用于 Agent（Supervisor/Selector/Analyzer/Manager/Evaluator）
- 千问用于 ASR（语音识别）和 TTS（语音合成）
- 需要在阿里云百炼平台申请 API Key

#### Claude

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `CLAUDE_API_KEY` | Claude API Key | - | `sk-ant-...` |
| `CLAUDE_BASE_URL` | Claude API 地址 | - | `https://api.anthropic.com` |

#### DeepSeek

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `DEEPSEEK_API_KEY` | DeepSeek API Key | - | `...` |
| `DEEPSEEK_BASE_URL` | DeepSeek API 地址 | - | `https://api.deepseek.com` |

#### Gemini

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `GEMINI_API_KEY` | Gemini API Key | - | `...` |
| `GEMINI_BASE_URL` | Gemini API 地址 | - | `https://generativelanguage.googleapis.com` |

### Agent 模型配置

| 环境变量 | 说明 | 默认值 | 推荐值 |
|---------|------|--------|--------|
| `SUPERVISOR` | Supervisor Agent 模型 | `qwen-plus` | `qwen-plus` / `qwen-max` |
| `SELECTOR` | Question Selector Agent 模型 | `qwen-plus` | `qwen-plus` |
| `MANAGER` | Stage Manager Agent 模型 | `qwen-plus` | `qwen-plus` |
| `ANALYZER` | Response Analyzer Agent 模型 | `qwen-plus` | `qwen-plus` |
| `EVALUATOR` | Evaluator Agent 模型 | `qwen-plus` | `qwen-plus` |

**可选模型**：
- `qwen-turbo` - 快速，成本低
- `qwen-plus` - 平衡性能和成本（推荐）
- `qwen-max` - 最强性能，成本高

### ASR/TTS 模型配置

| 环境变量 | 说明 | 默认值 | 可选值 |
|---------|------|--------|--------|
| `ASR_MODEL` | 语音识别模型 | `qwen3-asr-flash-realtime` | `qwen3-asr-flash-realtime` |
| `TTS_MODEL` | 语音合成模型 | `qwen-tts-flash-realtime` | `qwen-tts-flash-realtime` |
| `TTS_VOICE` | TTS 音色 | `zhifeng_emo` | `zhifeng_emo` / `zhiyan_emo` / `zhiyu_emo` |

**ASR 模型说明**：
- `qwen3-asr-flash-realtime` - 千问3 ASR 快速实时版（稳定版）
- 支持多语言、多方言识别
- 支持情感识别

**TTS 音色说明**：
- `zhifeng_emo` - 知风（男声，支持情感）
- `zhiyan_emo` - 知言（女声，支持情感）
- `zhiyu_emo` - 知语（女声，支持情感）

### Redis TTL 配置

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `RESUME_REDIS_TTL` | 简历缓存过期时间 | `168h` | `168h`（7天）/ `720h`（30天） |
| `INTERVIEW_STATE_TTL` | 面试状态过期时间 | `48h` | `48h`（2天）/ `168h`（7天） |

**格式说明**：使用 Go 的 `time.ParseDuration` 格式
- `h` - 小时
- `m` - 分钟
- `s` - 秒
- 示例：`24h`、`30m`、`1h30m`

---

## 快速开始

### 最小配置（本地开发）

```bash
# .env
APP_ENV=development
LOG_LEVEL=debug

# 数据库
POSTGRES_DSN=postgres://postgres:postgres@127.0.0.1:5432/ai_interview?sslmode=disable
REDIS_ADDR=127.0.0.1:6379

# 千问（必需）
QWEN_API_KEY=your_qwen_api_key
QWEN_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1

# Agent 模型
SUPERVISOR=qwen-plus
SELECTOR=qwen-plus
MANAGER=qwen-plus
ANALYZER=qwen-plus
EVALUATOR=qwen-plus

# ASR/TTS 模型
ASR_MODEL=qwen3-asr-flash-realtime
TTS_MODEL=qwen-tts-flash-realtime
TTS_VOICE=zhifeng_emo
```

### 生产环境配置

```bash
# .env
APP_ENV=production
LOG_LEVEL=info
LOG_FORMAT=json

# 数据库（使用生产环境地址）
POSTGRES_DSN=postgres://user:pass@prod-db:5432/ai_interview?sslmode=require
REDIS_ADDR=prod-redis:6379
REDIS_PASSWORD=strong_password

# JWT（使用强随机密钥）
JWT_SECRET=your-very-strong-random-secret-key-here

# 千问
QWEN_API_KEY=your_production_qwen_api_key
QWEN_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1

# 对象存储
S3_ENDPOINT=s3.amazonaws.com
S3_ACCESS_KEY=your_access_key
S3_SECRET_KEY=your_secret_key
S3_BUCKET=ai-interview-prod
S3_REGION=us-east-1
S3_USE_SSL=true
```

---

## 注意事项

1. **不要提交 .env 文件到版本库**
   - `.env` 包含敏感信息（API Key、密码等）
   - 已在 `.gitignore` 中排除

2. **生产环境必须设置强密钥**
   - `JWT_SECRET` 必须使用强随机字符串
   - 数据库密码必须足够复杂

3. **千问 API Key 申请**
   - 访问：https://bailian.console.aliyun.com/
   - 创建 API Key
   - 确保有足够的额度

4. **模型选择建议**
   - 开发环境：使用 `qwen-plus`（性价比高）
   - 生产环境：根据实际需求选择 `qwen-plus` 或 `qwen-max`

---

## 故障排查

### 配置未生效

```bash
# 检查环境变量是否加载
go run cmd/server/main.go

# 查看日志中的配置信息
```

### API Key 无效

```bash
# 检查 API Key 是否正确
echo $QWEN_API_KEY

# 测试 API Key
curl -H "Authorization: Bearer $QWEN_API_KEY" \
  https://dashscope.aliyuncs.com/compatible-mode/v1/models
```

### 数据库连接失败

```bash
# 测试 PostgreSQL 连接
psql "$POSTGRES_DSN"

# 测试 Redis 连接
redis-cli -h 127.0.0.1 -p 6379 ping
```
