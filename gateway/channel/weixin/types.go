package weixin

// 微信 iLink 协议结构体，对照 Tencent/openclaw-weixin src/api/types.ts。
// API 走 HTTP+JSON，proto 中的 bytes 字段在 JSON 里是 base64 字符串。

// 消息方向。
const (
	MessageTypeNone = 0
	MessageTypeUser = 1 // 用户发给 bot
	MessageTypeBot  = 2 // bot 发给用户
)

// 消息项类型。
const (
	MessageItemTypeNone  = 0
	MessageItemTypeText  = 1
	MessageItemTypeImage = 2
	MessageItemTypeVoice = 3
	MessageItemTypeFile  = 4
	MessageItemTypeVideo = 5
)

// BaseInfo 附加在每个 CGI 请求上，仅用于可观测，不参与鉴权/路由。
type BaseInfo struct {
	ChannelVersion string `json:"channel_version,omitempty"`
	BotAgent       string `json:"bot_agent,omitempty"`
}

// TextItem 文本消息内容。
type TextItem struct {
	Text string `json:"text,omitempty"`
}

// MessageItem 单条消息项。本期只处理 text，其余类型保留字段不展开。
type MessageItem struct {
	Type         int       `json:"type,omitempty"`
	CreateTimeMs int64     `json:"create_time_ms,omitempty"`
	UpdateTimeMs int64     `json:"update_time_ms,omitempty"`
	IsCompleted  bool      `json:"is_completed,omitempty"`
	MsgID        string    `json:"msg_id,omitempty"`
	TextItem     *TextItem `json:"text_item,omitempty"`
}

// WeixinMessage 统一消息（proto: WeixinMessage）。
type WeixinMessage struct {
	Seq          int64         `json:"seq,omitempty"`
	MessageID    int64         `json:"message_id,omitempty"`
	FromUserID   string        `json:"from_user_id,omitempty"`
	ToUserID     string        `json:"to_user_id,omitempty"`
	ClientID     string        `json:"client_id,omitempty"`
	CreateTimeMs int64         `json:"create_time_ms,omitempty"`
	UpdateTimeMs int64         `json:"update_time_ms,omitempty"`
	DeleteTimeMs int64         `json:"delete_time_ms,omitempty"`
	SessionID    string        `json:"session_id,omitempty"`
	GroupID      string        `json:"group_id,omitempty"`
	MessageType  int           `json:"message_type,omitempty"`
	MessageState int           `json:"message_state,omitempty"`
	ItemList     []MessageItem `json:"item_list,omitempty"`
	ContextToken string        `json:"context_token,omitempty"`
}

// GetUpdatesReq 拉取入站消息。get_updates_buf 为本地缓存的上下文游标，首次或重置后传 ""。
type GetUpdatesReq struct {
	GetUpdatesBuf string    `json:"get_updates_buf"`
	BaseInfo      *BaseInfo `json:"base_info,omitempty"`
}

// GetUpdatesResp 入站消息响应。
type GetUpdatesResp struct {
	Ret                  int             `json:"ret,omitempty"`
	ErrCode              int             `json:"errcode,omitempty"`
	ErrMsg               string          `json:"errmsg,omitempty"`
	Msgs                 []WeixinMessage `json:"msgs,omitempty"`
	GetUpdatesBuf        string          `json:"get_updates_buf,omitempty"`
	LongPollingTimeoutMs int             `json:"longpolling_timeout_ms,omitempty"`
}

// SendMessageReq 出站发送，包一条 WeixinMessage。
type SendMessageReq struct {
	Msg      *WeixinMessage `json:"msg,omitempty"`
	BaseInfo *BaseInfo      `json:"base_info,omitempty"`
}

// QRCodeResp get_bot_qrcode 响应。qrcode_img_content 是二维码内容 URL，由前端渲染。
type QRCodeResp struct {
	QRCode        string `json:"qrcode"`
	QRCodeImgContent string `json:"qrcode_img_content"`
}

// 扫码状态。
const (
	QRStatusWait              = "wait"                // 未扫
	QRStatusScaned            = "scaned"              // 已扫待确认
	QRStatusConfirmed         = "confirmed"           // 已确认，下发 token
	QRStatusExpired           = "expired"             // 二维码过期
	QRStatusScanedButRedirect = "scaned_but_redirect" // 需切换 IDC 主机继续轮询
	QRStatusNeedVerifyCode    = "need_verifycode"     // 需输入配对码
	QRStatusVerifyCodeBlocked = "verify_code_blocked" // 配对码多次错误被封
	QRStatusBindedRedirect    = "binded_redirect"     // 已绑定过，无需重复
)

// StatusResp get_qrcode_status 响应。
type StatusResp struct {
	Status       string `json:"status"`
	BotToken     string `json:"bot_token,omitempty"`
	ILinkBotID   string `json:"ilink_bot_id,omitempty"`
	BaseURL      string `json:"baseurl,omitempty"`
	ILinkUserID  string `json:"ilink_user_id,omitempty"`
	RedirectHost string `json:"redirect_host,omitempty"`
}

// qrCodeReq get_bot_qrcode 请求体。
type qrCodeReq struct {
	LocalTokenList []string `json:"local_token_list"`
}
