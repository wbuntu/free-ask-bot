package cluster

import (
	"os"
	"strconv"
	"strings"
	"time"

	"gitbub.com/wbuntu/free-ask-bot/internal/model"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/log"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/utils"
	"gitbub.com/wbuntu/free-ask-bot/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CreateClusterCtrl struct {
	model.BaseController
	request  model.CreateClusterReq
	response model.CreateClusterResp
}

func (ctrl *CreateClusterCtrl) Input() model.Request {
	return &ctrl.request
}

func (ctrl *CreateClusterCtrl) Output() model.Response {
	return &ctrl.response
}

// @Summary     创建集群
// @Description 根据配置参数自动创建集群
// @Tags        Cluster
// @Param       CreateClusterReq body     model.CreateClusterReq  true "请求"
// @Response    200              {object} model.CreateClusterResp "响应"
// @Router      /clusters [post]
func (ctrl *CreateClusterCtrl) Serve(c *gin.Context) {
	// 生成集群资源ID
	resourceID, err := storage.GenerateClusterResourceID()
	if err != nil {
		ctrl.Logger.Errorf("generate resourceID: %s", err)
		ctrl.response.Update(model.CodeInternalError, "generate resourceID")
		return
	}
	req := &ctrl.request
	cluster := &storage.Cluster{
		Name:        req.Name,
		Description: req.Description,
		ResourceID:  resourceID.ResourceID,
		TenantID:    req.TenantID,
		Type:        req.Type,
		Version:     req.Version,
		Runtime:     req.Runtime,
		Status:      storage.ClusterStatusCreating,
	}
	// 超过150分钟未Ready判定为异常
	limit := uint16(150)
	timtoutLimit := os.Getenv("TIMEOUT_LIMIT")
	if len(timtoutLimit) > 0 {
		t, _ := strconv.Atoi(timtoutLimit)
		limit = uint16(t)
	}
	task := &storage.Task{
		ResourceID: cluster.ResourceID,
		RequestID:  req.RequestID,
		Status:     storage.TaskStatusEnqueued,
		RetryType:  storage.TaskRetryTypeFixed,
		RetryDelay: 10,
		RetryCount: 0,
		RetryLimit: limit,
		Action:     storage.ClusterActionCreate,
	}
	// 在事务中创建集群、提交任务
	if err := storage.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(cluster).Error; err != nil {
			return errors.Wrap(err, "create cluster")
		}
		if err := tx.Create(task).Error; err != nil {
			return errors.Wrap(err, "create task")
		}
		return nil
	}); err != nil {
		ctrl.Logger.Errorf("transaction: %s", err)
		ctrl.response.Update(model.CodeInternalError, "commit request")
		return
	}
	ctrl.Logger.WithFields(log.Fields{
		"clusterID": task.ResourceID,
		"taskID":    task.ID,
	}).Info("create cluster task committed")
	// 返回响应
	ctrl.response.Data = resourceID.ResourceID
}

type DeleteClusterCtrl struct {
	model.BaseController
	request  model.BaseRequest
	response model.BaseResponse
}

func (ctrl *DeleteClusterCtrl) Input() model.Request {
	return &ctrl.request
}

func (ctrl *DeleteClusterCtrl) Output() model.Response {
	return &ctrl.response
}

