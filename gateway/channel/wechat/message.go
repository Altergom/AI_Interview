package wechat

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// WxMessage 微信服务号消息结构，兼容加密和明文两种模式。
type WxMessage struct {
	// 加密模式字段
	Encrypt   string `xml:"Encrypt"   json:"Encrypt,omitempty"`
	MsgSignature string `xml:"MsgSignature" json:"MsgSignature,omitempty"`

	// 标准消息字段
	ToUserName   string `xml:"ToUserName"   json:"ToUserName,omitempty"`
	FromUserName string `xml:"FromUserName" json:"FromUserName,omitempty"`
	CreateTime   int64  `xml:"CreateTime"   json:"CreateTime,omitempty"`
	MsgType      string `xml:"MsgType"      json:"MsgType,omitempty"`
	Content      string `xml:"Content"      json:"Content,omitempty"`
	MsgId        string `xml:"MsgId"        json:"MsgId,omitempty"`
	MediaId      string `xml:"MediaId"      json:"MediaId,omitempty"`
}

// parseXMLMessage 解析微信 XML 消息。
func parseXMLMessage(body []byte) (*WxMessage, error) {
	var msg WxMessage
	if err := xml.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("parse xml: %w", err)
	}
	return &msg, nil
}

// buildXMLReply 构建加密回复 XML。
func buildXMLReply(toUser, fromUser, replyText, encrypt, msgSignature, timestamp, nonce string) string {
	if encrypt != "" {
		return fmt.Sprintf(
			`<xml><Encrypt><![CDATA[%s]]></Encrypt><MsgSignature><![CDATA[%s]]></MsgSignature><TimeStamp>%s</TimeStamp><Nonce><![CDATA[%s]]></Nonce></xml>`,
			encrypt, msgSignature, timestamp, nonce,
		)
	}
	// 明文模式回复
	return fmt.Sprintf(
		`<xml><ToUserName><![CDATA[%s]]></ToUserName><FromUserName><![CDATA[%s]]></FromUserName><CreateTime>%s</CreateTime><MsgType><![CDATA[text]]></MsgType><Content><![CDATA[%s]]></Content></xml>`,
		toUser, fromUser, timestamp, strings.ReplaceAll(replyText, "]]>", "]]]]><![CDATA[>"),
	)
}
