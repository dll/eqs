package handler

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/dxf"
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

// PreviewFile 文件在线预览（图片/PDF 直接返回内容流，浏览器内联展示）
// GET /api/v1/file/:id/preview
// 权限与 DownloadFile 一致；仅允许预览图片与 PDF（避免任意文件下载）
func PreviewFile(c *gin.Context) {
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

	// 图片/PDF/DXF/DWG 允许内联预览（DXF 服务端自研转换；DWG 走第三方引擎适配器）
	previewable := map[string]bool{"jpg": true, "jpeg": true, "png": true, "pdf": true, "dxf": true, "dwg": true}
	if !previewable[file.FileType] {
		badRequest(c, "该文件类型不支持在线预览，请下载查看")
		return
	}

	// 权限校验（与 DownloadFile 相同）
	if userType != 3 && file.UploaderID != userID {
		var project model.Project
		if err := model.DB.First(&project, file.ProjectID).Error; err != nil || project.UserID != userID {
			allowed := false
			if file.OrderID > 0 {
				var order model.Order
				if model.DB.First(&order, file.OrderID).Error == nil {
					if order.SupplierID == userID {
						allowed = true
					} else if model.DB.First(&project, order.ProjectID).Error == nil && project.UserID == userID {
						allowed = true
					}
				}
			}
			if !allowed {
				forbidden(c, "无权预览该文件")
				return
			}
		}
	}

	serveFilePreview(c, file)
}

// PreviewFilePublic 公开预览（带签名 token，供服务商主页案例图等未登录公开展示场景）
// GET /api/v1/file/:id/preview/public?token=<exp.sign>
// 签名由 signPreviewToken 生成，绑定 file_id 且 24h 过期；仅允许预览类型，不开放下载。
func PreviewFilePublic(c *gin.Context) {
	fileID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "文件ID无效")
		return
	}
	token := c.Query("token")
	if token == "" || !verifyPreviewToken(token, fileID) {
		unauthorized(c, "预览链接无效或已过期")
		return
	}
	var file model.ProjectFile
	if err := model.DB.First(&file, fileID).Error; err != nil {
		notFound(c, "文件不存在")
		return
	}
	previewable := map[string]bool{"jpg": true, "jpeg": true, "png": true, "pdf": true, "dxf": true, "dwg": true}
	if !previewable[file.FileType] {
		badRequest(c, "该文件类型不支持在线预览，请下载查看")
		return
	}
	serveFilePreview(c, file)
}

// serveFilePreview 文件预览内容服务（鉴权由调用方完成）
func serveFilePreview(c *gin.Context, file model.ProjectFile) {
	// URL 指向外部时直接重定向
	if len(file.StorageKey) > 7 && (file.StorageKey[:7] == "http://" || file.StorageKey[:8] == "https://") {
		c.Redirect(302, file.StorageKey)
		return
	}
	// 本地文件：按存储相对路径读取并返回（内联展示）
	path := file.StorageKey
	if len(path) > 0 && path[0] != '/' {
		path = "./" + path
	}

	// DXF：自研渲染器转换为 SVG 内联预览（无第三方依赖）
	if file.FileType == "dxf" {
		data, err := os.ReadFile(path)
		if err != nil {
			notFound(c, "文件存储不存在")
			return
		}
		if res, err := dxf.Render(data); err == nil {
			c.Data(200, "image/svg+xml; charset=utf-8", []byte(res.SVG))
			return
		}
	}

	// DWG（或 DXF 自研渲染失败）：走第三方 CAD 渲染引擎适配器（CAD_CONVERT_API）
	if file.FileType == "dwg" || file.FileType == "dxf" {
		data, err := os.ReadFile(path)
		if err != nil {
			notFound(c, "文件存储不存在")
			return
		}
		if svg, err := convertCADExternal(data, file.OriginalName, file.FileType); err == nil {
			c.Data(200, "image/svg+xml; charset=utf-8", svg)
			return
		}
		badRequest(c, "CAD 文件暂不支持在线预览（未配置第三方渲染引擎），请下载后用 CAD 软件查看")
		return
	}

	c.FileAttachment(path, file.OriginalName)
}

// convertCADExternal 调用第三方 CAD 渲染转换服务（如 Aspose.CAD 封装服务 / 商业渲染网关），
// 返回 SVG。配置项 CAD_CONVERT_API 指向该服务的 HTTP 接口：
//   POST {api}  multipart: file=<原始文件>  form: format=svg  header: X-File-Type
// 响应体为 image/svg+xml 内容。
func convertCADExternal(data []byte, name, fileType string) ([]byte, error) {
	api := config.Get().CADConvertAPI
	if api == "" {
		return nil, fmt.Errorf("未配置 CAD_CONVERT_API")
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", name)
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(data); err != nil {
		return nil, err
	}
	if err := mw.WriteField("format", "svg"); err != nil {
		return nil, err
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", api, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-File-Type", fileType)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CAD 转换服务返回 %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

