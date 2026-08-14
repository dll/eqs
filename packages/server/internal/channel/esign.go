package channel

import (
	"fmt"
)

// EsignGateway 电子签网关抽象：统一签署流程接口。
// 目前提供 Mock 实现（生成模拟签署链接）；真实服务商（法大大/e签宝/上上签等）
// 需各自实现本接口并在 handler 侧按 ESIGN_PROVIDER 分发——API 差异较大且需司法
// 存证资质签约，故具体适配器留待服务商选定后按本接口开发（接口已就绪）。
type EsignGateway interface {
	// CreateSignOrder 创建签署任务，返回签署链接（发起方/接收方）
	CreateSignOrder(orderNo string, signerName, signerPhone string) (signURL string, err error)
	// NotifySignature 处理签署完成回调（验签逻辑由实现负责）
	NotifySignature(payload []byte, headers map[string]string) (orderNo string, signed bool, err error)
}

// NewEsignGateway 构造电子签网关（当前仅 mock；配置 ESIGN_PROVIDER=real 且接入
// 具体服务商后返回真实实现）
func NewEsignGateway(provider string) EsignGateway {
	return &MockEsignGateway{}
}

// MockEsignGateway 模拟电子签：返回演示签署链接，仅用于联调与演示
type MockEsignGateway struct{}

func (m *MockEsignGateway) CreateSignOrder(orderNo string, signerName, signerPhone string) (string, error) {
	return fmt.Sprintf("https://eqs-chzu.tech/esign/mock/%s?signer=%s", orderNo, signerPhone), nil
}

func (m *MockEsignGateway) NotifySignature(payload []byte, headers map[string]string) (string, bool, error) {
	return "", false, fmt.Errorf("mock 网关不支持回调")
}
