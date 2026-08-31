package channel

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ==================== 微信小程序登录（code2session）适配器 ====================
// V11：真实小程序登录走 api.weixin.qq.com/sns/jscode2session；
// 未配置 AppID/Secret 或 WX_MINI_MOCK=1 时降级为 mock（openid_<code>），保留开发/CI 可用。

// WxCode2SessionResp 微信 code2session 响应
type WxCode2SessionResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// Exchanger 微信小程序登录凭据交换抽象；便于测试注入。
type WxExchanger interface {
	Code2Session(code string) (*WxCode2SessionResp, error)
}

// 微信 code2session 接口地址（可变以便测试注入 URL 或跳过网络）
var wxCode2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"

// httpClient 可测试替换；默认为带超时的 client。
var wxLoginHTTPClient = &http.Client{Timeout: 8 * time.Second}

// mockExchanger 仿真实现：openid = "openid_" + code（与旧 WxLogin 行为完全一致）。
type mockExchanger struct{}

func (m *mockExchanger) Code2Session(code string) (*WxCode2SessionResp, error) {
	return &WxCode2SessionResp{OpenID: fmt.Sprintf("openid_%s", code)}, nil
}

// code2sessionExchanger 真实实现：调用微信 jscode2session 接口。
type code2sessionExchanger struct {
	appID  string
	secret string
}

func (e *code2sessionExchanger) Code2Session(code string) (*WxCode2SessionResp, error) {
	if e.appID == "" || e.secret == "" {
		return nil, errors.New("微信小程序未配置 WX_MINI_APPID / WX_MINI_SECRET")
	}
	api := wxCode2SessionURL + "?appid=" + url.QueryEscape(e.appID) +
		"&secret=" + url.QueryEscape(e.secret) +
		"&js_code=" + url.QueryEscape(code) +
		"&grant_type=authorization_code"
	resp, err := wxLoginHTTPClient.Get(api)
	if err != nil {
		return nil, fmt.Errorf("调用微信 code2session 失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("读取微信 code2session 响应失败: %w", err)
	}
	var out WxCode2SessionResp
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("解析微信 code2session 响应失败: %s", truncate(body, 200))
	}
	if out.ErrCode != 0 {
		return nil, fmt.Errorf("微信 code2session 返回错误 errcode=%d errmsg=%s", out.ErrCode, out.ErrMsg)
	}
	if out.OpenID == "" {
		return nil, errors.New("微信 code2session 响应缺少 openid")
	}
	return &out, nil
}

// NewWxExchanger 依据配置构造登录交换器。
//   - useMock=true（WX_MINI_MOCK=1）→ 强制 mock，即便已配置凭据；
//   - appID/secret 任一为空 → mock；
//   - 否则 → 真实 code2session。
func NewWxExchanger(appID, secret string, useMock bool) WxExchanger {
	if useMock || appID == "" || secret == "" {
		return &mockExchanger{}
	}
	return &code2sessionExchanger{appID: appID, secret: secret}
}
