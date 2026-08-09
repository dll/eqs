package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ==================== 性能监控 ====================

type requestStat struct {
	Count      int64         `json:"count"`
	ErrorCount int64         `json:"error_count"`
	TotalTime  time.Duration `json:"-"`
	AvgMs      float64       `json:"avg_ms"`
	P95Ms      float64       `json:"p95_ms"`
}

type monitor struct {
	mu       sync.RWMutex
	requests map[string]*requestStat
}

var reqMonitor = &monitor{requests: make(map[string]*requestStat)}

// MonitorMiddleware 记录请求耗时
func MonitorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		status := c.Writer.Status()

		reqMonitor.mu.Lock()
		stat, ok := reqMonitor.requests[path]
		if !ok {
			stat = &requestStat{}
			reqMonitor.requests[path] = stat
		}
		stat.Count++
		stat.TotalTime += duration
		stat.AvgMs = float64(stat.TotalTime.Milliseconds()) / float64(stat.Count)
		if status >= http.StatusBadRequest {
			stat.ErrorCount++
		}
		reqMonitor.mu.Unlock()
	}
}

// MonitorStats 性能统计接口（管理员）
func MonitorStats(c *gin.Context) {
	reqMonitor.mu.RLock()
	defer reqMonitor.mu.RUnlock()

	result := make(map[string]interface{}, len(reqMonitor.requests))
	for path, stat := range reqMonitor.requests {
		result[path] = gin.H{
			"count":       stat.Count,
			"error_count": stat.ErrorCount,
			"avg_ms":      stat.AvgMs,
			"error_rate":  float64(stat.ErrorCount) / float64(stat.Count) * 100,
		}
	}
	ok(c, gin.H{"stats": result, "timestamp": time.Now()})
}

// ==================== 版本检查限流 ====================

type versionLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

var versionLimit = &versionLimiter{
	requests: make(map[string][]time.Time),
	limit:    10,           // 每窗口最多 10 次
	window:   time.Hour,    // 1 小时窗口
}

// VersionRateLimit 版本检查限流中间件
func VersionRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		versionLimit.mu.Lock()
		timestamps := versionLimit.requests[ip]
		// 清理过期记录
		valid := timestamps[:0]
		for _, t := range timestamps {
			if now.Sub(t) < versionLimit.window {
				valid = append(valid, t)
			}
		}
		if len(valid) >= versionLimit.limit {
			versionLimit.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后再试"})
			c.Abort()
			return
		}
		valid = append(valid, now)
		versionLimit.requests[ip] = valid
		versionLimit.mu.Unlock()
		c.Next()
	}
}
