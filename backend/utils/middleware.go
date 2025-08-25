// backend/utils/middleware.go
package utils

import (
	"context"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// context key for request ID
type ctxKey string

const requestIDKey ctxKey = "request_id"

// GetRequestID 从上下文获取请求ID
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(requestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// CORSConfig 用于配置 CORS 策略
// 若 AllowedOrigins 包含 "*" 则放开所有来源
// 若为空则回退为 "*"
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposeHeaders    []string
	AllowCredentials bool
}

// MiddlewareConfig 其他中间件配置
type MiddlewareConfig struct {
	EnableAccessLog bool
}

// WithCORS 添加基础CORS支持（默认允许所有域，按需收紧）
func WithCORS(next http.Handler) http.Handler {
	return NewCORS(CORSConfig{})(next)
}

// NewCORS 根据配置返回 CORS 中间件
func NewCORS(cfg CORSConfig) func(http.Handler) http.Handler {
	// 默认值
	allowedOrigins := cfg.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
	allowedMethods := cfg.AllowedMethods
	if len(allowedMethods) == 0 {
		allowedMethods = []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"}
	}
	allowedHeaders := cfg.AllowedHeaders
	if len(allowedHeaders) == 0 {
		allowedHeaders = []string{"Content-Type", "Authorization", "X-Request-ID"}
	}
	exposeHeaders := cfg.ExposeHeaders
	if len(exposeHeaders) == 0 {
		exposeHeaders = []string{"X-Request-ID"}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowOrigin := "*"
			if !(len(allowedOrigins) == 1 && allowedOrigins[0] == "*") {
				if origin != "" && slices.Contains(allowedOrigins, origin) {
					allowOrigin = origin
				} else {
					// 若不在白名单，不设置允许来源，直接走后续逻辑（可按需返回403）
					allowOrigin = ""
				}
			}

			if allowOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			}
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ","))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(exposeHeaders, ", "))
			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// WithRequestID 确保每个请求拥有唯一ID，并写回响应头
func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if strings.TrimSpace(rid) == "" {
			// 生成16位随机ID
			if v, err := GenerateRandomString(16); err == nil {
				rid = v
			} else {
				rid = "unknown"
			}
		}
		w.Header().Set("X-Request-ID", rid)
		ctx := context.WithValue(r.Context(), requestIDKey, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// responseWriter 包装以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// WithLogging 简单访问日志中间件（始终开启版本）
func WithLogging(next http.Handler) http.Handler {
	return NewLogging(true)(next)
}

// NewLogging 根据开关返回日志中间件
func NewLogging(enabled bool) func(http.Handler) http.Handler {
	if !enabled {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			dur := time.Since(start)
			rid := GetRequestID(r.Context())
			// 继续使用标准log，因为这是HTTP中间件层
			log.Printf("%s %s %d %s rid=%s", r.Method, r.URL.Path, rw.status, dur, rid)
		})
	}
}

// WithRecover 防止panic导致进程崩溃
func WithRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// 继续使用标准log，因为这是HTTP中间件层的panic恢复
				log.Printf("panic recovered: %v", rec)
				Fail(w, http.StatusInternalServerError, ErrInternalServer, ErrorMessage(ErrInternalServer))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Chain 组合多个中间件，按给定顺序包裹
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// RequestIDMiddlewareWithGin 为Gin添加请求ID中间件
func RequestIDMiddlewareWithGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if strings.TrimSpace(rid) == "" {
			// 生成16位随机ID
			if v, err := GenerateRandomString(16); err == nil {
				rid = v
			} else {
				rid = "unknown"
			}
		}
		// 设置响应头
		c.Header("X-Request-ID", rid)
		// 设置到Gin上下文中
		c.Set(string(requestIDKey), rid)
		c.Next()
	}
}

