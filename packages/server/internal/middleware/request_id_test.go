package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_UsesSafeIncomingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/ping", func(c *gin.Context) {
		if got := c.GetString("request_id"); got != "client-123" {
			t.Fatalf("request_id = %q", got)
		}
		c.Status(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(requestIDHeader, " client-123 ")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Header().Get(requestIDHeader) != "client-123" {
		t.Fatalf("response request id = %q", w.Header().Get(requestIDHeader))
	}
}

func TestRequestID_ReplacesUnsafeIncomingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(requestIDHeader, "bad value")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	got := w.Header().Get(requestIDHeader)
	if got == "" || got == "bad value" || len(got) != 32 {
		t.Fatalf("unsafe request id was not replaced: %q", got)
	}
}
