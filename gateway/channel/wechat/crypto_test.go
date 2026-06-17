package wechat

import (
	"testing"
)

// testKey 是一个合法的 43 位 encodingAESKey（base64 decode 后 32 字节）。
const testKey = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG"
const testReceiveID = "gh_test1234567"
const testToken = "testtoken"

func TestVerifySignature(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		timestamp string
		nonce     string
		encrypt   string
		signature string
		want      bool
	}{
		{
			name:      "valid signature",
			token:     testToken,
			timestamp: "1409735669",
			nonce:     "534102190",
			encrypt:   "msg_encrypt",
			// 预期签名由 sort([testToken, "1409735669", "534102190", "msg_encrypt"]) SHA-1 得出
			signature: buildSignature(testToken, "1409735669", "534102190", "msg_encrypt"),
			want:      true,
		},
		{
			name:      "wrong token",
			token:     "wrongtoken",
			timestamp: "1409735669",
			nonce:     "534102190",
			encrypt:   "msg_encrypt",
			signature: buildSignature(testToken, "1409735669", "534102190", "msg_encrypt"),
			want:      false,
		},
		{
			name:      "tampered encrypt",
			token:     testToken,
			timestamp: "1409735669",
			nonce:     "534102190",
			encrypt:   "tampered_encrypt",
			signature: buildSignature(testToken, "1409735669", "534102190", "msg_encrypt"),
			want:      false,
		},
		{
			name:      "signature case insensitive",
			token:     testToken,
			timestamp: "1409735669",
			nonce:     "534102190",
			encrypt:   "msg_encrypt",
			signature: "  " + buildSignature(testToken, "1409735669", "534102190", "msg_encrypt") + "  ",
			want:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := verifySignature(tc.token, tc.timestamp, tc.nonce, tc.encrypt, tc.signature)
			if got != tc.want {
				t.Errorf("verifySignature() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBuildSignature_Symmetric(t *testing.T) {
	sig := buildSignature(testToken, "12345", "67890", "encrypt_content")
	if !verifySignature(testToken, "12345", "67890", "encrypt_content", sig) {
		t.Error("buildSignature and verifySignature are not symmetric")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	messages := []string{
		"hello world",
		"你好，这是一条中文消息",
		"<xml><Content>test</Content></xml>",
		"a", // 极短消息
	}

	for _, msg := range messages {
		t.Run(msg, func(t *testing.T) {
			encrypted, err := encryptMessage(testKey, testReceiveID, msg)
			if err != nil {
				t.Fatalf("encryptMessage: %v", err)
			}
			decrypted, err := decryptMessage(testKey, testReceiveID, encrypted)
			if err != nil {
				t.Fatalf("decryptMessage: %v", err)
			}
			if decrypted != msg {
				t.Errorf("round trip failed: got %q want %q", decrypted, msg)
			}
		})
	}
}

func TestDecryptMessage_WrongReceiveID(t *testing.T) {
	encrypted, err := encryptMessage(testKey, testReceiveID, "test message")
	if err != nil {
		t.Fatalf("encryptMessage: %v", err)
	}
	_, err = decryptMessage(testKey, "wrong_receive_id", encrypted)
	if err == nil {
		t.Error("expected error with wrong receiveID, got nil")
	}
}

func TestDecryptMessage_InvalidKey(t *testing.T) {
	_, err := decryptMessage("not_valid_key", testReceiveID, "anything")
	if err == nil {
		t.Error("expected error with invalid key, got nil")
	}
}

func TestDecryptMessage_InvalidBase64(t *testing.T) {
	_, err := decryptMessage(testKey, testReceiveID, "!!!not_base64!!!")
	if err == nil {
		t.Error("expected error with invalid base64, got nil")
	}
}
