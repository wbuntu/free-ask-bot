package leaderelection

import (
	"context"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

// 使用单例暴露package级别的func
var sharedCtrl *Controller

func Setup(ctx context.Context, cfg *rest.Config, onStartedLeading func(context.Context), onStoppedLeading func(), onNewLeader func(string)) error {
	var err error
	sharedCtrl, err = NewLeaderElectionCtrl(ctx, cfg, onStartedLeading, onStoppedLeading, onNewLeader)
	if err != nil {
		return errors.Wrap(err, "setup sharedCtrl")
	}
	return nil
}

func Leading() bool {
	return sharedCtrl.Leading()
}

func GetLeaderIdentity() string {
	return sharedCtrl.GetLeaderIdentity()
}

func RunOrDie() {
	sharedCtrl.RunOrDie()
}

func NewLeaderElectionCtrl(ctx context.Context, cfg *rest.Config, onStartedLeading func(context.Context), onStoppedLeading func(), onNewLeader func(string)) (*Controller, error) {
	identify, err := getInClusterIdentity()
	if err != nil {
		return nil, errors.Wrap(err, "getInClusterIdentity")
	}
	ns, err := getInClusterNamespace()
	if err != nil {
		return nil, errors.Wrap(err, "getInClusterNamespace")
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "get k8s client")
	}
	ctrl := &Controller{
		ctx:              ctx,
		identify:         identify,
		ns:               ns,
		leaseName:        "0ae6b428.wbuntu.com",
		cfg:              cfg,
		client:           client,
		onStartedLeading: onStartedLeading,
		onStoppedLeading: onStoppedLeading,
		onNewLeader:      onNewLeader,
	}
	return ctrl, nil
}

type Controller struct {
	ctx              context.Context
	identify         string
	ns               string
	leaseName        string
	leaderIdentity   string
	cfg              *rest.Config
	client           *kubernetes.Clientset
	onStartedLeading func(context.Context)
	onStoppedLeading func()
	onNewLeader      func(string)
}

func (c *Controller) Leading() bool {
	return c.identify == c.leaderIdentity
}

func (c *Controller) GetLeaderIdentity() string {
	return c.leaderIdentity
}

func (c *Controller) RunOrDie() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// we use the Lease lock type since edits to Leases are less common
			// and fewer objects in the cluster watch "all Leases".
			lock := &resourcelock.LeaseLock{
				LeaseMeta: metav1.ObjectMeta{
					Name:      c.leaseName,
					Namespace: c.ns,
				},
				Client: c.client.CoordinationV1(),
				LockConfig: resourcelock.ResourceLockConfig{
					Identity: c.identify,
				},
			}
			// start the leader election code loop
			leaderelection.RunOrDie(c.ctx, leaderelection.LeaderElectionConfig{
				Lock: lock,
				// IMPORTANT: you MUST ensure that any code you have that
				// is protected by the lease must terminate **before**
				// you call cancel. Otherwise, you could have a background
				// loop still running and another process could
				// get elected before your background loop finished, violating
				// the stated goal of the lease.
				ReleaseOnCancel: true,
				LeaseDuration:   15 * time.Second,
				RenewDeadline:   10 * time.Second,
				RetryPeriod:     2 * time.Second,
				Callbacks: leaderelection.LeaderCallbacks{
					OnStartedLeading: c.onStartedLeading,
					OnStoppedLeading: c.onStoppedLeading,
					OnNewLeader: func(identity string) {
						c.leaderIdentity = identity
						c.onNewLeader(identity)
					},
				},
			})
			time.Sleep(time.Millisecond * 200)
		}
	}
}
