# 微信渠道接入（iLink 协议）

网关的微信渠道基于 iLink bot 协议接入（参考实现 `Tencent/openclaw-weixin`），
通过**扫码登录获取 bot token**，再以 **HTTP 长轮询**拉取入站消息、HTTP 调用发送出站消息。
不走服务号 webhook，也不依赖 Redis / OAuth。

## 整体流程

```
前端                     网关 (gateway)                 iLink 后端
 │  POST /weixin/login        │                              │
 │ ─────────────────────────► │  get_bot_qrcode              │
 │                            │ ───────────────────────────► │
 │ ◄───────────────────────── │  {qrcode, qrcode_img_content}│
 │  {session_key, qrcode_url} │                              │
 │                            │                              │
 │  渲染二维码，用户扫码        │  后台轮询 get_qrcode_status   │
 │                            │ ───────────────────────────► │
 │  GET /weixin/login/status  │ ◄─────────── status=confirmed │
 │ ─────────────────────────► │  落 token 到 Store            │
 │ ◄───────────────────────── │                              │
 │  {status, account_id}      │                              │
 │                            │  Connector.Start: getupdates  │
 │                            │ ───────────────────────────► │ (长轮询)
 │                            │ ◄─────────── msgs[]           │
 │                            │  归一为 InboundEvent          │
```

## 登录 HTTP 接口

二维码由**后端透传 URL、前端渲染**（前端用 qrcode.js 等把 `qrcode_url` 画成图），后端不做终端渲染。

### 发起登录

```
POST /v1/gateway/weixin/login
```

响应：
```json
{
  "success": true,
  "data": {
    "session_key": "uuid",
    "qrcode": "服务端二维码标识",
    "qrcode_url": "https://weixin.qq.com/q/xxx"
  },
  "error": null
}
```

前端拿 `qrcode_url` 渲染二维码；用 `session_key` 轮询状态。后端已在后台代为长轮询 iLink。

### 查询扫码状态

```
GET /v1/gateway/weixin/login/status?session_key=<session_key>
```

响应（pending）：
```json
{ "success": true, "data": { "status": "wait" }, "error": null }
```

响应（confirmed）：
```json
{
  "success": true,
  "data": { "status": "confirmed", "account_id": "ilink_bot_id", "user_id": "ilink_user_id" },
  "error": null
}
```

`status` 取值：`wait`（未扫）/ `scaned`（已扫待确认）/ `confirmed`（已确认，token 落库）/
`expired`（二维码过期）。`confirmed` 后 token 存入内存 Store，可据 `account_id` 启动 Connector。
失败时 `data.error` 给出原因。

## 协议端点

基址：`https://ilinkai.weixin.qq.com`

| 端点 | 方法 | 用途 |
|------|------|------|
| `ilink/bot/get_bot_qrcode?bot_type=3` | POST | 拉登录二维码，body `{local_token_list:[]}` |
| `ilink/bot/get_qrcode_status?qrcode=X` | GET | 长轮询扫码状态（最长 hold 35s） |
| `ilink/bot/getupdates` | POST | 长轮询拉入站消息，body `{get_updates_buf, base_info}` |
| `ilink/bot/sendmessage` | POST | 发送出站消息，body `{msg, base_info}` |

## 鉴权

POST 请求头：

| 头 | 值 |
|----|----|
| `AuthorizationType` | `ilink_bot_token` |
| `Authorization` | `Bearer <bot_token>`（未登录时不带） |
| `X-WECHAT-UIN` | random uint32 → 十进制字符串 → base64 |
| `iLink-App-Id` | `bot` |
| `iLink-App-ClientVersion` | 版本号编码（uint32） |

`get_qrcode_status`（GET）只带 `iLink-App-Id` / `iLink-App-ClientVersion`，无需 token。

## 消息模型

入站 `getupdates` 返回 `msgs: WeixinMessage[]`，网关只处理 `message_type == 1`（用户消息）
且含 `text_item` 的文本消息，归一为 `inbound.InboundEvent`：

- `Channel = "wechat"`
- `AccountID = ilink_bot_id`
- `PeerID = from_user_id`
- `Payload.Type = "text"` / `Payload.Content = text_item.text`

游标 `get_updates_buf` 由响应回带、本地缓存，下次请求带上；首次传 `""`。
服务端通过 `longpolling_timeout_ms` 建议下次长轮询时长。

