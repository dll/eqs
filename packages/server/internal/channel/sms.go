package channel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SmsSender 短信发送适配器接口
type SmsSender interface {
	// SendVerificationCode 发送登录验证码
	SendVerificationCode(phone, code string) error
}

// NewSmsSender 依据配置构造短信发送器；未配置腾讯云短信凭据时返回 nil（由调用方降级处理）
func NewSmsSender(secretID, secretKey, sdkAppID, signName, templateID string) SmsSender {
	if secretID == "" || secretKey == "" || sdkAppID == "" || signName == "" || templateID == "" {
		return nil
	}
	return &TencentSmsSender{
		sign: TencentCloudSign{
			SecretID: secretID, SecretKey: secretKey,
			Service: "sms", Host: "sms.tencentcloudapi.com",
			Action: "SendSms", Region: "ap-guangzhou", Version: "2021-01-11",
		},
		sdkAppID:   sdkAppID,
		signName:   signName,
		templateID: templateID,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// TencentSmsSender 腾讯云短信（SendSms，TC3 签名）
type TencentSmsSender struct {
	sign       TencentCloudSign
	sdkAppID   string
	signName   string
	templateID string
	client     *http.Client
}

func (s *TencentSmsSender) SendVerificationCode(phone, code string) error {
	// 手机号标准化为 +86 前缀（腾讯云国际区号格式）
	normalized := phone
	if !strings.HasPrefix(normalized, "+") {
		normalized = "+86" + normalized
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"PhoneNumberSet":   []string{normalized},
		"SmsSdkAppId":      s.sdkAppID,
		"SignName":         s.signName,
		"TemplateId":       s.templateID,
		"TemplateParamSet": []string{code},
		"SessionContext":   "",
	})
	auth, _, headers, err := s.sign.Sign(payload)
	if err != nil {
		return err
	}
	headers["Authorization"] = auth

	req, err := http.NewRequest("POST", "https://sms.tencentcloudapi.com", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("腾讯云短信返回 %d: %s", resp.StatusCode, truncate(body, 200))
	}
	var r struct {
		Response struct {
			SendStatusSet []struct {
				Code string `json:"Code"`
				Msg  string `json:"Message"`
			} `json:"SendStatusSet"`
		} `json:"Response"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return fmt.Errorf("腾讯云短信响应解析失败: %s", truncate(body, 200))
	}
	if len(r.Response.SendStatusSet) == 0 || r.Response.SendStatusSet[0].Code != "Ok" {
		if len(r.Response.SendStatusSet) > 0 {
			return fmt.Errorf("腾讯云短信发送失败: %s", r.Response.SendStatusSet[0].Msg)
		}
		return fmt.Errorf("腾讯云短信发送失败: %s", truncate(body, 200))
	}
	return nil
}

func truncate(b []byte, n int) string {
	s := string(b)
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
