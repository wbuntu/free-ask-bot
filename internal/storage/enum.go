package storage

// K8sType k8s类型
type K8sType string

const (
	K8sTypeK8s K8sType = "k8s"
	K8sTypeK3s K8sType = "k3s"
)

// ClusterStatus 集群状态
type ClusterStatus uint8

// 中间态 0~99
const (
	ClusterStatusCreating  ClusterStatus = iota // 创建中
	ClusterStatusDeleting                       // 删除中
	ClusterStatusUpgrading                      // 升级中
)

// 稳态 >= 100
const (
	ClusterStatusRunning ClusterStatus = 100 + iota // 运行中
	ClusterStatusDeleted                            // 已删除
	ClusterStatusError                              // 异常
)

func (s ClusterStatus) String() string {
	switch s {
	case ClusterStatusCreating:
		return "Creating"
	case ClusterStatusDeleting:
		return "Deleting"
	case ClusterStatusUpgrading:
		return "Upgrading"
	case ClusterStatusRunning:
		return "Running"
	case ClusterStatusDeleted:
		return "Deleted"
	case ClusterStatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

const (
	ClusterActionCreate  = "CreateCluster"
	ClusterActionDelete  = "DeleteCluster"
	ClusterActionUpgrade = "UpgradeCluster"
)

type TaskStatus uint8

const (
	TaskStatusEnqueued TaskStatus = iota // 已入队列
	TaskStatusRetrying                   // 重试中
	TaskStatusSuccess                    // 已成功
	TaskStatusFail                       // 已失败
)

func (s TaskStatus) String() string {
	switch s {
	case TaskStatusEnqueued:
		return "Enqueued"
	case TaskStatusRetrying:
		return "Retrying"
	case TaskStatusSuccess:
		return "Success"
	case TaskStatusFail:
		return "Fail"
	default:
		return "Unknown"
	}
}

type TaskRetryType uint8

const (
	TaskRetryTypeFixed TaskRetryType = iota // 固定间隔
	TaskRetryTypePow                        // 指数递增
)
