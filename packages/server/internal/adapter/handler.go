package adapter

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type handler struct{ service *service }

func (h *handler) metadataResponse() response {
	return response{Metadata: Metadata{ProjectID: ProjectID, Source: Source, ObservedAt: h.service.now(), SchemaVersion: SchemaVersion}}
}

func (h *handler) unavailable(c *gin.Context) {
	body := h.metadataResponse()
	body.Status = "unavailable"
	body.Message = "Database connection failed"
	c.JSON(http.StatusServiceUnavailable, body)
}

func (h *handler) health(c *gin.Context) {
	if err := h.service.available(); err != nil {
		h.unavailable(c)
		return
	}
	body := h.metadataResponse()
	body.Status = "healthy"
	c.JSON(http.StatusOK, body)
}

func (h *handler) projects(c *gin.Context) {
	if err := h.service.available(); err != nil {
		h.unavailable(c)
		return
	}
	page, size, ok := h.pagination(c)
	if !ok {
		return
	}
	status := c.Query("status")
	data, err := h.service.projects(page, size, status)
	if err != nil {
		if strings.HasPrefix(err.Error(), "unsupported status") {
			h.badRequest(c, "status 必须是 active、completed 或 cancelled")
			return
		}
		h.unavailable(c)
		return
	}
	body := h.metadataResponse()
	body.Data = data
	c.JSON(http.StatusOK, body)
}

func (h *handler) projectSummary(c *gin.Context) {
	if err := h.service.available(); err != nil {
		h.unavailable(c)
		return
	}
	raw := strings.TrimPrefix(c.Param("id"), "proj_")
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		h.badRequest(c, "无效的项目 ID")
		return
	}
	data, err := h.service.projectSummary(uint(id))
	if errors.Is(err, errNotFound) {
		body := h.metadataResponse()
		body.Error = "not_found"
		body.Message = "项目不存在"
		c.JSON(http.StatusNotFound, body)
		return
	}
	if err != nil {
		h.unavailable(c)
		return
	}
	body := h.metadataResponse()
	body.Data = data
	c.JSON(http.StatusOK, body)
}

func (h *handler) alerts(c *gin.Context) {
	if err := h.service.available(); err != nil {
		h.unavailable(c)
		return
	}
	page, size, ok := h.pagination(c)
	if !ok {
		return
	}
	severity := c.Query("severity")
	if severity != "" && severity != "info" && severity != "warning" && severity != "error" && severity != "critical" {
		h.badRequest(c, "severity 必须是 info、warning、error 或 critical")
		return
	}
	data, err := h.service.alerts(page, size, severity)
	if err != nil {
		h.unavailable(c)
		return
	}
	body := h.metadataResponse()
	body.Data = data
	c.JSON(http.StatusOK, body)
}

func (h *handler) version(c *gin.Context) {
	if err := h.service.available(); err != nil {
		h.unavailable(c)
		return
	}
	eqsVersion, err := h.service.eqsVersion()
	if err != nil {
		h.unavailable(c)
		return
	}
	body := h.metadataResponse()
	body.AdapterVersion = AdapterVersion
	body.EQSVersion = eqsVersion
	c.JSON(http.StatusOK, body)
}

func (h *handler) badRequest(c *gin.Context, message string) {
	body := h.metadataResponse()
	body.Error = "bad_request"
	body.Message = message
	c.JSON(http.StatusBadRequest, body)
}

func (h *handler) pagination(c *gin.Context) (int, int, bool) {
	page, err := strconv.Atoi(defaultString(c.Query("page"), "1"))
	if err != nil || page < 1 {
		h.badRequest(c, "page 必须为正整数")
		return 0, 0, false
	}
	size, err := strconv.Atoi(defaultString(c.Query("page_size"), "20"))
	if err != nil || size < 1 || size > 100 {
		h.badRequest(c, "page_size 必须在 1 到 100 之间")
		return 0, 0, false
	}
	return page, size, true
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
