# 微信服务号接入指南

## 概览

微信渠道通过**微信服务号 Webhook** 接入。微信服务器将用户消息推送到我们的网关，网关解密、解析后进入面试流程，回复通过微信客服消息接口发送。

消息加密模式：**安全模式**（AES-256-CBC + SHA-1 签名）。

---

## 前置条件

- 已认证的微信服务号（订阅号无法使用客服消息接口）
- 服务器有公网域名（微信要求 HTTPS，80/443 端口）
- 网关服务可公网访问

---

## 配置步骤

### 1. 微信公众平台配置

进入 [微信公众平台](https://mp.weixin.qq.com) → 设置与开发 → 基本配置：

| 字段 | 说明 |
|------|------|
| URL | `https://your-domain/webhook/wechat` |
| Token | 自定义字符串，需与环境变量 `WECHAT_TOKEN` 一致 |
| EncodingAESKey | 点击「随机生成」，填入环境变量 `WECHAT_ENCODING_AES_KEY` |
| 消息加密方式 | 选择「安全模式」 |

填写完成后点击「提交」，微信会向 URL 发送 GET 请求验证，网关自动处理。

### 2. 获取服务号原始 ID

微信公众平台 → 设置与开发 → 账号信息 → 原始 ID（格式 `gh_xxxx`），填入环境变量 `WECHAT_RECEIVE_ID`。

### 3. 配置环境变量

```bash
WECHAT_TOKEN=your_token                        # 微信后台配置的 Token
WECHAT_ENCODING_AES_KEY=your_43_char_key       # 43 位 AES 密钥（不含末尾 =）
WECHAT_RECEIVE_ID=gh_xxxxxxxxxxxx              # 服务号原始 ID
WECHAT_CALLBACK_URL=                           # 客服消息回调基址（主动回复，后续补充）
```

### 4. 启动网关

```bash
cd gateway
WECHAT_TOKEN=xxx WECHAT_ENCODING_AES_KEY=xxx WECHAT_RECEIVE_ID=xxx go run cmd/main.go
```

---

## 消息流程

```
用户微信消息
    ↓
微信服务器 POST /webhook/wechat
    ↓
签名验证（SHA-1）
    ↓
AES-256-CBC 解密消息体
    ↓
解析 XML → InboundEvent
    ↓
会话路由 → 状态机 → Agent
    ↓
加密 XML 回复 / 客服消息接口回复
```

---

## 请求格式

### 微信推送（POST）

Query 参数：

| 参数 | 说明 |
|------|------|
| `timestamp` | 时间戳 |
| `nonce` | 随机数 |
| `msg_signature` | 消息签名（安全模式） |

Body（XML，安全模式）：

```xml
<xml>
  <ToUserName><![CDATA[gh_xxxx]]></ToUserName>
  <Encrypt><![CDATA[加密消息内容]]></Encrypt>
</xml>
```

### 签名算法

```
SHA-1(sort([token, timestamp, nonce, encrypt]).join(""))
```

### AES 解密

- 算法：AES-256-CBC
- 密钥：`base64.decode(encodingAESKey + "=")`（32 字节）
- IV：密钥前 16 字节
- 明文格式：`random(16) + msgLen(4,BigEndian) + message + receiveId`

---

## 服务器验证（首次配置）

微信在保存配置时发送 GET 请求：

```
GET /webhook/wechat?signature=xxx&timestamp=xxx&nonce=xxx&echostr=xxx
```

网关验证签名后解密 `echostr` 并原样返回，微信验证通过后配置生效。

---

## 相关代码

| 文件 | 职责 |
|------|------|
| `gateway/channel/wechat/adapter.go` | 适配器主体，Parse/Send/HandleVerify |
| `gateway/channel/wechat/crypto.go` | SHA-1 签名验证、AES 加解密 |
| `gateway/channel/wechat/message.go` | XML 消息结构定义和解析 |
| `gateway/handler/webhook.go` | Webhook 路由，含 GET 验证处理 |

---

## 后续待实现

- `Send()` 方法：接入微信客服消息接口（需要 access_token 管理）
- `encryptMessage` 中的随机 16 字节改用 `crypto/rand`
- access_token 自动刷新（有效期 2 小时）
