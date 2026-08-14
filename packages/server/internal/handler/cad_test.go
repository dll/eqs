package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// TestConvertCADExternal 第三方 CAD 转换服务适配器（配置后转发并返回转换结果）
func TestConvertCADExternal(t *testing.T) {
	// 未配置 → 返回错误
	config.ResetCache()
	t.Setenv("CAD_CONVERT_API", "")
	if _, err := convertCADExternal([]byte("dwg-bytes"), "a.dwg", "dwg"); err == nil {
		t.Fatal("未配置 CAD_CONVERT_API 应返回错误")
	}

	// 配置 Mock 转换服务 → 返回转换结果
	var gotFile []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Errorf("未收到文件字段: %v", err)
			w.WriteHeader(400)
			return
		}
		defer file.Close()
		buf := make([]byte, 64)
		n, _ := file.Read(buf)
		gotFile = buf[:n]
		if r.Header.Get("X-File-Type") != "dwg" {
			t.Errorf("缺少 X-File-Type 头")
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		fmt.Fprint(w, `<svg xmlns="http://www.w3.org/2000/svg"><rect width="10" height="10"/></svg>`)
	}))
	defer srv.Close()

	config.ResetCache()
	t.Setenv("CAD_CONVERT_API", srv.URL)
	out, err := convertCADExternal([]byte("dwg-bytes"), "a.dwg", "dwg")
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	if !strings.Contains(string(out), "<svg") {
		t.Fatalf("返回内容不是 SVG: %s", string(out))
	}
	if string(gotFile) != "dwg-bytes" {
		t.Fatalf("转发文件内容不符: %q", string(gotFile))
	}
}

// TestPreviewDWG_Unconfigured 未配置第三方引擎时 DWG 预览返回下载提示
func TestPreviewDWG_Unconfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	config.ResetCache()
	t.Setenv("CAD_CONVERT_API", "")

	// 写入一个本地 dwg 文件
	dir := "uploads/cad_test"
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	fpath := filepath.Join(dir, "sample.dwg")
	if err := os.WriteFile(fpath, []byte("AC1015 fake dwg"), 0o600); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("uploads/cad_test")

	file := model.ProjectFile{UploaderID: 1, OriginalName: "sample.dwg", FileType: "dwg", StorageKey: dir + "/sample.dwg", Version: 1}
	if err := model.DB.Create(&file).Error; err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	api := r.Group("/api/v1")
	auth := api.Group("")
	auth.Use(AuthTestMiddleware(1, 3)) // 管理员
	auth.GET("/file/:id/preview", PreviewFile)

	w := doJSON(t, r, "GET", fmt.Sprintf("/api/v1/file/%d/preview", file.ID), nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("未配置引擎的 DWG 预览应返回400，得到 %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "下载") {
		t.Fatalf("应返回下载提示，实际: %s", w.Body.String())
	}
}
