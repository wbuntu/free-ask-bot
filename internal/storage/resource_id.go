package storage

import (
	"time"

	"gitbub.com/wbuntu/gin-template/internal/pkg/utils"
)

type ResourceID struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	ResourceID string    `gorm:"not null;uniqueIndex;size:20;comment:'资源ID'"`
	CreatedAt  time.Time `gorm:"comment:'创建时间'"`
}

// createResourceID 通用的资源ID生成方法
func createResourceID(prefix string) (*ResourceID, error) {
	item := &ResourceID{}
	str := utils.GetRandString(10)
	item.ResourceID = prefix + "-" + str
	if err := DB().Create(item).Error; err != nil {
		return nil, handleStorageError(err)
	}
	return item, nil
}

// GenerateClusterResourceID 生成集群资源ID
func GenerateClusterResourceID() (*ResourceID, error) {
	return createResourceID("cluster")
}
