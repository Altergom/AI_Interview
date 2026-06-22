# AI Interview API

本文档按模块整理 HTTP API，并为每个接口提供入参/出参字段说明与可直接测试的 curl。

## 约定

- Base URL: `https://api.aiinterview.com/v1`
- 文档中接口路径不带 `/v1` 前缀（例如 `POST /auth/login`），实际请求请使用 `Base URL + Path`
- 认证：默认需要 `Authorization: Bearer {jwt_token}`，个别接口在说明中标注为“无需登录”
- 入参：
  - `POST` 默认 `Content-Type: application/json`，请求体为 JSON
  - `GET` 使用 Query 参数
  - 例外：文件/音频上传接口按各自章节的 `Content-Type` 执行
- 统一响应体（所有接口）：

成功：

```json
{
  "success": true,
  "data": {},
  "error": null
}
```

失败：

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": 1001,
    "message": "参数错误",
    "request_id": "req_123"
  }
}
```

## 认证模块

### 注册

- Method: `POST`
- Path: `/auth/register`
- Auth: 无需登录
- Content-Type: `application/json`

**入参（JSON Body）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| username | string | 是 | 用户名 | `张三` |
| email | string | 是 | 邮箱 | `zhangsan@example.com` |
| password | string | 是 | 密码 | `xxxxxxxx` |

**出参（data）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| user_id | string | 是 | 用户 ID | `uuid` |
| username | string | 是 | 用户名 | `张三` |
| token | string | 是 | JWT | `jwt_token` |

**curl**

```bash
curl -sS -X POST 'https://api.aiinterview.com/v1/auth/register' \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "张三",
    "email": "zhangsan@example.com",
    "password": "xxxxxxxx"
  }'
```

### 登录

- Method: `POST`
- Path: `/auth/login`
- Auth: 无需登录
- Content-Type: `application/json`

**入参（JSON Body）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| email | string | 是 | 邮箱 | `zhangsan@example.com` |
| password | string | 是 | 密码 | `xxxxxxxx` |

**出参（data）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| user_id | string | 是 | 用户 ID | `uuid` |
| username | string | 是 | 用户名 | `张三` |
| token | string | 是 | JWT | `jwt_token` |

**curl**

```bash
curl -sS -X POST 'https://api.aiinterview.com/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "zhangsan@example.com",
    "password": "xxxxxxxx"
  }'
```

### 游客模式

- Method: `POST`
- Path: `/auth/guest`
- Auth: 无需登录
- Content-Type: `application/json`

**入参（JSON Body）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| (空对象) | object | 是 | 固定传 `{}` | `{}` |

**出参（data）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| user_id | string | 是 | 游客用户 ID | `guest_uuid` |
| token | string | 是 | JWT | `jwt_token` |
| expires_at | string | 是 | 过期时间（RFC3339） | `2024-01-02T10:00:00Z` |

**curl**

```bash
curl -sS -X POST 'https://api.aiinterview.com/v1/auth/guest' \
  -H 'Content-Type: application/json' \
  -d '{}'
```

## 设备检测模块

### 检测设备状态

- Method: `POST`
- Path: `/device/check`
- Auth: 需要登录
- Content-Type: `application/json`

**Headers**

| Header | 必填 | 说明 | 示例 |
|---|---:|---|---|
| Authorization | 是 | `Bearer {jwt_token}` | `Bearer xxx` |

**入参（JSON Body）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| has_microphone | boolean | 是 | 是否有麦克风权限 | `true` |
| has_camera | boolean | 是 | 是否有摄像头权限 | `true` |
| browser | string | 是 | 浏览器信息 | `Chrome 120` |
| os | string | 是 | 操作系统信息 | `Windows 11` |

**出参（data）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| status | string | 是 | 设备状态 | `ready` |
| message | string | 是 | 提示信息 | `设备检测通过，可以开始面试` |

**curl**

```bash
curl -sS -X POST 'https://api.aiinterview.com/v1/device/check' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer jwt_token' \
  -d '{
    "has_microphone": true,
    "has_camera": true,
    "browser": "Chrome 120",
    "os": "Windows 11"
  }'
