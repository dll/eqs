package channel

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"
)

const (
	goldenBody      = `{"appid":"wx1234567890","mchid":"1900000001","description":"test","out_trade_no":"EQS000001","notify_url":"https://eqs.example.com/api/v1/pay/notify/wechat","amount":{"total":100}}`
	goldenSignature = "rdf7/9B0/ypGsjTwP5uaN/VGMujVi5IQBcGDZQ8OCN1X8fPawngsEfFr/dXacVMt3gY+HY7KW2UEsstxOeQPc54VPAPaUSc3qxiT8xFlk9NCPvEO/5zr1ERBlWX425mdz96gGFjCgg1FU7VGQHqFJ7etpb4t6gy5PHxzysDg40apjoQRJfwcidO/fLssCXuXWKvbT6p3mUe/OQU+X4tN2bznPQ1l94iFNON03qLe00c0/CtloqQ1oz+ncIXZLT42Zcc+zA8TVd7Zy8m1ve24s5JVb9ShySc7kBAAZQ5dqI7eaMNWT/QLJyoXmYDQEecD+fcVE0g0FpeSOyueuYlJNA=="
)

// TestWechatSign_Golden 黄金测试：微信支付 v3 签名与 Python 独立实现一致。
// 固定测试私钥见 testdata/merchant_private.pem，签名经 Python 独立计算后固化。
func TestWechatSign_Golden(t *testing.T) {
	priv, err := loadRSAPrivateKey(filepath.Join("testdata", "merchant_private.pem"))
	if err != nil {
		t.Fatalf("加载测试私钥失败: %v", err)
	}
	w := &WechatPayV3{
		mchID: "1900000001", mchSerialNo: "SER123",
		privateKey: priv,
	}

	sig, err := w.signMessage("POST", "/v3/pay/transactions/native", "1710000000", "abc123def456", goldenBody)
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}
	if sig != goldenSignature {
		t.Fatalf("签名不符\n期望: %s\n实际: %s", goldenSignature, sig)
	}

	auth, err := w.authHeader("POST", "/v3/pay/transactions/native", "1710000000", "abc123def456", goldenBody)
	if err != nil {
		t.Fatal(err)
	}
	wantAuth := `WECHATPAY2-SHA256-RSA2048 mchid="1900000001",nonce_str="abc123def456",timestamp="1710000000",serial_no="SER123",signature="` + goldenSignature + `"`
	if auth != wantAuth {
		t.Fatalf("Authorization 头不符\n期望: %s\n实际: %s", wantAuth, auth)
	}
}

// TestAesGCMDecrypt_Golden 黄金测试：支付回调 resource 解密与 Python 独立实现一致
func TestAesGCMDecrypt_Golden(t *testing.T) {
	const (
		apiV3Key   = "0123456789abcdef0123456789abcdef"
		nonce      = "abcdefghijkl"
		aad        = "transaction"
		ciphertext = "Eu6k3KOMCfFZPsMoOqloD2wBiy6brE06bs5X+PBm7J0U+ZLPGmHElHJDoIIiseio6grIxto6E6SwkG13jujZFvzDuf3KwURhJy1xCUcbhV6OnIQa9/ZDK+tzROARj0oFgL3AjtuEaHrSrgz6JcQZ326YxnxlEyUzqWG1f5o4P2u1fWl+fZ4a0exYpVmrRputx6cF0Oi0L3egkoei3oZuH6ebWPVlndf1tj4jdUbQykCNJ3U8zuvjmg3T7abeZZ7p3zpaFuRpB/I6hN2UHBHU+8pre4ij/L5HA/5tyLSiyGn5dlBYej9RCQSpPCr3rPfLZhVRSciqB9DyuVvRXAFs/bBum/bOvni+HxqRqiPm0+eB/H+Y7N2whgvc2igN8M0hd/ETE9zailLTCgKVUGNZiBdUKLe00J9CFYc/08UM5Tbkjo+8ixkNowOpRo6BxgiRoU+VuwpEMQlPg97+/exmoD2bC9jZCxcx9FyhwbXi9qNqGe8eEdkDa6aG+/GiwHvadzXdwGj7p3e8kPdEscg2UQzmnxzbHDoHJAwG4L3QFuaD"
	)
	plain, err := aesGCMDecrypt(apiV3Key, ciphertext, nonce, aad)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	var tx struct {
		OutTradeNo string `json:"out_trade_no"`
		Amount     struct {
			Total int64 `json:"total"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(plain, &tx); err != nil {
		t.Fatalf("解密内容解析失败: %v", err)
	}
	if tx.OutTradeNo != "EQS000001" || tx.Amount.Total != 100 {
		t.Fatalf("解密内容不符: %s", string(plain))
	}
}

// TestWechatNotify_VerifyAndDecrypt 验签+解密全链路（自洽：签名→验签）
func TestWechatNotify_VerifyAndDecrypt(t *testing.T) {
	priv, err := loadRSAPrivateKey(filepath.Join("testdata", "merchant_private.pem"))
	if err != nil {
		t.Fatal(err)
	}
	// 平台公钥 = 同一密钥对公钥（测试用）
	w := &WechatPayV3{apiV3Key: "0123456789abcdef0123456789abcdef", privateKey: priv, platformKey: &priv.PublicKey}

	// 构造合法回调
	plaintext := `{"out_trade_no":"EQS000001","transaction_id":"4200001234","trade_state":"SUCCESS","amount":{"total":100}}`
	ct, err := aesGCMEncryptForTest(w.apiV3Key, "abcdefghijkl", "transaction", []byte(plaintext))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]interface{}{
		"event_type": "TRANSACTION.SUCCESS",
		"resource": map[string]string{
			"algorithm": "AEAD_AES_256_GCM", "ciphertext": ct,
			"nonce": "abcdefghijkl", "associated_data": "transaction",
		},
	})
	// 签名（用同一私钥模拟平台签名）
	ts, nonce := "1710000100", "aaabbbcccddd"
	msg := ts + "\n" + nonce + "\n" + string(body) + "\n"
	sig := rsaSignForTest(priv, msg)

	hdr := http.Header{}
	hdr.Set("Wechatpay-Signature", sig)
	hdr.Set("Wechatpay-Timestamp", ts)
	hdr.Set("Wechatpay-Nonce", nonce)

	outTradeNo, total, txID, err := w.VerifyAndDecryptNotify(hdr, body)
	if err != nil {
		t.Fatalf("验签解密失败: %v", err)
	}
	if outTradeNo != "EQS000001" || total != 100 || txID != "4200001234" {
		t.Fatalf("回调解析不符: %s %d %s", outTradeNo, total, txID)
	}

	// 篡改 body → 验签失败
	hdrBad := hdr.Clone()
	if _, _, _, err := w.VerifyAndDecryptNotify(hdrBad, []byte(`{"x":1}`)); err == nil {
		t.Fatal("篡改回调应验签失败")
	}
}

// aesGCMEncryptForTest 测试辅助：AES-256-GCM 加密（与微信 API v3 构造一致）
func aesGCMEncryptForTest(apiV3Key, nonce, aad string, plain []byte) (string, error) {
	block, err := aes.NewCipher([]byte(apiV3Key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ct := gcm.Seal(nil, []byte(nonce), plain, []byte(aad))
	return base64.StdEncoding.EncodeToString(ct), nil
}

// rsaSignForTest 测试辅助：RSA-SHA256 签名（base64）
func rsaSignForTest(priv *rsa.PrivateKey, message string) string {
	digest := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, digest[:])
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(sig)
}
