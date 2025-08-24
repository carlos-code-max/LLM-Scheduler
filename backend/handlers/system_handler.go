package handlers

import (
	"llm-scheduler/database"
	"llm-scheduler/queue"
	"llm-scheduler/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SystemHandler 系统处理器
type SystemHandler struct {
	db           *gorm.DB
	redisClient  *redis.Client
	queueManager *queue.Manager
	logger       *logrus.Logger
}

// NewSystemHandler 创建系统处理器
func NewSystemHandler(db *gorm.DB, redisClient *redis.Client, queueManager *queue.Manager, logger *logrus.Logger) *SystemHandler {
	return &SystemHandler{
		db:           db,
		redisClient:  redisClient,
		queueManager: queueManager,
		logger:       logger,
	}
}

// HealthCheck 健康检查
func (h *SystemHandler) HealthCheck(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": gin.H{"database": "unknown", "redis": "unknown", "queue": "unknown"},
	}

	// 检查数据库连接
	if err := database.Health(h.db); err != nil {
		h.logger.WithError(err).Error("Database health check failed")
		health["database"] = "error"
		health["database_error"] = err.Error()
		health["status"] = "error"
	} else {
		health["database"] = "ok"
	}

	// 检查 Redis 连接
	if err := queue.HealthCheck(h.redisClient); err != nil {
		h.logger.WithError(err).Error("Redis health check failed")
		health["redis"] = "error"
		health["redis_error"] = err.Error()
		health["status"] = "error"
	} else {
		health["redis"] = "ok"
	}

	// 检查队列状态
	if queueStatus, err := h.queueManager.GetQueueStatus(c.Request.Context()); err != nil {
		h.logger.WithError(err).Error("Queue health check failed")
		health["queue"] = "error"
		health["queue_error"] = err.Error()
		health["status"] = "error"
	} else {
		health["queue"] = "ok"
		health["queue_status"] = queueStatus
	}

	if health["status"] == "ok" {
		utils.Success(c, health)
	} else {
		utils.InternalServerError(c, "系统健康检查失败")
		c.Header("Content-Type", "application/json")
		c.JSON(500, health)
	}
}

// GetSystemInfo 获取系统信息
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	info := map[string]interface{}{
		"version": "1.0.0",
		"environment": "development", // 可以从配置中获取
	}

	// 获取数据库统计
	if dbStats, err := database.GetStats(h.db); err == nil {
		info["database_stats"] = dbStats
	}

	// 获取 Redis 信息
	if redisInfo, err := queue.GetRedisInfo(h.redisClient); err == nil {
		info["redis_info"] = redisInfo
	}

	// 获取队列状态
	if queueStatus, err := h.queueManager.GetQueueStatus(c.Request.Context()); err == nil {
		info["queue_status"] = queueStatus
	}

	utils.Success(c, info)
}
