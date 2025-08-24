package worker

import (
	"context"
	"fmt"
	"time"

	"llm-scheduler/models"
	"llm-scheduler/queue"
	"llm-scheduler/services"

	"github.com/sirupsen/logrus"
)

// Worker 任务工作器
type Worker struct {
	id           string
	modelID      uint64
	queueManager *queue.Manager
	taskService  *services.TaskService
	modelService *services.ModelService
	logger       *logrus.Logger
	status       string
	currentTask  *uint64
	startTime    time.Time
	lastHeartbeat time.Time
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewWorker 创建新的 Worker
func NewWorker(
	id string,
	modelID uint64,
	queueManager *queue.Manager,
	taskService *services.TaskService,
	modelService *services.ModelService,
	logger *logrus.Logger,
) *Worker {
	return &Worker{
		id:           id,
		modelID:      modelID,
		queueManager: queueManager,
		taskService:  taskService,
		modelService: modelService,
		logger:       logger,
		status:       "idle",
		startTime:    time.Now(),
	}
}

// Start 启动 Worker
func (w *Worker) Start(ctx context.Context) error {
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.logger.WithFields(logrus.Fields{
		"worker_id": w.id,
		"model_id":  w.modelID,
	}).Info("Worker starting")

	// 心跳协程
	go w.heartbeat()

	// 主工作循环
	for {
		select {
		case <-w.ctx.Done():
			w.logger.WithField("worker_id", w.id).Info("Worker stopped")
			return nil
		default:
			if err := w.processNextTask(); err != nil {
				w.logger.WithError(err).WithField("worker_id", w.id).Error("Error processing task")
				// 短暂休息后继续
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// Stop 停止 Worker
func (w *Worker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}

// processNextTask 处理下一个任务
func (w *Worker) processNextTask() error {
	// 从队列获取任务
	queueItem, err := w.queueManager.DequeueTask(w.ctx, w.modelID)
	if err != nil {
		return fmt.Errorf("failed to dequeue task: %w", err)
	}

	if queueItem == nil {
		// 队列为空，休息一下
		time.Sleep(1 * time.Second)
		return nil
	}

	// 获取任务详情
	task, err := w.taskService.GetTask(queueItem.TaskID)
	if err != nil {
		w.logger.WithError(err).WithField("task_id", queueItem.TaskID).Error("Failed to get task")
		return err
	}

	// 处理任务
	return w.executeTask(task)
}

// executeTask 执行任务
func (w *Worker) executeTask(task *models.Task) error {
	w.status = "busy"
	w.currentTask = &task.ID
	defer func() {
		w.status = "idle"
		w.currentTask = nil
	}()

	w.logger.WithFields(logrus.Fields{
		"worker_id": w.id,
		"task_id":   task.ID,
		"task_type": task.Type,
	}).Info("Executing task")

	// 标记任务开始执行
	if err := w.taskService.StartTask(task.ID); err != nil {
		w.logger.WithError(err).Error("Failed to mark task as started")
		return err
	}

	// 获取模型信息
	model, err := w.modelService.GetModel(task.ModelID)
	if err != nil {
		w.taskService.FailTask(task.ID, "Failed to get model information")
		return fmt.Errorf("failed to get model: %w", err)
	}

	// 执行具体任务
	output, err := w.executeTaskByType(task, model)
	if err != nil {
		// 任务失败
		w.taskService.FailTask(task.ID, err.Error())
		w.modelService.IncrementRequestCount(model.ID, false)
		
		// 从处理队列中移除任务
		w.queueManager.CompleteTask(w.ctx, task.ID)
		
		return fmt.Errorf("task execution failed: %w", err)
	}

	// 任务成功完成
	if err := w.taskService.CompleteTask(task.ID, output); err != nil {
		w.logger.WithError(err).Error("Failed to mark task as completed")
	}
	
	w.modelService.IncrementRequestCount(model.ID, true)
	
	// 从处理队列中移除任务
	w.queueManager.CompleteTask(w.ctx, task.ID)

	w.logger.WithFields(logrus.Fields{
		"worker_id": w.id,
		"task_id":   task.ID,
		"task_type": task.Type,
	}).Info("Task completed successfully")

	return nil
}

// executeTaskByType 根据任务类型执行具体逻辑
func (w *Worker) executeTaskByType(task *models.Task, model *models.Model) (string, error) {
	switch task.Type {
	case "text-generation":
		return w.executeTextGeneration(task, model)
	case "translation":
		return w.executeTranslation(task, model)
	case "summarization":
		return w.executeSummarization(task, model)
	case "embedding":
		return w.executeEmbedding(task, model)
	default:
		return w.executeCustomTask(task, model)
	}
}

// executeTextGeneration 执行文本生成任务
func (w *Worker) executeTextGeneration(task *models.Task, model *models.Model) (string, error) {
	// 模拟处理时间
	time.Sleep(2 * time.Second)
	
	// 这里应该调用实际的 LLM API
	switch model.Type {
	case models.ModelTypeOpenAI:
		return w.callOpenAIAPI(task, model)
	case models.ModelTypeLocal:
		return w.callLocalAPI(task, model)
	default:
		return "", fmt.Errorf("unsupported model type: %s", model.Type)
	}
}

// executeTranslation 执行翻译任务
func (w *Worker) executeTranslation(task *models.Task, model *models.Model) (string, error) {
	time.Sleep(1 * time.Second)
	// 模拟翻译结果
	return fmt.Sprintf("翻译结果: %s", task.Input), nil
}

// executeSummarization 执行摘要任务
func (w *Worker) executeSummarization(task *models.Task, model *models.Model) (string, error) {
	time.Sleep(1 * time.Second)
	// 模拟摘要结果
	return fmt.Sprintf("摘要: %s", task.Input[:min(50, len(task.Input))]), nil
}

// executeEmbedding 执行向量化任务
func (w *Worker) executeEmbedding(task *models.Task, model *models.Model) (string, error) {
	time.Sleep(1 * time.Second)
	// 模拟向量化结果
	return "[0.1, 0.2, 0.3, ...]", nil
}

// executeCustomTask 执行自定义任务
func (w *Worker) executeCustomTask(task *models.Task, model *models.Model) (string, error) {
	time.Sleep(1 * time.Second)
	return fmt.Sprintf("自定义任务完成: %s", task.Input), nil
}

// callOpenAIAPI 调用 OpenAI API
func (w *Worker) callOpenAIAPI(task *models.Task, model *models.Model) (string, error) {
	// 这里应该实现实际的 OpenAI API 调用
	// 为了演示，我们模拟一个响应
	time.Sleep(3 * time.Second)
	
	apiKey, exists := model.GetConfigValue("api_key")
	if !exists || apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}
	
	// 模拟 API 调用结果
	return fmt.Sprintf("OpenAI 响应: 根据输入 '%s' 生成的内容", task.Input), nil
}

// callLocalAPI 调用本地 API
func (w *Worker) callLocalAPI(task *models.Task, model *models.Model) (string, error) {
	// 这里应该实现实际的本地模型 API 调用
	time.Sleep(5 * time.Second)
	
	host, _ := model.GetConfigValue("host")
	port, _ := model.GetConfigValue("port")
	
	if host == nil || port == nil {
		return "", fmt.Errorf("local model host/port not configured")
	}
	
	// 模拟本地 API 调用结果
	return fmt.Sprintf("本地模型响应: 基于输入 '%s' 的处理结果", task.Input), nil
}

// heartbeat 心跳
func (w *Worker) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.lastHeartbeat = time.Now()
			w.logger.WithField("worker_id", w.id).Debug("Worker heartbeat")
		}
	}
}

// GetStatus 获取 Worker 状态
func (w *Worker) GetStatus() models.WorkerStatus {
	return models.WorkerStatus{
		WorkerID:      w.id,
		ModelID:       w.modelID,
		Status:        w.status,
		CurrentTaskID: w.currentTask,
		StartTime:     w.startTime,
		LastHeartbeat: w.lastHeartbeat,
	}
}

// min 辅助函数，返回两个数的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
