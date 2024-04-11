package api

import (
	"context"
	"net/http"
	"time"

	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/config"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/log"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	srv    *http.Server
	tlsSrv *http.Server
	tlsCrt string
	tlsKey string
	ctx    context.Context
	cancel context.CancelFunc
	logger log.Logger
}

func (s *Server) Setup(ctx context.Context, c *config.Config) error {
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.logger = log.WithField("module", "api")
	if err := s.setupEngine(ctx, c); err != nil {
		return errors.Wrap(err, "setup engine")
	}
	return nil
}

func (s *Server) Serve() error {
	if s.tlsSrv != nil {
		go func() {
			log.WithField("module", "api").Infof("start listening tls: %s", s.tlsSrv.Addr)
			if err := s.tlsSrv.ListenAndServeTLS(s.tlsCrt, s.tlsKey); err != nil && err != http.ErrServerClosed {
				log.Fatalf("tlsSrv.ListenAndServeTLS: %s", err)
			}
		}()
	}
	log.WithField("module", "api").Infof("start listening: %s", s.srv.Addr)
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("srv.ListenAndServe: %s", err)
	}
	return nil
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(s.ctx, time.Second*5)
	defer cancel()
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := s.srv.Shutdown(gCtx); err != nil {
			return errors.Wrap(err, "shutdown srv")
		}
		return nil
	})
	if s.tlsSrv != nil {
		g.Go(func() error {
			if err := s.tlsSrv.Shutdown(gCtx); err != nil {
				return errors.Wrap(err, "shutdown tls srv")
			}
			return nil
		})
	}
	g.Wait()
	return nil
}
