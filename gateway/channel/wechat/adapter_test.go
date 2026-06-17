package wechat

import (
	"context"
	"strings"
	"testing"
)

var testAdapter = New(Config{
	Token:          testToken,
	EncodingAESKey: testKey,
	ReceiveID:      testReceiveID,
})

// ─── Parse ───────────────────────────────────────────────────────────────────

func TestParse_PlainText(t *testing.T) {
	body := []byte(`<xml>
		<ToUserName><![CDATA[gh_test1234567]]></ToUserName>
		<FromUserName><![CDATA[oUser123]]></FromUserName>
		<CreateTime>1409735669</CreateTime>
		<MsgType><![CDATA[text]]></MsgType>
		<Content><![CDATA[面试一下]]></Content>
		<MsgId>9876543210</MsgId>
	</xml>`)

	event, err := testAdapter.Parse(context.Background(), body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if event.Channel != "wechat" {
		t.Errorf("Channel got %q", event.Channel)
	}
	if event.PeerID != "oUser123" {
		t.Errorf("PeerID got %q", event.PeerID)
	}
	if event.AccountID != "gh_test1234567" {
		t.Errorf("AccountID got %q", event.AccountID)
	}
	if event.MessageID != "9876543210" {
		t.Errorf("MessageID got %q", event.MessageID)
	}
	if event.Payload.Type != "text" {
		t.Errorf("Payload.Type got %q", event.Payload.Type)
	}
	if event.Payload.Content != "面试一下" {
		t.Errorf("Payload.Content got %q", event.Payload.Content)
	}
}

func TestParse_Encrypted(t *testing.T) {
	// 构造一段合法的加密消息：先加密一段明文 XML，再包进外层 XML
	innerXML := `<xml><ToUserName><![CDATA[gh_test1234567]]></ToUserName><FromUserName><![CDATA[oEncUser]]></FromUserName><CreateTime>1000000000</CreateTime><MsgType><![CDATA[text]]></MsgType><Content><![CDATA[加密消息]]></Content><MsgId>111</MsgId></xml>`

	encrypted, err := encryptMessage(testKey, testReceiveID, innerXML)
	if err != nil {
		t.Fatalf("encryptMessage: %v", err)
	}

	outerXML := []byte(`<xml><ToUserName><![CDATA[gh_test1234567]]></ToUserName><Encrypt><![CDATA[` + encrypted + `]]></Encrypt></xml>`)

	event, err := testAdapter.Parse(context.Background(), outerXML)
	if err != nil {
		t.Fatalf("Parse encrypted: %v", err)
	}
	if event.PeerID != "oEncUser" {
		t.Errorf("PeerID got %q, want oEncUser", event.PeerID)
	}
	if event.Payload.Content != "加密消息" {
		t.Errorf("Payload.Content got %q", event.Payload.Content)
	}
}

func TestParse_InvalidXML(t *testing.T) {
	_, err := testAdapter.Parse(context.Background(), []byte("not xml"))
	if err == nil {
		t.Error("expected error for invalid XML, got nil")
	}
}

func TestParse_EmptyMsgType_DefaultsToText(t *testing.T) {
	body := []byte(`<xml>
		<ToUserName><![CDATA[gh_test1234567]]></ToUserName>
		<FromUserName><![CDATA[oUser]]></FromUserName>
		<Content><![CDATA[hello]]></Content>
	</xml>`)
	event, err := testAdapter.Parse(context.Background(), body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if event.Payload.Type != "text" {
		t.Errorf("default MsgType should be text, got %q", event.Payload.Type)
	}
}

// ─── HandleVerify ─────────────────────────────────────────────────────────────

func TestHandleVerify_Valid(t *testing.T) {
	// 用真实 echostr（加密后的随机字符串）来验证整个流程
	echostr, err := encryptMessage(testKey, testReceiveID, "random_echo_content")
	if err != nil {
		t.Fatalf("prepare echostr: %v", err)
	}
	timestamp := "1409735669"
	nonce := "534102190"
	sig := buildSignature(testToken, timestamp, nonce, echostr)

	plain, err := testAdapter.HandleVerify(timestamp, nonce, sig, echostr)
	if err != nil {
		t.Fatalf("HandleVerify: %v", err)
	}
	if plain != "random_echo_content" {
		t.Errorf("HandleVerify got %q, want random_echo_content", plain)
	}
}

func TestHandleVerify_WrongSignature(t *testing.T) {
	_, err := testAdapter.HandleVerify("123", "456", "wrong_sig", "echostr")
	if err == nil {
		t.Error("expected error for wrong signature, got nil")
	}
}

// ─── BuildEncryptedReply ──────────────────────────────────────────────────────

func TestBuildEncryptedReply_ContainsEncrypt(t *testing.T) {
	reply, err := testAdapter.BuildEncryptedReply("oUser123", "面试已开始，请输入您的姓名", "nonce123")
	if err != nil {
		t.Fatalf("BuildEncryptedReply: %v", err)
	}
	if !strings.Contains(reply, "<Encrypt>") {
		t.Errorf("reply missing <Encrypt>: %s", reply)
	}
	if !strings.Contains(reply, "<MsgSignature>") {
		t.Errorf("reply missing <MsgSignature>: %s", reply)
	}
}

// ─── VerifySignature ──────────────────────────────────────────────────────────

func TestVerifySignature_ViaAdapter(t *testing.T) {
	timestamp := "1409735669"
	nonce := "534102190"
	encrypt := "some_encrypt"
	sig := buildSignature(testToken, timestamp, nonce, encrypt)

	meta := map[string]string{
		"timestamp":     timestamp,
		"nonce":         nonce,
		"encrypt":       encrypt,
		"msg_signature": sig,
	}
	if err := testAdapter.VerifySignature(context.Background(), meta, nil); err != nil {
		t.Errorf("VerifySignature: %v", err)
	}
}

func TestVerifySignature_Invalid(t *testing.T) {
	meta := map[string]string{
		"timestamp":     "123",
		"nonce":         "456",
		"encrypt":       "enc",
		"msg_signature": "badsig",
	}
	if err := testAdapter.VerifySignature(context.Background(), meta, nil); err == nil {
		t.Error("expected error for bad signature, got nil")
	}
}
