package model

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

// P1-09：AES-256-GCM 敏感字段加密
// 密钥由 config 层注入（SHA-256 派生 32 字节）；未配置时返回原值（开发模式降级，
// 生产环境由 main.go 启动时拒绝空密钥）

// cipherVersion 密文版本前缀：区分「已加密」与「历史明文」，便于后续换算法/换密钥
const cipherVersion = "v1:"

var (
	aeadOnce sync.Once
	aead     cipher.AEAD
	aeadErr  error

	// encryptionKey 由 InitFieldCrypto 注入；回退读环境变量以兼容未走 config 的测试
	encryptionKey string
)

// InitFieldCrypto 注入敏感字段加密密钥，须在使用加密字段前调用（main 启动时）
func InitFieldCrypto(key string) {
	encryptionKey = key
	aeadOnce = sync.Once{}
	aead, aeadErr = nil, nil
}

// fieldAEAD 返回进程内复用的 AES-256-GCM；未配置密钥时返回 nil
// 密钥派生与 GCM 构造只做一次，避免每字段重复 SHA-256/密钥扩展/GHASH 预计算
func fieldAEAD() (cipher.AEAD, error) {
	aeadOnce.Do(func() {
		key := encryptionKey
		if key == "" {
			key = os.Getenv("DATA_ENCRYPTION_KEY")
		}
		if key == "" {
			return
		}
		sum := sha256.Sum256([]byte(key))
		block, err := aes.NewCipher(sum[:])
		if err != nil {
			aeadErr = err
			return
		}
		aead, aeadErr = cipher.NewGCM(block)
	})
	return aead, aeadErr
}

// EncryptField 加密敏感字段（AES-256-GCM，返回 base64）
func EncryptField(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	gcm, err := fieldAEAD()
	if err != nil {
		return "", err
	}
	if gcm == nil {
		// 开发模式：无密钥时原样返回
		return plain, nil
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	return cipherVersion + base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(plain), nil)), nil
}

// DecryptField 解密敏感字段
func DecryptField(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	// 无版本前缀说明是历史明文（或未配置密钥时写入的值），原样返回
	if !strings.HasPrefix(encoded, cipherVersion) {
		return encoded, nil
	}
	encoded = strings.TrimPrefix(encoded, cipherVersion)
	gcm, err := fieldAEAD()
	if err != nil {
		return "", err
	}
	if gcm == nil {
		return "", errors.New("存在加密字段但未配置 DATA_ENCRYPTION_KEY")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文过短")
	}
	plain, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// EncryptedString 透明加密字段：列中只存密文，Go 侧始终是明文
// 由 driver.Valuer/sql.Scanner 在数据库边界完成转换，所有读写路径（含 demo 播种、
// 管理端列表）自动获得加密与解密，无需各 handler 记得调用
type EncryptedString string

// Value 写库前加密
func (s EncryptedString) Value() (driver.Value, error) {
	enc, err := EncryptField(string(s))
	if err != nil {
		return nil, err
	}
	return enc, nil
}

// Scan 读库后解密；历史明文行无版本前缀，原样返回
func (s *EncryptedString) Scan(src interface{}) error {
	if src == nil {
		*s = ""
		return nil
	}
	var raw string
	switch v := src.(type) {
	case string:
		raw = v
	case []byte:
		raw = string(v)
	default:
		return fmt.Errorf("EncryptedString 不支持的类型 %T", src)
	}
	dec, err := DecryptField(raw)
	if err != nil {
		return err
	}
	*s = EncryptedString(dec)
	return nil
}

// GormDataType 让 GORM 以字符串列建表；密文比明文长，留足长度
func (EncryptedString) GormDataType() string { return "varchar(512)" }

// EncryptedFloat 透明加密的浮点字段（经纬度）：列中只存密文
type EncryptedFloat float64

// Value 写库前格式化为 6 位小数再加密
func (f EncryptedFloat) Value() (driver.Value, error) {
	enc, err := EncryptField(strconv.FormatFloat(float64(f), 'f', 6, 64))
	if err != nil {
		return nil, err
	}
	return enc, nil
}

// Scan 读库后解密并解析；历史明文行可能是数值类型，一并兼容
func (f *EncryptedFloat) Scan(src interface{}) error {
	switch v := src.(type) {
	case nil:
		*f = 0
		return nil
	case float64:
		*f = EncryptedFloat(v)
		return nil
	case int64:
		*f = EncryptedFloat(v)
		return nil
	}

	var raw string
	switch v := src.(type) {
	case string:
		raw = v
	case []byte:
		raw = string(v)
	default:
		return fmt.Errorf("EncryptedFloat 不支持的类型 %T", src)
	}
	dec, err := DecryptField(raw)
	if err != nil {
		return err
	}
	if dec == "" {
		*f = 0
		return nil
	}
	parsed, err := strconv.ParseFloat(dec, 64)
	if err != nil {
		return err
	}
	*f = EncryptedFloat(parsed)
	return nil
}

// GormDataType 密文以字符串存储
func (EncryptedFloat) GormDataType() string { return "varchar(512)" }

// MaskPhone 手机号脱敏（如 138****5678）
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
