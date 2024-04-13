/*
Copyright © 2022 wbuntu
*/
package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gitbub.com/wbuntu/free-ask-bot/internal/api"
	"gitbub.com/wbuntu/free-ask-bot/internal/daemon"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/bot"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/config"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/llm"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/log"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/utils"
	"gitbub.com/wbuntu/free-ask-bot/internal/storage"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func run(cmd *cobra.Command, args []string) error {
	//  1. 配置日志模块
	if err := log.Setup(&config.C); err != nil {
		return errors.Wrap(err, "setup log")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 2. 配置任务
	logger := log.WithField("module", "main")
	t := &task{
		ctx:    ctx,
		config: &config.C,
		logger: logger,
	}
	logger.WithField("config", t.config).Debug("print config")
	if err := t.setup(); err != nil {
		logger.Fatal(err)
	}
	// 3. 启动服务
	t.serve()
	// 4. 等待退出
	sigChan := make(chan os.Signal, 1)
	exitChan := make(chan struct{})
	// 监听 SIGINT 和 SIGTERM
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// 首次收到信号，打印日志
	logger.Info("signal: ", <-sigChan, " signal received")
	// 触发graceful shutdown
	go func() {
		t.shutdown()
		cancel()
		exitChan <- struct{}{}
	}()
	// 异步等待强制终止或主动退出
	select {
	case <-exitChan:
	case s := <-sigChan:
		logger.Info("signal:", s, " signal received: stopping immediately")
	}
	logger.Warn("free-ask-bot stopped")
	return nil
}

// task
type task struct {
	ctx           context.Context
	config        *config.Config
	serveFuncs    []func() error
	shutdownFuncs []func() error
	logger        log.Logger
}

func (t *task) setup() error {
	// 顺序执行初始化任务
	taskFuncs := []func(*task) error{
		printStartupLog,
		setupStorage,
		migrateStorage,
		setupDependency,
		setupDaemon,
		setupAPI,
	}
	for _, fn := range taskFuncs {
		if err := fn(t); err != nil {
			return errors.Errorf("%s: %s", utils.GetFunctionName(fn), err)
		}
	}
	return nil
}

func (t *task) serve() error {
	for i := range t.serveFuncs {
		go func(i int) {
			if err := t.serveFuncs[i](); err != nil {
				t.logger.Fatalf("executing serve func")
			}
		}(i)
	}
	t.logger.Info("start serving")
	return nil
}

func (t *task) shutdown() {
	t.logger.Warn("stop serving")
	for i := range t.shutdownFuncs {
		go func(i int) {
			if err := t.shutdownFuncs[i](); err != nil {
				t.logger.Errorf("executing shutdown func")
			}
		}(i)
	}
}

// printStartupLog 打印版本
func printStartupLog(t *task) error {
	t.logger.WithField("version", t.config.Version).Info("starting free-ask-bot")
	return nil
}

// setupStorage 初始化数据存储
func setupStorage(t *task) error {
	if err := storage.Setup(
		t.ctx,
		t.config,
	); err != nil {
		return errors.Wrap(err, "setup storage")
	}
	t.logger.Info("setup storage success")
	return nil
}

// migrateStorage 数据迁移
func migrateStorage(t *task) error {
	if err := storage.Migrate(
		t.ctx,
		t.config,
	); err != nil {
		return errors.Wrap(err, "migrate storage")
	}
	t.logger.Info("migrate storage success")
	return nil
}

// setupDependency 初始化依赖项
func setupDependency(t *task) error {
	if err := llm.Setup(t.ctx, t.config); err != nil {
		return errors.Wrap(err, "setup llm")
	}
	t.serveFuncs = append(t.serveFuncs, llm.Serve)
	t.shutdownFuncs = append(t.shutdownFuncs, llm.Shutdown)
	t.logger.Info("setup llm success")
	if err := bot.Setup(t.ctx, t.config); err != nil {
		return errors.Wrap(err, "setup bot")
	}
	t.serveFuncs = append(t.serveFuncs, bot.Serve)
	t.shutdownFuncs = append(t.shutdownFuncs, bot.Shutdown)
	t.logger.Info("setup bot success")
	return nil
}

// setupDaemon 初始化 Daemon 模块
func setupDaemon(t *task) error {
	srv := &daemon.Server{}
	if err := srv.Setup(
		t.ctx,
		t.config,
	); err != nil {
		return errors.Wrap(err, "setup daemon")
	}
	t.serveFuncs = append(t.serveFuncs, srv.Serve)
	t.shutdownFuncs = append(t.shutdownFuncs, srv.Shutdown)
	t.logger.Info("setup daemon server success")
	return nil
}

// setupAPI 初始化 API 模块
func setupAPI(t *task) error {
	srv := &api.Server{}
	if err := srv.Setup(
		t.ctx,
		t.config,
	); err != nil {
		return errors.Wrap(err, "setup api")
	}
	t.serveFuncs = append(t.serveFuncs, srv.Serve)
	t.shutdownFuncs = append(t.shutdownFuncs, srv.Shutdown)
	t.logger.Info("setup api server success")
	return nil
}
