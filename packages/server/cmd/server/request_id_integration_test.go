package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

func TestSetupRouterAddsRequestIDResponseHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := model.InitTestDB()
	r := setupRouter(db, config.Load())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config/public", nil)
	req.Header.Set("X-Request-ID", "integration-request-1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got != "integration-request-1" {
		t.Fatalf("X-Request-ID response header = %q", got)
	}
}
