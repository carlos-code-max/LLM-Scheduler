package models

import (
	"time"

	"gorm.io/gorm"
)

// TaskStatus 任务状态枚举
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// TaskPriority 任务优先级枚举
type TaskPriority int

const (
	TaskPriorityLow    TaskPriority = 1
	TaskPriorityMedium TaskPriority = 2
	TaskPriorityHigh   TaskPriority = 3
)

// Task 任务表结构
type Task struct {
	ID           uint64       `json:"id" gorm:"primaryKey;autoIncrement"`
	ModelID      uint64       `json:"model_id" gorm:"not null;index:idx_model_status"`
	Type         string       `json:"type" gorm:"type:varchar(50);not null;index"`
	Input        string       `json:"input" gorm:"type:text;not null"`
	Output       *string      `json:"output" gorm:"type:text"`
	Status       TaskStatus   `json:"status" gorm:"type:enum('pending','running','completed','failed','cancelled');default:pending;index:idx_status_priority"`
	Priority     TaskPriority `json:"priority" gorm:"type:tinyint;default:1;index:idx_status_priority"`
	RetryCount   int          `json:"retry_count" gorm:"default:0"`
	MaxRetries   int          `json:"max_retries" gorm:"default:3"`
	ErrorMessage *string      `json:"error_message" gorm:"type:text"`
	StartedAt    *time.Time   `json:"started_at"`
	CompletedAt  *time.Time   `json:"completed_at"`
	CreatedAt    time.Time    `json:"created_at" gorm:"index:idx_created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`

	// 关联关系
	Model *Model    `json:"model,omitempty" gorm:"foreignKey:ModelID"`
	Logs  []TaskLog `json:"logs,omitempty" gorm:"foreignKey:TaskID"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}

// GetProcessingTimeMS 获取处理时间（毫秒）
func (t *Task) GetProcessingTimeMS() int64 {
	if t.StartedAt == nil || t.CompletedAt == nil {
		return 0
	}
	return t.CompletedAt.Sub(*t.StartedAt).Milliseconds()
}

// CanRetry 检查是否可以重试
func (t *Task) CanRetry() bool {
	return t.Status == TaskStatusFailed && t.RetryCount < t.MaxRetries
}

// IsCompleted 检查任务是否已完成
func (t *Task) IsCompleted() bool {
	return t.Status == TaskStatusCompleted || 
		   t.Status == TaskStatusFailed || 
		   t.Status == TaskStatusCancelled
}

// GetPriorityString 获取优先级字符串表示
func (t *Task) GetPriorityString() string {
	switch t.Priority {
	case TaskPriorityHigh:
		return "high"
	case TaskPriorityMedium:
		return "medium"
	case TaskPriorityLow:
		return "low"
	default:
		return "unknown"
	}
}

// BeforeCreate GORM 钩子：创建前
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.Status == "" {
		t.Status = TaskStatusPending
	}
	if t.Priority == 0 {
		t.Priority = TaskPriorityMedium
	}
	return nil
}

// BeforeUpdate GORM 钩子：更新前
func (t *Task) BeforeUpdate(tx *gorm.DB) error {
	// 状态变更时自动设置时间戳
	if t.Status == TaskStatusRunning && t.StartedAt == nil {
		now := time.Now()
		t.StartedAt = &now
	}
	if (t.Status == TaskStatusCompleted || t.Status == TaskStatusFailed || t.Status == TaskStatusCancelled) && t.CompletedAt == nil {
		now := time.Now()
		t.CompletedAt = &now
	}
	return nil
}

// TaskCreateRequest 创建任务请求结构
type TaskCreateRequest struct {
	ModelID  uint64       `json:"model_id" binding:"required"`
	Type     string       `json:"type" binding:"required"`
	Input    string       `json:"input" binding:"required"`
	Priority TaskPriority `json:"priority"`
}

// TaskUpdateRequest 更新任务请求结构
type TaskUpdateRequest struct {
	Priority *TaskPriority `json:"priority"`
	Status   *TaskStatus   `json:"status"`
}

// TaskListRequest 任务列表请求结构
type TaskListRequest struct {
	ModelID  *uint64     `form:"model_id"`
	Status   *TaskStatus `form:"status"`
	Type     *string     `form:"type"`
	Priority *TaskPriority `form:"priority"`
	Page     int         `form:"page,default=1"`
	PageSize int         `form:"page_size,default=20"`
	OrderBy  string      `form:"order_by,default=created_at"`
	Order    string      `form:"order,default=desc"`
}

// TaskStats 任务统计信息
type TaskStats struct {
	TotalTasks       int64   `json:"total_tasks"`
	PendingTasks     int64   `json:"pending_tasks"`
	RunningTasks     int64   `json:"running_tasks"`
	CompletedTasks   int64   `json:"completed_tasks"`
	FailedTasks      int64   `json:"failed_tasks"`
	CancelledTasks   int64   `json:"cancelled_tasks"`
	SuccessRate      float64 `json:"success_rate"`
	AvgProcessingMS  int64   `json:"avg_processing_ms"`
}
