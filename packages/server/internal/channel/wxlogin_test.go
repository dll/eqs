package channel

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestWxExchanger_Mock 无凭据/强制 mock 时返回 openid_<code>（保持旧行为）
func TestWxExchanger_Mock(t *testing.T) {
	cases := []struct {
		name   string
		appID  string
		secret string
		mock   bool
		code   string
		want   string
	}{
		{"无凭据自动降级 mock", "", "", false, "abc", "openid_abc"},
		{"仅 AppID 缺 Secret 降级", "wx123", "", false, "x", "openid_x"},
		{"强制 mock 即便有凭据", "wx123", "sec", true, "y", "openid_y"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewWxExchanger(tc.appID, tc.secret, tc.mock)
			resp, err := e.Code2Session(tc.code)
			if err != nil {
				t.Fatalf("mock 不应报错: %v", err)
			}
			if resp.OpenID != tc.want {
				t.Fatalf("openid 不符: got %q want %q", resp.OpenID, tc.want)
			}
		})
	}
}

// TestWxExchanger_Real_Success 模拟微信返回合法 openid
func TestWxExchanger_Real_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sns/jscode2session" {
			t.Errorf("路径不符: %s", r.URL.Path)
		}
		if r.URL.Query().Get("grant_type") != "authorization_code" {
			t.Errorf("缺 grant_type")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"openid": "real_openid_001", "session_key": "sk", "unionid": "union_1", "errcode": 0,
		})
	}))
	defer srv.Close()

	oldURL := wxCode2SessionURL
	oldClient := wxLoginHTTPClient
	wxCode2SessionURL = srv.URL + "/sns/jscode2session"
	wxLoginHTTPClient = srv.Client()
	defer func() {
		wxCode2SessionURL = oldURL
		wxLoginHTTPClient = oldClient
	}()

	e := &code2sessionExchanger{appID: "wx123", secret: "sec"}
	resp, err := e.Code2Session("js_code")
	if err != nil {
		t.Fatalf("code2session 不应报错: %v", err)
	}
	if resp.OpenID != "real_openid_001" {
		t.Fatalf("openid 不符: got %q", resp.OpenID)
	}
}

// TestWxExchanger_Real_WechatError 微信侧返回错误码
func TestWxExchanger_Real_WechatError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 40029, "errmsg": "invalid code",
		})
	}))
	defer srv.Close()

	oldURL := wxCode2SessionURL
	oldClient := wxLoginHTTPClient
	wxCode2SessionURL = srv.URL + "/sns/jscode2session"
	wxLoginHTTPClient = srv.Client()
	defer func() {
		wxCode2SessionURL = oldURL
		wxLoginHTTPClient = oldClient
	}()

	e := &code2sessionExchanger{appID: "wx123", secret: "sec"}
	if _, err := e.Code2Session("bad_code"); err == nil {
		t.Fatal("微信返回错误码时应报错")
	}
}

// TestWxExchanger_Real_HTTPError 网络/接口错误
func TestWxExchanger_Real_HTTPError(t *testing.T) {
	// 指向关闭的 server，触发连接错误
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	addr := srv.URL + "/sns/jscode2session"
	srv.Close()

	oldURL := wxCode2SessionURL
	oldClient := wxLoginHTTPClient
	wxCode2SessionURL = addr
	wxLoginHTTPClient = &http.Client{}
	defer func() {
		wxCode2SessionURL = oldURL
		wxLoginHTTPClient = oldClient
	}()

	e := &code2sessionExchanger{appID: "wx123", secret: "sec"}
	if _, err := e.Code2Session("code"); err == nil {
		t.Fatal("网络错误时应报错")
	}
}
