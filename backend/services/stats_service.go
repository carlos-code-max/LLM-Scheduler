package services

import (
	"database/sql"
	"fmt"
	"time"

	"llm-scheduler/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// StatsService 统计服务
type StatsService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewStatsService 创建统计服务
func NewStatsService(db *gorm.DB, logger *logrus.Logger) *StatsService {
	return &StatsService{
		db:     db,
		logger: logger,
	}
}

// GetDashboardStats 获取 Dashboard 统计数据
func (s *StatsService) GetDashboardStats() (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// 获取任务统计
	taskStats, err := s.getTaskStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get task stats: %w", err)
	}
	stats.TaskStats = *taskStats

	// 获取模型统计
	modelStats, err := s.getModelStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get model stats: %w", err)
	}
	stats.ModelStats = modelStats

	// 获取队列状态（这里先返回空值，实际应该从队列管理器获取）
	stats.QueueStatus = models.QueueStatus{}

	// 获取 Worker 状态（这里先返回空值，实际应该从 Worker 管理器获取）
	stats.WorkerStatus = []models.WorkerStatus{}

	// 获取系统统计
	systemStats, err := s.getTodaySystemStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get system stats: %w", err)
	}
	stats.SystemStats = *systemStats

	// 获取最近任务
	recentTasks, err := s.getRecentTasks(10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent tasks: %w", err)
	}
	stats.RecentTasks = recentTasks

	return stats, nil
}

// getTaskStats 获取任务统计
func (s *StatsService) getTaskStats() (*models.TaskStats, error) {
	var stats models.TaskStats

	// 总任务数
	s.db.Model(&models.Task{}).Count(&stats.TotalTasks)
	
	// 各状态任务数
	s.db.Model(&models.Task{}).Where("status = ?", models.TaskStatusPending).Count(&stats.PendingTasks)
	s.db.Model(&models.Task{}).Where("status = ?", models.TaskStatusRunning).Count(&stats.RunningTasks)
	s.db.Model(&models.Task{}).Where("status = ?", models.TaskStatusCompleted).Count(&stats.CompletedTasks)
	s.db.Model(&models.Task{}).Where("status = ?", models.TaskStatusFailed).Count(&stats.FailedTasks)
	s.db.Model(&models.Task{}).Where("status = ?", models.TaskStatusCancelled).Count(&stats.CancelledTasks)

	// 计算成功率
	if stats.TotalTasks > 0 {
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	// 平均处理时间
	var avgMs sql.NullFloat64
	s.db.Model(&models.Task{}).
		Select("AVG(TIMESTAMPDIFF(MICROSECOND, started_at, completed_at) / 1000)").
		Where("started_at IS NOT NULL AND completed_at IS NOT NULL").
		Scan(&avgMs)
	
	if avgMs.Valid {
		stats.AvgProcessingMS = int64(avgMs.Float64)
	}

	return &stats, nil
}

// getModelStats 获取模型统计
func (s *StatsService) getModelStats() ([]models.ModelStats, error) {
	var stats []models.ModelStats
	
	query := `
		SELECT 
			m.*,
			COALESCE(pending_tasks, 0) as pending_tasks,
			COALESCE(running_tasks, 0) as running_tasks,
			ROUND(
				CASE WHEN m.total_requests > 0 
				THEN (m.success_requests * 100.0 / m.total_requests) 
				ELSE 0 END, 2
			) as success_rate,
			COALESCE(avg_response_ms, 0) as avg_response_ms
		FROM models m
		LEFT JOIN (
			SELECT 
				model_id,
				SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_tasks,
				SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running_tasks,
				AVG(CASE 
					WHEN started_at IS NOT NULL AND completed_at IS NOT NULL 
					THEN TIMESTAMPDIFF(MICROSECOND, started_at, completed_at) / 1000
					ELSE NULL 
				END) as avg_response_ms
			FROM tasks 
			GROUP BY model_id
		) t ON m.id = t.model_id
		ORDER BY m.id
	`

	if err := s.db.Raw(query).Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get model stats: %w", err)
	}

	return stats, nil
}

// getTodaySystemStats 获取今日系统统计
func (s *StatsService) getTodaySystemStats() (*models.SystemStats, error) {
	today := time.Now().Format("2006-01-02")
	
	var stats models.SystemStats
	err := s.db.Where("stat_date = ?", today).First(&stats).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果今日统计不存在，创建一个默认的
			stats = models.SystemStats{
				StatDate:            time.Now(),
				TotalTasks:          0,
				CompletedTasks:      0,
				FailedTasks:         0,
				AvgProcessingTimeMs: 0,
				QueueLength:         0,
				ActiveModels:        0,
			}
			return &stats, nil
		}
		return nil, fmt.Errorf("failed to get today system stats: %w", err)
	}

	return &stats, nil
}

// getRecentTasks 获取最近任务
func (s *StatsService) getRecentTasks(limit int) ([]models.Task, error) {
	var tasks []models.Task
	err := s.db.Preload("Model").
		Order("created_at DESC").
		Limit(limit).
		Find(&tasks).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get recent tasks: %w", err)
	}

	return tasks, nil
}

