# Session Manager 实现总结

## ✅ 已完成的功能

### 1. **核心数据结构**

#### InterviewSession
```go
type InterviewSession struct {
    InterviewID string                // 面试会话ID
    UserID      string                // 用户ID
    Stage       domain.InterviewStage // 当前阶段
    CreatedAt   time.Time             // 创建时间
    UpdatedAt   time.Time             // 更新时间
    History     []Message             // 历史对话
    Stats       SessionStats          // 统计信息
    Context     map[string]any        // 上下文信息
}
```

#### Message
```go
type Message struct {
    Role      string    // user | assistant
    Content   string    // 消息内容
    Timestamp time.Time // 时间戳
}
```

#### SessionStats
```go
type SessionStats struct {
    QuestionCount  int // 已问问题数
    AlgorithmCount int // 已做算法题数
    TotalRounds    int // 总对话轮数
}
```

---

### 2. **基础 CRUD 操作**

#### CreateSession
- 创建新的面试会话
- 初始化为自我介绍阶段
- 设置 TTL（自动过期）

#### GetSession
- 从 Redis 获取会话数据
- JSON 反序列化
- 错误处理

#### saveSession
- 保存会话到 Redis
- JSON 序列化
- 更新 UpdatedAt 时间戳

#### DeleteSession
- 删除会话
- 清理 Redis 数据

---

### 3. **历史对话管理**

#### AddMessage
```go
func (sm *SessionManager) AddMessage(
    ctx context.Context,
    interviewID, role, content string,
) error
```
- 添加对话消息
- 自动更新 TotalRounds
- 记录时间戳

#### GetHistory
```go
func (sm *SessionManager) GetHistory(
    ctx context.Context,
    interviewID string,
) ([]Message, error)
```
- 获取完整历史对话
- 按时间顺序排列

---

### 4. **阶段管理**

#### UpdateStage
```go
func (sm *SessionManager) UpdateStage(
    ctx context.Context,
    interviewID string,
    newStage domain.InterviewStage,
) error
```
- 更新面试阶段
- 支持所有阶段切换

#### GetStage
```go
func (sm *SessionManager) GetStage(
    ctx context.Context,
    interviewID string,
) (domain.InterviewStage, error)
```
- 获取当前阶段

---

### 5. **统计信息管理**

#### IncrementQuestionCount
- 增加问题计数
- 用于技术问答阶段

#### IncrementAlgorithmCount
- 增加算法题计数
- 用于算法题阶段

#### GetStats
- 获取完整统计信息

---

### 6. **上下文管理**

#### UpdateContext
```go
func (sm *SessionManager) UpdateContext(
    ctx context.Context,
    interviewID string,
    key string,
    value any,
) error
```
- 更新自定义上下文
- 支持任意类型数据

#### GetContext
- 获取完整上下文

---

### 7. **Graph 集成方法**

#### GetGraphContext
```go
func (sm *SessionManager) GetGraphContext(
    ctx context.Context,
    interviewID string,
) (map[string]any, error)
```

**功能**：
- 转换历史对话为 Graph 所需格式
- 包含统计信息
- 合并自定义上下文

**返回格式**：
```go
{
    "history": []map[string]string{
        {"role": "user", "content": "..."},
        {"role": "assistant", "content": "..."},
    },
    "question_count": 3,
    "algorithm_count": 1,
    "total_rounds": 5,
    // ... 其他自定义上下文
}
```

---

#### UpdateFromGraphOutput
```go
func (sm *SessionManager) UpdateFromGraphOutput(
    ctx context.Context,
    interviewID string,
    userInput string,
    aiResponse string,
    newStage domain.InterviewStage,
    graphContext map[string]any,
) error
```

**功能**：
- 添加用户消息和 AI 回复到历史
- 更新面试阶段
- 更新统计信息
- 合并 Graph 返回的上下文

**使用场景**：
```go
// Graph 执行后
output, _ := graph.Invoke(ctx, input)

// 更新会话
sm.UpdateFromGraphOutput(
    ctx,
    interviewID,
    userInput,
    output.Text,
    output.NewStage,
    output.Context,
)
```

---

#### ShouldAdvanceStage
```go
func (sm *SessionManager) ShouldAdvanceStage(
    ctx context.Context,
    interviewID string,
) (bool, error)
```

**功能**：
- 基于规则判断是否应该切换阶段
- 作为 Supervisor 判断的补充

**规则**：
- 自我介绍阶段：3+ 轮对话
- 技术问答阶段：5+ 个问题
- 算法题阶段：1+ 道题
- 反问阶段：不自动切换

---

### 8. **辅助功能**

#### ExtendTTL
- 延长会话过期时间
- 用于活跃会话

#### sessionKey
- 生成 Redis Key
- 格式：`interview:session:{interviewID}`

