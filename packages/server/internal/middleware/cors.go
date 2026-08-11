package middleware

import (
	"net/url"
	"slices"
	"strings"

	"github.com/eqs/server/internal/config"
	"github.com/gin-gonic/gin"
)

// CORS 跨域配置
// P1 修复：生产环境按域名白名单限制来源，不再通配 *；
// 开发/测试环境保留 * 便于本地联调。配置经 config.Get() 缓存读取，热路径不重读环境变量。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		cfg := config.Get()

		// 生产环境且来自非白名单来源：不返回 CORS 头，浏览器将拦截该请求
		allowed := !cfg.IsProduction() || origin == "" ||
			isSameOrigin(origin, c.Request.Host) ||
			slices.Contains(cfg.CORSAllowedOrigins, origin)
		if !allowed {
			c.AbortWithStatus(403)
			return
		}

		// 有 Origin 时回显来源，否则通配（开发环境 / 无来源请求）
		allowOrigin := "*"
		if origin != "" {
			allowOrigin = origin
		}
		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isSameOrigin 判断请求 Origin 与后端看到的 Host 是否同源。
// nginx 代理用 $host 转发时会去掉端口（如 129.211.223.113），
// 而浏览器 Origin 通常带端口（http://129.211.223.113:8091），
// 因此按主机名比较而非整串相等，避免同源请求被误判为跨域。
func isSameOrigin(origin, host string) bool {
	o := originHost(origin)
	h := originHost(host)
	return o != "" && o == h
}

// originHost 从 URL 或 host[:port] 中提取小写主机名（兼容无 scheme、IPv6 括号形式）
func originHost(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	u, err := url.Parse(s)
	if err != nil {
		return ""
	}
	if host := u.Hostname(); host != "" {
		return strings.ToLower(host)
	}
	// url.Parse 对 host[:port]（无 scheme）会解析失败或放入 Path，回退直接按端口剥离
	if i := strings.IndexAny(s, "/?"); i >= 0 {
		s = s[:i]
	}
	if i := strings.LastIndex(s, ":"); i >= 0 {
		s = s[:i]
	}
	return strings.ToLower(s)
}
