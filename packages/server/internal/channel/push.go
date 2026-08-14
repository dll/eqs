package channel

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// md5Hex 个推鉴权签名：md5(appkey + masterSecret + timestamp)
func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// Pusher App 推送适配器接口
type Pusher interface {
	// PushByAlias 按别名推送（alias = 用户手机号，客户端登录后上报）
	PushByAlias(alias, title, content string) error
}

// NewPusher 依据配置构造推送器；未配置 uni-push 凭据时返回 nil（降级 SSE/轮询）
func NewPusher(appID, appKey, masterSecret string) Pusher {
	if appID == "" || appKey == "" || masterSecret == "" {
		return nil
	}
	return &UniPushSender{
		appID: appID, appKey: appKey, masterSecret: masterSecret,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// UniPushSender DCloud uni-push（个推）服务端推送
// 文档：https://docs.getui.com/getui/server/rest_v2/push/
type UniPushSender struct {
	appID        string
	appKey       string
	masterSecret string
	client       *http.Client
}

type uniPushResp struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

// authToken 获取推送鉴权 token
func (u *UniPushSender) authToken() (string, error) {
	body, _ := json.Marshal(map[string]string{
		"sign": md5Hex(u.appKey + u.masterSecret + u.now()),
		"timestamp": u.now(),
		"appkey":    u.appKey,
	})
	req, err := http.NewRequest("POST", "https://restapi.getui.com/v2/"+u.appID+"/auth", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	resp, err := u.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var r uniPushResp
	if err := json.Unmarshal(data, &r); err != nil {
		return "", fmt.Errorf("uni-push 鉴权响应解析失败: %s", truncate(data, 200))
	}
	if r.Code != 0 || r.Data.Token == "" {
		return "", fmt.Errorf("uni-push 鉴权失败: code=%d msg=%s", r.Code, r.Msg)
	}
	return r.Data.Token, nil
}

// PushByAlias 按别名单推
func (u *UniPushSender) PushByAlias(alias, title, content string) error {
	token, err := u.authToken()
	if err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]interface{}{
		"request_id": fmt.Sprintf("eqs%d", time.Now().UnixNano()),
		"audience":   map[string]interface{}{"alias": []string{alias}},
		"push_message": map[string]interface{}{
			"notification": map[string]interface{}{
				"title": title, "body": content,
				"click_type": "intent",
				"intent":     "intent://io.dcloud.uniapp.eqs/#Intent;scheme=eqs;launchFlags=0x4000000;end;",
			},
		},
	})
	req, err := http.NewRequest("POST", "https://restapi.getui.com/v2/"+u.appID+"/push/single/alias", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	req.Header.Set("token", token)
	resp, err := u.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var r uniPushResp
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("uni-push 推送响应解析失败: %s", truncate(data, 200))
	}
	if r.Code != 0 {
		return fmt.Errorf("uni-push 推送失败: code=%d msg=%s", r.Code, r.Msg)
	}
	return nil
}

func (u *UniPushSender) now() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}
