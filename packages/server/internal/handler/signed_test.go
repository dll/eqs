package handler

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// TestSignPreviewToken 公开预览签名：绑定文件、过期、防篡改
func TestSignPreviewToken(t *testing.T) {
	config.ResetCache()
	t.Setenv("JWT_SECRET", "test-secret")

	tok := signPreviewToken(42)
	if tok == "" {
		t.Fatal("签名不应为空")
	}
	if !verifyPreviewToken(tok, 42) {
		t.Fatal("合法签名应通过校验")
	}
	if verifyPreviewToken(tok, 43) {
		t.Fatal("文件ID不匹配应拒绝")
	}
	if verifyPreviewToken(tok+"x", 42) {
		t.Fatal("篡改签名应拒绝")
	}
	if verifyPreviewToken("0.abc", 42) {
		t.Fatal("过期签名应拒绝")
	}
}

// TestPreviewFilePublic 公开预览接口：签名有效放行，缺失/无效拒绝
func TestPreviewFilePublic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	config.ResetCache()
	t.Setenv("JWT_SECRET", "test-secret")

	// 外部 URL 图片（放行时以 302 重定向体现）
	file := model.ProjectFile{UploaderID: 1, OriginalName: "case.png", FileType: "png", StorageKey: "https://cdn.example.com/case.png", Version: 1}
	if err := model.DB.Create(&file).Error; err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	api := r.Group("/api/v1")
	api.GET("/file/:id/preview/public", PreviewFilePublic)

	// 无 token → 401
	w := doJSON(t, r, "GET", fmt.Sprintf("/api/v1/file/%d/preview/public", file.ID), nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("无 token 应401，得到 %d", w.Code)
	}
	// 有效 token → 302（外部 URL 重定向）
	tok := signPreviewToken(file.ID)
	w = doJSON(t, r, "GET", fmt.Sprintf("/api/v1/file/%d/preview/public?token=%s", file.ID, tok), nil)
	if w.Code != http.StatusFound {
		t.Fatalf("有效 token 应302，得到 %d: %s", w.Code, w.Body.String())
	}
	// 错误 token → 401
	w = doJSON(t, r, "GET", fmt.Sprintf("/api/v1/file/%d/preview/public?token=%s", file.ID, "999.bad"), nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("错误 token 应401，得到 %d", w.Code)
	}
}
