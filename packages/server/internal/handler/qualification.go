package handler

import (
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ListQualifications 查询服务方资质列表
func ListQualifications(c *gin.Context) {
	supplierID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "服务方ID无效")
		return
	}

	var quals []model.SupplierQualification
	model.DB.Where("supplier_id = ?", supplierID).Order("created_at DESC").Find(&quals)
	ok(c, gin.H{"qualifications": quals})
}

type SubmitQualificationRequest struct {
	QualificationType string `json:"qualification_type" binding:"required"`
	CertificateNo     string `json:"certificate_no" binding:"required"`
	Level             string `json:"level"`
	Scope             string `json:"scope"`
	EvidenceFileID    uint   `json:"evidence_file_id"`
}

// SubmitQualification 服务方提交资质用于核验
func SubmitQualification(c *gin.Context) {
	supplierID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "服务方ID无效")
		return
	}

	var req SubmitQualificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	qual := model.SupplierQualification{
		SupplierID:        supplierID,
		QualificationType: req.QualificationType,
		CertificateNo:     req.CertificateNo,
		Level:             req.Level,
		Scope:             req.Scope,
		EvidenceFileID:    req.EvidenceFileID,
		VerificationStatus: "pending",
	}
	if err := model.DB.Create(&qual).Error; err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"qualification": qual})
}

// ReviewQualification 平台核验资质（OCR/人工），结果更新核验状态
func ReviewQualification(c *gin.Context) {
	qualID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "资质ID无效")
		return
	}
	reviewerID := c.GetUint("user_id")

	var req struct {
		Verified *bool  `json:"verified" binding:"required"`
		Comment  string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var qual model.SupplierQualification
	if err := model.DB.First(&qual, qualID).Error; err != nil {
		notFound(c, "资质不存在")
		return
	}

	status := "rejected"
	if *req.Verified {
		status = "approved"
	}
	now := time.Now()
	model.DB.Model(&qual).Updates(map[string]interface{}{
		"verification_status": status,
		"reviewed_by":         reviewerID,
		"reviewed_at":         now,
	})
	WriteAudit(c, "qualification.review", "qualification", qualID, gin.H{"status": status, "reviewer_id": reviewerID})
	ok(c, gin.H{"qualification": qual, "status": status})
}