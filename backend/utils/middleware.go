package utils

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LoggerMiddleware 日志中间件
func LoggerMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return gin.LoggerWithWriter(gin.DefaultWriter)
}

// RequestLoggerMiddleware 请求日志中间件
func RequestLoggerMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		
		// 处理请求
		c.Next()
		
		// 记录请求日志
		duration := time.Since(startTime)
		logger.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"duration":   duration,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}).Info("HTTP request completed")
	}
}

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// 处理错误
		for _, err := range c.Errors {
			logger.WithFields(logrus.Fields{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"error":  err.Error(),
			}).Error("Request error")
		}
	}
}

// RateLimitMiddleware 限流中间件（简单实现）
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以实现限流逻辑
		// 例如使用 Redis 存储访问频率
		c.Next()
	}
}

// AuthMiddleware 认证中间件（预留）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以实现认证逻辑
		// 例如验证 JWT Token
		c.Next()
	}
}
