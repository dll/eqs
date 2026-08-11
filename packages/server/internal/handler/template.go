package handler

import (
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ListDeliveryTemplates 标准交付模板列表（按服务类型，仅有效版本）
// GET /api/v1/delivery-templates?service_type=cost
func ListDeliveryTemplates(c *gin.Context) {
	q := model.DB.Model(&model.DeliveryTemplate{}).Where("status = ?", "active")
	if st := c.Query("service_type"); st != "" {
		q = q.Where("service_type = ?", st)
	}
	var templates []model.DeliveryTemplate
	q.Order("service_type ASC, version DESC").Find(&templates)
	ok(c, gin.H{"templates": templates, "count": len(templates)})
}

// GetDeliveryTemplate 模板详情（含 checklist）
// GET /api/v1/delivery-templates/:id
func GetDeliveryTemplate(c *gin.Context) {
	tplID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "模板ID无效")
		return
	}
	var tpl model.DeliveryTemplate
	if err := model.DB.First(&tpl, tplID).Error; err != nil {
		notFound(c, "模板不存在")
		return
	}
	ok(c, gin.H{"template": tpl})
}

// ValidateDeliveryChecklist 按模板清单校验必交材料
// POST /api/v1/delivery-templates/:id/validate  body: {"checklist_result": {"材料A": true, ...}}
func ValidateDeliveryChecklist(c *gin.Context) {
	tplID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "模板ID无效")
		return
	}
	var req struct {
		ChecklistResult map[string]bool `json:"checklist_result"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var tpl model.DeliveryTemplate
	if err := model.DB.First(&tpl, tplID).Error; err != nil {
		notFound(c, "模板不存在")
		return
	}
	// 解析模板 checklist
	checklist := parseChecklist(tpl.Checklist)
	missing := make([]string, 0)
	for _, item := range checklist {
		if !req.ChecklistResult[item] {
			missing = append(missing, item)
		}
	}
	ok(c, gin.H{"complete": len(missing) == 0, "missing": missing, "total": len(checklist)})
}

// parseChecklist 解析 JSON 数组 checklist（兼容字符串数组）
func parseChecklist(s string) []string {
	var items []string
	if err := parseJSONString(s, &items); err != nil || len(items) == 0 {
		return []string{}
	}
	return items
}
