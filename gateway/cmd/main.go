package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	"gateway/agent"
	"gateway/channel/feishu"
	"gateway/channel/qqbot"
	"gateway/config"
	"gateway/handler"
	"gateway/outbound"
	"gateway/session"
	"gateway/statemachine"
)

func main() {
	cfg := config.Load()

	// 渠道连接器：每个渠道一个常驻长连接（拉取/WS），自管重连。
	feishuConn := feishu.New()
	qqbotConn := qqbot.New()

	// 内部层
	sessionMgr := session.NewManager()
	_ = statemachine.NewFSM()
	_ = agent.NewClient()
	_ = outbound.NewDispatcher(feishuConn, qqbotConn)

	// TODO: 启动各 connector 的 Start(ctx, eventCh) goroutine，
	// 由统一消费循环把 InboundEvent 喂入 session -> fsm -> agent。

	// Handler
	sessionH := handler.NewSessionHandler(sessionMgr)
	manageH := handler.NewManageHandler(sessionMgr)

	// HTTP 服务
	h := server.Default(server.WithHostPorts(cfg.HTTPAddr))

	v1 := h.Group("/v1/gateway")
	v1.GET("/session/:session_id", sessionH.Get)
	v1.GET("/sessions", sessionH.List)
	v1.POST("/session/:session_id/handoff", manageH.Handoff)
	v1.POST("/session/:session_id/terminate", manageH.Terminate)

	h.Spin()
}
