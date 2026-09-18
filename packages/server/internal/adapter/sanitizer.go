package adapter

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

var (
	phonePattern  = regexp.MustCompile(`1[3-9][0-9]{9}`)
	idPattern     = regexp.MustCompile(`[0-9]{17}[0-9Xx]`)
	secretPattern = regexp.MustCompile(`(?i)(token|secret|api[_-]?key)[[:space:]]*[:=][[:space:]]*[^[:space:],;]+`)
)

func sanitizeText(value string) string {
	// 证件号中可能包含符合手机号模式的连续数字，必须先处理更长的证件号。
	value = idPattern.ReplaceAllStringFunc(value, func(id string) string {
		return id[:3] + "***********" + id[len(id)-4:]
	})
	value = phonePattern.ReplaceAllStringFunc(value, func(phone string) string {
		return phone[:3] + "****" + phone[7:]
	})
	return secretPattern.ReplaceAllString(value, "$1=***")
}

func maskAddress(address string) string {
	address = sanitizeText(strings.TrimSpace(address))
	if address == "" {
		return ""
	}
	for _, marker := range []string{"区", "县", "市"} {
		if index := strings.Index(address, marker); index >= 0 {
			return address[:index+len(marker)] + "***"
		}
	}
	runes := []rune(address)
	if len(runes) > 6 {
		return string(runes[:6]) + "***"
	}
	return "***"
}

func maskProvider(name string) string {
	name = sanitizeText(strings.TrimSpace(name))
	for _, suffix := range []string{"设计院", "地勘院", "监理公司", "事务所", "有限公司", "公司"} {
		if strings.Contains(name, suffix) {
			return "某" + suffix + "***"
		}
	}
	if name == "" {
		return "未披露***"
	}
	return "服务机构***"
}

func moneyRange(value float64) string {
	if value <= 0 {
		return "未披露"
	}
	unit := 10000.0
	lower := math.Floor(value/unit) * unit
	upper := math.Ceil(value/unit) * unit
	if lower == upper {
		upper += unit
	}
	return fmt.Sprintf("%.0f万-%.0f万", lower/unit, upper/unit)
}

func budgetRange(minimum, maximum float64) string {
	if minimum <= 0 && maximum <= 0 {
		return "未披露"
	}
	if maximum < minimum {
		minimum, maximum = maximum, minimum
	}
	return fmt.Sprintf("%.0f万-%.0f万", math.Floor(minimum/10000), math.Ceil(maximum/10000))
}
