package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eqs/server/internal/config"
)

// 公开预览签名：供服务商主页案例图等"未登录可浏览"场景使用。
// 签名绑定 file_id + 过期时间（HMAC-SHA256 + JWTSecret），过期自动失效，
// 避免将受保护文件直接暴露为永久公开 URL。

// previewTokenTTL 公开预览签名有效期
const previewTokenTTL = 24 * time.Hour

// signPreviewToken 生成公开预览签名，返回 "exp.sign" 形式
func signPreviewToken(fileID uint) string {
	exp := time.Now().Add(previewTokenTTL).Unix()
	payload := fmt.Sprintf("preview|%d|%d", fileID, exp)
	mac := hmac.New(sha256.New, []byte(config.Get().JWTSecret))
	mac.Write([]byte(payload))
	tok := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%d.%s", exp, tok)
}

// verifyPreviewToken 校验公开预览签名（过期/篡改/文件不匹配均拒绝）
func verifyPreviewToken(token string, fileID uint) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}
	exp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || exp < time.Now().Unix() {
		return false
	}
	payload := fmt.Sprintf("preview|%d|%d", fileID, exp)
	mac := hmac.New(sha256.New, []byte(config.Get().JWTSecret))
	mac.Write([]byte(payload))
	expect := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expect), []byte(parts[1]))
}
