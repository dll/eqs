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

// publicCacheLoadMu 用作单飞锁：冷启动首次加载时只允许一个 goroutine 执行 DB 查询，
// 消除 R2-2 中「空缓存下多 goroutine 重复 loadPublicCache」的竞态。
// 采用独立互斥量而非 sync.Once，避免 invalidate 时复制已使用的 sync.Once 带来的数据竞争。
var publicCacheLoadMu sync.Mutex

func loadPublicCache() {
	publicCacheLoadMu.Lock()
	defer publicCacheLoadMu.Unlock()
	// 双检：拿到单飞锁后若缓存已被并发回填，则直接返回，避免二次重复加载
	publicCache.mu.RLock()
	filled := len(publicCache.configs) > 0
	publicCache.mu.RUnlock()
	if filled {
		return
	}
	var configs []model.SystemConfig
	model.DB.Where("is_public = ?", true).Find(&configs)
	m := make(map[string]interface{}, len(configs))
	for _, cfg := range configs {
		m[cfg.ConfigKey] = parseConfigValue(cfg.ConfigValue, cfg.ValueType)
	}
	publicCache.mu.Lock()
	publicCache.configs = m
	publicCache.updated = time.Now()
	publicCache.mu.Unlock()
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
	publicCache.mu.RLock()
	defer publicCache.mu.RUnlock()
	return publicCache.configs
}
