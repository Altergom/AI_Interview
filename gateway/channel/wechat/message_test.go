package wechat

import (
	"strings"
	"testing"
)

func TestParseXMLMessage_PlainText(t *testing.T) {
	body := []byte(`<xml>
		<ToUserName><![CDATA[gh_test1234567]]></ToUserName>
		<FromUserName><![CDATA[oUser123]]></FromUserName>
		<CreateTime>1409735669</CreateTime>
		<MsgType><![CDATA[text]]></MsgType>
		<Content><![CDATA[你好]]></Content>
		<MsgId>1234567890</MsgId>
	</xml>`)

	msg, err := parseXMLMessage(body)
	if err != nil {
		t.Fatalf("parseXMLMessage: %v", err)
	}
	if msg.ToUserName != "gh_test1234567" {
		t.Errorf("ToUserName got %q", msg.ToUserName)
	}
	if msg.FromUserName != "oUser123" {
		t.Errorf("FromUserName got %q", msg.FromUserName)
	}
	if msg.MsgType != "text" {
		t.Errorf("MsgType got %q", msg.MsgType)
	}
	if msg.Content != "你好" {
		t.Errorf("Content got %q", msg.Content)
	}
	if msg.MsgId != "1234567890" {
		t.Errorf("MsgId got %q", msg.MsgId)
	}
	if msg.CreateTime != 1409735669 {
		t.Errorf("CreateTime got %d", msg.CreateTime)
	}
}

func TestParseXMLMessage_Encrypted(t *testing.T) {
	body := []byte(`<xml>
		<ToUserName><![CDATA[gh_test1234567]]></ToUserName>
		<Encrypt><![CDATA[some_encrypted_content]]></Encrypt>
	</xml>`)

	msg, err := parseXMLMessage(body)
	if err != nil {
		t.Fatalf("parseXMLMessage: %v", err)
	}
	if msg.Encrypt != "some_encrypted_content" {
		t.Errorf("Encrypt got %q", msg.Encrypt)
	}
}

func TestParseXMLMessage_InvalidXML(t *testing.T) {
	body := []byte(`not xml at all`)
	_, err := parseXMLMessage(body)
	if err == nil {
		t.Error("expected error for invalid XML, got nil")
	}
}

func TestBuildXMLReply_Encrypted(t *testing.T) {
	result := buildXMLReply("", "", "", "enc_content", "sig123", "1234567890", "nonce456")
	if !strings.Contains(result, "<Encrypt><![CDATA[enc_content]]></Encrypt>") {
		t.Errorf("missing Encrypt field in: %s", result)
	}
	if !strings.Contains(result, "<MsgSignature><![CDATA[sig123]]></MsgSignature>") {
		t.Errorf("missing MsgSignature in: %s", result)
	}
	if !strings.Contains(result, "<TimeStamp>1234567890</TimeStamp>") {
		t.Errorf("missing TimeStamp in: %s", result)
	}
}

func TestBuildXMLReply_Plaintext(t *testing.T) {
	result := buildXMLReply("oUser123", "gh_test", "hello reply", "", "", "1234567890", "")
	if !strings.Contains(result, "<ToUserName><![CDATA[oUser123]]></ToUserName>") {
		t.Errorf("missing ToUserName in: %s", result)
	}
	if !strings.Contains(result, "<Content><![CDATA[hello reply]]></Content>") {
		t.Errorf("missing Content in: %s", result)
	}
	if !strings.Contains(result, "<MsgType><![CDATA[text]]></MsgType>") {
		t.Errorf("missing MsgType in: %s", result)
	}
}

func TestBuildXMLReply_CDATAEscape(t *testing.T) {
	// ]]> 在 CDATA 内需要转义
	result := buildXMLReply("u", "s", "bad]]>content", "", "", "123", "")
	if strings.Contains(result, "bad]]>content") {
		t.Error("]]> in content should be escaped")
	}
}
