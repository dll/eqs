package channel

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// PaymentGateway 支付网关适配器接口（微信支付 v3 / Mock）
type PaymentGateway interface {
	// CreateNativeOrder 创建 Native（扫码）支付订单，返回 code_url
	CreateNativeOrder(outTradeNo string, amountFen int64, description, notifyURL string) (codeURL string, err error)
	// CreateJSAPIOrder 创建 JSAPI（公众号/小程序）支付订单，返回 prepay_id
	CreateJSAPIOrder(openID, outTradeNo string, amountFen int64, description, notifyURL string) (prepayID string, err error)
	// VerifyAndDecryptNotify 验签并解密支付回调，返回商户订单号与支付金额（分）
	VerifyAndDecryptNotify(header http.Header, body []byte) (outTradeNo string, totalFen int64, transactionID string, err error)
	// Refund 发起退款（totalFen=原单金额，refundFen=退款金额，单位分）
	Refund(outTradeNo, outRefundNo string, totalFen, refundFen int64) error
}

// NewPaymentGateway 依据配置构造支付网关：PAYMENT_PROVIDER=wechat 且凭据齐全时返回
// 微信支付 v3 网关；否则返回 nil（调用方使用 Mock）。
func NewPaymentGateway(provider, appID, mchID, apiV3Key, mchSerialNo, privateKeyFile, platformCertFile string) (PaymentGateway, error) {
	if provider != "wechat" {
		return nil, nil
	}
	if appID == "" || mchID == "" || apiV3Key == "" || mchSerialNo == "" {
		return nil, fmt.Errorf("微信支付配置不完整（WXPAY_APPID/WXPAY_MCHID/WXPAY_API_V3_KEY/WXPAY_MCH_SERIAL_NO）")
	}
	priv, err := loadRSAPrivateKey(privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("加载商户私钥失败: %w", err)
	}
	pub, err := loadRSAPublicKey(platformCertFile)
	if err != nil {
		return nil, fmt.Errorf("加载平台证书失败: %w", err)
	}
	return &WechatPayV3{
		mchID: mchID, appID: appID, apiV3Key: apiV3Key, mchSerialNo: mchSerialNo,
		privateKey: priv, platformKey: pub,
		client: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// WechatPayV3 微信支付 API v3 适配器（纯标准库：RSA-SHA256 签名 + AES-256-GCM 回调解密）
type WechatPayV3 struct {
	mchID        string
	appID        string
	apiV3Key     string
	mchSerialNo  string
	privateKey   *rsa.PrivateKey
	platformKey  *rsa.PublicKey
	client       *http.Client
}

const wechatBase = "https://api.mch.weixin.qq.com"

// signRequest 生成 WECHATPAY2-SHA256-RSA2048 请求签名头
func (w *WechatPayV3) signRequest(method, url, body string) (string, error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", err
	}
	nonce := hex.EncodeToString(nonceBytes)
	return w.authHeader(method, url, ts, nonce, body)
}

// signMessage 计算 RSA-SHA256 签名（base64），供内部与黄金测试复用
func (w *WechatPayV3) signMessage(method, url, ts, nonce, body string) (string, error) {
	message := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", method, url, ts, nonce, body)
	digest := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPKCS1v15(rand.Reader, w.privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// authHeader 组装完整 Authorization 头
func (w *WechatPayV3) authHeader(method, url, ts, nonce, body string) (string, error) {
	signature, err := w.signMessage(method, url, ts, nonce, body)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",timestamp="%s",serial_no="%s",signature="%s"`,
		w.mchID, nonce, ts, w.mchSerialNo, signature), nil
}

// postJSON 带签名发送 JSON 请求
func (w *WechatPayV3) postJSON(path string, body interface{}) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	auth, err := w.signRequest("POST", path, string(data))
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", wechatBase+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "eqs-server/1.0")
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("微信支付接口 %s 返回 %d: %s", path, resp.StatusCode, truncate(respBody, 300))
	}
	return respBody, nil
}

// CreateNativeOrder 统一下单（Native 扫码支付）
func (w *WechatPayV3) CreateNativeOrder(outTradeNo string, amountFen int64, description, notifyURL string) (string, error) {
	body := map[string]interface{}{
		"appid": w.appID, "mchid": w.mchID,
		"description":  description,
		"out_trade_no": outTradeNo,
		"notify_url":   notifyURL,
		"amount":       map[string]int64{"total": amountFen},
	}
	respBody, err := w.postJSON("/v3/pay/transactions/native", body)
	if err != nil {
		return "", err
	}
	var r struct {
		CodeURL string `json:"code_url"`
	}
	if err := json.Unmarshal(respBody, &r); err != nil {
		return "", fmt.Errorf("解析下单响应失败: %s", truncate(respBody, 200))
	}
	if r.CodeURL == "" {
		return "", fmt.Errorf("下单响应缺少 code_url: %s", truncate(respBody, 200))
	}
	return r.CodeURL, nil
}

