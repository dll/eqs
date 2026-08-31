package config

import (
	"os"
	"strings"
	"sync"
)

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
	PaymentProvider             string
	PaymentMchID                string
	PaymentAPIKey               string
	PaymentNotifySecret         string
	ESignProvider               string
	ESignAppID                  string
	ESignAPIKey                 string
	QualificationVerifyProvider string

	// V10 外部通道凭据（代码已就绪，仅剩填入；未配置时自动降级）
	// 腾讯云短信
	TencentSMSSecretID  string
	TencentSMSSecretKey string
	SMSSDKAppID         string
	SMSSignName         string
	SMSTemplateID       string
	// 腾讯云 OCR
	TencentOCRSecretID  string
	TencentOCRSecretKey string
	// 微信支付 v3
	WXPayAppID           string
	WXPayMchID           string
	WXPayAPIV3Key        string
	WXPayMchSerialNo     string
	WXPayMchPrivateKeyFile string
	WXPayPlatformCertFile  string
	WXPayNotifyURL       string
	// uni-push（App 推送）
	PushAppID       string
	PushAppKey      string
	PushMasterSecret string

	// P1-09 敏感字段加密密钥（未配置时字段以明文存取，仅允许开发环境）
	DataEncryptionKey string

	// CORS 信任来源白名单（生产环境生效；逗号分隔，默认信任 eqs-chzu.tech 及 www）
	CORSAllowedOrigins []string

	// UploadDir 本地文件上传目录（资质扫描件等附件；相对工作目录或绝对路径）
	UploadDir string

	// CADConvertAPI 第三方 CAD 在线渲染转换服务地址（DWG 预览用）。
	// 未配置时 DWG 返回"请下载查看"提示；配置后 /file/:id/preview 将文件转发至
	// 该服务（multipart file + format=svg）并内联返回转换结果。
	// 典型部署：自建 Aspose.CAD 封装服务 / 商业 CAD 渲染网关。
	CADConvertAPI string

	// 中国国产 AI 模型
	BaiduAPIKey    string
	BaiduSecretKey string
	AliyunAPIKey   string
	IflyAPIKey     string
	IflyAPISecret  string

	// V11：微信小程序登录（code2session）
	// 未配置 AppID/Secret 时自动降级为 mock（保留 openid_<code> 行为）；
	// WX_MINI_MOCK=1 可强制 mock（开发/CI 用），即使已配置凭据也不真实请求。
	WXMiniAppID   string
	WXMiniSecret  string
	WXMiniMock    bool
}

// Load 从环境变量构造完整配置（每次调用都会重新解析环境变量）
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

		PaymentProvider:             getEnv("PAYMENT_PROVIDER", "mock"),
		PaymentMchID:                getEnv("PAYMENT_MCH_ID", ""),
		PaymentAPIKey:               getEnv("PAYMENT_API_KEY", ""),
		PaymentNotifySecret:         getEnv("PAYMENT_NOTIFY_SECRET", ""),
		ESignProvider:               getEnv("ESIGN_PROVIDER", "mock"),
		ESignAppID:                  getEnv("ESIGN_APP_ID", ""),
		ESignAPIKey:                 getEnv("ESIGN_API_KEY", ""),
		QualificationVerifyProvider: getEnv("QUALIFICATION_VERIFY_PROVIDER", "manual"),

		// V10 外部通道凭据
		TencentSMSSecretID:  getEnv("TENCENT_SMS_SECRET_ID", ""),
		TencentSMSSecretKey: getEnv("TENCENT_SMS_SECRET_KEY", ""),
		SMSSDKAppID:         getEnv("SMS_SDK_APP_ID", ""),
		SMSSignName:         getEnv("SMS_SIGN_NAME", ""),
		SMSTemplateID:       getEnv("SMS_TEMPLATE_ID", ""),
		TencentOCRSecretID:  getEnv("TENCENT_OCR_SECRET_ID", ""),
		TencentOCRSecretKey: getEnv("TENCENT_OCR_SECRET_KEY", ""),
		WXPayAppID:          getEnv("WXPAY_APPID", ""),
		WXPayMchID:          getEnv("WXPAY_MCHID", ""),
		WXPayAPIV3Key:       getEnv("WXPAY_API_V3_KEY", ""),
		WXPayMchSerialNo:    getEnv("WXPAY_MCH_SERIAL_NO", ""),
		WXPayMchPrivateKeyFile: getEnv("WXPAY_MCH_PRIVATE_KEY_FILE", ""),
		WXPayPlatformCertFile:  getEnv("WXPAY_PLATFORM_CERT_FILE", ""),
		WXPayNotifyURL:      getEnv("WXPAY_NOTIFY_URL", ""),
		PushAppID:           getEnv("PUSH_APP_ID", ""),
		PushAppKey:          getEnv("PUSH_APP_KEY", ""),
		PushMasterSecret:    getEnv("PUSH_MASTER_SECRET", ""),

		DataEncryptionKey: getEnv("DATA_ENCRYPTION_KEY", ""),

		// CORS 白名单：CORS_ALLOW_ORIGINS 逗号分隔；未配置时默认信任生产域名
		CORSAllowedOrigins: parseOrigins(getEnv("CORS_ALLOW_ORIGINS", "https://eqs-chzu.tech,https://www.eqs-chzu.tech")),

		// 本地上传目录（资质扫描件附件；相对工作目录或绝对路径）
		UploadDir: getEnv("UPLOAD_DIR", "uploads/qualifications"),

		// 第三方 CAD 渲染转换服务（DWG 预览；未配置则提示下载）
		CADConvertAPI: getEnv("CAD_CONVERT_API", ""),

		BaiduAPIKey:    getEnv("BAIDU_API_KEY", ""),
		BaiduSecretKey: getEnv("BAIDU_SECRET_KEY", ""),
		AliyunAPIKey:   getEnv("ALIYUN_API_KEY", ""),
		IflyAPIKey:     getEnv("IFLY_API_KEY", ""),
		IflyAPISecret:  getEnv("IFLY_API_SECRET", ""),

		WXMiniAppID:  getEnv("WX_MINI_APPID", ""),
		WXMiniSecret: getEnv("WX_MINI_SECRET", ""),
		WXMiniMock:   getEnv("WX_MINI_MOCK", "") == "1",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

var (
	cachedOnce sync.Once
	cachedCfg  *Config
)

// Get 返回进程内缓存的配置（首次调用时从环境变量构造一次，之后复用同一实例）。
// 配置在进程生命周期内恒定，热路径（如 CORS 中间件、错误处理）应使用 Get 而非每次 Load。
func Get() *Config {
	cachedOnce.Do(func() {
		cachedCfg = Load()
	})
	return cachedCfg
}

// ResetCache 清空配置缓存（测试使用；生产代码不应调用，配置进程内恒定）。
func ResetCache() {
	cachedOnce = sync.Once{}
	cachedCfg = nil
}

// parseOrigins 解析逗号分隔的 CORS 白名单为切片
func parseOrigins(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// IsProduction 是否生产环境
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// IsTest 是否测试环境（允许固定验证码等测试便利）
func (c *Config) IsTest() bool {
	return c.AppEnv == "test" || c.AppEnv == "testing"
}
