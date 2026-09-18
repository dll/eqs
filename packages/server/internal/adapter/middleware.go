package adapter

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// gate 对适配器前缀统一执行只读和项目隔离门禁。
func gate(now func() response) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/adapter/v1") {
			c.Next()
			return
		}

		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			body := now()
			body.Error = "method_not_allowed"
			body.Message = "适配器仅允许只读访问"
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, body)
			return
		}

		requestedProject := c.Query("project_id")
		if requestedProject == "" {
			requestedProject = c.GetHeader("X-Project-ID")
		}
		if requestedProject != "" && !strings.EqualFold(requestedProject, ProjectID) {
			body := now()
			body.Error = "cross_project_forbidden"
			body.Message = "拒绝跨项目访问"
			c.AbortWithStatusJSON(http.StatusForbidden, body)
			return
		}

		c.Next()
	}
}
