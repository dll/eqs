package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ok 统一成功响应
func ok(c *gin.Context, data gin.H) {
	c.JSON(http.StatusOK, data)
}

// created 统一创建成功响应
func created(c *gin.Context, data gin.H) {
	c.JSON(http.StatusOK, data)
}

// fail 统一失败响应，errCode 为业务错误码，message_key 供前端翻译
func fail(c *gin.Context, status int, errCode string, message string) {
	c.JSON(status, gin.H{
		"error":       errCode,
		"message":     message,
		"message_key": errCode,
	})
}

// badRequest 参数错误
func badRequest(c *gin.Context, message string) {
	fail(c, http.StatusBadRequest, "bad_request", message)
}

// unauthorized 未授权
func unauthorized(c *gin.Context, message string) {
	fail(c, http.StatusUnauthorized, "unauthorized", message)
}

// notFound 资源不存在
func notFound(c *gin.Context, message string) {
	fail(c, http.StatusNotFound, "not_found", message)
}

// serverError 服务器错误
func serverError(c *gin.Context, err error) {
	if err != nil {
		fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	fail(c, http.StatusInternalServerError, "internal_error", "服务器内部错误")
}

// forbidden 无权限
func forbidden(c *gin.Context, message string) {
	fail(c, http.StatusForbidden, "forbidden", message)
}