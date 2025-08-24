package services

import (
	"fmt"

	"llm-scheduler/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ModelService 模型服务
type ModelService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewModelService 创建模型服务
func NewModelService(db *gorm.DB, logger *logrus.Logger) *ModelService {
	return &ModelService{
		db:     db,
		logger: logger,
	}
}

// CreateModel 创建模型
func (s *ModelService) CreateModel(req *models.Model) (*models.Model, error) {
	// 检查模型名称是否已存在
	var existingModel models.Model
	if err := s.db.Where("name = ?", req.Name).First(&existingModel).Error; err == nil {
		return nil, fmt.Errorf("model with name '%s' already exists", req.Name)
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing model: %w", err)
	}

	// 设置默认值
	if req.Status == "" {
		req.Status = models.ModelStatusOffline
	}
	if req.MaxWorkers <= 0 {
		req.MaxWorkers = 1
	}

	// 创建模型
	if err := s.db.Create(req).Error; err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"model_id":   req.ID,
		"model_name": req.Name,
		"model_type": req.Type,
	}).Info("Model created")

	return req, nil
}

// GetModel 获取模型详情
func (s *ModelService) GetModel(id uint64) (*models.Model, error) {
	var model models.Model
	if err := s.db.First(&model, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("model not found")
		}
		return nil, fmt.Errorf("failed to get model: %w", err)
	}
	return &model, nil
}

// GetModelByName 根据名称获取模型
func (s *ModelService) GetModelByName(name string) (*models.Model, error) {
	var model models.Model
	if err := s.db.Where("name = ?", name).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("model not found")
		}
		return nil, fmt.Errorf("failed to get model: %w", err)
	}
	return &model, nil
}

// ListModels 获取模型列表
func (s *ModelService) ListModels(modelType *models.ModelType, status *models.ModelStatus) ([]models.Model, error) {
	var models_list []models.Model
	query := s.db

	if modelType != nil {
		query = query.Where("type = ?", *modelType)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Find(&models_list).Error; err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	return models_list, nil
}

// UpdateModel 更新模型
func (s *ModelService) UpdateModel(id uint64, updates *models.Model) (*models.Model, error) {
	var model models.Model
	if err := s.db.First(&model, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("model not found")
		}
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	// 更新字段
	updateMap := make(map[string]interface{})
	
	if updates.Name != "" && updates.Name != model.Name {
		// 检查新名称是否已存在
		var existingModel models.Model
		if err := s.db.Where("name = ? AND id != ?", updates.Name, id).First(&existingModel).Error; err == nil {
			return nil, fmt.Errorf("model with name '%s' already exists", updates.Name)
		} else if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check existing model: %w", err)
		}
		updateMap["name"] = updates.Name
	}
	
	if updates.Type != "" {
		updateMap["type"] = updates.Type
	}
	
	if updates.Config != nil {
		updateMap["config"] = updates.Config
	}
	
	if updates.Status != "" {
		updateMap["status"] = updates.Status
	}
	
	if updates.MaxWorkers > 0 {
		updateMap["max_workers"] = updates.MaxWorkers
	}

	if len(updateMap) > 0 {
		if err := s.db.Model(&model).Updates(updateMap).Error; err != nil {
			return nil, fmt.Errorf("failed to update model: %w", err)
		}
		
		s.logger.WithFields(logrus.Fields{
			"model_id":   id,
			"model_name": model.Name,
			"updates":    updateMap,
		}).Info("Model updated")
	}

	return s.GetModel(id)
}

// DeleteModel 删除模型
func (s *ModelService) DeleteModel(id uint64) error {
	// 检查是否有正在执行的任务
	var runningTaskCount int64
	if err := s.db.Model(&models.Task{}).
		Where("model_id = ? AND status IN (?)", 
			id, []models.TaskStatus{models.TaskStatusPending, models.TaskStatusRunning}).
		Count(&runningTaskCount).Error; err != nil {
		return fmt.Errorf("failed to check running tasks: %w", err)
	}

	if runningTaskCount > 0 {
		return fmt.Errorf("cannot delete model with %d running/pending tasks", runningTaskCount)
	}

	// 删除模型
	if err := s.db.Delete(&models.Model{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}

	s.logger.WithField("model_id", id).Info("Model deleted")
	return nil
}

// UpdateModelStatus 更新模型状态
func (s *ModelService) UpdateModelStatus(id uint64, status models.ModelStatus) error {
	if err := s.db.Model(&models.Model{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update model status: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"model_id": id,
		"status":   status,
	}).Info("Model status updated")

	return nil
}

// IncrementWorkerCount 增加 Worker 数量
func (s *ModelService) IncrementWorkerCount(id uint64) error {
	if err := s.db.Model(&models.Model{}).
		Where("id = ?", id).
		UpdateColumn("current_workers", gorm.Expr("current_workers + 1")).Error; err != nil {
		return fmt.Errorf("failed to increment worker count: %w", err)
	}
	return nil
}

// DecrementWorkerCount 减少 Worker 数量
func (s *ModelService) DecrementWorkerCount(id uint64) error {
	if err := s.db.Model(&models.Model{}).
		Where("id = ? AND current_workers > 0", id).
		UpdateColumn("current_workers", gorm.Expr("current_workers - 1")).Error; err != nil {
		return fmt.Errorf("failed to decrement worker count: %w", err)
	}
	return nil
}

// IncrementRequestCount 增加请求计数
func (s *ModelService) IncrementRequestCount(id uint64, success bool) error {
	updates := map[string]interface{}{
		"total_requests": gorm.Expr("total_requests + 1"),
	}
	
	if success {
		updates["success_requests"] = gorm.Expr("success_requests + 1")
	}

	if err := s.db.Model(&models.Model{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to increment request count: %w", err)
	}
	return nil
}

// GetAvailableModels 获取可用的模型（在线且有空闲 Worker）
func (s *ModelService) GetAvailableModels() ([]models.Model, error) {
	var models_list []models.Model
	if err := s.db.Where("status = ? AND current_workers < max_workers", 
		models.ModelStatusOnline).Find(&models_list).Error; err != nil {
		return nil, fmt.Errorf("failed to get available models: %w", err)
	}
	return models_list, nil
}

// GetModelStats 获取模型统计信息
func (s *ModelService) GetModelStats() ([]models.ModelStats, error) {
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
	`

	if err := s.db.Raw(query).Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get model stats: %w", err)
	}

	return stats, nil
}
