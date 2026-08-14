// Package channel 外部服务通道适配层（短信/OCR/支付/推送/电子签）。
// 所有适配器均可独立单元测试（签名黄金测试），未配置凭据时自动降级
// （内置模拟 / 规则 / SSE 轮询），保证"代码完整、仅剩凭据填写即可结项"。
package channel

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// TencentCloudSign 腾讯云 API v3 签名（TC3-HMAC-SHA256），纯标准库实现。
// 参考：https://cloud.tencent.com/document/api/598/32801
type TencentCloudSign struct {
	SecretID  string
	SecretKey string
	Service   string // 如 sms / ocr
	Host      string // 如 sms.tencentcloudapi.com
	Action    string // 如 SendSms（签名时转小写作为 x-tc-action）
	Region    string // 如 ap-guangzhou
	Version   string // 如 2021-01-11
	Now       time.Time
}

// Sign 对 JSON payload 生成 Authorization 头与请求头（x-tc-action / x-tc-timestamp / x-tc-region / Host）
func (s *TencentCloudSign) Sign(payload []byte) (authorization, actionHeader string, headers map[string]string, err error) {
	if s.SecretID == "" || s.SecretKey == "" {
		return "", "", nil, fmt.Errorf("腾讯云密钥未配置")
	}
	now := s.Now
	if now.IsZero() {
		now = time.Now()
	}
	timestamp := now.Unix()
	date := now.UTC().Format("2006-01-02")

	hashedPayload := sha256Hex(payload)
	canonicalHeaders := fmt.Sprintf("content-type:application/json; charset=utf-8\nhost:%s\nx-tc-action:%s\n",
		strings.ToLower(s.Host), strings.ToLower(s.Action))
	signedHeaders := "content-type;host;x-tc-action"

	canonicalRequest := fmt.Sprintf("POST\n/\n\n%s\n%s\n%s", canonicalHeaders, signedHeaders, hashedPayload)
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, s.Service)
	hashedCanonical := sha256Hex([]byte(canonicalRequest))
	stringToSign := fmt.Sprintf("TC3-HMAC-SHA256\n%d\n%s\n%s", timestamp, credentialScope, hashedCanonical)

	secretDate := hmacSHA256([]byte("TC3"+s.SecretKey), date)
	secretService := hmacSHA256(secretDate, s.Service)
	secretSigning := hmacSHA256(secretService, "tc3_request")
	signature := hex.EncodeToString(hmacSHA256(secretSigning, stringToSign))

	authorization = fmt.Sprintf("TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.SecretID, credentialScope, signedHeaders, signature)
	actionHeader = strings.ToLower(s.Action)
	headers = map[string]string{
		"Authorization":  authorization,
		"Content-Type":   "application/json; charset=utf-8",
		"Host":           strings.ToLower(s.Host),
		"X-TC-Action":    strings.ToLower(s.Action),
		"X-TC-Timestamp": fmt.Sprintf("%d", timestamp),
		"X-TC-Version":   s.Version,
		"X-TC-Region":    s.Region,
	}
	return authorization, actionHeader, headers, nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, msg string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(msg))
	return mac.Sum(nil)
}
