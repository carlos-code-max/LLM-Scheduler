package database

import (
	"fmt"
	"time"

	"llm-scheduler/config"
	"llm-scheduler/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Initialize 初始化数据库连接
func Initialize(cfg *config.Config) (*gorm.DB, error) {
	// 构建 DSN
	dsn := cfg.Database.GetDSN()
	
	// GORM 配置
	gormConfig := &gorm.Config{}
	
	// 根据环境设置日志级别
	switch cfg.App.Env {
	case "development":
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	case "production":
		gormConfig.Logger = logger.Default.LogMode(logger.Warn)
	default:
		gormConfig.Logger = logger.Default.LogMode(logger.Silent)
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层 sql.DB 实例进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 自动迁移数据库表结构
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// migrate 执行数据库迁移
func migrate(db *gorm.DB) error {
	// 按依赖关系顺序迁移
	err := db.AutoMigrate(
		&models.Model{},
		&models.Task{},
		&models.TaskLog{},
		&models.SystemStats{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	// 创建索引
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// createIndexes 创建额外的索引
func createIndexes(db *gorm.DB) error {
	// 任务表复合索引
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tasks_model_status ON tasks(model_id, status)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tasks_status_priority ON tasks(status, priority DESC)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at DESC)
	`).Error; err != nil {
		return err
	}

	// 模型表索引
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_models_type_status ON models(type, status)
	`).Error; err != nil {
		return err
	}

	// 任务日志表索引
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_task_logs_task_created ON task_logs(task_id, created_at DESC)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_task_logs_level_created ON task_logs(level, created_at DESC)
	`).Error; err != nil {
		return err
	}

	return nil
}

// Health 健康检查
func Health(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// GetStats 获取数据库连接统计
func GetStats(db *gorm.DB) (map[string]interface{}, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration,
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}, nil
}
