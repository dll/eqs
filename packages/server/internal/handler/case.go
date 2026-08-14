package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ListProviderCases 服务商公开案例（服务超市/主页展示，仅 published）
// GET /api/v1/provider/:id/cases（公开接口）
// 返回含 image_urls：为成果图生成 24h 有效签名公开预览链接（未登录可浏览）
func ListProviderCases(c *gin.Context) {
	supplierID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "服务方ID无效")
		return
	}
	var cases []model.CaseShowcase
	model.DB.Where("supplier_id = ? AND status = ?", supplierID, "published").
		Order("created_at DESC").Limit(20).Find(&cases)

	type caseItem struct {
		model.CaseShowcase
		ImageURLs []string `json:"image_urls"`
	}
	items := make([]caseItem, 0, len(cases))
	for _, cs := range cases {
		items = append(items, caseItem{CaseShowcase: cs, ImageURLs: signedCaseImageURLs(cs.ImageFileIDs)})
	}
	ok(c, gin.H{"cases": items, "count": len(items)})
}

// parseFileIDList 解析 JSON 数组形式的 file_id 列表
func parseFileIDList(s string) []uint {
	var ids []uint
	if err := json.Unmarshal([]byte(s), &ids); err != nil {
		return nil
	}
	return ids
}

// signedCaseImageURLs 为案例成果图生成签名公开预览链接
func signedCaseImageURLs(idsJSON string) []string {
	ids := parseFileIDList(idsJSON)
	if len(ids) == 0 {
		return nil
	}
	urls := make([]string, 0, len(ids))
	for _, fid := range ids {
		urls = append(urls, fmt.Sprintf("/api/v1/file/%d/preview/public?token=%s", fid, signPreviewToken(fid)))
	}
	return urls
}

// ListMyCases 我的案例（服务方本人或管理员）
// GET /api/v1/case/mine
func ListMyCases(c *gin.Context) {
	userID := c.GetUint("user_id")
	var cases []model.CaseShowcase
	model.DB.Where("supplier_id = ?", userID).Order("created_at DESC").Find(&cases)
	ok(c, gin.H{"cases": cases, "count": len(cases)})
}

type CreateCaseRequest struct {
	ProjectID    uint     `json:"project_id"`
	OrderID      uint     `json:"order_id"`
	Title        string   `json:"title" binding:"required"`
	Description  string   `json:"description"`
	ServiceType  string   `json:"service_type"`
	ImageFileIDs []uint   `json:"image_file_ids"`
}

// CreateCase 服务方创建案例（可关联自己的已完成订单，一键沉淀成果）
// POST /api/v1/case/create
func CreateCase(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	// 关联订单校验：只能沉淀本人作为服务方的订单，且订单已完成（status=3）
	if req.OrderID > 0 {
		var order model.Order
		if err := model.DB.First(&order, req.OrderID).Error; err != nil {
			badRequest(c, "关联订单不存在")
			return
		}
		if order.SupplierID != userID && !isAdmin(c) {
			forbidden(c, "只能沉淀本人承接的订单")
			return
		}
		if order.Status != 3 {
			badRequest(c, "仅已完成订单可沉淀为案例")
			return
		}
		if req.ProjectID == 0 {
			req.ProjectID = order.ProjectID
		}
		if req.ServiceType == "" {
			var project model.Project
			if err := model.DB.First(&project, order.ProjectID).Error; err == nil {
				req.ServiceType = project.ServiceType
			}
		}
	}

	ids, _ := json.Marshal(req.ImageFileIDs)
	if len(req.ImageFileIDs) == 0 {
		ids = []byte("[]")
	}

	cs := model.CaseShowcase{
		SupplierID:   userID,
		ProjectID:    req.ProjectID,
		OrderID:      req.OrderID,
		Title:        req.Title,
		Description:  req.Description,
		ServiceType:  req.ServiceType,
		ImageFileIDs: string(ids),
		Status:       "published",
	}
	if err := model.DB.Create(&cs).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "case.create", "case", cs.ID, gin.H{"order_id": req.OrderID, "service_type": req.ServiceType})
	ok(c, gin.H{"case": cs, "message": "案例已发布"})
}

// UpdateCase 编辑案例（仅本人或管理员）
// PUT /api/v1/case/:id
func UpdateCase(c *gin.Context) {
	caseID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "案例ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var cs model.CaseShowcase
	if err := model.DB.First(&cs, caseID).Error; err != nil {
		notFound(c, "案例不存在")
		return
	}
	if cs.SupplierID != userID && !isAdmin(c) {
		forbidden(c, "仅本人可编辑案例")
		return
	}

	var req CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	ids, _ := json.Marshal(req.ImageFileIDs)
	if len(req.ImageFileIDs) == 0 {
		ids = []byte("[]")
	}
	updates := map[string]interface{}{
		"title":          req.Title,
		"description":    req.Description,
		"service_type":   req.ServiceType,
		"image_file_ids": string(ids),
		"updated_at":     time.Now(),
	}
	if err := model.DB.Model(&cs).Updates(updates).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "case.update", "case", caseID, gin.H{})
	ok(c, gin.H{"message": "案例已更新"})
}

// DeleteCase 删除案例（仅本人或管理员）
// DELETE /api/v1/case/:id
func DeleteCase(c *gin.Context) {
	caseID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "案例ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var cs model.CaseShowcase
	if err := model.DB.First(&cs, caseID).Error; err != nil {
		notFound(c, "案例不存在")
		return
	}
	if cs.SupplierID != userID && !isAdmin(c) {
		forbidden(c, "仅本人可删除案例")
		return
	}
	if err := model.DB.Delete(&cs).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "case.delete", "case", caseID, gin.H{})
	ok(c, gin.H{"message": "案例已删除"})
}
