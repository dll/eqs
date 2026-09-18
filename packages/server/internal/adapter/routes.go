package adapter

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Register 注册 AOC 只读适配器路由，不改变现有业务路由与模型。
func Register(router *gin.Engine, db *gorm.DB) {
	h := &handler{service: &service{db: db, now: time.Now}}
	router.Use(gate(h.metadataResponse))
	group := router.Group("/adapter/v1")
	group.GET("/health", h.health)
	group.HEAD("/health", h.health)
	group.GET("/projects", h.projects)
	group.HEAD("/projects", h.projects)
	group.GET("/projects/:id/summary", h.projectSummary)
	group.HEAD("/projects/:id/summary", h.projectSummary)
	group.GET("/alerts", h.alerts)
	group.HEAD("/alerts", h.alerts)
	group.GET("/version", h.version)
	group.HEAD("/version", h.version)
}
