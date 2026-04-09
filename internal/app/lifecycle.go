package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/amirhossein-shakeri/zhinux-platform/logging"
	"github.com/amirhossein-shakeri/zhinux-platform/shutdown"
	gogrpc "google.golang.org/grpc"
)

func (a *App) Start() error {
	if a.httpServer != nil {
		go func() {
			a.logger.Info("http server started", logging.KV("addr", a.cfg.Base.HTTPListenAddr))
			err := a.httpServer.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				a.errCh <- fmt.Errorf("http server failed: %w", err)
			}
		}()
	}

	if a.grpcServer != nil {
		if err := a.grpcServer.Start(); err != nil {
			return fmt.Errorf("start grpc server: %w", err)
		}

		go func() {
			a.logger.Info("grpc server started", logging.KV("addr", a.grpcServer.Address()))
			err := a.grpcServer.Wait()
			if err != nil && !errors.Is(err, gogrpc.ErrServerStopped) {
				a.errCh <- fmt.Errorf("grpc server failed: %w", err)
			}
		}()
	}

	return nil
}

func (a *App) Wait(ctx context.Context) error {
	select {
	case err := <-a.errCh:
		if err == nil {
			return nil
		}
		return err
	case <-ctx.Done():
		return nil
	}
}

func (a *App) Shutdown() error {
	return shutdown.Run(a.logger,
		shutdown.Hook{
			Name: "postgres-pool",
			Stop: func(context.Context) error {
				if a.postgresPool != nil {
					a.postgresPool.Close()
				}
				return nil
			},
		},
		shutdown.Hook{
			Name: "cache",
			Stop: func(context.Context) error {
				if a.cache != nil {
					return a.cache.Close()
				}
				return nil
			},
		},
		shutdown.Hook{
			Name:    "grpc-server",
			Timeout: a.cfg.Base.ShutdownTimeout,
			Stop: func(ctx context.Context) error {
				if a.grpcServer == nil {
					return nil
				}
				return a.grpcServer.GracefulStop(ctx, a.cfg.Runtime.GRPC.DrainTimeout)
			},
		},
		shutdown.Hook{
			Name:    "http-server",
			Timeout: a.cfg.Base.ShutdownTimeout,
			Stop: func(ctx context.Context) error {
				if a.httpServer == nil {
					return nil
				}
				return a.httpServer.Shutdown(ctx)
			},
		},
	)
}
