package services

import (
	"context"
	"fmt"
	"time"

	"llm-scheduler/models"
	"llm-scheduler/queue"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TaskService 任务服务
type TaskService struct {
	db           *gorm.DB
	queueManager *queue.Manager
	logger       *logrus.Logger
}

// NewTaskService 创建任务服务
func NewTaskService(db *gorm.DB, queueManager *queue.Manager, logger *logrus.Logger) *TaskService {
	return &TaskService{
		db:           db,
		queueManager: queueManager,
		logger:       logger,
	}
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(ctx context.Context, req *models.TaskCreateRequest) (*models.Task, error) {
	// 验证模型是否存在
	var model models.Model
	if err := s.db.First(&model, req.ModelID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("model not found")
		}
		return nil, fmt.Errorf("failed to query model: %w", err)
	}

	// 创建任务
	task := &models.Task{
		ModelID:  req.ModelID,
		Type:     req.Type,
		Input:    req.Input,
		Priority: req.Priority,
		Status:   models.TaskStatusPending,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// 将任务加入队列
	if err := s.queueManager.EnqueueTask(ctx, task); err != nil {
		s.logger.WithError(err).Error("Failed to enqueue task")
		// 任务创建成功但入队失败，更新状态
		s.db.Model(task).Update("status", models.TaskStatusFailed)
		s.db.Model(task).Update("error_message", "Failed to enqueue task")
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	// 记录日志
	s.addTaskLog(task.ID, models.LogLevelInfo, "Task created and enqueued", nil)

	s.logger.WithFields(logrus.Fields{
		"task_id":  task.ID,
		"model_id": task.ModelID,
		"type":     task.Type,
		"priority": task.Priority,
	}).Info("Task created")

	return task, nil
}

// GetTask 获取任务详情
func (s *TaskService) GetTask(id uint64) (*models.Task, error) {
	var task models.Task
	err := s.db.Preload("Model").Preload("Logs").First(&task, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

// ListTasks 获取任务列表
func (s *TaskService) ListTasks(req *models.TaskListRequest) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	query := s.db.Model(&models.Task{}).Preload("Model")

	// 过滤条件
	if req.ModelID != nil {
		query = query.Where("model_id = ?", *req.ModelID)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}
	if req.Priority != nil {
		query = query.Where("priority = ?", *req.Priority)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	// 分页和排序
	offset := (req.Page - 1) * req.PageSize
	orderBy := req.OrderBy
	if orderBy == "" {
		orderBy = "created_at"
	}
	order := req.Order
	if order == "" {
		order = "desc"
	}

	err := query.Order(fmt.Sprintf("%s %s", orderBy, order)).
		Limit(req.PageSize).
		Offset(offset).
		Find(&tasks).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, total, nil
}

// UpdateTask 更新任务
func (s *TaskService) UpdateTask(id uint64, req *models.TaskUpdateRequest) (*models.Task, error) {
	var task models.Task
	if err := s.db.First(&task, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	updates := make(map[string]interface{})
	
	if req.Priority != nil {
		updates["priority"] = *req.Priority
		s.addTaskLog(id, models.LogLevelInfo, 
			fmt.Sprintf("Priority updated to %d", *req.Priority), nil)
	}
	
	if req.Status != nil {
		updates["status"] = *req.Status
		s.addTaskLog(id, models.LogLevelInfo, 
			fmt.Sprintf("Status updated to %s", *req.Status), nil)
	}

	if len(updates) > 0 {
		if err := s.db.Model(&task).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update task: %w", err)
		}
	}

	return s.GetTask(id)
}

// CancelTask 取消任务
func (s *TaskService) CancelTask(ctx context.Context, id uint64) error {
	var task models.Task
	if err := s.db.First(&task, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("task not found")
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	// 只有 pending 和 running 状态的任务可以取消
	if task.Status != models.TaskStatusPending && task.Status != models.TaskStatusRunning {
		return fmt.Errorf("task cannot be cancelled in current status: %s", task.Status)
	}

	// 更新状态
	if err := s.db.Model(&task).Updates(map[string]interface{}{
		"status":       models.TaskStatusCancelled,
		"completed_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	// 如果任务在处理中，从处理队列中移除
	if task.Status == models.TaskStatusRunning {
		s.queueManager.CompleteTask(ctx, id)
	}

	s.addTaskLog(id, models.LogLevelInfo, "Task cancelled by user", nil)
	
	s.logger.WithField("task_id", id).Info("Task cancelled")
	
	return nil
}

// RetryTask 重试任务
func (s *TaskService) RetryTask(ctx context.Context, id uint64) error {
	var task models.Task
	if err := s.db.First(&task, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("task not found")
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	// 只有失败的任务可以重试
	if task.Status != models.TaskStatusFailed {
		return fmt.Errorf("task cannot be retried in current status: %s", task.Status)
	}

	// 检查重试次数
	if task.RetryCount >= task.MaxRetries {
		return fmt.Errorf("task has exceeded maximum retry count")
	}

	// 重置任务状态
	updates := map[string]interface{}{
		"status":        models.TaskStatusPending,
		"error_message": nil,
		"started_at":    nil,
		"completed_at":  nil,
		"retry_count":   task.RetryCount + 1,
	}

	if err := s.db.Model(&task).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update task for retry: %w", err)
	}

	// 重新入队
	task.Status = models.TaskStatusPending
	task.RetryCount++
	if err := s.queueManager.EnqueueTask(ctx, &task); err != nil {
		return fmt.Errorf("failed to enqueue retry task: %w", err)
	}

	s.addTaskLog(id, models.LogLevelInfo, 
		fmt.Sprintf("Task retried (attempt %d/%d)", task.RetryCount+1, task.MaxRetries), nil)
	
	s.logger.WithFields(logrus.Fields{
		"task_id":      id,
		"retry_count":  task.RetryCount + 1,
		"max_retries":  task.MaxRetries,
	}).Info("Task retried")
	
	return nil
}

// StartTask 开始执行任务
func (s *TaskService) StartTask(id uint64) error {
	updates := map[string]interface{}{
		"status":     models.TaskStatusRunning,
		"started_at": time.Now(),
	}

	if err := s.db.Model(&models.Task{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to start task: %w", err)
	}

	s.addTaskLog(id, models.LogLevelInfo, "Task execution started", nil)
	return nil
}

// CompleteTask 完成任务
func (s *TaskService) CompleteTask(id uint64, output string) error {
	updates := map[string]interface{}{
		"status":       models.TaskStatusCompleted,
		"output":       output,
		"completed_at": time.Now(),
	}

	if err := s.db.Model(&models.Task{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	s.addTaskLog(id, models.LogLevelInfo, "Task completed successfully", nil)
	return nil
}

// FailTask 任务失败
func (s *TaskService) FailTask(id uint64, errorMsg string) error {
	updates := map[string]interface{}{
		"status":        models.TaskStatusFailed,
		"error_message": errorMsg,
		"completed_at":  time.Now(),
	}

	if err := s.db.Model(&models.Task{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to fail task: %w", err)
	}

	s.addTaskLog(id, models.LogLevelError, "Task failed", map[string]interface{}{
		"error": errorMsg,
	})
	return nil
}

// GetTaskStats 获取任务统计
func (s *TaskService) GetTaskStats() (*models.TaskStats, error) {
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
	var avgMs float64
	s.db.Model(&models.Task{}).
		Select("AVG(TIMESTAMPDIFF(MICROSECOND, started_at, completed_at) / 1000)").
		Where("started_at IS NOT NULL AND completed_at IS NOT NULL").
		Scan(&avgMs)
	stats.AvgProcessingMS = int64(avgMs)

	return &stats, nil
}

// addTaskLog 添加任务日志
func (s *TaskService) addTaskLog(taskID uint64, level models.LogLevel, message string, data models.LogData) {
	log := &models.TaskLog{
		TaskID:  taskID,
		Level:   level,
		Message: message,
		Data:    data,
	}
	
	if err := s.db.Create(log).Error; err != nil {
		s.logger.WithError(err).Error("Failed to create task log")
	}
}