出站发送构造 `WeixinMessage{to_user_id, message_type=2, context_token, item_list:[{type:1, text_item:{text}}]}`。
`context_token` 是会话上下文令牌，回复时需回传（来自入站消息或 `SendOpts.ContextToken`）。

## 代码结构

```
gateway/channel/weixin/
├── types.go      # 协议结构体与常量
├── client.go     # iLink HTTP 客户端（4 个端点 + 鉴权头）
├── store.go      # 多账号 token 存储（本期内存版）
├── login.go      # 扫码登录状态机（后台轮询 + 落 token）
└── connector.go  # 实现 channel.ChannelConnector（getupdates 长轮询循环 + 重连）

gateway/handler/weixin_login.go  # 登录 HTTP 接口
```

## 本期范围与未做项

已落地：扫码登录、token 存储、入站文本长轮询、出站文本发送、连接状态与重连退避。

未做（后续阶段）：
- 图片/语音/文件等富媒体（CDN AES-128-ECB 加解密、`getuploadurl`）
- 配对码（`need_verifycode`）交互——后端服务无终端 stdin，当前直接判失败
- token 持久化（重启后需重新扫码）
- Connector 与 session → fsm → agent 真实链路对接
- 登录确认后自动拉起 Connector（当前 main.go 留 TODO）

## 联调失败记录

### 失败现象

- 前端二维码弹层空白，扫码入口拿到了数据但页面不显示二维码。
- 扫码确认后，状态接口返回 `status=failed`，错误为“扫码已确认，但返回的账号信息不完整”。
- 日志里出现 `status=confirmed`、`token_ready=true`，但 `account=""`、`weixin_user=""`，导致绑定不落库，长连接不启动。

### 原因

- `get_bot_qrcode` 的可扫码链接实际可能落在 `qrcodeImage` 字段里，前端只按 `qrcodeUrl` 渲染，导致二维码显示失败。
- 登录状态机把 `confirmed` 直接当成“账号信息必须完整返回”，对确认态容错不够，字段未解析到时被误判成失败。
- `get_qrcode_status` 确认成功后，官方实现使用的账号字段是 `ilink_bot_id`、`ilink_user_id`，不是仅有 `account_id`、`user_id`。字段映射不全会出现 token 已拿到、账号却为空的假象。

### 解决方式

- 后端对二维码返回做兼容：如果 `qrcodeImage` 实际是 URL，则转为 `qrcodeUrl` 供前端渲染；前端也增加同样兜底。
- 登录状态机改为：`confirmed` 但账号信息未补齐时进入 `binding` 状态继续轮询，不直接判失败。
- `get_qrcode_status` 字段解析补齐 `ilink_bot_id` / `ilink_user_id`，并保留对 `account_id` / `user_id` / `bot_id` 的兼容读取。
- 增加登录阶段调试日志，重点输出 `status`、`account`、`weixin_user`、`token_ready`，用于定位扫码确认后的真实返回。

## 回包成功但微信不显示

### 失败现象

- `sendmessage` HTTP 返回 `200`，响应体是 `{}`，服务端日志也显示发送成功。
- 重新绑定后的第一条回复偶尔能在微信里看到，后续回复大多不显示。
- 仅调整 `context_token` 的传法（空值、透传、固定值）不能稳定解决问题。

### 原因

- 问题不在网络，也不在 `sendmessage` 接口是否可达，而在于**没有完整按官方插件的会话语义发送消息**。
- 仅调用 `sendmessage` 不够，官方实现还依赖完整运行时链路：`notifystart`、入站 `context_token` 缓存、出站 `client_id`、标准 `msg` 结构。
- 我们早期实现是“收到消息后直接拼一个回包请求”，协议字段能过，但微信侧不会稳定把它识别为当前会话中的合法 bot 回复。

### 解决方式

- 按官方实现补齐启动和停止通知：连接启动先调 `ilink/bot/msg/notifystart`，连接结束调 `ilink/bot/msg/notifystop`。
- 入站消息按 `accountId + from_user_id` 缓存并持久化 `context_token`，出站时原样带回。
- 出站 `sendmessage` 改为对齐官方结构，至少包含：
  - `from_user_id: ""`
  - `to_user_id`
  - `client_id`
  - `message_type: 2`
  - `message_state: 2`
  - `item_list`
  - `context_token`
- 移除多套试验发送逻辑，收敛为单一路径回复。
