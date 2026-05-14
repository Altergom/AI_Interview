package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"

	jwtutil "ai_interview/internal/auth"
	biz "ai_interview/internal/errors"
	"ai_interview/internal/log"
	authmw "ai_interview/internal/middleware/auth"
	"ai_interview/internal/middleware/ratelimit"
)

// wsUpgrader Hertz WebSocket 升级器。
// CheckOrigin 返回 true 表示骨架阶段允许所有来源；
// 生产阶段应改为按配置白名单校验。
var wsUpgrader = &websocket.HertzUpgrader{
	CheckOrigin: func(ctx *app.RequestContext) bool { return true },
}

// wsInterviewHandler 持有面试服务和 JWT Secret（握手时鉴权用）。
type wsInterviewHandler struct {
	jwtSecret string
	limiter   *ratelimit.Limiter // 连接维度限流
}

// ServeWS 处理 GET /v1/interview/ws/{interview_id}。
//
// 鉴权流程（握手阶段，HTTP 升级前）：
//  1. 优先读 Authorization: Bearer <token>
//  2. 若头部为空，则读 query 参数 token=<token>（浏览器 WebSocket API 无法自定义头）
//  3. 验证 JWT，失败返回 401，连接不升级
//
// 限流：握手前按 IP + USER 双维度检查，超限返回 429。
func (h *wsInterviewHandler) ServeWS(ctx context.Context, c *app.RequestContext) {
	interviewID := c.Param("interview_id")
	if interviewID == "" {
		c.JSON(http.StatusBadRequest, newWSErrResp(biz.CodeBadRequest, "missing interview_id"))
		return
	}

	// ── 1. JWT 鉴权（支持 header 和 query param 两种方式）──────────────────
	token := extractWSToken(c)
	if token == "" {
		log.Warnf("[WS] missing token, interview_id=%s", interviewID)
		c.JSON(http.StatusUnauthorized, newWSErrResp(biz.CodeUnauthorized, biz.CodeUnauthorized.Message()))
		return
	}

	claims, err := jwtutil.ValidateToken(h.jwtSecret, token)
	if err != nil {
		log.Warnf("[WS] invalid token, interview_id=%s: %v", interviewID, err)
		c.JSON(http.StatusUnauthorized, newWSErrResp(biz.CodeUnauthorized, "invalid or expired token"))
		return
	}

	userID := claims.UserID
	ip := wsClientIP(c)

	// ── 2. 限流（IP + USER 维度，在升级前拦截）──────────────────────────────
	if h.limiter != nil {
		if !h.limiter.Allow(ctx, "interview.ws", ratelimit.DimensionIP, ip) {
			log.Infof("[WS] rate limited by ip, ip=%s", ip)
			c.JSON(http.StatusTooManyRequests, newWSErrResp(biz.CodeRateLimitExceeded, biz.CodeRateLimitExceeded.Message()))
			return
		}
		if !h.limiter.Allow(ctx, "interview.ws", ratelimit.DimensionUser, userID) {
			log.Infof("[WS] rate limited by user, user_id=%s", userID)
			c.JSON(http.StatusTooManyRequests, newWSErrResp(biz.CodeRateLimitExceeded, biz.CodeRateLimitExceeded.Message()))
			return
		}
	}

	// ── 3. 升级为 WebSocket ──────────────────────────────────────────────────
	if err := wsUpgrader.Upgrade(c, func(conn *websocket.Conn) {
		h.handleConn(ctx, conn, userID, interviewID)
	}); err != nil {
		log.Errorf("[WS] upgrade failed, interview_id=%s: %v", interviewID, err)
		// 升级失败时连接已被关闭，无法再写 HTTP 响应
	}
}

// handleConn 在已建立的 WebSocket 连接上运行消息循环。
// 骨架阶段：鉴权通过即建立连接，推送 intro 阶段事件，然后进入消息分发循环。
// 后续模块（ASR / LLM / TTS）接入时在对应 TODO 处注入。
func (h *wsInterviewHandler) handleConn(
	ctx context.Context,
	conn *websocket.Conn,
	userID, interviewID string,
) {
	defer conn.Close()

	log.Infof("[WS] connected, user_id=%s interview_id=%s", userID, interviewID)

	// 注入 user_id 到 context，供后续 service 层使用
	ctx = authmw.WithUserID(ctx, userID)

	// 向前端发送连接确认：告知当前阶段为 intro
	if err := sendDownMsg(conn, DownMsg{
		Type:    DownMsgStageChange,
		Payload: StageChangePayload{Stage: "intro", QuestionsAsked: 0},
	}); err != nil {
		log.Warnf("[WS] send welcome failed, user_id=%s: %v", userID, err)
		return
	}

	// ── 消息循环 ──────────────────────────────────────────────────────────────
	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			// 正常断连（1000/1001）或网络错误，结束循环
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
				websocket.CloseNoStatusReceived,
			) {
				log.Warnf("[WS] unexpected close, user_id=%s interview_id=%s: %v", userID, interviewID, err)
			} else {
				log.Infof("[WS] disconnected, user_id=%s interview_id=%s", userID, interviewID)
			}
			return
		}

		switch msgType {
		case websocket.BinaryMessage:
			// 二进制帧 → audio_chunk（PCM 16kHz/16bit/mono）
			h.handleAudioChunk(ctx, conn, userID, interviewID, data)

		case websocket.TextMessage:
			// 文本帧 → JSON 包装的控制/代码消息
			h.handleTextMsg(ctx, conn, userID, interviewID, data)

		default:
			// ping/pong 由底层自动处理，其他类型忽略
		}
	}
}

