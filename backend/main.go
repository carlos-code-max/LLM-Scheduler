package main

import (
	"context"
	"fmt"
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
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config: ", err)
	}

	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err == nil {
		logger.SetLevel(level)
	}

	logger.Info("Starting LLM Scheduler Server...")
	logger.Infof("Version: %s, Environment: %s", cfg.App.Version, cfg.App.Env)

	db, err := database.Init(cfg)
	if err != nil {
		panic(err)
	}

	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	redisClient, err := queue.InitRedis(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize Redis: ", err)
	}
	defer redisClient.Close()

	queueManager := queue.NewManager(redisClient, cfg, logger)

	taskService := services.NewTaskService(db, queueManager, logger)
	modelService := services.NewModelService(db, logger)
	statsService := services.NewStatsService(db, logger)

	workerManager := worker.NewManager(cfg, db, queueManager, taskService, modelService, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := workerManager.Start(ctx); err != nil {
			logger.Error("Worker manager error: ", err)
		}
	}()

	if cfg.App.Env == "live" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(utils.LoggerMiddleware(logger))

	// CORS
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

	routes.RegisterRoutes(router, taskService, modelService, statsService, queueManager, logger)
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		logger.Infof("Server starting on http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	cancel()
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: ", err)
	}

	logger.Info("Server exited")
}