// CreateJSAPIOrder 统一下单（JSAPI 公众号/小程序支付）
func (w *WechatPayV3) CreateJSAPIOrder(openID, outTradeNo string, amountFen int64, description, notifyURL string) (string, error) {
	body := map[string]interface{}{
		"appid": w.appID, "mchid": w.mchID,
		"description":  description,
		"out_trade_no": outTradeNo,
		"notify_url":   notifyURL,
		"amount":       map[string]int64{"total": amountFen},
		"payer":        map[string]string{"openid": openID},
	}
	respBody, err := w.postJSON("/v3/pay/transactions/jsapi", body)
	if err != nil {
		return "", err
	}
	var r struct {
		PrepayID string `json:"prepay_id"`
	}
	if err := json.Unmarshal(respBody, &r); err != nil {
		return "", fmt.Errorf("解析下单响应失败: %s", truncate(respBody, 200))
	}
	if r.PrepayID == "" {
		return "", fmt.Errorf("下单响应缺少 prepay_id: %s", truncate(respBody, 200))
	}
	return r.PrepayID, nil
}

// VerifyAndDecryptNotify 回调验签（平台证书公钥 RSA-SHA256）+ 解密 resource（AES-256-GCM）
func (w *WechatPayV3) VerifyAndDecryptNotify(header http.Header, body []byte) (string, int64, string, error) {
	signature := header.Get("Wechatpay-Signature")
	timestamp := header.Get("Wechatpay-Timestamp")
	nonce := header.Get("Wechatpay-Nonce")
	if signature == "" || timestamp == "" || nonce == "" {
		return "", 0, "", fmt.Errorf("回调缺少签名头")
	}
	// 1. 验签：message = timestamp\nnonce\nbody\n
	message := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, string(body))
	digest := sha256.Sum256([]byte(message))
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return "", 0, "", fmt.Errorf("回调签名 base64 解码失败")
	}
	if err := rsa.VerifyPKCS1v15(w.platformKey, crypto.SHA256, digest[:], sigBytes); err != nil {
		return "", 0, "", fmt.Errorf("回调验签失败: %v", err)
	}

	// 2. 解密 resource（AEAD_AES_256_GCM）
	var res struct {
		Resource struct {
			Ciphertext      string `json:"ciphertext"`
			Nonce           string `json:"nonce"`
			AssociatedData  string `json:"associated_data"`
			Algorithm       string `json:"algorithm"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", 0, "", fmt.Errorf("回调 JSON 解析失败")
	}
	if res.Resource.Algorithm != "AEAD_AES_256_GCM" {
		return "", 0, "", fmt.Errorf("不支持的加密算法: %s", res.Resource.Algorithm)
	}
	plain, err := aesGCMDecrypt(w.apiV3Key, res.Resource.Ciphertext, res.Resource.Nonce, res.Resource.AssociatedData)
	if err != nil {
		return "", 0, "", fmt.Errorf("回调解密失败: %v", err)
	}
	var tx struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionID string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
		Amount        struct {
			Total int64 `json:"total"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(plain, &tx); err != nil {
		return "", 0, "", fmt.Errorf("解密内容解析失败")
	}
	if tx.OutTradeNo == "" {
		return "", 0, "", fmt.Errorf("解密内容缺少商户订单号")
	}
	return tx.OutTradeNo, tx.Amount.Total, tx.TransactionID, nil
}

// Refund 退款（v3 退款接口）
func (w *WechatPayV3) Refund(outTradeNo, outRefundNo string, totalFen, refundFen int64) error {
	body := map[string]interface{}{
		"out_trade_no":  outTradeNo,
		"out_refund_no": outRefundNo,
		"amount": map[string]interface{}{
			"refund": refundFen, "total": totalFen, "currency": "CNY",
		},
	}
	_, err := w.postJSON("/v3/refund/domestic/refunds", body)
	return err
}

// aesGCMDecrypt AES-256-GCM 解密（微信支付 API v3 回调 resource）
func aesGCMDecrypt(apiV3Key, ciphertextB64, nonce, aad string) ([]byte, error) {
	key := []byte(apiV3Key)
	if len(key) != 32 {
		return nil, fmt.Errorf("API v3 密钥必须为 32 字节")
	}
	ct, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("nonce 长度不符: %d", len(nonce))
	}
	return gcm.Open(nil, []byte(nonce), ct, []byte(aad))
}

// loadRSAPrivateKey 从 PEM 文件加载商户私钥（PKCS1/PKCS8）
func loadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("PEM 解析失败")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rk, ok := key.(*rsa.PrivateKey); ok {
			return rk, nil
		}
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("无法解析 RSA 私钥")
}

// loadRSAPublicKey 从 PEM 文件加载平台证书公钥（证书或公钥 PEM）
func loadRSAPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("PEM 解析失败")
	}
	if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
		if rk, ok := cert.PublicKey.(*rsa.PublicKey); ok {
			return rk, nil
		}
	}
	if pub, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if rk, ok := pub.(*rsa.PublicKey); ok {
			return rk, nil
		}
	}
	return nil, fmt.Errorf("无法解析 RSA 公钥/证书")
}

// isConfigured 供外部判断微信支付网关是否已启用
func (w *WechatPayV3) isConfigured() bool { return w != nil && w.mchID != "" }
