package handler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ==================== 版本 ====================

// VersionCheck 版本检查
func VersionCheck(c *gin.Context) {
	current := c.Query("current")
	platform := c.DefaultQuery("platform", "h5")

	var latest model.SystemVersion
	err := model.DB.Where("platform IN (?)", []string{platform, "all"}).
		Order("build DESC").First(&latest).Error
	if err != nil {
		ok(c, gin.H{"update_available": false, "message": "当前已是最新版本"})
		return
	}

	needUpdate := false
	if current != "" {
		needUpdate = compareVersions(current, latest.Version) < 0
	}

	ok(c, gin.H{
		"update_available": needUpdate,
		"version":          latest.Version,
		"build":            latest.Build,
		"update_url":       latest.UpdateURL,
		"release_notes":    latest.ReleaseNotes,
		"mandatory":        latest.Mandatory,
		"message":          fmt.Sprintf("最新版本 %s", latest.Version),
	})
}

// VersionLatest 最新版本信息
func VersionLatest(c *gin.Context) {
	platform := c.DefaultQuery("platform", "all")
	var latest model.SystemVersion
	err := model.DB.Where("platform IN (?)", []string{platform, "all"}).
		Order("build DESC").First(&latest).Error
	if err != nil {
		notFound(c, "暂无版本信息")
		return
	}
	ok(c, gin.H{"version": latest})
}

// compareVersions 比较版本号（语义化）
func compareVersions(a, b string) int {
	parse := func(v string) []int {
		var parts []int
		for _, p := range splitVersion(v) {
			n, _ := strconv.Atoi(p)
			parts = append(parts, n)
		}
		for len(parts) < 3 {
			parts = append(parts, 0)
		}
		return parts
	}
	pa, pb := parse(a), parse(b)
	for i := 0; i < 3; i++ {
		if pa[i] < pb[i] {
			return -1
		}
		if pa[i] > pb[i] {
			return 1
		}
	}
	return 0
}

func splitVersion(v string) []string {
	var parts []string
	cur := ""
	for _, ch := range v {
		if ch >= '0' && ch <= '9' {
			cur += string(ch)
		} else if cur != "" {
			parts = append(parts, cur)
			cur = ""
		}
	}
	if cur != "" {
		parts = append(parts, cur)
	}
	return parts
}

// AdminPublishVersion 发布新版本（管理员）
func AdminPublishVersion(c *gin.Context) {
	var req struct {
		Version      string `json:"version" binding:"required"`
		Build        int    `json:"build"`
		Platform     string `json:"platform"`
		UpdateURL    string `json:"update_url"`
		ReleaseNotes string `json:"release_notes"`
		Mandatory    bool   `json:"mandatory"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	if req.Platform == "" {
		req.Platform = "all"
	}

	version := model.SystemVersion{
		Version:      req.Version,
		Build:        req.Build,
		Platform:     req.Platform,
		UpdateURL:    req.UpdateURL,
		ReleaseNotes: req.ReleaseNotes,
		Mandatory:    req.Mandatory,
		ReleasedAt:   time.Now(),
	}
	if err := model.DB.Create(&version).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "version.publish", "version", version.ID, gin.H{"version": req.Version, "platform": req.Platform})
	ok(c, gin.H{"version": version, "message": "版本已发布"})
}

// AdminListVersions 版本历史（管理员）
func AdminListVersions(c *gin.Context) {
	var versions []model.SystemVersion
	model.DB.Order("build DESC").Find(&versions)
	ok(c, gin.H{"versions": versions, "count": len(versions)})
}
