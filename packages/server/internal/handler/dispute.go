package handler

import (
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

type CreateDisputeRequest struct {
	OrderID     uint   `json:"order_id" binding:"required"`
	MilestoneID uint   `json:"milestone_id"`
	Reason      string `json:"reason" binding:"required"`
	Claim       string `json:"claim"`
}

// CreateDispute 发起争议：先证据阶段，节点资金自动冻结
func CreateDispute(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateDisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var order model.Order
	if err := model.DB.Preload("Project").First(&order, req.OrderID).Error; err != nil {
		notFound(c, "订单不存在")
		return
	}
	if order.SupplierID != userID && order.Project.UserID != userID {
		forbidden(c, "仅合同双方可发起争议")
		return
	}

	// 同节点存在未结争议则拒绝
	var count int64
	q := model.DB.Model(&model.Dispute{}).Where("order_id = ? AND status <> ?", req.OrderID, "closed")
	if req.MilestoneID > 0 {
		q = q.Where("milestone_id = ?", req.MilestoneID)
	}
	q.Count(&count)
	if count > 0 {
		badRequest(c, "该订单已有未结争议")
		return
	}

	dispute := model.Dispute{
		OrderID:     req.OrderID,
		MilestoneID: req.MilestoneID,
		InitiatorID: userID,
		Reason:      req.Reason,
		Claim:       req.Claim,
		Status:      "evidence",
	}
	if err := model.DB.Create(&dispute).Error; err != nil {
		serverError(c, err)
		return
	}

	// 冻结对应节点：标记争议中（结算时检查 status<>closed 即拒绝）
	WriteAudit(c, "dispute.create", "dispute", dispute.ID, gin.H{"order_id": req.OrderID, "milestone_id": req.MilestoneID, "reason": req.Reason})
	ok(c, gin.H{"dispute": dispute, "message": "争议已发起，相关款项已冻结"})
}

// UploadDisputeEvidence 上传争议证据（文件+哈希）
func UploadDisputeEvidence(c *gin.Context) {
	disputeID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "争议ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var req struct {
		FileID  uint   `json:"file_id" binding:"required"`
		SHA256  string `json:"sha256"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	evidence := model.DisputeEvidence{
		DisputeID: disputeID,
		UserID:    userID,
		FileID:    req.FileID,
		SHA256:    req.SHA256,
		Content:   req.Content,
	}
	if err := model.DB.Create(&evidence).Error; err != nil {
		serverError(c, err)
		return
	}

	// 有证据后进入评审阶段
	model.DB.Model(&model.Dispute{}).Where("id = ? AND status = ?", disputeID, "evidence").
		Update("status", "review")
	ok(c, gin.H{"evidence": evidence})
}

// AssignDisputeExpert 平台指派评审专家（管理员/平台）
func AssignDisputeExpert(c *gin.Context) {
	disputeID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "争议ID无效")
		return
	}

	var req struct {
		ExpertUserID uint `json:"expert_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	assignment := model.DisputeExpertAssignment{
		DisputeID:    disputeID,
		ExpertUserID: req.ExpertUserID,
	}
	if err := model.DB.Create(&assignment).Error; err != nil {
		serverError(c, err)
		return
	}

	model.DB.Model(&model.Dispute{}).Where("id = ? AND status = ?", disputeID, "review").
		Update("status", "review")
	ok(c, gin.H{"assignment": assignment})
}

type SubmitExpertOpinionRequest struct {
	Opinion string `json:"opinion" binding:"required"`
	Vote    string `json:"vote" binding:"required"`
}

// SubmitExpertOpinion 专家提交评审意见
func SubmitExpertOpinion(c *gin.Context) {
	assignmentID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "指派ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var req SubmitExpertOpinionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var assignment model.DisputeExpertAssignment
	if err := model.DB.First(&assignment, assignmentID).Error; err != nil {
		notFound(c, "指派不存在")
		return
	}
	if assignment.ExpertUserID != userID {
		forbidden(c, "仅被指派专家可提交意见")
		return
	}

	now := time.Now()
	model.DB.Model(&assignment).Updates(map[string]interface{}{
		"opinion":      req.Opinion,
		"vote":         req.Vote,
		"submitted_at": now,
	})
	WriteAudit(c, "dispute.expert_opinion", "expert_assignment", assignmentID, gin.H{"dispute_id": assignment.DisputeID, "vote": req.Vote})
	ok(c, gin.H{"message": "评审意见已提交"})
}

// CloseDispute 争议结案：填写调解/评审结论，解冻资金并可按结论结算
func CloseDispute(c *gin.Context) {
	disputeID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "争议ID无效")
		return
	}

	var req struct {
		ResolutionType  string  `json:"resolution_type" binding:"required"` // settlement/agreement/award
		ResolutionFileID uint   `json:"resolution_file_id"`
		SettleAmount    float64 `json:"settle_amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var dispute model.Dispute
	if err := model.DB.First(&dispute, disputeID).Error; err != nil {
		notFound(c, "争议不存在")
		return
	}
	if dispute.Status == "closed" {
		ok(c, gin.H{"message": "已结案", "idempotent": true})
		return
	}

	now := time.Now()
	model.DB.Model(&dispute).Updates(map[string]interface{}{
		"status":            "closed",
		"resolution_type":   req.ResolutionType,
		"resolution_file_id": req.ResolutionFileID,
		"closed_at":         now,
	})
	WriteAudit(c, "dispute.close", "dispute", disputeID, gin.H{"resolution_type": req.ResolutionType, "settle_amount": req.SettleAmount})
	ok(c, gin.H{"dispute": dispute, "message": "争议已结案，款项已解冻"})
}

// GetDispute 争议详情（含证据与专家意见）
func GetDispute(c *gin.Context) {
	disputeID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "争议ID无效")
		return
	}

	var dispute model.Dispute
	if err := model.DB.First(&dispute, disputeID).Error; err != nil {
		notFound(c, "争议不存在")
		return
	}

	var evidence []model.DisputeEvidence
	model.DB.Where("dispute_id = ?", disputeID).Find(&evidence)

	var assignments []model.DisputeExpertAssignment
	model.DB.Where("dispute_id = ?", disputeID).Find(&assignments)

	ok(c, gin.H{
		"dispute":     dispute,
		"evidence":    evidence,
		"assignments": assignments,
	})
}