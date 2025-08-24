package routes

import (
	"llm-scheduler/handlers"
	"llm-scheduler/queue"
	"llm-scheduler/services"
	"llm-scheduler/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(
	router *gin.Engine,
	taskService *services.TaskService,
	modelService *services.ModelService,
	statsService *services.StatsService,
	queueManager *queue.Manager,
	logger *logrus.Logger,
) {
	// 获取依赖（这里需要修改，实际应该从参数传入）
	var db *gorm.DB
	var redisClient *redis.Client
	
	// 创建处理器
	taskHandler := handlers.NewTaskHandler(taskService, logger)
	modelHandler := handlers.NewModelHandler(modelService, logger)
	statsHandler := handlers.NewStatsHandler(statsService, logger)
	systemHandler := handlers.NewSystemHandler(db, redisClient, queueManager, logger)

	// 添加中间件
	router.Use(utils.RequestLoggerMiddleware(logger))
	router.Use(utils.ErrorHandlerMiddleware(logger))

	// API 版本分组
	v1 := router.Group("/api/v1")
	{
		// 系统相关路由
		system := v1.Group("/system")
		{
			system.GET("/health", systemHandler.HealthCheck)
			system.GET("/info", systemHandler.GetSystemInfo)
		}

		// 任务相关路由
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", taskHandler.CreateTask)           // 创建任务
			tasks.GET("", taskHandler.ListTasks)            // 获取任务列表
			tasks.GET("/:id", taskHandler.GetTask)          // 获取任务详情
			tasks.PUT("/:id", taskHandler.UpdateTask)       // 更新任务
			tasks.DELETE("/:id", taskHandler.CancelTask)    // 取消任务
			tasks.POST("/:id/retry", taskHandler.RetryTask) // 重试任务
			tasks.GET("/stats", taskHandler.GetTaskStats)   // 任务统计
		}

		// 模型相关路由
		models := v1.Group("/models")
		{
			models.POST("", modelHandler.CreateModel)                    // 创建模型
			models.GET("", modelHandler.ListModels)                     // 获取模型列表
			models.GET("/available", modelHandler.GetAvailableModels)   // 获取可用模型
			models.GET("/stats", modelHandler.GetModelStats)            // 模型统计
			models.GET("/:id", modelHandler.GetModel)                   // 获取模型详情
			models.PUT("/:id", modelHandler.UpdateModel)                // 更新模型
			models.DELETE("/:id", modelHandler.DeleteModel)             // 删除模型
			models.PUT("/:id/status", modelHandler.UpdateModelStatus)   // 更新模型状态
		}

		// 统计相关路由
		stats := v1.Group("/stats")
		{
			stats.GET("/dashboard", statsHandler.GetDashboardStats)      // Dashboard 统计
			stats.GET("/tasks/date", statsHandler.GetTaskStatsByDate)    // 按日期统计任务
			stats.GET("/tasks/model", statsHandler.GetTaskStatsByModel)  // 按模型统计任务
			stats.GET("/tasks/type", statsHandler.GetTaskStatsByType)    // 按类型统计任务
		}
	}

	// 根路径重定向到健康检查
	router.GET("/", func(c *gin.Context) {
		utils.Success(c, gin.H{
			"name":    "LLM Scheduler",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// 404 处理
	router.NoRoute(func(c *gin.Context) {
		utils.NotFound(c, "API接口不存在")
	})
}
