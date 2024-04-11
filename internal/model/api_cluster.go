package model

import (
	"gitbub.com/wbuntu/free-ask-bot/internal/storage"
)

type CreateClusterReq struct {
	BaseRequest
	Name        string          `json:"name" example:"imortal-cluster-name"`                               // 名称，支持 1～127 位字符，必须以字母或中文开头，可以包含字母、数字、下划线（_）、中划线（-）、点(.)
	Description string          `json:"description" example:"amazing-cluster-description"`                 // 描述，支持 0~255 位字符
	Type        storage.K8sType `json:"k8sType" example:"k8s" binding:"required,oneof=k8s k3s"`            // 类型：k8s、k3s
	Version     string          `json:"version" example:"1.22.5" binding:"required"`                       // 版本
	Runtime     string          `json:"runtime" example:"cri-o" binding:"required,oneof=cri-o containerd"` // 容器运行时
}

type CreateClusterResp struct {
	BaseResponse
	Data string `json:"data" example:"cluster-sedqqz7kavbh"` // 集群资源ID
}

type ListClusterReq struct {
	BaseRequest
	PageNo      int    `json:"pageNo" form:"pageNo" binding:"gte=1"`     // 分页页码
	PageSize    int    `json:"pageSize" form:"pageSize" binding:"gte=1"` // 分页大小
	FilterKey   string `json:"filterKey" form:"filterKey"`               // 过滤字段
	FilterValue string `json:"filterValue" form:"filterValue"`           // 过滤字段的值
}

type ListClusterResp struct {
	BaseResponse
	Data       []ClusterSummary `json:"data"`                     // 集群概要列表
	TotalCount int              `json:"totalCount" example:"100"` // 集群总数
}

type ClusterSummary struct {
	Name        string          `json:"name" example:"imortal-cluster-name"`               // 名称
	Description string          `json:"description" example:"amazing-cluster-description"` // 描述
	ResourceID  string          `json:"resourceID" example:"cluster-sedqqz7ka"`            // 集群ID
	Type        storage.K8sType `json:"k8sType" example:"k8s"`                             // 类型：k8s、k3s
	Version     string          `json:"version" example:"1.22.5"`                          // 版本
	Runtime     string          `json:"runtime" example:"cri-o"`                           // 容器运行时
	Status      string          `json:"status" exmaple:"Creating"`                         // 集群状态
	CreateTime  string          `json:"createTime" example:"2006-01-02 15:04:05"`          // 创建时间
}

type GetClusterReq struct {
	BaseRequest
}

type GetClusterResp struct {
	BaseResponse
	Data *ClusterDetail `json:"data"` // 集群详情
}

type ClusterDetail struct {
	Name        string          `json:"name" example:"imortal-cluster-name"`               // 名称
	Description string          `json:"description" example:"amazing-cluster-description"` // 描述
	ResourceID  string          `json:"resourceID" example:"cluster-sedqqz7ka"`            // 集群ID
	Type        storage.K8sType `json:"k8sType" example:"k8s"`                             // 类型：k8s、k3s
	Version     string          `json:"version" example:"1.22.5"`                          // 版本
	Runtime     string          `json:"runtime" example:"cri-o"`                           // 容器运行时
	Status      string          `json:"status" exmaple:"Creating"`                         // 集群状态
	CreateTime  string          `json:"createTime" example:"2006-01-02 15:04:05"`          // 创建时间
}
