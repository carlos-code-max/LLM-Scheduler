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

type Worker struct {
	id            string
	modelID       uint64
	queueManager  *queue.Manager
	taskService   *services.TaskService
	modelService  *services.ModelService
	logger        *logrus.Logger
	status        string
	currentTask   *uint64
	startTime     time.Time
	lastHeartbeat time.Time
	ctx           context.Context
	cancel        context.CancelFunc
}

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

func (w *Worker) Start(ctx context.Context) error {
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.logger.WithFields(logrus.Fields{
		"worker_id": w.id,
		"model_id":  w.modelID,
	}).Info("Worker starting")

	go w.heartbeat()

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

func (w *Worker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}

func (w *Worker) processNextTask() error {
	queueItem, err := w.queueManager.DequeueTask(w.ctx, w.modelID)
	if err != nil {
		return fmt.Errorf("failed to dequeue task: %w", err)
	}

	if queueItem == nil {
		time.Sleep(1 * time.Second)
		return nil
	}

	task, err := w.taskService.GetTask(queueItem.TaskID)
	if err != nil {
		w.logger.WithError(err).WithField("task_id", queueItem.TaskID).Error("Failed to get task")
		return err
	}

	return w.executeTask(task)
}

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
		_ = w.taskService.FailTask(task.ID, err.Error())
		_ = w.modelService.IncrementRequestCount(model.ID, false)

		// 从处理队列中移除任务
		_ = w.queueManager.CompleteTask(w.ctx, task.ID)

		return fmt.Errorf("task execution failed: %w", err)
	}

	// 任务成功完成
	if err := w.taskService.CompleteTask(task.ID, output); err != nil {
		w.logger.WithError(err).Error("Failed to mark task as completed")
	}

	_ = w.modelService.IncrementRequestCount(model.ID, true)

	// 从处理队列中移除任务
	_ = w.queueManager.CompleteTask(w.ctx, task.ID)

	w.logger.WithFields(logrus.Fields{
		"worker_id": w.id,
		"task_id":   task.ID,
		"task_type": task.Type,
	}).Info("Task completed successfully")

	return nil
}

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

func (w *Worker) executeTextGeneration(task *models.Task, model *models.Model) (string, error) {
	switch model.Type {
	case models.ModelTypeOpenAI:
		return w.callOpenAIAPI(task, model)
	case models.ModelTypeLocal:
		return w.callLocalAPI(task, model)
	default:
		return "", fmt.Errorf("unsupported model type: %s", model.Type)
	}
}

func (w *Worker) executeTranslation(task *models.Task, model *models.Model) (string, error) {
	time.Sleep(1 * time.Second)
	// 模拟翻译结果
	return fmt.Sprintf("translation result: %s", task.Input), nil
}

func (w *Worker) executeSummarization(task *models.Task, model *models.Model) (string, error) {
	time.Sleep(1 * time.Second)
	// 模拟摘要结果
	return fmt.Sprintf("summarization result: %s", task.Input[:min(50, len(task.Input))]), nil
}

func (w *Worker) executeEmbedding(task *models.Task, model *models.Model) (string, error) {
	time.Sleep(1 * time.Second)
	// 模拟向量化结果
	return "[0.1, 0.2, 0.3, ...]", nil
}

func (w *Worker) executeCustomTask(task *models.Task, model *models.Model) (string, error) {
	time.Sleep(1 * time.Second)
	return fmt.Sprintf("custom task done: %s", task.Input), nil
}

func (w *Worker) callOpenAIAPI(task *models.Task, model *models.Model) (string, error) {
	// 这里应该实现实际的 OpenAI API 调用
	time.Sleep(3 * time.Second)

	apiKey, exists := model.GetConfigValue("api_key")
	if !exists || apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	// 模拟 API 调用结果
	return fmt.Sprintf("OpenAI 响应: 根据输入 '%s' 生成的内容", task.Input), nil
}

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
