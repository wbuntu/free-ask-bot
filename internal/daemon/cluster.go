package daemon

import (
	"context"
	"time"

	"gitbub.com/wbuntu/gin-template/internal/pkg/log"
	"gitbub.com/wbuntu/gin-template/internal/pkg/queue"
	"gitbub.com/wbuntu/gin-template/internal/pkg/utils"
	"gitbub.com/wbuntu/gin-template/internal/storage"
	"github.com/pkg/errors"
)

func clusterExecutorProducer(ctx context.Context) ([]string, error) {
	resourceIDs, err := storage.ListRunnableClusterResoureID()
	if err != nil {
		return nil, errors.Wrap(err, "list runnable clusters")
	}
	return resourceIDs, nil
}

func clusterExecutorConsumer(ctx context.Context, resourceID string) (time.Duration, error) {
	// 获取集群
	cluster, err := storage.GetClusterByResourceID(resourceID)
	if err != nil && err != storage.ErrDoesNotExist {
		return queue.NextDurationNone, errors.Wrap(err, "get cluster")
	}
	if cluster == nil {
		return queue.NextDurationNone, nil
	}
	// 获取任务
	task, err := storage.GetNextTaskByResourceID(resourceID)
	if err != nil && err != storage.ErrDoesNotExist {
		return queue.NextDurationNone, errors.Wrap(err, "get next task")
	}
	// 任务处理
	if task != nil {
		// 任务存在，执行任务，出错时重试
		logger := log.G(ctx).WithFields(log.Fields{
			"taskID":    task.ID,
			"clusterID": task.ResourceID,
			"requestID": task.RequestID,
			"action":    task.Action,
		})
		if err := handleClusterTask(log.S(ctx, logger), logger, cluster, task); err != nil {
			logger.Errorf("handle cluster task: %s", err)
			return task.NextRetryDuration(), nil
		}
	} else {
		// 任务不存在，执行同步
		logger := log.G(ctx).WithFields(log.Fields{
			"clusterID": cluster.ResourceID,
			"action":    "sync",
			"requestID": utils.UUID(),
		})
		if err := handleClusterSynchronization(log.S(ctx, logger), logger, cluster); err != nil {
			logger.Errorf("handle cluster sync: %s", err)
		}
	}
	return queue.NextDurationNone, nil
}

func handleClusterTask(ctx context.Context, logger log.Logger, cluster *storage.Cluster, task *storage.Task) error {
	return nil
}

func handleClusterSynchronization(ctx context.Context, logger log.Logger, cluster *storage.Cluster) error {
	return nil
}
