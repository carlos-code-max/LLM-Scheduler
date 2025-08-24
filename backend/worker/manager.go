package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"llm-scheduler/config"
	"llm-scheduler/models"
	"llm-scheduler/queue"
	"llm-scheduler/services"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Manager Worker 管理器
type Manager struct {
	config       *config.Config
	db           *gorm.DB
	queueManager *queue.Manager
	taskService  *services.TaskService
	modelService *services.ModelService
	logger       *logrus.Logger
	workers      map[string]*Worker
	workersMutex sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewManager 创建 Worker 管理器
func NewManager(
	cfg *config.Config,
	db *gorm.DB,
	queueManager *queue.Manager,
	taskService *services.TaskService,
	modelService *services.ModelService,
	logger *logrus.Logger,
) *Manager {
	return &Manager{
		config:       cfg,
		db:           db,
		queueManager: queueManager,
		taskService:  taskService,
		modelService: modelService,
		logger:       logger,
		workers:      make(map[string]*Worker),
	}
}

// Start 启动 Worker 管理器
func (m *Manager) Start(ctx context.Context) error {
	m.ctx, m.cancel = context.WithCancel(ctx)
	
	m.logger.Info("Starting worker manager")

	// 启动延迟任务处理协程
	go m.processDelayedTasks()
	
	// 启动清理卡住任务的协程
	go m.cleanupStuckTasks()
	
	// 启动 Worker 监控协程
	go m.monitorWorkers()

	// 启动默认 Worker 池
	if err := m.startDefaultWorkers(); err != nil {
		return fmt.Errorf("failed to start default workers: %w", err)
	}

	// 等待上下文取消
	<-m.ctx.Done()
	
	m.logger.Info("Stopping worker manager")
	m.stopAllWorkers()
	
	return nil
}

// Stop 停止 Worker 管理器
func (m *Manager) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
}

// startDefaultWorkers 启动默认 Worker
func (m *Manager) startDefaultWorkers() error {
	// 获取所有在线模型
	models, err := m.modelService.GetAvailableModels()
	if err != nil {
		return fmt.Errorf("failed to get available models: %w", err)
	}

	for _, model := range models {
		// 为每个模型启动 Worker
		workerCount := model.MaxWorkers
		if workerCount <= 0 {
			workerCount = 1
		}
		
		for i := 0; i < workerCount; i++ {
			if err := m.startWorker(&model); err != nil {
				m.logger.WithError(err).WithFields(logrus.Fields{
					"model_id":   model.ID,
					"model_name": model.Name,
				}).Error("Failed to start worker")
			}
		}
	}

	return nil
}

// startWorker 启动单个 Worker
func (m *Manager) startWorker(model *models.Model) error {
	workerID := fmt.Sprintf("worker-%d-%d", model.ID, time.Now().UnixNano())
	
	worker := NewWorker(
		workerID,
		model.ID,
		m.queueManager,
		m.taskService,
		m.modelService,
		m.logger,
	)
	
	m.workersMutex.Lock()
	m.workers[workerID] = worker
	m.workersMutex.Unlock()

	// 在新协程中启动 Worker
	go func() {
		if err := worker.Start(m.ctx); err != nil {
			m.logger.WithError(err).WithField("worker_id", workerID).Error("Worker stopped with error")
		}
		
		// Worker 停止后从管理器中移除
		m.workersMutex.Lock()
		delete(m.workers, workerID)
		m.workersMutex.Unlock()
		
		// 减少模型的当前 Worker 数量
		m.modelService.DecrementWorkerCount(model.ID)
	}()

	// 增加模型的当前 Worker 数量
	m.modelService.IncrementWorkerCount(model.ID)
	
	m.logger.WithFields(logrus.Fields{
		"worker_id":  workerID,
		"model_id":   model.ID,
		"model_name": model.Name,
	}).Info("Worker started")

	return nil
}

// stopAllWorkers 停止所有 Worker
func (m *Manager) stopAllWorkers() {
	m.workersMutex.Lock()
	defer m.workersMutex.Unlock()

	for _, worker := range m.workers {
		worker.Stop()
	}
	
	// 等待所有 Worker 停止
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for len(m.workers) > 0 {
		select {
		case <-timeout:
			m.logger.Warn("Timeout waiting for workers to stop")
			return
		case <-ticker.C:
			// 继续等待
		}
	}
	
	m.logger.Info("All workers stopped")
}

// processDelayedTasks 处理延迟任务
func (m *Manager) processDelayedTasks() {
	ticker := time.NewTicker(10 * time.Second) // 每10秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if err := m.queueManager.ProcessDelayedTasks(m.ctx); err != nil {
				m.logger.WithError(err).Error("Failed to process delayed tasks")
			}
		}
	}
}

// cleanupStuckTasks 清理卡住的任务
func (m *Manager) cleanupStuckTasks() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if err := m.queueManager.CleanupStuckTasks(m.ctx); err != nil {
				m.logger.WithError(err).Error("Failed to cleanup stuck tasks")
			}
		}
	}
}

// monitorWorkers 监控 Worker 状态
func (m *Manager) monitorWorkers() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkWorkerHealth()
		}
	}
}

// checkWorkerHealth 检查 Worker 健康状态
func (m *Manager) checkWorkerHealth() {
	m.workersMutex.RLock()
	workerCount := len(m.workers)
	m.workersMutex.RUnlock()

	// 获取在线模型
	models, err := m.modelService.GetAvailableModels()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get available models for health check")
		return
	}

	expectedWorkers := 0
	for _, model := range models {
		expectedWorkers += model.MaxWorkers
	}

	if workerCount < expectedWorkers {
		m.logger.WithFields(logrus.Fields{
			"current_workers":  workerCount,
			"expected_workers": expectedWorkers,
		}).Warn("Worker count is below expected")
		
		// 尝试启动缺失的 Worker
		// 这里可以添加自动恢复逻辑
	}
}

// GetWorkerStatus 获取 Worker 状态
func (m *Manager) GetWorkerStatus() []models.WorkerStatus {
	m.workersMutex.RLock()
	defer m.workersMutex.RUnlock()

	var status []models.WorkerStatus
	for _, worker := range m.workers {
		status = append(status, worker.GetStatus())
	}

	return status
}

// GetWorkerCount 获取 Worker 数量
func (m *Manager) GetWorkerCount() int {
	m.workersMutex.RLock()
	defer m.workersMutex.RUnlock()
	return len(m.workers)
}
