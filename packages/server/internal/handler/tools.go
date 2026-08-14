package handler

import (
	"math"

	"github.com/gin-gonic/gin"
)

// ==================== 轻量计价工具（工程量清单计价估算） ====================
// 面向甲方/服务方的快捷估价：分部分项 → 措施费 → 规费 → 税金 → 含税总价。
// 费率可配置（默认：措施 5%、规费 2%、税金 9%，对应建筑业增值税一般计税口径）。

// EstimateItem 清单项
type EstimateItem struct {
	Name      string  `json:"name"`
	Unit      string  `json:"unit"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

// EstimateRates 取费费率（百分比）
type EstimateRates struct {
	MeasureRate  float64 `json:"measure_rate"`  // 措施项目费率 %
	OverheadRate float64 `json:"overhead_rate"` // 规费率 %
	TaxRate      float64 `json:"tax_rate"`      // 增值税率 %
}

// CostEstimateResult 计价估算结果
type CostEstimateResult struct {
	ItemsCount      int     `json:"items_count"`
	Subtotal        float64 `json:"subtotal"`        // 分部分项工程费
	MeasureFee      float64 `json:"measure_fee"`     // 措施项目费
	OverheadFee     float64 `json:"overhead_fee"`    // 规费
	Tax             float64 `json:"tax"`             // 税金
	Total           float64 `json:"total"`           // 含税总价
	MeasureRate     float64 `json:"measure_rate"`
	OverheadRate    float64 `json:"overhead_rate"`
	TaxRate         float64 `json:"tax_rate"`
	PerItemBreakdown []EstimateItemBreakdown `json:"per_item_breakdown"`
}

// EstimateItemBreakdown 单项小计
type EstimateItemBreakdown struct {
	Name     string  `json:"name"`
	Unit     string  `json:"unit"`
	Quantity float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Amount   float64 `json:"amount"`
}

// calculateCostEstimate 纯计算函数（可单测）
func calculateCostEstimate(items []EstimateItem, rates EstimateRates) CostEstimateResult {
	if rates.MeasureRate <= 0 {
		rates.MeasureRate = 5
	}
	if rates.OverheadRate <= 0 {
		rates.OverheadRate = 2
	}
	if rates.TaxRate <= 0 {
		rates.TaxRate = 9
	}

	var subtotal float64
	breakdown := make([]EstimateItemBreakdown, 0, len(items))
	for _, it := range items {
		amount := it.Quantity * it.UnitPrice
		subtotal += amount
		breakdown = append(breakdown, EstimateItemBreakdown{
			Name: it.Name, Unit: it.Unit, Quantity: it.Quantity, UnitPrice: it.UnitPrice, Amount: amount,
		})
	}
	measure := subtotal * rates.MeasureRate / 100
	// 其他项目费按 0（预留）
	other := 0.0
	overhead := (subtotal + measure + other) * rates.OverheadRate / 100
	tax := (subtotal + measure + other + overhead) * rates.TaxRate / 100
	total := subtotal + measure + other + overhead + tax

	round2 := func(v float64) float64 { return math.Round(v*100) / 100 }
	return CostEstimateResult{
		ItemsCount:       len(items),
		Subtotal:         round2(subtotal),
		MeasureFee:       round2(measure),
		OverheadFee:      round2(overhead),
		Tax:              round2(tax),
		Total:            round2(total),
		MeasureRate:      rates.MeasureRate,
		OverheadRate:     rates.OverheadRate,
		TaxRate:          rates.TaxRate,
		PerItemBreakdown: breakdown,
	}
}

// CostEstimate 清单计价估算接口（登录用户）
// POST /api/v1/tools/cost-estimate
func CostEstimate(c *gin.Context) {
	var req struct {
		Items []EstimateItem `json:"items"`
		Rates EstimateRates  `json:"rates"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	if len(req.Items) == 0 {
		badRequest(c, "请至少填写一项清单")
		return
	}
	for _, it := range req.Items {
		if it.Quantity < 0 || it.UnitPrice < 0 {
			badRequest(c, "工程量与单价不能为负数")
			return
		}
	}
	result := calculateCostEstimate(req.Items, req.Rates)
	WriteAudit(c, "tools.cost_estimate", "system", 0, gin.H{"items": len(req.Items), "total": result.Total})
	ok(c, gin.H{"result": result})
}
