package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// LogLevel 日志级别枚举
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogData 日志附加数据，存储为 JSON
type LogData map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (ld *LogData) Scan(value interface{}) error {
	if value == nil {
		*ld = make(LogData)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal LogData: %v", value)
	}

	return json.Unmarshal(bytes, ld)
}

// Value 实现 driver.Valuer 接口
func (ld LogData) Value() (driver.Value, error) {
	if ld == nil {
		return nil, nil
	}
	return json.Marshal(ld)
}

// TaskLog 任务日志表结构
type TaskLog struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID    uint64    `json:"task_id" gorm:"not null;index:idx_task_created"`
	Level     LogLevel  `json:"level" gorm:"type:enum('info','warn','error','debug');default:info;index:idx_level_created"`
	Message   string    `json:"message" gorm:"type:text;not null"`
	Data      LogData   `json:"data" gorm:"type:json"`
	CreatedAt time.Time `json:"created_at" gorm:"index:idx_task_created,idx_level_created"`

	// 关联关系
	Task *Task `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

// TableName 指定表名
func (TaskLog) TableName() string {
	return "task_logs"
}

// SetData 设置附加数据
func (tl *TaskLog) SetData(key string, value interface{}) {
	if tl.Data == nil {
		tl.Data = make(LogData)
	}
	tl.Data[key] = value
}

// GetData 获取附加数据
func (tl *TaskLog) GetData(key string) (interface{}, bool) {
	if tl.Data == nil {
		return nil, false
	}
	value, exists := tl.Data[key]
	return value, exists
}

// SystemStats 系统统计表结构
type SystemStats struct {
	ID                   uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	StatDate             time.Time `json:"stat_date" gorm:"type:date;uniqueIndex;not null"`
	TotalTasks           int       `json:"total_tasks" gorm:"default:0"`
	CompletedTasks       int       `json:"completed_tasks" gorm:"default:0"`
	FailedTasks          int       `json:"failed_tasks" gorm:"default:0"`
	AvgProcessingTimeMs  int       `json:"avg_processing_time_ms" gorm:"default:0"`
	QueueLength          int       `json:"queue_length" gorm:"default:0"`
	ActiveModels         int       `json:"active_models" gorm:"default:0"`
	CreatedAt            time.Time `json:"created_at"`
}

// TableName 指定表名
func (SystemStats) TableName() string {
	return "system_stats"
}

// QueueStatus 队列状态信息
type QueueStatus struct {
	HighPriorityCount   int64 `json:"high_priority_count"`
	MediumPriorityCount int64 `json:"medium_priority_count"`
	LowPriorityCount    int64 `json:"low_priority_count"`
	ProcessingCount     int64 `json:"processing_count"`
	DelayedCount        int64 `json:"delayed_count"`
	TotalCount          int64 `json:"total_count"`
}

// WorkerStatus Worker 状态信息
type WorkerStatus struct {
	WorkerID      string    `json:"worker_id"`
	ModelID       uint64    `json:"model_id"`
	ModelName     string    `json:"model_name"`
	Status        string    `json:"status"`
	CurrentTaskID *uint64   `json:"current_task_id"`
	StartTime     time.Time `json:"start_time"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

// DashboardStats Dashboard 统计数据
type DashboardStats struct {
	TaskStats     TaskStats       `json:"task_stats"`
	ModelStats    []ModelStats    `json:"model_stats"`
	QueueStatus   QueueStatus     `json:"queue_status"`
	WorkerStatus  []WorkerStatus  `json:"worker_status"`
	SystemStats   SystemStats     `json:"system_stats"`
	RecentTasks   []Task          `json:"recent_tasks"`
}