// GetTaskStatsByDate 按日期获取任务统计
func (s *StatsService) GetTaskStatsByDate(days int) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as total,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
			AVG(CASE 
				WHEN started_at IS NOT NULL AND completed_at IS NOT NULL 
				THEN TIMESTAMPDIFF(MICROSECOND, started_at, completed_at) / 1000
				ELSE NULL 
			END) as avg_processing_ms
		FROM tasks 
		WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`

	var results []map[string]interface{}
	if err := s.db.Raw(query, days).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get task stats by date: %w", err)
	}

	return results, nil
}

// GetTaskStatsByModel 按模型获取任务统计
func (s *StatsService) GetTaskStatsByModel() ([]map[string]interface{}, error) {
	query := `
		SELECT 
			m.name as model_name,
			m.type as model_type,
			COUNT(t.id) as total_tasks,
			SUM(CASE WHEN t.status = 'completed' THEN 1 ELSE 0 END) as completed_tasks,
			SUM(CASE WHEN t.status = 'failed' THEN 1 ELSE 0 END) as failed_tasks,
			SUM(CASE WHEN t.status = 'pending' THEN 1 ELSE 0 END) as pending_tasks,
			SUM(CASE WHEN t.status = 'running' THEN 1 ELSE 0 END) as running_tasks,
			ROUND(
				CASE WHEN COUNT(t.id) > 0 
				THEN (SUM(CASE WHEN t.status = 'completed' THEN 1 ELSE 0 END) * 100.0 / COUNT(t.id))
				ELSE 0 END, 2
			) as success_rate,
			AVG(CASE 
				WHEN t.started_at IS NOT NULL AND t.completed_at IS NOT NULL 
				THEN TIMESTAMPDIFF(MICROSECOND, t.started_at, t.completed_at) / 1000
				ELSE NULL 
			END) as avg_processing_ms
		FROM models m
		LEFT JOIN tasks t ON m.id = t.model_id
		GROUP BY m.id, m.name, m.type
		ORDER BY total_tasks DESC
	`

	var results []map[string]interface{}
	if err := s.db.Raw(query).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get task stats by model: %w", err)
	}

	return results, nil
}

// GetTaskStatsByType 按任务类型获取统计
func (s *StatsService) GetTaskStatsByType() ([]map[string]interface{}, error) {
	query := `
		SELECT 
			type,
			COUNT(*) as total_tasks,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_tasks,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_tasks,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_tasks,
			SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running_tasks,
			ROUND(
				CASE WHEN COUNT(*) > 0 
				THEN (SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) * 100.0 / COUNT(*))
				ELSE 0 END, 2
			) as success_rate,
			AVG(CASE 
				WHEN started_at IS NOT NULL AND completed_at IS NOT NULL 
				THEN TIMESTAMPDIFF(MICROSECOND, started_at, completed_at) / 1000
				ELSE NULL 
			END) as avg_processing_ms
		FROM tasks
		GROUP BY type
		ORDER BY total_tasks DESC
	`

	var results []map[string]interface{}
	if err := s.db.Raw(query).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get task stats by type: %w", err)
	}

	return results, nil
}

// UpdateDailyStats 更新每日统计
func (s *StatsService) UpdateDailyStats() error {
	today := time.Now().Format("2006-01-02")
	
	// 计算今日统计数据
	var totalTasks, completedTasks, failedTasks int64
	var avgProcessingMs sql.NullFloat64
	
	todayStart := time.Now().Truncate(24 * time.Hour)
	todayEnd := todayStart.Add(24 * time.Hour)
	
	s.db.Model(&models.Task{}).
		Where("created_at >= ? AND created_at < ?", todayStart, todayEnd).
		Count(&totalTasks)
	
	s.db.Model(&models.Task{}).
		Where("created_at >= ? AND created_at < ? AND status = ?", 
			todayStart, todayEnd, models.TaskStatusCompleted).
		Count(&completedTasks)
		
	s.db.Model(&models.Task{}).
		Where("created_at >= ? AND created_at < ? AND status = ?", 
			todayStart, todayEnd, models.TaskStatusFailed).
		Count(&failedTasks)
	
	s.db.Model(&models.Task{}).
		Select("AVG(TIMESTAMPDIFF(MICROSECOND, started_at, completed_at) / 1000)").
		Where("created_at >= ? AND created_at < ? AND started_at IS NOT NULL AND completed_at IS NOT NULL", 
			todayStart, todayEnd).
		Scan(&avgProcessingMs)

	// 获取活跃模型数量
	var activeModels int64
	s.db.Model(&models.Model{}).
		Where("status = ?", models.ModelStatusOnline).
		Count(&activeModels)

	// 更新或创建统计记录
	stats := models.SystemStats{
		StatDate:            todayStart,
		TotalTasks:          int(totalTasks),
		CompletedTasks:      int(completedTasks),
		FailedTasks:         int(failedTasks),
		AvgProcessingTimeMs: int(avgProcessingMs.Float64),
		QueueLength:         0, // 需要从队列管理器获取
		ActiveModels:        int(activeModels),
	}

	if err := s.db.Where("stat_date = ?", today).
		Assign(&stats).
		FirstOrCreate(&stats).Error; err != nil {
		return fmt.Errorf("failed to update daily stats: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"date":            today,
		"total_tasks":     totalTasks,
		"completed_tasks": completedTasks,
		"failed_tasks":    failedTasks,
		"active_models":   activeModels,
	}).Info("Daily stats updated")

	return nil
}
