# AI Interview API 文档

> 本文档描述所有 HTTP 端点、SSE 事件格式及 Kitex RPC IDL 定义。
> Base URL: `https://api.aiinterview.com/v1`
> 所有请求需携带 Header: `Authorization: Bearer {jwt_token}`（游客模式除外）

---

## 目录

1. [认证模块](#认证模块)
2. [设备检测模块](#设备检测模块)
3. [简历模块](#简历模块)
4. [面试配置模块](#面试配置模块)
5. [面试模块](#面试模块)
6. [代码提交模块](#代码提交模块)
7. [报告模块](#报告模块)
8. [问卷模块](#问卷模块)
9. [SSE 事件格式](#sse-事件格式)
10. [Kitex RPC IDL](#kitex-rpc-idl)
11. [错误码](#错误码)

---

## 认证模块

### 注册

```
POST /auth/register
```

**Request Body**
```json
{
  "username": "张三",
  "email": "zhangsan@example.com",
  "password": "xxxxxxxx"
}
```

**Response 200**
```json
{
  "code": 0,
  "data": {
    "user_id": "uuid",
    "username": "张三",
    "token": "jwt_token"
  }
}
```

---

### 登录

```
POST /auth/login
```

**Request Body**
```json
{
  "email": "zhangsan@example.com",
  "password": "xxxxxxxx"
}
```

**Response 200**
```json
{
  "code": 0,
  "data": {
    "user_id": "uuid",
    "username": "张三",
    "token": "jwt_token"
  }
}
```

---

### 游客模式

```
POST /auth/guest
```

**Request Body**
```json
{}
```

**Response 200**
```json
{
  "code": 0,
  "data": {
    "user_id": "guest_uuid",
    "token": "jwt_token",
    "expires_at": "2024-01-02T10:00:00Z"
  }
}
```

**说明**
- 生成临时游客账号，有效期 24 小时
- 游客数据（简历、面试记录）24 小时后自动清理
- 游客可随时通过注册转为正式用户

---

## 设备检测模块

### 检测设备状态

```
POST /device/check
```

**Request Body**
```json
{
  "has_microphone": true,
  "has_camera": true,
  "browser": "Chrome 120",
  "os": "Windows 11"
}
```

**Response 200**
```json
{
  "code": 0,
  "data": {
    "status": "ready",
    "message": "设备检测通过，可以开始面试"
  }
}
```

**说明**
- 前端完成设备权限获取后，调用此接口确认设备状态
- 服务端记录设备信息用于后续故障排查
- 如果设备检测失败，用户可选择重试或退出

---

## 简历模块

### 解析简历（辅助功能）

```
POST /resume/parse
Content-Type: multipart/form-data
```

**Request Form**
| 字段 | 类型 | 说明 |
|---|---|---|
| file | File | PDF 格式简历文件 |

**Response 200**
```json
{
  "code": 0,
  "data": {
    "skills": ["Java", "Redis", "MySQL"],
    "projects": [
      {
        "name": "电商平台",
        "tech_stack": ["SpringBoot", "Redis"],
        "description": "负责订单模块开发",
        "highlights": ["QPS提升3倍", "响应时间降低50%"]
      }
    ],
    "internships": [
      {
        "company": "字节跳动",
        "position": "后端开发实习生",
        "duration": "2023.06-2023.09"
      }
    ],
    "education": {
      "school": "xxx大学",
      "major": "计算机科学与技术",
      "degree": "本科",
      "graduation": "2025-06"
    }
  }
}
```

**说明**
- 上传 PDF 后立即返回解析结果（同步）
- 前端将解析结果填充到表单栏框
- 用户可以查看、修改后再提交
- 解析失败返回错误码 3001，用户可选择手动填写

---

### 提交简历信息

```
POST /resume/submit
```

**Request Body**
```json
{
  "user_id": "uuid",
  "skills": ["Java", "Redis", "MySQL"],
  "projects": [
    {
      "name": "电商平台",
      "tech_stack": ["SpringBoot", "Redis"],
      "description": "负责订单模块开发",
      "highlights": ["QPS提升3倍"]
    }
  ],
  "internships": [
    {
      "company": "字节跳动",
      "position": "后端开发实习生",
      "duration": "2023.06-2023.09"
    }
  ],
  "education": {
    "school": "xxx大学",
    "major": "计算机科学与技术",
    "degree": "本科",
    "graduation": "2025-06"
  }
}
```

**Response 200**
```json
{
  "code": 0,
  "data": {
    "resume_id": "uuid",
    "message": "简历信息保存成功"
  }
}
```

**说明**
- 无论是手动填写还是 PDF 解析，最终都调用此接口提交
- 提交的是用户确认后的最终数据
- 简历信息存入 Redis，key: `resume:{user_id}`，TTL 7天

---

## 面试配置模块

### 选择面试岗位和方向

```
POST /interview/config
```

**Request Body**
```json
{
  "user_id": "uuid",
  "position": "golang",
  "direction": "backend"
}
```

**position 枚举**
| 值 | 说明 |
|---|---|
| golang | Golang开发 |
| java | Java开发 |
| frontend | 前端工程师 |
| test | 测试开发 |

**direction 枚举**
| 值 | 说明 |
|---|---|
| backend | 软件开发 |
| cloud | 云平台运维开发 |
| agent | Agent开发 |
| server | 服务端开发 |

**Response 200**
```json
{
  "code": 0,
  "data": {
    "config_id": "uuid",
    "message": "配置保存成功"
  }
}
```

**说明**
- 岗位决定题库范围
- 方向影响面试侧重点
- 配置信息存入 Redis，关联 user_id

---

## 面试模块

### 创建面试

```
POST /interview/create
```

**Request Body**
```json
{
  "user_id": "uuid"
}
```

**Response 200**
```json
{
  "code": 0,
  "data": {
    "interview_id": "uuid",
    "stage": "intro",
    "created_at": "2024-01-01T10:00:00Z"
  }
}
```

---

### 建立 SSE 连接

```
GET /interview/stream?interview_id={interview_id}
Accept: text/event-stream
```

**说明**
- 建立后服务端通过此连接推送所有 AI 回复（文字流 + 音频流）
- 断线重连时携带相同 `interview_id`，服务端自动恢复面试状态
- SSE 事件格式见 [SSE 事件格式](#sse-事件格式)

---

### 发送音频流

```
POST /interview/audio
Content-Type: application/octet-stream
```

**Request Headers**
| Header | 说明 |
|---|---|
| X-Interview-Id | 当前面试 ID |
| X-Turn-Id | 当前轮次 ID |

**Request Body**
- 原始音频数据（PCM / WebM）

**Response 200**
```json
{
  "code": 0,
  "data": {
    "turn_id": "uuid",
    "status": "received"
  }
}
```

**说明**
- 前端通过 VAD 检测说话结束后，将本轮音频数据 POST 至此接口
- 服务端触发 ASR → Router → Interview Agent 链路
- AI 回复通过 SSE 推送，不在此接口同步返回

---

### 结束面试

```
POST /interview/finish
```

**Request Body**
```json
{
  "interview_id": "uuid"
}
```

**Response 200**
```json
{
  "code": 0,
  "data": {
    "interview_id": "uuid",
    "finished_at": "2024-01-01T11:00:00Z",
    "duration_seconds": 3600
  }
}
```

**说明**
- 调用后 Interview Service 向 MQ 发布 `interview_finished` 消息
- Report Worker 异步消费，生成评分报告

---

### 查询面试状态

```
GET /interview/state?interview_id={interview_id}
```

**Response 200**
```json
{
  "code": 0,
  "data": {
    "interview_id": "uuid",
    "stage": "questioning",
    "questions_asked": 3,
    "current_question_followups": 1,
    "started_at": "2024-01-01T10:00:00Z"
  }
}
```

**stage 枚举**
| 值 | 说明 |
|---|---|
| intro | 自我介绍阶段 |
| questioning | 技术提问阶段 |
| algorithm | 算法题阶段 |
| closing | 反问阶段 |
| finished | 面试结束 |

---

## 代码提交模块

### 提交代码

```
POST /interview/code/submit
```

**Request Body**
```json
{
  "interview_id": "uuid",
  "question_id": "uuid",
  "language": "java",
  "code": "class Solution {\n    public int[] twoSum...\n}"
}
```

**language 枚举**
| 值 | 说明 |
|---|---|
| java | Java |
| python | Python |
| go | Go |
| cpp | C++ |

**Response 200**
```json
{
  "code": 0,
  "data": {
    "status": "judging",
    "message": "代码评估中"
  }
}
```

**说明**
- 代码提交后触发 Code Judge Agent → Interview Agent 链路
- 评估结果和 AI 追问通过 SSE 推送，不在此接口同步返回

---

## 问卷模块

### 获取问卷（面试结束后）

```
GET /questionnaire?interview_id={interview_id}
```

**Response 200**
```json
{
  "code": 0,
  "data": {
    "interview_id": "uuid",
    "turns": [
      {
        "turn_id": "001",
        "stage": "questioning",
        "question": "请介绍一下HashMap的扩容机制",
        "user_answer": "用户的回答文本..."
      },
      {
        "turn_id": "002",
        "stage": "questioning",
        "question": "你提到了负载因子，那如果调成0.5会有什么影响？",
        "user_answer": "用户的回答文本..."
      }
    ]
  }
}
```

---

### 提交问卷

```
POST /questionnaire/submit
```

**Request Body**
```json
{
  "interview_id": "uuid",
  "answers": [
    {
      "turn_id": "001",
      "quality": "good",
      "feedback": "追问很有启发"
    },
    {
      "turn_id": "002",
      "quality": "bad",
      "feedback": "追问方向感觉不对"
    }
  ]
}
```

**quality 枚举**
| 值 | 说明 |
|---|---|
| good | 这轮对话质量好 |
| bad | 这轮对话质量差 |

**Response 200**
```json
{
  "code": 0,
  "data": {
    "message": "问卷提交成功，感谢你的反馈"
  }
}
```

---

## 报告模块

### 查询报告状态

```
GET /report/status?interview_id={interview_id}
```

**Response 200**
```json
{
  "code": 0,
  "data": {
    "status": "generating",
    "message": "报告生成中，预计需要1-2分钟"
  }
}
```

**status 枚举**
| 值 | 说明 |
|---|---|
| pending | 等待生成 |
| generating | 生成中 |
| done | 生成完成 |
| failed | 生成失败 |

---

### 获取报告

```
GET /report?interview_id={interview_id}
```

**Response 200**
```json
{
  "code": 0,
  "data": {
    "interview_id": "uuid",
    "dimensions": {
      "knowledge_depth": 8,
      "expression": 7,
      "problem_solving": 9,
      "code_quality": 8,
      "stress_response": 6
    },
    "summary": "候选人基础扎实，算法能力突出，建议加强系统设计深度。",
    "strong_points": ["算法思路清晰", "表达流畅", "代码风格规范"],
    "weak_points": ["系统设计深度不足", "边界条件考虑不周"],
    "created_at": "2024-01-01T11:05:00Z"
  }
}
```

---

## SSE 事件格式

所有 SSE 事件通过 `GET /interview/stream` 推送，格式如下：

```
event: {event_type}
data: {json_payload}
```

---

### `text.delta` 文字流（AI 回复逐字输出）

```
event: text.delta
data: {"turn_id": "uuid", "delta": "你提到了"}
```

---

### `text.done` 文字流结束

```
event: text.done
data: {"turn_id": "uuid", "full_text": "你提到了负载因子，那如果调成0.5会有什么影响？"}
```

---

### `audio.delta` 音频流（TTS 流式推送）

```
event: audio.delta
data: {"turn_id": "uuid", "audio_base64": "base64编码音频数据"}
```

---

### `audio.done` 音频流结束

```
event: audio.done
data: {"turn_id": "uuid"}
```

---

### `stage.changed` 面试阶段切换

```
event: stage.changed
data: {"from": "intro", "to": "questioning"}
```

---

### `code.judged` 代码评估完成

```
event: code.judged
data: {
  "correctness": true,
  "time_complexity": "O(n)",
  "space_complexity": "O(1)",
  "issues": ["边界条件未处理"]
}
```

---

### `resume.parsed` 简历解析完成

```
event: resume.parsed
data: {"status": "done"}
```

---

### `report.ready` 报告生成完成

```
event: report.ready
data: {"interview_id": "uuid"}
```

---

### `interview.finished` 面试结束确认

```
event: interview.finished
data: {"interview_id": "uuid", "finished_at": "2024-01-01T11:00:00Z"}
```

---

### `error` 错误事件

```
event: error
data: {"code": 5001, "message": "ASR识别失败，请重试"}
```

---

## Kitex RPC IDL

### Interview Service

```thrift
namespace go interview

struct CreateInterviewRequest {
  1: required string user_id
}

struct CreateInterviewResponse {
  1: required string interview_id
  2: required string stage
  3: required string created_at
}

struct ProcessTurnRequest {
  1: required string interview_id
  2: required string turn_id
  3: required string asr_text
}

struct ProcessTurnResponse {
  1: required string reply_text
  2: required string stage
}

struct SubmitCodeRequest {
  1: required string interview_id
  2: required string question_id
  3: required string language
  4: required string code
}

struct SubmitCodeResponse {
  1: required string status
}

struct FinishInterviewRequest {
  1: required string interview_id
}

struct FinishInterviewResponse {
  1: required string interview_id
  2: required string finished_at
  3: required i64 duration_seconds
}

service InterviewService {
  CreateInterviewResponse CreateInterview(1: CreateInterviewRequest req)
  ProcessTurnResponse ProcessTurn(1: ProcessTurnRequest req)
  SubmitCodeResponse SubmitCode(1: SubmitCodeRequest req)
  FinishInterviewResponse FinishInterview(1: FinishInterviewRequest req)
}
```

---

### User / Resume Service

```thrift
namespace go user

struct RegisterRequest {
  1: required string username
  2: required string email
  3: required string password
}

struct RegisterResponse {
  1: required string user_id
  2: required string token
}

struct LoginRequest {
  1: required string email
  2: required string password
}

struct LoginResponse {
  1: required string user_id
  2: required string token
}

struct GetResumeRequest {
  1: required string user_id
}

struct Project {
  1: required string name
  2: required list<string> tech_stack
  3: required string description
  4: required list<string> highlights
}

struct Education {
  1: required string school
  2: required string major
  3: required string graduation
}

struct GetResumeResponse {
  1: required string user_id
  2: required list<string> skills
  3: required list<Project> projects
  4: required list<string> internships
  5: required Education education
}

service UserService {
  RegisterResponse Register(1: RegisterRequest req)
  LoginResponse Login(1: LoginRequest req)
  GetResumeResponse GetResume(1: GetResumeRequest req)
}
```

---

### Record / Storage Service

```thrift
namespace go record

struct SaveTurnRequest {
  1: required string interview_id
  2: required string turn_id
  3: required string stage
  4: required string question
  5: required string user_answer
  6: required string asr_raw
}

struct SaveTurnResponse {
  1: required bool success
}

struct Turn {
  1: required string turn_id
  2: required string stage
  3: required string question
  4: required string user_answer
  5: required string asr_raw
  6: required string created_at
}

struct GetInterviewRecordRequest {
  1: required string interview_id
}

struct GetInterviewRecordResponse {
  1: required string interview_id
  2: required list<Turn> turns
}

service RecordService {
  SaveTurnResponse SaveTurn(1: SaveTurnRequest req)
  GetInterviewRecordResponse GetInterviewRecord(1: GetInterviewRecordRequest req)
}
```

---

## 错误码

| 错误码 | 说明 |
|---|---|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 未授权，token 无效或过期 |
| 2001 | 用户不存在 |
| 2002 | 邮箱已注册 |
| 3001 | 简历解析失败 |
| 3002 | 简历格式不支持（仅支持 PDF） |
| 4001 | 面试不存在 |
| 4002 | 面试已结束，无法继续操作 |
| 4003 | 当前阶段不支持代码提交 |
| 5001 | ASR 识别失败 |
| 5002 | LLM 调用失败 |
| 5003 | TTS 合成失败 |
| 6001 | 报告生成中，尚未完成 |
| 6002 | 报告不存在 |