// @Summary     删除集群
// @Description 根据集群ID自动删除集群
// @Tags        Cluster
// @Param       clusterId path     string             true "集群资源ID" extensions(x-example=cluster-sedqqz7ka)
// @Response    200       {object} model.BaseResponse "响应"
// @Router      /clusters/{clusterId} [delete]
func (ctrl *DeleteClusterCtrl) Serve(c *gin.Context) {
	// 获取资源ID
	clusterID := c.Param("clusterId")
	// 获取集群
	cluster, err := storage.GetClusterByResourceID(clusterID)
	if err != nil {
		ctrl.Logger.WithField("clusterID", clusterID).Errorf("get cluster: %s", err)
		if err == storage.ErrDoesNotExist {
			ctrl.response.Update(model.CodeNotExists, "cluster not found")
		} else {
			ctrl.response.Update(model.CodeInternalError, "get cluster")
		}
		return
	}
	// 检查状态，稳态才可以提交任务
	if cluster.Status < storage.ClusterStatusRunning {
		ctrl.Logger.WithField("clusterID", clusterID).Error("cluster status pending")
		ctrl.response.Update(model.CodeForbidOperate, "cluster status pending")
		return
	}
	// 超过30分钟未删除判定为异常
	task := &storage.Task{
		ResourceID: cluster.ResourceID,
		RequestID:  ctrl.request.RequestID,
		Status:     storage.TaskStatusEnqueued,
		RetryType:  storage.TaskRetryTypeFixed,
		RetryDelay: 10,
		RetryCount: 0,
		RetryLimit: 10,
		Action:     storage.ClusterActionDelete,
	}
	// 在事务中更改集群状态、提交任务
	cluster.Status = storage.ClusterStatusDeleting
	if err := storage.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Updates(cluster).Error; err != nil {
			return errors.Wrap(err, "update cluster")
		}
		if err := tx.Create(task).Error; err != nil {
			return errors.Wrap(err, "create task")
		}
		return nil
	}); err != nil {
		ctrl.Logger.Errorf("transaction: %s", err)
		ctrl.response.Update(model.CodeInternalError, "commit request")
		return
	}
	ctrl.Logger.WithFields(log.Fields{
		"clusterID": task.ResourceID,
		"taskID":    task.ID,
	}).Info("delete cluster task committed")
}

type GetClusterCtrl struct {
	model.BaseController
	request  model.GetClusterReq
	response model.GetClusterResp
}

func (ctrl *GetClusterCtrl) Input() model.Request {
	return &ctrl.request
}

func (ctrl *GetClusterCtrl) Output() model.Response {
	return &ctrl.response
}

// @Summary     集群详情
// @Description 根据集群ID获取集群详细信息
// @Tags        Cluster
// @Param       clusterId path     string               true "集群资源ID" extensions(x-example=cluster-sedqqz7ka)
// @Response    200       {object} model.GetClusterResp "响应"
// @Router      /clusters/{clusterId} [get]
func (ctrl *GetClusterCtrl) Serve(c *gin.Context) {
	// 获取资源ID
	resourceID := c.Param("clusterId")
	// 获取集群
	cluster, err := storage.GetClusterByResourceID(resourceID)
	if err != nil {
		ctrl.Logger.WithField("clusterID", resourceID).Errorf("get cluster: %s", err)
		if err == storage.ErrDoesNotExist {
			ctrl.response.Update(model.CodeNotExists, "cluster not found")
		} else {
			ctrl.response.Update(model.CodeInternalError, "get cluster")
		}
		return
	}
	ctrl.response.Data = &model.ClusterDetail{
		Name:        cluster.Name,
		Description: cluster.Description,
		ResourceID:  cluster.ResourceID,
		Type:        cluster.Type,
		Version:     cluster.Version,
		Runtime:     cluster.Runtime,
		Status:      cluster.Status.String(),
		CreateTime:  utils.FormatTime(cluster.CreatedAt),
	}
}

type ListClusterCtrl struct {
	model.BaseController
	request  model.ListClusterReq
	response model.ListClusterResp
}

func (ctrl *ListClusterCtrl) Input() model.Request {
	return &ctrl.request
}

func (ctrl *ListClusterCtrl) Output() model.Response {
	return &ctrl.response
}

