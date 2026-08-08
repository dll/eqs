package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eqs/server/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func setup() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestAuth_MissingToken(t *testing.T) {
	r := setup()
	cfg := config.Load()
	r.Use(Auth(cfg))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望401，得到 %d", w.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	r := setup()
	cfg := config.Load()
	r.Use(Auth(cfg))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望401，得到 %d", w.Code)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	r := setup()
	cfg := config.Load()
	r.Use(Auth(cfg))

	var gotUserID uint
	var gotUserType int
	r.GET("/ping", func(c *gin.Context) {
		gotUserID = c.GetUint("user_id")
		gotUserType = c.GetInt("user_type")
		c.Status(http.StatusOK)
	})

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   42,
		"user_type": 2,
		"exp":       time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString([]byte(cfg.JWTSecret))

	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望200，得到 %d", w.Code)
	}
	if gotUserID != 42 || gotUserType != 2 {
		t.Fatalf("期望注入 user_id=42 type=2，得到 %d %d", gotUserID, gotUserType)
	}
}

func TestAuth_NonMapClaims(t *testing.T) {
	r := setup()
	cfg := config.Load()
	r.Use(Auth(cfg))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	// 使用自定义 Claims 类型（非 MapClaims）
	type customClaims struct {
		jwt.RegisteredClaims
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims{})
	tokenStr, _ := token.SignedString([]byte(cfg.JWTSecret))

	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("非MapClaims应401，得到 %d", w.Code)
	}
}

func TestLogger_Run(t *testing.T) {
	r := setup()
	r.Use(Logger())
	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望200，得到 %d", w.Code)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
}

func TestCORS_Headers(t *testing.T) {
	r := setup()
	r.Use(CORS())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "https://eqs.example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望200，得到 %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("缺少CORS头: %v", w.Header())
	}
}

func TestCORS_OptionsPreflight(t *testing.T) {
	r := setup()
	r.Use(CORS())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("OPTIONS", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("预检应204，得到 %d", w.Code)
	}
}