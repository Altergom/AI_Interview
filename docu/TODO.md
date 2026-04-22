# AI Interview TODO List

> 按模块拆分的开发任务清单，每个任务包含明确的输入、输出和技术要点，方便 AI Coding 辅助实现。

---

## 前端层（React Frontend）

### WebRTC / 音频采集
- [ ] 使用 `getUserMedia` 采集麦克风音频流
- [ ] 集成 VAD（语音活动检测），判断用户停止说话
- [ ] 音频分片，通过 WebRTC 或 AudioWorklet 实时传输至后端
- [ ] 设备权限异常处理（用户拒绝授权、设备不存在）

### SSE Client
- [ ] 建立 SSE 长连接，携带 `interview_id` 和 JWT token
- [ ] 断线自动重连，重连时恢复面试状态
- [ ] 接收文字流，实时渲染对话记录
- [ ] 接收 TTS 音频流，使用 Web Audio API 流式播放

### 代码编辑器
- [ ] 集成 Monaco Editor
- [ ] 支持语言切换（Java / Python / Go / C++）
- [ ] 实现提交按钮，HTTP POST 上传 `{code, interview_id, question_id}`
- [ ] 算法阶段显示编辑器，其他阶段隐藏

### 页面与路由
- [ ] 简历上传页：支持 PDF 上传，上传成功后跳转等待页
- [ ] 等待页：显示面试注意事项，等待简历解析完成
- [ ] 面试间页：左侧对话记录 + 右侧代码编辑器（算法阶段）+ 顶部阶段进度
- [ ] 问卷页：逐轮展示 ASR 文本，用户对每轮打标 good/bad + 文字反馈
- [ ] 报告页：展示多维度评分雷达图 + 优劣势总结

---

## 网关层（Hertz Server）

- [ ] 初始化 Hertz 服务，配置路由
- [ ] 实现 HAuth 中间件，JWT 鉴权
- [ ] 实现 SSE 端点 `GET /interview/stream`，维护长连接
- [ ] 实现 HTTP POST 端点 `POST /interview/code/submit`，接收代码提交
- [ ] 实现 Request Dispatcher，根据请求类型通过 Kitex RPC 转发至对应服务
- [ ] 全局错误处理与日志中间件

---

## 业务服务层

### Interview Service
- [ ] 定义 Kitex RPC IDL（创建面试、推进面试、结束面试）
- [ ] 实现 `CreateInterview`：初始化面试状态写入 Redis，关联 user_id 和 resume
- [ ] 实现 `ProcessTurn`：接收 ASR 文本，调用 Eino Core Graph，返回 AI 回复
- [ ] 实现 `SubmitCode`：接收代码，转发至 Eino Core Code Judge Agent
- [ ] 实现 `FinishInterview`：更新面试状态，向 MQ 发布 `interview_finished` 消息
- [ ] 面试状态写入 Redis，key: `interview:{interview_id}:state`

**Redis 面试状态结构**
```json
{
  "interview_id": "xxx",
  "stage": "questioning",
  "questions_asked": 3,
  "current_question_followups": 1,
  "started_at": "2024-01-01T10:00:00Z"
}
```

---

### User / Resume Service
- [ ] 定义 Kitex RPC IDL（注册、登录、上传简历、获取简历）
- [ ] 实现用户注册/登录，JWT 签发
- [ ] 实现简历上传：接收 PDF → 存 S3 → 异步触发信息提取 Agent
- [ ] 实现 `GetResume`：从 Redis 返回结构化简历 JSON
- [ ] 简历解析结果存 Redis，key: `resume:{user_id}`，TTL 7天

---

### Record / Storage Service
- [ ] 定义 Kitex RPC IDL（存储对话轮次、查询面试记录）
- [ ] 实现 `SaveTurn`：将每轮 ASR 文本存入 PostgreSQL `interview_turns` 表
- [ ] 实现 `GetInterviewRecord`：按 interview_id 返回完整对话记录
- [ ] 音频文件上传至 S3，路径: `/audio/{interview_id}/{turn_id}.wav`

