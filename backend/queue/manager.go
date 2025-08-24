package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"llm-scheduler/config"
	"llm-scheduler/models"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// Manager 队列管理器
type Manager struct {
	client *redis.Client
	config *config.Config
	logger *logrus.Logger
}

// QueueItem 队列项目
type QueueItem struct {
	TaskID    uint64    `json:"task_id"`
	ModelID   uint64    `json:"model_id"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

// NewManager 创建队列管理器
func NewManager(client *redis.Client, cfg *config.Config, logger *logrus.Logger) *Manager {
	return &Manager{
		client: client,
		config: cfg,
		logger: logger,
	}
}

// EnqueueTask 将任务加入队列
func (m *Manager) EnqueueTask(ctx context.Context, task *models.Task) error {
	queueKey := m.getQueueKey(models.TaskPriority(task.Priority))
	
	item := QueueItem{
		TaskID:    task.ID,
		ModelID:   task.ModelID,
		Priority:  int(task.Priority),
		CreatedAt: task.CreatedAt,
	}
	
	itemBytes, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal queue item: %w", err)
	}

	// 使用 Redis List 作为队列，LPUSH 保证 FIFO
	if err := m.client.LPush(ctx, queueKey, itemBytes).Err(); err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	m.logger.WithFields(logrus.Fields{
		"task_id":  task.ID,
		"model_id": task.ModelID,
		"priority": task.Priority,
		"queue":    queueKey,
	}).Info("Task enqueued")

	return nil
}

// DequeueTask 从队列中获取任务
func (m *Manager) DequeueTask(ctx context.Context, modelID uint64) (*QueueItem, error) {
	// 按优先级顺序检查队列
	queues := []string{
		m.config.Queue.HighPriorityQueue,
		m.config.Queue.MediumPriorityQueue,
		m.config.Queue.LowPriorityQueue,
	}

	for _, queueKey := range queues {
		// 使用 BRPOP 阻塞式获取任务，超时时间设为 1 秒
		result, err := m.client.BRPop(ctx, 1*time.Second, queueKey).Result()
		if err != nil {
			if err == redis.Nil {
				// 队列为空，继续检查下一个队列
				continue
			}
			return nil, fmt.Errorf("failed to dequeue from %s: %w", queueKey, err)
		}

		if len(result) != 2 {
			continue
		}

		var item QueueItem
		if err := json.Unmarshal([]byte(result[1]), &item); err != nil {
			m.logger.WithError(err).Error("Failed to unmarshal queue item")
			continue
		}

		// 检查是否是指定模型的任务
		if modelID != 0 && item.ModelID != modelID {
			// 如果不是指定模型的任务，将任务放回队列末尾
			if err := m.client.LPush(ctx, queueKey, result[1]).Err(); err != nil {
				m.logger.WithError(err).Error("Failed to requeue task")
			}
			continue
		}

		// 将任务移到处理中队列
		if err := m.moveToProcessing(ctx, &item); err != nil {
			m.logger.WithError(err).Error("Failed to move task to processing queue")
			// 将任务放回原队列
			m.client.LPush(ctx, queueKey, result[1])
			return nil, err
		}

		m.logger.WithFields(logrus.Fields{
			"task_id":  item.TaskID,
			"model_id": item.ModelID,
			"priority": item.Priority,
			"queue":    queueKey,
		}).Info("Task dequeued")

		return &item, nil
	}

	// 所有队列都为空
	return nil, nil
}

// moveToProcessing 将任务移到处理中队列
func (m *Manager) moveToProcessing(ctx context.Context, item *QueueItem) error {
	itemBytes, err := json.Marshal(item)
	if err != nil {
		return err
	}

	// 使用有序集合存储处理中的任务，score 为开始处理时间
	score := float64(time.Now().Unix())
	return m.client.ZAdd(ctx, m.config.Queue.ProcessingQueue, &redis.Z{
		Score:  score,
		Member: itemBytes,
	}).Err()
}

// CompleteTask 完成任务，从处理中队列移除
func (m *Manager) CompleteTask(ctx context.Context, taskID uint64) error {
	// 从处理中队列中移除任务
	processingKey := m.config.Queue.ProcessingQueue
	
	// 获取所有处理中的任务
	results, err := m.client.ZRange(ctx, processingKey, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, result := range results {
		var item QueueItem
		if err := json.Unmarshal([]byte(result), &item); err != nil {
			continue
		}

		if item.TaskID == taskID {
			return m.client.ZRem(ctx, processingKey, result).Err()
		}
	}

	return nil
}

// RequeueTask 重新将任务加入队列（用于重试失败的任务）
func (m *Manager) RequeueTask(ctx context.Context, item *QueueItem, delay time.Duration) error {
	// 如果有延迟，使用延迟队列
	if delay > 0 {
		return m.enqueueDelayed(ctx, item, delay)
	}

	// 否则直接加入对应优先级队列
	queueKey := m.getQueueKey(models.TaskPriority(item.Priority))
	
	itemBytes, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return m.client.LPush(ctx, queueKey, itemBytes).Err()
}

// enqueueDelayed 将任务加入延迟队列
func (m *Manager) enqueueDelayed(ctx context.Context, item *QueueItem, delay time.Duration) error {
	itemBytes, err := json.Marshal(item)
	if err != nil {
		return err
	}

	// 使用有序集合存储延迟任务，score 为执行时间
	executeAt := time.Now().Add(delay)
	score := float64(executeAt.Unix())

	return m.client.ZAdd(ctx, m.config.Queue.DelayedQueue, &redis.Z{
		Score:  score,
		Member: itemBytes,
	}).Err()
}

// ProcessDelayedTasks 处理延迟任务，将到期任务移到正常队列
func (m *Manager) ProcessDelayedTasks(ctx context.Context) error {
	delayedKey := m.config.Queue.DelayedQueue
	now := float64(time.Now().Unix())

	// 获取所有到期的延迟任务
	results, err := m.client.ZRangeByScore(ctx, delayedKey, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%f", now),
	}).Result()
	if err != nil {
		return err
	}

	for _, result := range results {
		var item QueueItem
		if err := json.Unmarshal([]byte(result), &item); err != nil {
			m.logger.WithError(err).Error("Failed to unmarshal delayed task")
			continue
		}

		// 将任务移到正常队列
		queueKey := m.getQueueKey(models.TaskPriority(item.Priority))
		if err := m.client.LPush(ctx, queueKey, result).Err(); err != nil {
			m.logger.WithError(err).Error("Failed to move delayed task to queue")
			continue
		}

		// 从延迟队列中移除
		if err := m.client.ZRem(ctx, delayedKey, result).Err(); err != nil {
			m.logger.WithError(err).Error("Failed to remove task from delayed queue")
		}

		m.logger.WithField("task_id", item.TaskID).Info("Delayed task moved to queue")
	}

	return nil
}

// CleanupStuckTasks 清理卡住的任务
func (m *Manager) CleanupStuckTasks(ctx context.Context) error {
	processingKey := m.config.Queue.ProcessingQueue
	timeout := m.config.Queue.TaskTimeout

	// 获取超时的处理中任务
	cutoff := float64(time.Now().Add(-timeout).Unix())
	results, err := m.client.ZRangeByScore(ctx, processingKey, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%f", cutoff),
	}).Result()
	if err != nil {
		return err
	}

	for _, result := range results {
		var item QueueItem
		if err := json.Unmarshal([]byte(result), &item); err != nil {
			continue
		}

		// 将超时任务重新加入队列或标记为失败
		m.logger.WithField("task_id", item.TaskID).Warn("Found stuck task, requeueing")
		
		// 重新加入延迟队列，等待重试
		if err := m.enqueueDelayed(ctx, &item, m.config.Queue.RetryDelay); err != nil {
			m.logger.WithError(err).Error("Failed to requeue stuck task")
		}

		// 从处理中队列移除
		m.client.ZRem(ctx, processingKey, result)
	}

	return nil
}

// GetQueueStatus 获取队列状态
func (m *Manager) GetQueueStatus(ctx context.Context) (*models.QueueStatus, error) {
	status := &models.QueueStatus{}

	// 获取各队列长度
	highCount, _ := m.client.LLen(ctx, m.config.Queue.HighPriorityQueue).Result()
	mediumCount, _ := m.client.LLen(ctx, m.config.Queue.MediumPriorityQueue).Result()
	lowCount, _ := m.client.LLen(ctx, m.config.Queue.LowPriorityQueue).Result()
	processingCount, _ := m.client.ZCard(ctx, m.config.Queue.ProcessingQueue).Result()
	delayedCount, _ := m.client.ZCard(ctx, m.config.Queue.DelayedQueue).Result()

	status.HighPriorityCount = highCount
	status.MediumPriorityCount = mediumCount
	status.LowPriorityCount = lowCount
	status.ProcessingCount = processingCount
	status.DelayedCount = delayedCount
	status.TotalCount = highCount + mediumCount + lowCount + processingCount + delayedCount

	return status, nil
}

// getQueueKey 根据优先级获取队列键名
func (m *Manager) getQueueKey(priority models.TaskPriority) string {
	switch priority {
	case models.TaskPriorityHigh:
		return m.config.Queue.HighPriorityQueue
	case models.TaskPriorityMedium:
		return m.config.Queue.MediumPriorityQueue
	case models.TaskPriorityLow:
		return m.config.Queue.LowPriorityQueue
	default:
		return m.config.Queue.MediumPriorityQueue
	}
}
