package handler

import (
	"fmt"
	"strings"
)

// publicConfigKeyBlocked 判断配置键是否明确属于凭据或密钥类配置。
// 公开配置采用允许通过、敏感键拒绝的最小门禁；私密配置不受此规则影响。
func publicConfigKeyBlocked(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, marker := range []string{
		"secret", "password", "passwd", "token", "api_key", "apikey", "access_key",
		"private_key", "credential", "encryption_key", "jwt", "authorization",
	} {
		if strings.Contains(key, marker) {
			return true
		}
	}
	// 兼容常见的点号、短横线命名（例如 payment.key、wx-key）。
	return strings.Contains(key, ".key") || strings.HasSuffix(key, "_key") || strings.HasSuffix(key, "-key")
}

// publicConfigValueBlocked 拦截即使键名不敏感也不应进入公开缓存的凭据形态。
// 不做通用高熵检测，避免误伤公开的版本号、URL 或业务 JSON。
func publicConfigValueBlocked(value string) bool {
	trimmed := strings.TrimSpace(value)
	upper := strings.ToUpper(trimmed)
	if strings.Contains(upper, "-----BEGIN ") && strings.Contains(upper, "PRIVATE KEY-----") {
		return true
	}
	// JWT 的三个 base64url 段通常以 eyJ 开头；仅对完整三段形态拦截。
	parts := strings.Split(trimmed, ".")
	return len(parts) == 3 && strings.HasPrefix(parts[0], "eyJ") && parts[1] != "" && parts[2] != ""
}

func validatePublicConfig(key, value string) error {
	if publicConfigKeyBlocked(key) || publicConfigValueBlocked(value) {
		return fmt.Errorf("配置键或配置值包含敏感信息，不允许公开")
	}
	return nil
}
