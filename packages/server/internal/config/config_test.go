package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// 清空环境变量确保走默认值
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("DB_HOST")

	cfg := Load()
	if cfg.ServerPort != "8080" {
		t.Fatalf("默认端口应为8080，得到 %s", cfg.ServerPort)
	}
	if cfg.DBHost != "localhost" {
		t.Fatalf("默认DB主机应为localhost，得到 %s", cfg.DBHost)
	}
	if cfg.PaymentProvider != "mock" {
		t.Fatalf("支付通道默认应为mock，得到 %s", cfg.PaymentProvider)
	}
	if cfg.ESignProvider != "mock" {
		t.Fatalf("签约通道默认应为mock，得到 %s", cfg.ESignProvider)
	}
	if cfg.JWTSecret != "eqs-secret-key" {
		t.Fatalf("JWT密钥默认值异常: %s", cfg.JWTSecret)
	}
}

func TestLoad_WithEnv(t *testing.T) {
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("DB_HOST", "db.internal")
	os.Setenv("PAYMENT_PROVIDER", "tenpay")
	os.Setenv("ESIGN_PROVIDER", "esign")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("PAYMENT_PROVIDER")
		os.Unsetenv("ESIGN_PROVIDER")
	}()

	cfg := Load()
	if cfg.ServerPort != "9090" {
		t.Fatalf("期望端口9090，得到 %s", cfg.ServerPort)
	}
	if cfg.DBHost != "db.internal" {
		t.Fatalf("期望DB主机db.internal，得到 %s", cfg.DBHost)
	}
	if cfg.PaymentProvider != "tenpay" {
		t.Fatalf("期望支付通道tenpay，得到 %s", cfg.PaymentProvider)
	}
	if cfg.ESignProvider != "esign" {
		t.Fatalf("期望签约通道esign，得到 %s", cfg.ESignProvider)
	}
}