// ErrorHandlerMiddlewareWithGin 错误处理中间件，捕获内部错误并添加到响应头
func ErrorHandlerMiddlewareWithGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建自定义的ResponseWriter来捕获错误
		blw := &bodyLogWriter{body: &strings.Builder{}, ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// 检查是否有错误发生
		if len(c.Errors) > 0 {
			// 获取最后一个错误
			lastError := c.Errors.Last()
			if lastError != nil {
				// 将错误信息添加到响应头
				c.Header("X-Error-Message", lastError.Error())
				c.Header("X-Error-Type", "internal")
			}
		}

		// 检查HTTP状态码，如果是5xx错误，也添加错误头
		if c.Writer.Status() >= 500 {
			if c.GetHeader("X-Error-Message") == "" {
				c.Header("X-Error-Message", "Internal server error")
				c.Header("X-Error-Type", "server")
			}
		}
	}
}

// bodyLogWriter 用于捕获响应体的writer
type bodyLogWriter struct {
	gin.ResponseWriter
	body *strings.Builder
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// LoggingMiddlewareWithGin 根据开关返回 Gin 日志中间件
func LoggingMiddlewareWithGin(enabled bool) gin.HandlerFunc {
	if !enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		dur := time.Since(start)

		// 获取状态码
		status := c.Writer.Status()

		// 获取请求ID（如果存在）
		ridValue, exists := c.Get(string(requestIDKey))
		rid := "unknown"
		if exists {
			if ridStr, ok := ridValue.(string); ok {
				rid = ridStr
			}
		}

		// 继续使用标准log，因为这是Gin中间件层
		log.Printf("%s %s %d %s rid=%s", c.Request.Method, c.Request.URL.Path, status, dur, rid)
	}
}

// RateLimitMiddlewareWithGin 创建基于IP的速率限制 Gin 中间件
func RateLimitMiddlewareWithGin(r rate.Limit, b int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(r, b)

	return func(c *gin.Context) {
		// 添加空指针检查
		if limiter == nil {
			// 如果limiter为nil，记录错误并继续处理请求（不进行限流）
			log.Printf("Warning: rate limiter is nil, skipping rate limiting")
			c.Next()
			return
		}

		ip := c.ClientIP()
		limiterInstance := limiter.GetLimiter(ip)
		if limiterInstance == nil {
			// 如果特定IP的limiter为nil，记录错误并继续处理请求
			log.Printf("Warning: limiter for IP %s is nil, skipping rate limiting", ip)
			c.Next()
			return
		}

		// 添加对Allow方法调用的保护
		var allowed bool
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Warning: rate limiter Allow() panic recovered: %v", err)
					// 即使限流器出错，也允许请求通过
					allowed = true
				}
			}()
			allowed = limiterInstance.Allow()
		}()

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    ErrTooManyRequests,
				"message": "Too many requests",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// NewCORSWithGin 根据配置返回 Gin CORS 中间件
func NewCORSWithGin(cfg CORSConfig) gin.HandlerFunc {
	// 默认值
	allowedOrigins := cfg.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
	allowedMethods := cfg.AllowedMethods
	if len(allowedMethods) == 0 {
		allowedMethods = []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"}
	}
	allowedHeaders := cfg.AllowedHeaders
	if len(allowedHeaders) == 0 {
		allowedHeaders = []string{"Content-Type", "Authorization", "X-Request-ID"}
	}
	exposeHeaders := cfg.ExposeHeaders
	if len(exposeHeaders) == 0 {
		exposeHeaders = []string{"X-Request-ID"}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowOrigin := "*"
		if !(len(allowedOrigins) == 1 && allowedOrigins[0] == "*") {
			if origin != "" && slices.Contains(allowedOrigins, origin) {
				allowOrigin = origin
			} else {
				// 若不在白名单，不设置允许来源，直接走后续逻辑（可按需返回403）
				allowOrigin = ""
			}
		}

		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
		}
		c.Header("Access-Control-Allow-Methods", strings.Join(allowedMethods, ","))
		c.Header("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
		c.Header("Access-Control-Expose-Headers", strings.Join(exposeHeaders, ", "))
		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
