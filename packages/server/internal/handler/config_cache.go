package handler

import (
	"sync"
	"time"

	"github.com/eqs/server/internal/model"
)

// ==================== 配置缓存 ====================

type configCache struct {
	mu      sync.RWMutex
	configs map[string]interface{}
	updated time.Time
}

var publicCache = &configCache{configs: make(map[string]interface{})}

func loadPublicCache() {
	publicCache.mu.Lock()
	defer publicCache.mu.Unlock()
	var configs []model.SystemConfig
	model.DB.Where("is_public = ?", true).Find(&configs)
	m := make(map[string]interface{}, len(configs))
	for _, cfg := range configs {
		m[cfg.ConfigKey] = parseConfigValue(cfg.ConfigValue, cfg.ValueType)
	}
	publicCache.configs = m
	publicCache.updated = time.Now()
}

func invalidatePublicCache() {
	publicCache.mu.Lock()
	publicCache.configs = make(map[string]interface{})
	publicCache.mu.Unlock()
}

func getPublicCached() map[string]interface{} {
	publicCache.mu.RLock()
	if len(publicCache.configs) > 0 {
		defer publicCache.mu.RUnlock()
		return publicCache.configs
	}
	publicCache.mu.RUnlock()
	loadPublicCache()
	return publicCache.configs
}
