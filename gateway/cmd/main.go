package main

import (
	"os"

	"github.com/cloudwego/hertz/pkg/app/server"

	"gateway/agent"
	"gateway/channel/feishu"
	"gateway/channel/qqbot"
	"gateway/channel/wechat"
	"gateway/config"
	"gateway/handler"
	"gateway/outbound"
	"gateway/session"
	"gateway/statemachine"
)

func main() {
	cfg := config.Load()

	// 渠道适配器，配置从环境变量读取
	wechatAdapter := wechat.New(wechat.Config{
		Token:          os.Getenv("WECHAT_TOKEN"),
		EncodingAESKey: os.Getenv("WECHAT_ENCODING_AES_KEY"),
		ReceiveID:      os.Getenv("WECHAT_RECEIVE_ID"),
		CallbackURL:    os.Getenv("WECHAT_CALLBACK_URL"),
	})
	feishuAdapter := feishu.New()
	qqbotAdapter := qqbot.New()

	// 内部层
	sessionMgr := session.NewManager()
	_ = statemachine.NewFSM()
	_ = agent.NewClient()
	_ = outbound.NewDispatcher(wechatAdapter, feishuAdapter, qqbotAdapter)

	// Handler
	webhookH := handler.NewWebhookHandler(wechatAdapter, feishuAdapter, qqbotAdapter)
	sessionH := handler.NewSessionHandler(sessionMgr)
	manageH := handler.NewManageHandler(sessionMgr)

	// HTTP 服务
	h := server.Default(server.WithHostPorts(cfg.HTTPAddr))

	h.GET("/webhook/:channel", webhookH.Handle)
	h.POST("/webhook/:channel", webhookH.Handle)

	v1 := h.Group("/v1/gateway")
	v1.GET("/session/:session_id", sessionH.Get)
	v1.GET("/sessions", sessionH.List)
	v1.POST("/session/:session_id/handoff", manageH.Handoff)
	v1.POST("/session/:session_id/terminate", manageH.Terminate)

	h.Spin()
}