```

## 简历模块

### 解析简历（文件上传）

- Method: `POST`
- Path: `/resume/parse`
- Auth: 需要登录
- Content-Type: `multipart/form-data`

**Headers**

| Header | 必填 | 说明 | 示例 |
|---|---:|---|---|
| Authorization | 是 | `Bearer {jwt_token}` | `Bearer xxx` |

**入参（Form）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| file | file | 是 | PDF 简历文件 | `resume.pdf` |

**出参（data）**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| skills | string[] | 是 | 技能列表 |
| projects | object[] | 是 | 项目经历 |
| internships | object[] | 是 | 实习经历 |
| education | object | 是 | 教育信息 |

`projects[]` 字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| name | string | 是 | 项目名称 |
| tech_stack | string[] | 是 | 技术栈 |
| description | string | 是 | 描述 |
| highlights | string[] | 是 | 亮点 |

`internships[]` 字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| company | string | 是 | 公司 |
| position | string | 是 | 岗位 |
| duration | string | 是 | 时间区间 |

`education` 字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| school | string | 是 | 学校 |
| major | string | 是 | 专业 |
| degree | string | 是 | 学历 |
| graduation | string | 是 | 毕业时间 |

**curl**

```bash
curl -sS -X POST 'https://api.aiinterview.com/v1/resume/parse' \
  -H 'Authorization: Bearer jwt_token' \
  -F 'file=@/path/to/resume.pdf'
```

### 提交简历信息

- Method: `POST`
- Path: `/resume/submit`
- Auth: 需要登录
- Content-Type: `application/json`

**Headers**

| Header | 必填 | 说明 | 示例 |
|---|---:|---|---|
| Authorization | 是 | `Bearer {jwt_token}` | `Bearer xxx` |

**入参（JSON Body）**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| user_id | string | 是 | 用户 ID |
| skills | string[] | 是 | 技能列表 |
| projects | object[] | 是 | 项目经历 |
| internships | object[] | 是 | 实习经历 |
| education | object | 是 | 教育信息 |

`projects[] / internships[] / education` 字段同 “解析简历”章节。

**出参（data）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| resume_id | string | 是 | 简历 ID | `uuid` |
| message | string | 是 | 提示信息 | `简历信息保存成功` |

**curl**

```bash
curl -sS -X POST 'https://api.aiinterview.com/v1/resume/submit' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer jwt_token' \
  -d '{
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
  }'
```

## 面试配置模块

### 选择面试岗位和方向

- Method: `POST`
- Path: `/interview/config`
- Auth: 需要登录
- Content-Type: `application/json`

**Headers**

| Header | 必填 | 说明 | 示例 |
|---|---:|---|---|
| Authorization | 是 | `Bearer {jwt_token}` | `Bearer xxx` |

**入参（JSON Body）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| user_id | string | 是 | 用户 ID | `uuid` |
| position | string | 是 | 岗位枚举 | `golang` |
| direction | string | 是 | 方向枚举 | `backend` |

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

**出参（data）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| config_id | string | 是 | 配置 ID | `uuid` |
| message | string | 是 | 提示信息 | `配置保存成功` |

**curl**

```bash
curl -sS -X POST 'https://api.aiinterview.com/v1/interview/config' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer jwt_token' \
  -d '{
    "user_id": "uuid",
    "position": "golang",
    "direction": "backend"
  }'
```

## 面试模块

### 创建面试

- Method: `POST`
- Path: `/interview/create`
- Auth: 需要登录
- Content-Type: `application/json`

**Headers**

| Header | 必填 | 说明 | 示例 |
|---|---:|---|---|
| Authorization | 是 | `Bearer {jwt_token}` | `Bearer xxx` |

**入参（JSON Body）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| user_id | string | 是 | 用户 ID | `uuid` |

**出参（data）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| interview_id | string | 是 | 面试 ID | `uuid` |
| stage | string | 是 | 当前阶段 | `intro` |
| created_at | string | 是 | 创建时间（RFC3339） | `2024-01-01T10:00:00Z` |

**curl**

```bash
curl -sS -X POST 'https://api.aiinterview.com/v1/interview/create' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer jwt_token' \
  -d '{
    "user_id": "uuid"
  }'
```

### 建立 SSE 连接

- Method: `GET`
- Path: `/interview/stream`
- Auth: 需要登录
- Accept: `text/event-stream`

**Query**

| 参数 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| interview_id | string | 是 | 面试 ID | `uuid` |

**curl**

```bash
curl -N -sS 'https://api.aiinterview.com/v1/interview/stream?interview_id=uuid' \
  -H 'Accept: text/event-stream' \
  -H 'Authorization: Bearer jwt_token'
