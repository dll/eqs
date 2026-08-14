package channel

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OcrClient 图片文字识别适配器接口
type OcrClient interface {
	// RecognizeText 识别图片中的文字，返回文本行列表
	RecognizeText(image []byte) ([]string, error)
}

// NewOcrClient 依据配置构造 OCR 客户端；未配置腾讯云凭据时返回 nil（降级规则/AI 审核）
func NewOcrClient(secretID, secretKey string) OcrClient {
	if secretID == "" || secretKey == "" {
		return nil
	}
	return &TencentOcrClient{
		sign: TencentCloudSign{
			SecretID: secretID, SecretKey: secretKey,
			Service: "ocr", Host: "ocr.tencentcloudapi.com",
			Action: "GeneralBasicOCR", Region: "ap-guangzhou", Version: "2018-11-19",
		},
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// TencentOcrClient 腾讯云通用印刷体识别（GeneralBasicOCR）
type TencentOcrClient struct {
	sign   TencentCloudSign
	client *http.Client
}

func (o *TencentOcrClient) RecognizeText(image []byte) ([]string, error) {
	if len(image) > 7*1024*1024 {
		return nil, fmt.Errorf("图片超过 7MB 限制")
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"ImageBase64":  base64.StdEncoding.EncodeToString(image),
		"LanguageType": "zh",
	})
	auth, _, headers, err := o.sign.Sign(payload)
	if err != nil {
		return nil, err
	}
	headers["Authorization"] = auth

	req, err := http.NewRequest("POST", "https://ocr.tencentcloudapi.com", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("腾讯云 OCR 返回 %d: %s", resp.StatusCode, truncate(body, 200))
	}
	var r struct {
		Response struct {
			TextDetections []struct {
				DetectedText string `json:"DetectedText"`
			} `json:"TextDetections"`
			Error *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("腾讯云 OCR 响应解析失败")
	}
	if r.Response.Error != nil {
		return nil, fmt.Errorf("腾讯云 OCR 失败: %s %s", r.Response.Error.Code, r.Response.Error.Message)
	}
	lines := make([]string, 0, len(r.Response.TextDetections))
	for _, d := range r.Response.TextDetections {
		if t := trimSpace(d.DetectedText); t != "" {
			lines = append(lines, t)
		}
	}
	return lines, nil
}

func trimSpace(s string) string {
	out := []rune{}
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
