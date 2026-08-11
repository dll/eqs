package model

import (
	"strings"
	"sync"
	"testing"
)

// resetFieldAEAD 重置缓存的 AEAD 与密钥，使各用例可独立切换密钥配置
func resetFieldAEAD() {
	encryptionKey = ""
	aeadOnce = sync.Once{}
	aead = nil
	aeadErr = nil
}

const testKey = "test_key_for_field_encryption"

func TestEncryptField_RoundTrip(t *testing.T) {
	InitFieldCrypto(testKey)
	defer resetFieldAEAD()

	plain := "GD-2026-0001"
	enc, err := EncryptField(plain)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if enc == plain {
		t.Fatal("配置密钥后不应返回明文")
	}
	if !strings.HasPrefix(enc, cipherVersion) {
		t.Fatalf("密文应带版本前缀 %q，得到 %q", cipherVersion, enc)
	}

	dec, err := DecryptField(enc)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if dec != plain {
		t.Fatalf("往返不一致: %q != %q", dec, plain)
	}
}

// 相同明文两次加密应产生不同密文（nonce 随机），但都能解回原值
func TestEncryptField_NonceIsRandom(t *testing.T) {
	InitFieldCrypto(testKey)
	defer resetFieldAEAD()

	first, err := EncryptField("118.310000")
	if err != nil {
		t.Fatalf("首次加密失败: %v", err)
	}
	second, err := EncryptField("118.310000")
	if err != nil {
		t.Fatalf("二次加密失败: %v", err)
	}
	if first == second {
		t.Fatal("相同明文的密文不应相同")
	}
	for _, enc := range []string{first, second} {
		if dec, err := DecryptField(enc); err != nil || dec != "118.310000" {
			t.Fatalf("解密结果异常: %q, err=%v", dec, err)
		}
	}
}

func TestEncryptField_EmptyString(t *testing.T) {
	InitFieldCrypto(testKey)
	defer resetFieldAEAD()

	if enc, err := EncryptField(""); err != nil || enc != "" {
		t.Fatalf("空串应原样返回: %q, err=%v", enc, err)
	}
	if dec, err := DecryptField(""); err != nil || dec != "" {
		t.Fatalf("空串应原样返回: %q, err=%v", dec, err)
	}
}

// 未配置密钥时降级为原样返回（开发模式）
func TestEncryptField_NoKeyPassesThrough(t *testing.T) {
	InitFieldCrypto("")
	defer resetFieldAEAD()

	if enc, err := EncryptField("ZZ-DEMO-001"); err != nil || enc != "ZZ-DEMO-001" {
		t.Fatalf("无密钥应返回明文: %q, err=%v", enc, err)
	}
	if dec, err := DecryptField("ZZ-DEMO-001"); err != nil || dec != "ZZ-DEMO-001" {
		t.Fatalf("无密钥应返回原值: %q, err=%v", dec, err)
	}
}

// 历史明文行没有版本前缀，解密应原样返回而不是报错
func TestDecryptField_PlaintextPassthrough(t *testing.T) {
	InitFieldCrypto(testKey)
	defer resetFieldAEAD()

	dec, err := DecryptField("ZZ-DEMO-001")
	if err != nil || dec != "ZZ-DEMO-001" {
		t.Fatalf("无前缀明文应原样返回: %q, err=%v", dec, err)
	}
}

// 带版本前缀但内容非 base64 的密文应报错，调用方据此走 fallback
func TestDecryptField_InvalidInput(t *testing.T) {
	InitFieldCrypto(testKey)
	defer resetFieldAEAD()

	if _, err := DecryptField(cipherVersion + "not@valid@base64"); err == nil {
		t.Fatal("非法密文应报错")
	}
	if _, err := DecryptField(cipherVersion + "c2hvcnQ="); err == nil {
		t.Fatal("过短密文应报错")
	}
}

// 无密钥却存在带前缀密文：应报错而非静默返回
func TestDecryptField_EncryptedButNoKey(t *testing.T) {
	InitFieldCrypto(testKey)
	enc, _ := EncryptField("secret-value")
	InitFieldCrypto("")
	defer resetFieldAEAD()

	if _, err := DecryptField(enc); err == nil {
		t.Fatal("有密文但无密钥时应报错")
	}
}

func TestMaskPhone(t *testing.T) {
	cases := []struct{ in, want string }{
		{"13812345678", "138****5678"},
		{"1381234", "138****1234"},
		{"123456", "123456"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := MaskPhone(tc.in); got != tc.want {
			t.Errorf("MaskPhone(%q) = %q，期望 %q", tc.in, got, tc.want)
		}
	}
	if strings.Contains(MaskPhone("13812345678"), "1234567") {
		t.Fatal("脱敏后不应残留中段数字")
	}
}

func TestEncryptedString_ValueAndScan(t *testing.T) {
	InitFieldCrypto(testKey)
	defer resetFieldAEAD()

	var s EncryptedString = "GD-2026-0001"
	val, err := s.Value()
	if err != nil {
		t.Fatalf("Value 失败: %v", err)
	}
	raw, ok := val.(string)
	if !ok || raw == string(s) || !strings.HasPrefix(raw, cipherVersion) {
		t.Fatalf("Value 应返回带前缀密文: %q", val)
	}

	var back EncryptedString
	if err := back.Scan(raw); err != nil {
		t.Fatalf("Scan 失败: %v", err)
	}
	if back != s {
		t.Fatalf("往返不一致: %q != %q", back, s)
	}

	// 历史明文行原样扫入
	var legacy EncryptedString
	if err := legacy.Scan("ZZ-DEMO-001"); err != nil || legacy != "ZZ-DEMO-001" {
		t.Fatalf("明文行应原样扫入: %q, err=%v", legacy, err)
	}
}

func TestEncryptedFloat_ValueAndScan(t *testing.T) {
	InitFieldCrypto(testKey)
	defer resetFieldAEAD()

	var f EncryptedFloat = 118.31
	val, err := f.Value()
	if err != nil {
		t.Fatalf("Value 失败: %v", err)
	}
	raw, ok := val.(string)
	if !ok || !strings.HasPrefix(raw, cipherVersion) {
		t.Fatalf("Value 应返回带前缀密文: %q", val)
	}

	var back EncryptedFloat
	if err := back.Scan(raw); err != nil {
		t.Fatalf("Scan 失败: %v", err)
	}
	if float64(back) != 118.31 {
		t.Fatalf("往返不一致: %v", float64(back))
	}

	// 兼容历史数值行（旧明文列由 DECIMAL 读回时可能是 float64/int64）
	var legacy EncryptedFloat
	if err := legacy.Scan(float64(32.3)); err != nil || float64(legacy) != 32.3 {
		t.Fatalf("数值行应原样扫入: %v, err=%v", float64(legacy), err)
	}
}