```

### 发送音频流（音频上传）

- Method: `POST`
- Path: `/interview/audio`
- Auth: 需要登录
- Content-Type: `application/octet-stream`

**Headers**

| Header | 必填 | 说明 | 示例 |
|---|---:|---|---|
| Authorization | 是 | `Bearer {jwt_token}` | `Bearer xxx` |
| X-Interview-Id | 是 | 当前面试 ID | `uuid` |
| X-Turn-Id | 是 | 当前轮次 ID | `uuid` |

**入参（Body）**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| (raw) | bytes | 是 | 原始音频数据（PCM / WebM） |

**出参（data）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| turn_id | string | 是 | 轮次 ID | `uuid` |
| status | string | 是 | 接收状态 | `received` |

**curl**

```bash
curl -sS -X POST 'https://api.aiinterview.com/v1/interview/audio' \
  -H 'Authorization: Bearer jwt_token' \
  -H 'Content-Type: application/octet-stream' \
  -H 'X-Interview-Id: uuid' \
  -H 'X-Turn-Id: uuid' \
  --data-binary '@/path/to/audio.bin'
```

### 结束面试

- Method: `POST`
- Path: `/interview/finish`
- Auth: 需要登录
- Content-Type: `application/json`

**Headers**

| Header | 必填 | 说明 | 示例 |
|---|---:|---|---|
| Authorization | 是 | `Bearer {jwt_token}` | `Bearer xxx` |

**入参（JSON Body）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| interview_id | string | 是 | 面试 ID | `uuid` |

**出参（data）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| interview_id | string | 是 | 面试 ID | `uuid` |
| finished_at | string | 是 | 结束时间（RFC3339） | `2024-01-01T11:00:00Z` |
| duration_seconds | number | 是 | 面试时长（秒） | `3600` |

**curl**

```bash
curl -sS -X POST 'https://api.aiinterview.com/v1/interview/finish' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer jwt_token' \
  -d '{
    "interview_id": "uuid"
  }'
```

### 查询面试状态

- Method: `GET`
- Path: `/interview/state`
- Auth: 需要登录

**Query**

| 参数 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| interview_id | string | 是 | 面试 ID | `uuid` |

**出参（data）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| interview_id | string | 是 | 面试 ID | `uuid` |
| stage | string | 是 | 当前阶段 | `questioning` |
| questions_asked | number | 是 | 已提问数量 | `3` |
| current_question_followups | number | 是 | 当前问题追问次数 | `1` |
| started_at | string | 是 | 开始时间（RFC3339） | `2024-01-01T10:00:00Z` |

**stage 枚举**

| 值 | 说明 |
|---|---|
| intro | 自我介绍阶段 |
| questioning | 技术提问阶段 |
| algorithm | 算法题阶段 |
| closing | 反问阶段 |
| finished | 面试结束 |

**curl**

```bash
curl -sS 'https://api.aiinterview.com/v1/interview/state?interview_id=uuid' \
  -H 'Authorization: Bearer jwt_token'
```

## 代码提交模块

### 提交代码

- Method: `POST`
- Path: `/interview/code/submit`
- Auth: 需要登录
- Content-Type: `application/json`

**Headers**

| Header | 必填 | 说明 | 示例 |
|---|---:|---|---|
| Authorization | 是 | `Bearer {jwt_token}` | `Bearer xxx` |

**入参（JSON Body）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| interview_id | string | 是 | 面试 ID | `uuid` |
| question_id | string | 是 | 题目 ID | `uuid` |
| language | string | 是 | 语言枚举 | `java` |
| code | string | 是 | 代码内容 | `class Solution {...}` |

**language 枚举**

| 值 | 说明 |
|---|---|
| java | Java |
| python | Python |
| go | Go |
| cpp | C++ |

**出参（data）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| status | string | 是 | 状态 | `judging` |
| message | string | 是 | 提示信息 | `代码评估中` |

**curl**

```bash
curl -sS -X POST 'https://api.aiinterview.com/v1/interview/code/submit' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer jwt_token' \
  -d '{
    "interview_id": "uuid",
    "question_id": "uuid",
    "language": "java",
    "code": "class Solution {\\n  public int[] twoSum(int[] nums, int target) {\\n    return null;\\n  }\\n}"
  }'
```

## 问卷模块

### 获取问卷（面试结束后）

- Method: `GET`
- Path: `/questionnaire`
- Auth: 需要登录

**Query**

| 参数 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| interview_id | string | 是 | 面试 ID | `uuid` |

**出参（data）**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| interview_id | string | 是 | 面试 ID |
| turns | object[] | 是 | 轮次列表 |

`turns[]` 字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| turn_id | string | 是 | 轮次 ID |
| stage | string | 是 | 阶段 |
| question | string | 是 | 问题 |
| user_answer | string | 是 | 用户回答文本 |

**curl**

```bash
curl -sS 'https://api.aiinterview.com/v1/questionnaire?interview_id=uuid' \
  -H 'Authorization: Bearer jwt_token'