**PostgreSQL interview_turns 表结构**
```sql
CREATE TABLE interview_turns (
  id UUID PRIMARY KEY,
  interview_id UUID NOT NULL,
  turn_id VARCHAR(10) NOT NULL,
  stage VARCHAR(50) NOT NULL,
  question TEXT,
  user_answer TEXT,
  asr_raw TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);
```

---

### Report Worker
- [ ] 消费 MQ `interview_finished` 消息
- [ ] 调用 Record Service 拉取完整对话记录
- [ ] 构造评分 prompt，调用 LLM 生成多维度评分报告
- [ ] 评分结果存 PostgreSQL `reports` 表
- [ ] 报告生成完成后通过 SSE 或 WebSocket 推送通知前端

**PostgreSQL reports 表结构**
```sql
CREATE TABLE reports (
  id UUID PRIMARY KEY,
  interview_id UUID NOT NULL,
  knowledge_depth INT,
  expression INT,
  problem_solving INT,
  code_quality INT,
  stress_response INT,
  summary TEXT,
  weak_points JSONB,
  strong_points JSONB,
  created_at TIMESTAMP DEFAULT NOW()
);
```

---

### 问卷系统
- [ ] 定义问卷提交 API `POST /questionnaire/submit`
- [ ] 接收用户对每轮对话的标注（good/bad + 文字反馈）
- [ ] 存入 PostgreSQL `questionnaire_results` 表

**PostgreSQL questionnaire_results 表结构**
```sql
CREATE TABLE questionnaire_results (
  id UUID PRIMARY KEY,
  interview_id UUID NOT NULL,
  turn_id VARCHAR(10) NOT NULL,
  quality VARCHAR(10) NOT NULL, -- 'good' | 'bad'
  feedback TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);
```

- [ ] 实现定期 job，筛选 `quality='good'` 的数据，生成 JSONL 文件上传 S3

**JSONL 格式**
```json
{
  "messages": [
    {"role": "system", "content": "你是一个技术面试官..."},
    {"role": "assistant", "content": "面试官的问题"},
    {"role": "user", "content": "用户的回答"},
    {"role": "assistant", "content": "面试官的追问"}
  ],
  "quality": "good",
  "user_feedback": "追问很有启发"
}
```

---

## AI层（Eino Core）

### ASR Node
- [ ] 接入流式 ASR SDK（火山引擎 / 阿里云）
- [ ] 实现音频流分片输入，实时输出文字
- [ ] 集成 VAD，检测说话结束，输出完整句子至 Router Node
- [ ] 异常处理：ASR 识别失败时返回错误提示

---

### Router Node
- [ ] 从 Redis 读取当前面试状态
- [ ] 实现阶段切换逻辑：

```
intro → questioning:    自我介绍结束（VAD检测停止说话）
questioning → algorithm: questions_asked >= N（可配置）
algorithm → closing:    代码提交且追问结束
closing → end:          用户说结束
```

- [ ] 实现题目状态机：

```
出题 → 等待回答 → 判断追问（followups < 2）→ 追问
                              ↓（followups >= 2）
                           关闭当前题 → RAG检索 → 出下一题
```

- [ ] 判断是否为代码提交事件，触发 Code Judge Agent
- [ ] 更新 Redis 面试状态

---

### 信息提取 Agent
- [ ] 实现简历解析：PDF 文本 → 结构化 JSON（技术栈、项目、实习、教育）
- [ ] 实现自我介绍补充提取：ASR 文本 → 提取额外信息 → merge 进 Redis 上下文
- [ ] prompt 指定输出严格 JSON 格式，做好解析异常处理

---

### RAG 检索
- [ ] 初始化 VectorDB（Milvus / Qdrant），导入八股题库
- [ ] 题目 embedding，按技术标签建立索引
- [ ] 实现检索接口：输入用户技术栈 → 返回 Top-K 候选题目
- [ ] 题库维护脚本：支持新增、更新、删除题目

