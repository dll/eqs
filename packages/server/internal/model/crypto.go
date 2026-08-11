package model

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

// P1-09：AES-256-GCM 敏感字段加密
// 密钥来自环境变量 DATA_ENCRYPTION_KEY（32 字节）；未配置时返回原值（开发模式降级）

// encryptionKey 获取加密密钥（从 DATA_ENCRYPTION_KEY 派生 32 字节）
func encryptionKey() []byte {
	key := os.Getenv("DATA_ENCRYPTION_KEY")
	if key == "" {
		return nil
	}
	// 派生 32 字节（SHA-256）
	sum := sha256.Sum256([]byte(key))
	return sum[:]
}

// EncryptField 加密敏感字段（AES-256-GCM，返回 base64）
func EncryptField(plain string) (string, error) {
	key := encryptionKey()
	if key == nil {
		// 开发模式：无密钥时原样返回
		return plain, nil
	}
	if plain == "" {
		return "", nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptField 解密敏感字段
func DecryptField(encoded string) (string, error) {
	key := encryptionKey()
	if key == nil {
		return encoded, nil
	}
	if encoded == "" {
		return "", nil
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文过短")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// HashField 敏感字段哈希（用于唯一索引，SHA-256 hex）
func HashField(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// MaskPhone 手机号脱敏（如 138****5678）
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