```

### 提交问卷

- Method: `POST`
- Path: `/questionnaire/submit`
- Auth: 需要登录
- Content-Type: `application/json`

**Headers**

| Header | 必填 | 说明 | 示例 |
|---|---:|---|---|
| Authorization | 是 | `Bearer {jwt_token}` | `Bearer xxx` |

**入参（JSON Body）**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| interview_id | string | 是 | 面试 ID |
| answers | object[] | 是 | 答案列表 |

`answers[]` 字段：

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| turn_id | string | 是 | 轮次 ID | `001` |
| quality | string | 是 | 质量枚举 | `good` |
| feedback | string | 否 | 反馈内容 | `追问很有启发` |

**quality 枚举**

| 值 | 说明 |
|---|---|
| good | 这轮对话质量好 |
| bad | 这轮对话质量差 |

**出参（data）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| message | string | 是 | 提示信息 | `问卷提交成功，感谢你的反馈` |

**curl**

```bash
curl -sS -X POST 'https://api.aiinterview.com/v1/questionnaire/submit' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer jwt_token' \
  -d '{
    "interview_id": "uuid",
    "answers": [
      { "turn_id": "001", "quality": "good", "feedback": "追问很有启发" },
      { "turn_id": "002", "quality": "bad", "feedback": "追问方向感觉不对" }
    ]
  }'
```

## 报告模块

### 查询报告状态

- Method: `GET`
- Path: `/report/status`
- Auth: 需要登录

**Query**

| 参数 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| interview_id | string | 是 | 面试 ID | `uuid` |

**出参（data）**

| 字段 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| status | string | 是 | 状态枚举 | `generating` |
| message | string | 是 | 提示信息 | `报告生成中` |

**status 枚举**

| 值 | 说明 |
|---|---|
| pending | 等待生成 |
| generating | 生成中 |
| done | 生成完成 |
| failed | 生成失败 |

**curl**

```bash
curl -sS 'https://api.aiinterview.com/v1/report/status?interview_id=uuid' \
  -H 'Authorization: Bearer jwt_token'
```

### 获取报告

- Method: `GET`
- Path: `/report`
- Auth: 需要登录

**Query**

| 参数 | 类型 | 必填 | 说明 | 示例 |
|---|---|---:|---|---|
| interview_id | string | 是 | 面试 ID | `uuid` |

**出参（data）**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| interview_id | string | 是 | 面试 ID |
| dimensions | object | 是 | 维度评分 |
| summary | string | 是 | 总结 |
| strong_points | string[] | 是 | 优点 |
| weak_points | string[] | 是 | 不足 |
| created_at | string | 是 | 创建时间（RFC3339） |

`dimensions` 字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| knowledge_depth | number | 是 | 知识深度 |
| expression | number | 是 | 表达能力 |
| problem_solving | number | 是 | 解决问题 |
| code_quality | number | 是 | 代码质量 |
| stress_response | number | 是 | 抗压表现 |

**curl**

```bash
curl -sS 'https://api.aiinterview.com/v1/report?interview_id=uuid' \
  -H 'Authorization: Bearer jwt_token'
```

## SSE 事件（面试实时推送）

通过 `GET /interview/stream` 推送，基本格式：

```text
event: {event_type}
data: {json_payload}
```

### text.delta

```text
event: text.delta
data: {"turn_id":"uuid","delta":"你提到了"}
```

### text.done

```text
event: text.done
data: {"turn_id":"uuid","full_text":"..."}
```

### audio.delta

```text
event: audio.delta
data: {"turn_id":"uuid","audio_base64":"..."}
```

### audio.done

```text
event: audio.done
data: {"turn_id":"uuid"}
```

### stage.changed

```text
event: stage.changed
data: {"from":"intro","to":"questioning"}
```

### code.judged

```text
event: code.judged
data: {"correctness":true,"time_complexity":"O(n)","space_complexity":"O(1)","issues":["..."]}
```

### resume.parsed

```text
event: resume.parsed
data: {"status":"done"}
```

### report.ready

```text
event: report.ready
data: {"interview_id":"uuid"}
```

### interview.finished

```text
event: interview.finished
data: {"interview_id":"uuid","finished_at":"2024-01-01T11:00:00Z"}
```

### error

```text
event: error
data: {"code":5001,"message":"ASR识别失败，请重试"}
```

