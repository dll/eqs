package config

import "os"

type Config struct {
	ServerPort  string
	AppEnv      string
	DBDriver    string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	RedisAddr   string
	RedisPass   string
	JWTSecret   string
	SMSAppKey   string
	OSSEndpoint string
	OSSBucket   string

	// V6 交易与电子签约
	PaymentProvider            string
	PaymentMchID               string
	PaymentAPIKey              string
	PaymentNotifySecret        string
	ESignProvider              string
	ESignAppID                 string
	ESignAPIKey                string
	QualificationVerifyProvider string

	// P1-09 敏感字段加密密钥（未配置时字段以明文存取，仅允许开发环境）
	DataEncryptionKey string

	// 中国国产 AI 模型
	BaiduAPIKey    string
	BaiduSecretKey string
	AliyunAPIKey   string
	IflyAPIKey     string
	IflyAPISecret  string
}

func Load() *Config {
	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		AppEnv:      getEnv("APP_ENV", "development"),
		DBDriver:    getEnv("DB_DRIVER", "mysql"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "3306"),
		DBUser:      getEnv("DB_USER", "root"),
		DBPassword:  getEnv("DB_PASSWORD", "root"),
		DBName:      getEnv("DB_NAME", "eqs"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:   getEnv("REDIS_PASS", ""),
		JWTSecret:   getEnv("JWT_SECRET", "eqs-secret-key"),
		SMSAppKey:   getEnv("SMS_APP_KEY", ""),
		OSSEndpoint: getEnv("OSS_ENDPOINT", ""),
		OSSBucket:   getEnv("OSS_BUCKET", ""),

		PaymentProvider: getEnv("PAYMENT_PROVIDER", "mock"),
		PaymentMchID: getEnv("PAYMENT_MCH_ID", ""),
		PaymentAPIKey: getEnv("PAYMENT_API_KEY", ""),
		PaymentNotifySecret: getEnv("PAYMENT_NOTIFY_SECRET", ""),
		ESignProvider: getEnv("ESIGN_PROVIDER", "mock"),
		ESignAppID: getEnv("ESIGN_APP_ID", ""),
		ESignAPIKey: getEnv("ESIGN_API_KEY", ""),
		QualificationVerifyProvider: getEnv("QUALIFICATION_VERIFY_PROVIDER", "manual"),

		DataEncryptionKey: getEnv("DATA_ENCRYPTION_KEY", ""),

		BaiduAPIKey:    getEnv("BAIDU_API_KEY", ""),
		BaiduSecretKey: getEnv("BAIDU_SECRET_KEY", ""),
		AliyunAPIKey:   getEnv("ALIYUN_API_KEY", ""),
		IflyAPIKey:     getEnv("IFLY_API_KEY", ""),
		IflyAPISecret:  getEnv("IFLY_API_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// IsProduction 是否生产环境
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// IsTest 是否测试环境（允许固定验证码等测试便利）
func (c *Config) IsTest() bool {
	return c.AppEnv == "test" || c.AppEnv == "testing"
}