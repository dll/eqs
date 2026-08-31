package handler

import (
	"net/http"
	"testing"

	"github.com/eqs/server/internal/config"
)

// TestWxLogin_MockSwitch 强制 WX_MINI_MOCK=1：即便配置了凭据也走 mock（openid_<code>）
func TestWxLogin_MockSwitch(t *testing.T) {
	// 配了凭据但强制 mock，应成功且 openid 为 openid_<code>
	t.Setenv("WX_MINI_APPID", "wx_placeholder_appid")
	t.Setenv("WX_MINI_SECRET", "placeholder_secret")
	t.Setenv("WX_MINI_MOCK", "1")
	config.ResetCache()
	defer config.ResetCache()

	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/auth/wechat-login", map[string]interface{}{
		"code": "switch_code", "user_type": 2,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("强制 mock 登录应200，得到 %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	if out["token"] == nil || out["token"] == "" {
		t.Fatalf("未返回 token: %v", out)
	}
}

// TestWxLogin_RealButUnconfigured 配了 AppID/Secret + mock 关闭：真实通道在测试环境会失败——
// 该用例保证"误配真实通道但微信不可达"时返回业务失败而非 panic。
func TestWxLogin_RealButUnavailable(t *testing.T) {
	t.Setenv("WX_MINI_APPID", "wx_t")
	t.Setenv("WX_MINI_SECRET", "sec_t")
	t.Setenv("WX_MINI_MOCK", "0")
	config.ResetCache()
	defer config.ResetCache()

	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/auth/wechat-login", map[string]interface{}{
		"code": "any", "user_type": 2,
	})
	// 真实通道不可达 → 应返回业务错误（4xx），不能 5xx/panic
	if w.Code == http.StatusInternalServerError {
		t.Fatalf("真实通道失败不应返回500: %d %s", w.Code, w.Body.String())
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("真实通道不可达应400，得到 %d %s", w.Code, w.Body.String())
	}
}
