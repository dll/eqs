package handler

import (
	"github.com/gin-gonic/gin"
)

// ==================== 多端 ====================

// PlatformLinks 各端访问地址（从配置中心读取）
func PlatformLinks(c *gin.Context) {
	cfgs := getPublicCached()
	urls, _ := cfgs["multiplatform.urls"].(map[string]interface{})
	if urls == nil {
		urls = make(map[string]interface{})
	}

	getURL := func(key string) string {
		if v, ok := urls[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}

	ok(c, gin.H{"platforms": []gin.H{
		{"id": "h5", "name": "H5", "url": getURL("h5")},
		{"id": "mp-weixin", "name": "微信小程序", "url": getURL("mp-weixin")},
		{"id": "app-ios", "name": "iOS App", "url": getURL("app-ios")},
		{"id": "app-android", "name": "Android App", "url": getURL("app-android")},
	}})
}
