package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"llm-scheduler/config"
	"llm-scheduler/database"
	"llm-scheduler/queue"
	"llm-scheduler/routes"
	"llm-scheduler/services"
	"llm-scheduler/utils"
	"llm-scheduler/worker"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 初始化日志
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config: ", err)
	}

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err == nil {
		logger.SetLevel(level)
	}

	logger.Info("Starting LLM Scheduler Server...")
	logger.Infof("Version: %s, Environment: %s", cfg.App.Version, cfg.App.Env)

	// 初始化数据库
	db, err := database.Initialize(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize database: ", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// 初始化 Redis
	redisClient, err := queue.InitRedis(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize Redis: ", err)
	}
	defer redisClient.Close()

	// 初始化队列管理器
	queueManager := queue.NewManager(redisClient, cfg, logger)

	// 初始化服务
	taskService := services.NewTaskService(db, queueManager, logger)
	modelService := services.NewModelService(db, logger)
	statsService := services.NewStatsService(db, logger)

	// 初始化 Worker 管理器
	workerManager := worker.NewManager(cfg, db, queueManager, taskService, modelService, logger)

	// 启动 Worker 池
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go func() {
		if err := workerManager.Start(ctx); err != nil {
			logger.Error("Worker manager error: ", err)
		}
	}()

	// 设置 Gin 模式
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 路由
	router := gin.New()

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(utils.LoggerMiddleware(logger))

	// CORS 配置
	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		ExposeHeaders:    cfg.CORS.ExposeHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
	}
	if cfg.CORS.MaxAge != "" {
		if duration, err := time.ParseDuration(cfg.CORS.MaxAge); err == nil {
			corsConfig.MaxAge = duration
		}
	}
	router.Use(cors.New(corsConfig))

	// 注册路由
	routes.RegisterRoutes(router, taskService, modelService, statsService, queueManager, logger)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 启动服务器
	go func() {
		logger.Infof("Server starting on http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: ", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 停止 Worker 管理器
	cancel()

	// 关闭 HTTP 服务器
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: ", err)
	}

	logger.Info("Server exited")
}
