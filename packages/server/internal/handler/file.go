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

	// P0-05：仅项目参与方可查看文件
	var proj model.Project
	if err := model.DB.First(&proj, projectID).Error; err != nil || !canAccessProject(c, &proj) {
		forbidden(c, "无权查看该项目文件")
		return
	}

	var files []model.ProjectFile
	model.DB.Where("project_id = ?", projectID).Order("created_at DESC").Find(&files)
	ok(c, gin.H{"files": files})
}

// DownloadFile 下载文件
// GET /api/v1/file/:id/download
// 权限：登录用户且为项目成员/订单参与方/管理员。MVP 阶段 StorageKey 为占位，
// 若为 URL 直接 302 重定向；否则返回文件元信息（客户端据 storage_key 拼接下载地址）。
func DownloadFile(c *gin.Context) {
	userID := c.GetUint("user_id")
	userType := c.GetInt("user_type")

	fileID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "文件ID无效")
		return
	}

	var file model.ProjectFile
	if err := model.DB.First(&file, fileID).Error; err != nil {
		notFound(c, "文件不存在")
		return
	}

	// 权限校验：管理员放行；否则必须是项目创建者、上传者或订单参与方
	if userType != 3 && file.UploaderID != userID {
		// 校验是否项目创建者
		var project model.Project
		if err := model.DB.First(&project, file.ProjectID).Error; err == nil && project.UserID == userID {
			// 项目创建者可下载
		} else {
			// 校验订单参与方（服务方=orders.supplier_id；甲方=项目创建者）
			allowed := false
			if file.OrderID > 0 {
				var order model.Order
				if model.DB.First(&order, file.OrderID).Error == nil {
					if order.SupplierID == userID {
						allowed = true
					} else {
						var op model.Project
						if model.DB.First(&op, order.ProjectID).Error == nil && op.UserID == userID {
							allowed = true
						}
					}
				}
			}
			if !allowed {
				forbidden(c, "无权下载该文件")
				return
			}
		}
	}

	if file.StorageKey != "" && (len(file.StorageKey) > 7 && (file.StorageKey[:7] == "http://" || file.StorageKey[:8] == "https://")) {
		c.Redirect(302, file.StorageKey)
		return
	}

	ok(c, gin.H{
		"file": gin.H{
			"id":            file.ID,
			"original_name": file.OriginalName,
			"file_type":     file.FileType,
			"storage_key":   file.StorageKey,
			"version":       file.Version,
			"sha256":        file.SHA256,
		},
	})
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