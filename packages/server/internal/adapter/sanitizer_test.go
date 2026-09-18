package adapter

import "testing"

func TestSanitizers(t *testing.T) {
	if got := sanitizeText("手机号 13812345678，身份证 110101199001011234 token=abc"); got != "手机号 138****5678，身份证 110***********1234 token=***" {
		t.Fatalf("脱敏结果不符合基线: %q", got)
	}
	if got := maskAddress("北京市朝阳区幸福路 10 号"); got != "北京市朝阳区***" {
		t.Fatalf("地址脱敏结果不符合基线: %q", got)
	}
	if got := maskProvider("某地勘院有限公司"); got != "某地勘院***" {
		t.Fatalf("服务方脱敏结果不符合基线: %q", got)
	}
}

func TestBudgetAndMoneyRangeDoNotExposeExactAmount(t *testing.T) {
	for _, value := range []string{budgetRange(123456, 234567), moneyRange(123456)} {
		if value == "123456" || value == "234567" || value == "123456.00" {
			t.Fatalf("不应返回精确金额: %q", value)
		}
	}
}
