package storage

import (
	"database/sql"
	"math"
	"time"

	"gorm.io/datatypes"
)

type Task struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	ResourceID string         `gorm:"not null;index;size:20;comment:'资源ID'"`
	RequestID  string         `gorm:"not null;comment:'请求ID'"`
	Status     TaskStatus     `gorm:"not null;comment:'任务状态 0-已入队列 1-重试中 2-已成功 3-已失败'"`
	Action     string         `gorm:"not null;comment:'任务操作'"`
	RetryType  TaskRetryType  `gorm:"not null;comment:'重试类型 0-固定间隔 1-指数递增'"`
	RetryDelay int            `gorm:"not null;comment:'重试延迟'"`
	RetryCount uint16         `gorm:"not null;comment:'任务重试次数'"`
	RetryLimit uint16         `gorm:"not null;comment:'任务重试次数限制'"`
	RetryAt    sql.NullTime   `gorm:"comment:'重试时间'"`
	Config     datatypes.JSON `gorm:"comment:'任务配置'"`
	CreatedAt  time.Time      `gorm:"comment:'创建时间'"`
	UpdatedAt  time.Time      `gorm:"comment:'更新时间'"`
}

func (t *Task) NextRetryDuration() time.Duration {
	// 超过重试次数或任务已完成，则停止重试
	if t.RetryCount >= t.RetryLimit || t.Status == TaskStatusSuccess {
		return -1
	}
	switch t.RetryType {
	// 固定间隔
	case TaskRetryTypeFixed:
		return time.Duration(t.RetryDelay) * time.Second
	// 指数递增
	case TaskRetryTypePow:
		return time.Duration(math.Pow(2, float64(t.RetryCount))) * time.Duration(t.RetryDelay) * time.Second
	}
	// 无间隔
	return 0
}

func GetNextTaskByResourceID(resourceID string) (*Task, error) {
	items := []Task{}
	if err := DB().
		Where("resource_id = ? and status < ?", resourceID, TaskStatusSuccess).
		Order("id asc").
		Limit(1).
		Find(&items).
		Error; err != nil {
		return nil, handleStorageError(err)
	}
	if len(items) == 0 {
		return nil, ErrDoesNotExist
	}
	return &items[0], nil
}

func UpdateTaskStatus(item *Task) error {
	if err := DB().Model(item).Updates(map[string]interface{}{
		"retry_count": item.RetryCount,
		"retry_at":    item.RetryAt,
		"status":      item.Status,
	}).Error; err != nil {
		return handleStorageError(err)
	}
	return nil
}

type TaskLog struct {
	ID      uint64    `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	TaskID  uint64    `gorm:"not null;index;comment:'任务ID'"`
	Reason  string    `gorm:"not null;comment:'原因'"`
	Message string    `gorm:"not null;comment:'信息'"`
	StartAt time.Time `gorm:"not null;comment:'起始时间'"`
	EndAt   time.Time `gorm:"not null;comment:'结束时间'"`
}

func CreateTaskLog(item *TaskLog) error {
	if err := DB().Create(item).Error; err != nil {
		return handleStorageError(err)
	}
	return nil
}

func ListTaskLog(taskID uint64, offset int, limit int) ([]TaskLog, error) {
	var items []TaskLog
	if err := DB().
		Offset(offset).
		Limit(limit).
		Where("task_id = ?", taskID).
		Order("id desc").
		Find(&items).Error; err != nil {
		return nil, handleStorageError(err)
	}
	return items, nil
}
