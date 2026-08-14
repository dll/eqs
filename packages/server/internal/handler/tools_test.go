package handler

import "testing"

func TestCalculateCostEstimate(t *testing.T) {
	items := []EstimateItem{
		{Name: "挖基础土方", Unit: "m³", Quantity: 100, UnitPrice: 30},
		{Name: "C30混凝土", Unit: "m³", Quantity: 50, UnitPrice: 480},
	}
	res := calculateCostEstimate(items, EstimateRates{})

	// 分部分项 = 100*30 + 50*480 = 3000 + 24000 = 27000
	if res.Subtotal != 27000 {
		t.Errorf("分部分项费 = %v, 期望 27000", res.Subtotal)
	}
	// 措施 5% → 1350；规费 2% → (27000+1350)*2% = 567；税金 9% → (27000+1350+567)*9% = 2602.53
	// 总价 = 27000+1350+567+2602.53 = 31519.53
	if res.MeasureFee != 1350 {
		t.Errorf("措施费 = %v, 期望 1350", res.MeasureFee)
	}
	if res.OverheadFee != 567 {
		t.Errorf("规费 = %v, 期望 567", res.OverheadFee)
	}
	if res.Tax != 2602.53 {
		t.Errorf("税金 = %v, 期望 2602.53", res.Tax)
	}
	if res.Total != 31519.53 {
		t.Errorf("含税总价 = %v, 期望 31519.53", res.Total)
	}
	if len(res.PerItemBreakdown) != 2 {
		t.Errorf("清单明细数 = %d, 期望 2", len(res.PerItemBreakdown))
	}
	if res.MeasureRate != 5 || res.OverheadRate != 2 || res.TaxRate != 9 {
		t.Errorf("默认费率错误: %+v", res)
	}
}
