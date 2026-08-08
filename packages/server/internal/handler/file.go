package handler

import (
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

type UploadFileRequest struct {
	ProjectID    uint   `json:"project_id"`
	OrderID      uint   `json:"order_id"`
	OriginalName string `json:"original_name" binding:"required"`
	FileType     string `json:"file_type" binding:"required"`
	StorageKey   string `json:"storage_key" binding:"required"`
	SHA256       string `json:"sha256"`
}

// UploadFile 登记上传文件（MVP 直接存 COS storage_key，篡改校验用 SHA256）
func UploadFile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req UploadFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	file := model.ProjectFile{
		ProjectID:    req.ProjectID,
		OrderID:      req.OrderID,
		UploaderID:   userID,
		OriginalName: req.OriginalName,
		FileType:     req.FileType,
		StorageKey:   req.StorageKey,
		SHA256:       req.SHA256,
		Version:      1,
	}
	if err := model.DB.Create(&file).Error; err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"file": file})
}

// ListFiles 列出项目文件
func ListFiles(c *gin.Context) {
	projectID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}

	var files []model.ProjectFile
	model.DB.Where("project_id = ?", projectID).Order("created_at DESC").Find(&files)
	ok(c, gin.H{"files": files})
}

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
	model.DB.Model(&annotation).Update("status", "resolved")
	ok(c, gin.H{"message": "批注已标记为已解决"})
}