**题库数据结构**
```json
{
  "question_id": "xxx",
  "question": "讲一下HashMap的扩容机制",
  "tags": ["Java", "集合", "中等"],
  "standard_answer": "...",
  "follow_up_hints": [
    "线程安全怎么处理",
    "和ConcurrentHashMap区别"
  ]
}
```

---

### Interview Agent
- [ ] 设计提问阶段 system prompt（面试官角色、追问方向参考 follow_up_hints）
- [ ] 设计反问阶段 system prompt（公司面试官角色扮演）
- [ ] 实现对话 history 管理：从 Redis 读取 → 拼接 → 调用 LLM → 回写 Redis
- [ ] 接收 Code Judge Agent 结构化结果，根据 `correctness` 决定追问方向：
    - `correctness=true` → 追问复杂度优化
    - `correctness=false` → 引导用户找 bug
- [ ] 对话 history 超长时做裁剪（保留 system prompt + 最近 N 轮）

---

### Code Judge Agent
- [ ] 接收代码文本 + 题目信息
- [ ] 设计评估 prompt，指定输出结构化 JSON
- [ ] 返回结构化评估结果至 Interview Agent

**输出结构**
```json
{
  "correctness": true,
  "time_complexity": "O(n)",
  "space_complexity": "O(1)",
  "issues": ["边界条件未处理"]
}
```

---

### LLM Node
- [ ] 接入 Doubao API，开启 streaming 模式
- [ ] 第一个 token 输出时立即触发 TTS Node（流水线并行）
- [ ] 备用 GPT-4o 接入，支持快速切换

---

### TTS Node
- [ ] 接入流式 TTS SDK（火山引擎 / 阿里云）
- [ ] 接收 LLM 流式文字，实时合成音频
- [ ] 通过 SSE 将音频流推送至前端

---

## 存储层

### PostgreSQL
- [ ] 初始化数据库，创建所有表（users / interviews / interview_turns / reports / questionnaire_results）
- [ ] 配置连接池
- [ ] 编写 migration 脚本

### Redis
- [ ] 初始化 Redis 连接
- [ ] 封装面试状态读写方法
- [ ] 封装对话 history 读写方法（list 结构，append + 裁剪）
- [ ] 封装结构化简历读写方法
- [ ] 配置合理 TTL

### S3 / Object Storage
- [ ] 配置 S3 客户端
- [ ] 实现简历上传方法
- [ ] 实现音频文件上传方法
- [ ] 实现 JSONL 文件上传方法
- [ ] 配置 Bucket 权限策略

---

## 消息队列

- [ ] 选型并初始化 MQ（Kafka / RabbitMQ）
- [ ] 定义 Topic: `interview_finished`
- [ ] Interview Service 实现 Producer，面试结束时发布消息
- [ ] Report Worker 实现 Consumer，消费消息触发报告生成
- [ ] 消息消费失败时实现重试机制

---

## 基础设施

- [ ] 编写 Docker Compose，本地一键启动所有服务（PostgreSQL / Redis / S3 / MQ）
- [ ] 配置各服务环境变量（API Keys / DB连接 / MQ地址）
- [ ] 编写各服务 Dockerfile
- [ ] 配置日志采集
- [ ] 配置健康检查端点

---

## 开发顺序建议

```
1. 存储层初始化（PostgreSQL / Redis / S3）
2. User/Resume Service（用户注册登录、简历上传）
3. 信息提取 Agent（简历解析）
4. ASR Node + TTS Node（语音链路跑通）
5. Interview Agent + LLM Node（基础对话跑通）
6. Router Node（阶段状态机）
7. RAG 检索 + 题库导入
8. Code Judge Agent + 代码编辑器前端
9. Record Service + 问卷系统
10. MQ + Report Worker（异步报告生成）
11. 前端完整页面
12. 流式 pipeline 联调（ASR→LLM→TTS 延迟优化）
```