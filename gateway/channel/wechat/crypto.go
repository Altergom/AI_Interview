package wechat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

// verifySignature 验证微信服务号消息签名。
// 签名算法：SHA-1(sort([token, timestamp, nonce, encrypt]).join(""))
func verifySignature(token, timestamp, nonce, encrypt, signature string) bool {
	arr := []string{token, timestamp, nonce, encrypt}
	sort.Strings(arr)
	h := sha1.New()
	h.Write([]byte(strings.Join(arr, "")))
	expected := fmt.Sprintf("%x", h.Sum(nil))
	return expected == strings.TrimSpace(strings.ToLower(signature))
}

// decryptMessage 解密微信 AES-256-CBC 加密消息。
// 密钥由 encodingAESKey+"=" base64 解码得到（32字节），IV 取密钥前 16 字节。
// 明文格式：random(16) + msgLen(4,BigEndian) + message + receiveId
func decryptMessage(encodingAESKey, receiveID, encrypt string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil || len(key) != 32 {
		return "", fmt.Errorf("invalid encodingAESKey")
	}

	cipherText, err := base64.StdEncoding.DecodeString(encrypt)
	if err != nil {
		return "", fmt.Errorf("base64 decode encrypt: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	iv := key[:16]
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(cipherText))
	mode.CryptBlocks(plaintext, cipherText)

	// 去除 PKCS7 padding
	pad := int(plaintext[len(plaintext)-1])
	if pad < 1 || pad > 32 {
		return "", fmt.Errorf("invalid PKCS7 padding")
	}
	plaintext = plaintext[:len(plaintext)-pad]

	// 解析内容：跳过前 16 字节随机数
	payload := plaintext[16:]
	if len(payload) < 4 {
		return "", fmt.Errorf("payload too short")
	}
	msgLen := int(binary.BigEndian.Uint32(payload[:4]))
	if len(payload) < 4+msgLen {
		return "", fmt.Errorf("payload length mismatch")
	}
	message := string(payload[4 : 4+msgLen])
	gotReceiveID := string(payload[4+msgLen:])
	if gotReceiveID != receiveID {
		return "", fmt.Errorf("receiveId mismatch: got %q want %q", gotReceiveID, receiveID)
	}

	return message, nil
}

// buildSignature 构建微信消息签名（与 verifySignature 对称）。
func buildSignature(token, timestamp, nonce, encrypt string) string {
	arr := []string{token, timestamp, nonce, encrypt}
	sort.Strings(arr)
	h := sha1.New()
	h.Write([]byte(strings.Join(arr, "")))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// encryptMessage 加密回复消息（用于加密模式回复）。
func encryptMessage(encodingAESKey, receiveID, plaintext string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil || len(key) != 32 {
		return "", fmt.Errorf("invalid encodingAESKey")
	}

	random := bytes.Repeat([]byte("0"), 16) // 骨架用固定值，生产替换为 crypto/rand
	msg := []byte(plaintext)
	corp := []byte(receiveID)

	msgLenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(msgLenBuf, uint32(len(msg)))

	payload := bytes.Join([][]byte{random, msgLenBuf, msg, corp}, nil)

	// PKCS7 padding
	blockSize := 32
	pad := blockSize - len(payload)%blockSize
	payload = append(payload, bytes.Repeat([]byte{byte(pad)}, pad)...)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	iv := key[:16]
	mode := cipher.NewCBCEncrypter(block, iv)
	cipherText := make([]byte, len(payload))
	mode.CryptBlocks(cipherText, payload)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}
