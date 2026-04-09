package app

import (
	"net/http"

	appconfig "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/config"
	outboundports "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/ports/outbound"
	platformcache "github.com/amirhossein-shakeri/zhinux-platform/cache"
	grpcx "github.com/amirhossein-shakeri/zhinux-platform/grpc"
	"github.com/amirhossein-shakeri/zhinux-platform/health"
	"github.com/amirhossein-shakeri/zhinux-platform/logging"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BuildInfo struct {
	Version   string
	Commit    string
	BuildDate string
}

type Repositories struct {
	Database       outboundports.DatabaseRepository
	BackupPlan     outboundports.BackupPlanRepository
	BackupJob      outboundports.BackupJobRepository
	BackupArtifact outboundports.BackupArtifactRepository
	RestoreJob     outboundports.RestoreJobRepository
}

type App struct {
	cfg    appconfig.Config
	build  BuildInfo
	logger logging.Logger

	httpServer *http.Server
	grpcServer *grpcx.Runtime
	health     *health.Registry

	postgresPool *pgxpool.Pool
	cache        platformcache.Cache

	repositories Repositories
	errCh        chan error
}