// handleAudioChunk 处理音频帧（骨架：仅记录日志）。
// TODO: 将 PCM 帧送入 Qwen3 实时 ASR（语音面试模块实现）
func (h *wsInterviewHandler) handleAudioChunk(
	_ context.Context,
	_ *websocket.Conn,
	userID, interviewID string,
	pcm []byte,
) {
	log.Debugf("[WS] audio_chunk, user_id=%s interview_id=%s bytes=%d", userID, interviewID, len(pcm))
}

// handleTextMsg 解析并分发文本帧。
func (h *wsInterviewHandler) handleTextMsg(
	ctx context.Context,
	conn *websocket.Conn,
	userID, interviewID string,
	data []byte,
) {
	var msg UpMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Warnf("[WS] invalid json from user_id=%s: %v", userID, err)
		if err := sendDownMsg(conn, DownMsg{
			Type:    DownMsgError,
			Payload: ErrorPayload{Code: int(biz.CodeBadRequest), Message: "invalid message format"},
		}); err != nil {
			log.Warnf("[WS] send error msg failed user_id=%s: %v", userID, err)
		}
		return
	}

	switch msg.Type {
	case UpMsgControl:
		h.handleControl(ctx, conn, userID, interviewID, msg.Payload)
	case UpMsgCodeSubmit:
		h.handleCodeSubmit(ctx, conn, userID, interviewID, msg.Payload)
	default:
		log.Warnf("[WS] unknown msg type=%s user_id=%s", msg.Type, userID)
	}
}

// handleControl 处理控制指令（骨架：日志 + 阶段事件）。
// TODO: 接入面试阶段状态机（Interview Agent 模块实现）
func (h *wsInterviewHandler) handleControl(
	_ context.Context,
	conn *websocket.Conn,
	userID, interviewID string,
	raw json.RawMessage,
) {
	var p ControlPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		log.Warnf("[WS] bad control payload user_id=%s: %v", userID, err)
		return
	}
	log.Infof("[WS] control action=%s user_id=%s interview_id=%s", p.Action, userID, interviewID)

	if p.Action == "stop" {
		if err := sendDownMsg(conn, DownMsg{
			Type:    DownMsgStageChange,
			Payload: StageChangePayload{Stage: "end", QuestionsAsked: 0},
		}); err != nil {
			log.Warnf("[WS] send stage_change failed user_id=%s: %v", userID, err)
		}
	}
}

// handleCodeSubmit 处理代码提交（骨架：日志）。
// TODO: 接入 Code Judge Agent（Interview Agent 模块实现）
func (h *wsInterviewHandler) handleCodeSubmit(
	_ context.Context,
	_ *websocket.Conn,
	userID, interviewID string,
	raw json.RawMessage,
) {
	var p CodeSubmitPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		log.Warnf("[WS] bad code_submit payload user_id=%s: %v", userID, err)
		return
	}
	log.Infof("[WS] code_submit lang=%s user_id=%s interview_id=%s len=%d",
		p.Language, userID, interviewID, len(p.Code))
}

// sendDownMsg 将下行消息序列化为 JSON 文本帧发送。
func sendDownMsg(conn *websocket.Conn, msg DownMsg) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

// SendTTSAudio 将 TTS PCM 音频以二进制帧推送给前端（供语音面试模块调用）。
func SendTTSAudio(conn *websocket.Conn, pcm []byte) error {
	return conn.WriteMessage(websocket.BinaryMessage, pcm)
}

// ─── 辅助函数 ─────────────────────────────────────────────────────────────────

// extractWSToken 从 WebSocket 握手请求中提取 JWT token。
// 浏览器原生 WebSocket API 不支持自定义 header，故同时支持 query 参数。
func extractWSToken(c *app.RequestContext) string {
	if auth := string(c.GetHeader("Authorization")); strings.HasPrefix(auth, "Bearer ") {
		if t := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer ")); t != "" {
			return t
		}
	}
	if t := string(c.Query("token")); t != "" {
		return t
	}
	return ""
}

// newWSErrResp 构造与 handler.Result 格式兼容的错误响应体（握手阶段用，升级前）。
func newWSErrResp(code biz.ErrorCode, msg string) map[string]any {
	return map[string]any{
		"success": false,
		"data":    nil,
		"error":   map[string]any{"code": int(code), "message": msg},
	}
}

// wsClientIP 从 Hertz context 中取客户端 IP。
func wsClientIP(c *app.RequestContext) string {
	if xff := string(c.GetHeader("X-Forwarded-For")); xff != "" {
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if xri := string(c.GetHeader("X-Real-IP")); xri != "" {
		return strings.TrimSpace(xri)
	}
	return c.ClientIP()
}
