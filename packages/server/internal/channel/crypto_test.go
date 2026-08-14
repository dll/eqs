package channel

import (
	"testing"
	"time"
)

// TestTencentCloudSign_Golden 黄金测试：与独立 Python 实现计算的签名一致。
// 测试向量来源：腾讯云 API v3 文档示例密钥 + 固定时间戳，签名经 Python 独立计算后固化。
func TestTencentCloudSign_Golden(t *testing.T) {
	s := TencentCloudSign{
		SecretID:  "TESTONLY_TC3_SECRETID_FOR_UNIT",
		SecretKey: "TESTONLY_TC3_SECRETKEY_FOR_UNIT",
		Service:   "sms",
		Host:      "sms.tencentcloudapi.com",
		Action:    "SendSms",
		Region:    "ap-guangzhou",
		Version:   "2021-01-11",
		Now:       time.Unix(1551113065, 0),
	}
	payload := []byte(`{"phone":"13800138000","code":"123456"}`)

	auth, action, headers, err := s.Sign(payload)
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}
	const wantAuth = "TC3-HMAC-SHA256 Credential=TESTONLY_TC3_SECRETID_FOR_UNIT/2019-02-25/sms/tc3_request, SignedHeaders=content-type;host;x-tc-action, Signature=3426921fcb6e150e021e532ac0f1cb263ba9091e717a5fed878d13d40b129345"
	if auth != wantAuth {
		t.Fatalf("Authorization 不符\n期望: %s\n实际: %s", wantAuth, auth)
	}
	if action != "sendsms" {
		t.Errorf("x-tc-action = %s, 期望 sendsms", action)
	}
	if headers["X-TC-Timestamp"] != "1551113065" {
		t.Errorf("时间戳头不符: %s", headers["X-TC-Timestamp"])
	}
	if headers["Host"] != "sms.tencentcloudapi.com" {
		t.Errorf("Host 头不符: %s", headers["Host"])
	}
}

func TestTencentCloudSign_NoSecret(t *testing.T) {
	s := TencentCloudSign{Service: "sms", Host: "sms.tencentcloudapi.com"}
	if _, _, _, err := s.Sign([]byte(`{}`)); err == nil {
		t.Fatal("未配置密钥应返回错误")
	}
}
