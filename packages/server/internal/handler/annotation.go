package handler

import (
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ==================== 文件批注 ====================

type AddAnnotationRequest struct {
	FileID  uint    `json:"file_id" binding:"required"`
	PageNo  int     `json:"page_no"`
	XRatio  float64 `json:"x_ratio"`
	YRatio  float64 `json:"y_ratio"`
	Content string  `json:"content" binding:"required"`
}

// AddAnnotation 为文件添加批注
func AddAnnotation(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req AddAnnotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var file model.ProjectFile
	if err := model.DB.First(&file, req.FileID).Error; err != nil {
		notFound(c, "文件不存在")
		return
	}
	// P0-05：仅项目参与方可批注
	var proj model.Project
	if err := model.DB.First(&proj, file.ProjectID).Error; err != nil || !canAccessProject(c, &proj) {
		forbidden(c, "无权批注该项目文件")
		return
	}

	annotation := model.FileAnnotation{
		FileID:   req.FileID,
		AuthorID: userID,
		PageNo:   req.PageNo,
		XRatio:   req.XRatio,
		YRatio:   req.YRatio,
		Content:  req.Content,
	}
	if err := model.DB.Create(&annotation).Error; err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"annotation": annotation})
}

// ListAnnotations 查看文件批注
func ListAnnotations(c *gin.Context) {
	fileID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "文件ID无效")
		return
	}

	// P0-05：仅项目参与方可查看批注
	var file model.ProjectFile
	if err := model.DB.First(&file, fileID).Error; err != nil {
		notFound(c, "文件不存在")
		return
	}
	var proj model.Project
	if err := model.DB.First(&proj, file.ProjectID).Error; err != nil || !canAccessProject(c, &proj) {
		forbidden(c, "无权查看该项目批注")
		return
	}

	var annotations []model.FileAnnotation
	model.DB.Where("file_id = ? AND status = ?", fileID, "active").Order("created_at ASC").Find(&annotations)
	ok(c, gin.H{"annotations": annotations})
}

// ResolveAnnotation 解决/关闭批注
func ResolveAnnotation(c *gin.Context) {
	annotationID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "批注ID无效")
		return
	}

	var annotation model.FileAnnotation
	if err := model.DB.First(&annotation, annotationID).Error; err != nil {
		notFound(c, "批注不存在")
		return
	}
	// P0-05：仅项目参与方可解决批注
	var file model.ProjectFile
	if err := model.DB.First(&file, annotation.FileID).Error; err != nil {
		notFound(c, "文件不存在")
		return
	}
	var proj model.Project
	if err := model.DB.First(&proj, file.ProjectID).Error; err != nil || !canAccessProject(c, &proj) {
		forbidden(c, "无权操作该项目批注")
		return
	}
	model.DB.Model(&annotation).Update("status", "resolved")
	ok(c, gin.H{"message": "批注已标记为已解决"})
}
