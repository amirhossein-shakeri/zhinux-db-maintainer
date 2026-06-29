package app

import (
	"context"
	"fmt"
	"strings"

	appconfig "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/config"
	"github.com/amirhossein-shakeri/zhinux-platform/logging"
)

func New(ctx context.Context, build BuildInfo) (*App, error) {
	cfg, err := appconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	logger, err := logging.NewLogger(logging.LoggerOptions{
		Level:       cfg.Base.LogLevel,
		Service:     cfg.Base.ServiceName,
		Backend:     logging.BackendSlog,
		Development: cfg.Base.Environment != "production",
	})
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	build = NormalizeBuildInfo(build)
	app := &App{
		cfg:    cfg,
		build:  build,
		logger: logger,
		errCh:  make(chan error, 2),
	}

	if err := app.initDependencies(ctx); err != nil {
		return nil, err
	}

	if err := app.initServers(); err != nil {
		return nil, err
	}

	app.logger.Info(
		"bootstrap initialized",
		logging.KV("version", build.Version),
		logging.KV("commit", build.Commit),
		logging.KV("build_date", build.BuildDate),
		logging.KV("environment", cfg.Base.Environment),
		logging.KV("grpc_enabled", cfg.Runtime.GRPC.Enabled),
		logging.KV("grpc_listen_addr", cfg.Base.GRPCListenAddr),
		logging.KV("http_enabled", cfg.Runtime.HTTP.Enabled),
		logging.KV("http_listen_addr", cfg.Base.HTTPListenAddr),
	)

	return app, nil
}

func NormalizeBuildInfo(build BuildInfo) BuildInfo {
	if strings.TrimSpace(build.Version) == "" {
		build.Version = "dev"
	}

	if strings.TrimSpace(build.Commit) == "" {
		build.Commit = "unknown"
	}

	if strings.TrimSpace(build.BuildDate) == "" {
		build.BuildDate = "unknown"
	}

	return build
}