// @Summary     集群列表
// @Description 分页获取集群列表
// @Tags        Cluster
// @Param       pageNo      query    int                   true  "分页号，默认为1"   extensions(x-example=1)
// @Param       pageSize    query    int                   true  "分页大小，默认为10" extensions(x-example=10)
// @Param       filterKey   query    string                false "查询条件，默认为空"
// @Param       filterValue query    string                false "查询值，默认为空"
// @Response    200         {object} model.ListClusterResp "响应"
// @Router      /clusters [get]
func (ctrl *ListClusterCtrl) Serve(c *gin.Context) {
	req := ctrl.request
	filter, err := buildClusterFilter(req.FilterKey, req.FilterValue)
	if err != nil {
		ctrl.Logger.Errorf("build cluster filter: %s", err)
		ctrl.response.Update(model.CodeParamError, "invalid filterKey or filterValue")
		return
	}

	items, err := storage.ListCluster(req.TenantID, (req.PageNo-1)*req.PageSize, req.PageSize, filter)
	if err != nil {
		ctrl.Logger.Errorf("list cluster: %s", err)
		ctrl.response.Update(model.CodeInternalError, "list cluster")
		return
	}
	count, err := storage.CountCluster(req.TenantID, filter)
	if err != nil {
		ctrl.Logger.Errorf("count cluster: %s", err)
		ctrl.response.Update(model.CodeInternalError, "count cluster")
		return
	}
	ctrl.response.Data = make([]model.ClusterSummary, 0)
	for _, cluster := range items {
		ctrl.response.Data = append(ctrl.response.Data, model.ClusterSummary{
			Name:        cluster.Name,
			Description: cluster.Description,
			ResourceID:  cluster.ResourceID,
			Type:        cluster.Type,
			Version:     cluster.Version,
			Runtime:     cluster.Runtime,
			Status:      cluster.Status.String(),
			CreateTime:  utils.FormatTime(cluster.CreatedAt),
		})
	}
	ctrl.response.TotalCount = int(count)
}

func buildClusterFilter(filterKey string, filterValue string) (*storage.Filter, error) {
	if len(filterKey) == 0 {
		return nil, nil
	}
	if len(filterKey) > 0 && len(filterValue) == 0 {
		return nil, errors.New("empty filterValue")
	}
	var filter *storage.Filter
	switch filterKey {
	case "name":
		// 模糊查询
		filter = &storage.Filter{
			Query: "name LIKE ?",
			Args:  []interface{}{"%" + filterValue + "%"},
		}
	case "resourceID":
		// 文本匹配
		filter = &storage.Filter{
			Query: "resource_id = ?",
			Args:  []interface{}{filterValue},
		}
	case "type":
		// 多选
		items := strings.Split(filterValue, ",")
		if len(items) == 0 {
			return nil, errors.New("invalid type")
		}
		filter = &storage.Filter{
			Query: "type IN ?",
			Args:  []interface{}{},
		}
		filter.Args = append(filter.Args, items)
	case "version":
		// 多选
		items := strings.Split(filterValue, ",")
		if len(items) == 0 {
			return nil, errors.New("invalid version")
		}
		filter = &storage.Filter{
			Query: "version IN ?",
			Args:  []interface{}{},
		}
		filter.Args = append(filter.Args, items)
	case "status":
		// 多选
		items := strings.Split(filterValue, ",")
		if len(items) == 0 {
			return nil, errors.Errorf("invalid status: %s", filterValue)
		}
		for i := range items {
			if _, err := strconv.Atoi(items[i]); err != nil {
				return nil, errors.Errorf("invalid status: %s", filterValue)
			}
		}
		filter = &storage.Filter{
			Query: "status IN ?",
			Args:  []interface{}{},
		}
		filter.Args = append(filter.Args, items)
	case "createTime":
		// 多选
		items := strings.Split(filterValue, ",")
		if len(items) != 2 {
			return nil, errors.New("invalid createTime")
		}
		start, err := strconv.Atoi(items[0])
		if err != nil {
			return nil, errors.New("invalid createTime: startTime")
		}
		end, err := strconv.Atoi(items[1])
		if err != nil {
			return nil, errors.New("invalid createTime: endTime")
		}
		filter = &storage.Filter{
			Query: "created_at BETWEEN ? AND ?",
			Args: []interface{}{
				time.Unix(int64(start), 0),
				time.Unix(int64(end), 0),
			},
		}
	case "tenantID":
		// 文本匹配
		filter = &storage.Filter{
			Query: "tenant_id = ?",
			Args:  []interface{}{filterValue},
		}
	default:
		return nil, errors.New("unsupported filterKey")
	}
	return filter, nil
}
