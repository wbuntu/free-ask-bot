package daemon

import (
	"context"
	"time"

	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/config"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/leaderelection"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/log"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/queue"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

type Server struct {
	ctx                  context.Context
	cancel               context.CancelFunc
	logger               log.Logger
	enableLeaderElection bool
}

func (s *Server) Setup(ctx context.Context, c *config.Config) error {
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.logger = log.WithField("module", "daemon")
	s.enableLeaderElection = c.General.EnableLeaderElection
	if s.enableLeaderElection {
		// 使用kube-apiserver做选举
		cfg, err := rest.InClusterConfig()
		if err != nil {
			return errors.Wrap(err, "get in cluster config")
		}
		if err := leaderelection.Setup(
			s.ctx,
			cfg,
			s.onStartedLeading,
			s.onStoppedLeading,
			s.onNewLeader,
		); err != nil {
			return errors.Wrap(err, "setup leader election")
		}
	}
	return nil
}

func (s *Server) Serve() error {
	s.logger.Info("start running daemon")
	if s.enableLeaderElection {
		leaderelection.RunOrDie()
	} else {
		s.onStartedLeading(s.ctx)
	}
	return nil
}

func (s *Server) Shutdown() error {
	s.cancel()
	return nil
}

func (s *Server) onStartedLeading(ctx context.Context) {
	// 使用传入的ctx启动工作队列，在选举失败时会自动cancel
	s.logger.WithField("job", "leader_election").Info("start leading")
	// 创建一个延迟对接并启动
	delayQueue := queue.NewDeleyQueue(
		log.S(ctx, s.logger.WithField("job", "cluster_executor")),
		"cluster_executor",
		time.Second*10,
		32,
		clusterExecutorProducer,
		clusterExecutorConsumer,
	)
	go delayQueue.Run()
}

func (s *Server) onStoppedLeading() {
	s.logger.WithField("job", "leader_election").Info("stop leading")
}

func (s *Server) onNewLeader(identity string) {
	s.logger.WithField("job", "leader_election").Infof("new leader selected: %s", identity)
}
