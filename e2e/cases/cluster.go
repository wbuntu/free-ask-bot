package cases

import (
	"gitbub.com/wbuntu/free-ask-bot/internal/model"
)

// StandardCreateClusterRequest 标准集群创建请求
var StandardCreateClusterRequest = &model.CreateClusterReq{
	Name:        "e2e-cluster-name",
	Description: "e2e-cluster-description",
	Type:        "k8s",
	Version:     "1.22.5",
	Runtime:     "cri-o",
}
