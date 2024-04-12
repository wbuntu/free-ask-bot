package storage

import (
	"time"
)

type Cluster struct {
	ID          uint64        `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	Name        string        `gorm:"not null;comment:'名称'"`
	Description string        `gorm:"comment:'集群描述信息'"`
	ResourceID  string        `gorm:"not null;uniqueIndex;size:20;comment:'资源ID'"`
	TenantID    string        `gorm:"not null;index;size:20;comment:'租户ID'"`
	Type        K8sType       `gorm:"not null;comment:'k8s类型 k8s k3s'"`
	Version     string        `gorm:"not null;comment:'版本'"`
	Runtime     string        `gorm:"not null;comment:'运行时'"`
	Status      ClusterStatus `gorm:"not null;index;comment:'状态'"`
	CreatedAt   time.Time     `gorm:"comment:'创建时间'"`
	UpdatedAt   time.Time     `gorm:"comment:'更新时间'"`
}

// GetClusterByResourceID 根据资源ID获取集群
func GetClusterByResourceID(resourceID string) (*Cluster, error) {
	item := &Cluster{}
	if err := DB().
		Where("resource_id = ? and status != ?", resourceID, ClusterStatusDeleted).
		Take(item).
		Error; err != nil {
		return nil, handleStorageError(err)
	}
	return item, nil
}

// ListRunnableClusterResoureID 列出未删除集群ID
func ListRunnableClusterResoureID() ([]string, error) {
	var resourceIDs []string
	if err := DB().
		Raw("SELECT resource_id FROM clusters WHERE status != ? order by id desc", ClusterStatusDeleted).
		Scan(&resourceIDs).
		Error; err != nil {
		return nil, handleStorageError(err)
	}
	return resourceIDs, nil
}

// ListCluster 分页列出集群
func ListCluster(tenantID string, offset int, limit int, filter *Filter) ([]Cluster, error) {
	items := []Cluster{}
	var err error
	if filter != nil {
		err = DB().
			Limit(limit).
			Offset(offset).
			Where("tenant_id = ? and status != ?", tenantID, ClusterStatusDeleted).
			Where(filter.Query, filter.Args...).
			Order("id desc").
			Find(&items).Error
	} else {
		err = DB().
			Limit(limit).
			Offset(offset).
			Where("tenant_id = ? and status != ?", tenantID, ClusterStatusDeleted).
			Order("id desc").
			Find(&items).Error
	}
	if err != nil {
		return nil, handleStorageError(err)
	}
	return items, nil
}

// CountCluster 计算租户集群总数
func CountCluster(tenantID string, filter *Filter) (int, error) {
	var count int64
	var err error
	if filter != nil {
		err = DB().
			Model(&Cluster{}).
			Where("tenant_id = ? and status != ?", tenantID, ClusterStatusDeleted).
			Where(filter.Query, filter.Args...).
			Count(&count).Error
	} else {
		err = DB().
			Model(&Cluster{}).
			Where("tenant_id = ? and status != ?", tenantID, ClusterStatusDeleted).
			Count(&count).Error
	}
	if err != nil {
		return 0, handleStorageError(err)
	}
	return int(count), nil
}