---

## 📊 数据流

### 创建会话
```
CreateSession
  ↓
初始化 InterviewSession
  ↓
保存到 Redis (TTL: 48h)
```

### 对话流程
```
用户输入
  ↓
GetGraphContext (获取历史和上下文)
  ↓
调用 Graph
  ↓
UpdateFromGraphOutput (更新会话)
  ↓
保存到 Redis
```

### 阶段切换
```
ShouldAdvanceStage (规则判断)
  ↓
或 Supervisor 判断
  ↓
UpdateStage (更新阶段)
  ↓
保存到 Redis
```

---

## 🧪 测试覆盖

### 基础功能测试
- ✅ TestSessionManager_CreateAndGetSession
- ✅ TestSessionManager_AddMessage
- ✅ TestSessionManager_UpdateStage
- ✅ TestSessionManager_Stats
- ✅ TestSessionManager_Context
- ✅ TestSessionManager_DeleteSession

### Graph 集成测试
- ✅ TestSessionManager_GetGraphContext
- ✅ TestSessionManager_UpdateFromGraphOutput
- ✅ TestSessionManager_ShouldAdvanceStage

**测试结果**：
```
=== RUN   TestSessionManager
--- PASS: TestSessionManager (1.632s)
PASS
ok  	ai_interview/internal/service	1.632s
```

---

## 💡 使用示例

### 创建会话
```go
sm := NewSessionManager(redisClient, 48*time.Hour)

err := sm.CreateSession(ctx, "interview-123", "user-456")
```

### 对话流程
```go
// 1. 获取 Graph 上下文
graphCtx, _ := sm.GetGraphContext(ctx, interviewID)

// 2. 调用 Graph
output, _ := graph.Invoke(ctx, GraphInput{
    Text:        userInput,
    InterviewID: interviewID,
    Stage:       currentStage,
    Context:     graphCtx,
})

// 3. 更新会话
sm.UpdateFromGraphOutput(
    ctx,
    interviewID,
    userInput,
    output.Text,
    output.NewStage,
    output.Context,
)
```

### 阶段切换
```go
// 方式 1：基于规则
shouldAdvance, _ := sm.ShouldAdvanceStage(ctx, interviewID)
if shouldAdvance {
    sm.UpdateStage(ctx, interviewID, nextStage)
}

// 方式 2：基于 Supervisor 判断
// (已在 UpdateFromGraphOutput 中处理)
```

---

## 🎯 核心优势

### 1. **完整的状态管理**
- 历史对话
- 面试阶段
- 统计信息
- 自定义上下文

### 2. **与 Graph 无缝集成**
- `GetGraphContext` - 提供 Graph 所需数据
- `UpdateFromGraphOutput` - 处理 Graph 输出

### 3. **灵活的阶段切换**
- 支持 Supervisor 判断
- 支持规则判断
- 支持手动切换

### 4. **可靠的数据持久化**
- Redis 存储
- JSON 序列化
- TTL 自动过期

### 5. **完善的测试覆盖**
- 9 个测试用例
- 覆盖所有核心功能
- 使用 miniredis 进行单元测试

---

## 📝 Redis 数据结构

### Key 格式
```
interview:session:{interviewID}
```

### Value 格式（JSON）
```json
{
  "interview_id": "interview-123",
  "user_id": "user-456",
  "stage": "intro",
  "created_at": "2026-05-02T20:00:00Z",
  "updated_at": "2026-05-02T20:05:00Z",
  "history": [
    {
      "role": "user",
      "content": "你好",
      "timestamp": "2026-05-02T20:00:00Z"
    },
    {
      "role": "assistant",
      "content": "你好，请介绍一下你自己",
      "timestamp": "2026-05-02T20:00:01Z"
    }
  ],
  "stats": {
    "question_count": 0,
    "algorithm_count": 0,
    "total_rounds": 1
  },
  "context": {
    "tech_stack": ["Go", "Redis"]
  }
}
```

### TTL
- 默认：48 小时（可配置）
- 可通过 `ExtendTTL` 延长

---

## 🚀 下一步

Session Manager 已经完成，现在可以：

1. **实现 InterviewService**
   - 集成 Graph 和 Session Manager
   - 处理音频输入
   - 推送 AI 回复

2. **实现 HTTP Handler**
   - WebSocket/SSE 连接
   - 音频上传
   - 实时推送

3. **端到端测试**
   - 完整的面试流程测试
   - 性能测试

---

## 📚 文件清单

- ✅ `session_manager.go` - Session Manager 实现（新增，250+ 行）
- ✅ `session_manager_test.go` - 单元测试（新增，200+ 行）

---

## ✅ 验证

```bash
✅ 编译通过
✅ 所有测试通过（9 个测试用例）
✅ 代码质量良好
```
