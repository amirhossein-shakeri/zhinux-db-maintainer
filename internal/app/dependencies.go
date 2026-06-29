package app

import (
	"context"
	"fmt"
	"time"

	postgresbackup "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/adapters/persistence/postgres/backup"
	postgresdatabase "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/adapters/persistence/postgres/database"
	postgresrestore "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/adapters/persistence/postgres/restore"
	platformcache "github.com/amirhossein-shakeri/zhinux-platform/cache"
	"github.com/amirhossein-shakeri/zhinux-platform/health"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (a *App) initDependencies(ctx context.Context) error {
	if err := a.initPostgres(ctx); err != nil {
		return err
	}
	if err := a.initCache(); err != nil {
		a.postgresPool.Close()
		return err
	}
	a.initRepositories()
	a.initHealthChecks()
	return nil
}

func (a *App) initPostgres(ctx context.Context) error {
	poolConfig, err := pgxpool.ParseConfig(a.cfg.Runtime.Postgres.DSN)
	if err != nil {
		return fmt.Errorf("parse postgres dsn: %w", err)
	}
	poolConfig.MaxConns = a.cfg.Runtime.Postgres.MaxConns
	poolConfig.MinConns = a.cfg.Runtime.Postgres.MinConns
	poolConfig.MaxConnLifetime = a.cfg.Runtime.Postgres.MaxConnLifetime
	poolConfig.MaxConnIdleTime = a.cfg.Runtime.Postgres.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = a.cfg.Runtime.Postgres.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("open postgres pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return fmt.Errorf("ping postgres: %w", err)
	}

	a.postgresPool = pool
	return nil
}

func (a *App) initCache() error {
	cacheBackend := platformcache.Options{
		Backend: platformcache.BackendInMemory,
		InMemory: platformcache.InMemoryOptions{
			Namespace:       a.cfg.Base.ServiceName,
			DefaultTTL:      5 * time.Minute,
			CleanupInterval: 30 * time.Second,
		},
	}

	if a.cfg.Runtime.Redis.Enabled {
		cacheBackend = platformcache.Options{
			Backend: platformcache.BackendRedis,
			Redis: platformcache.RedisOptions{
				Address:      a.cfg.Runtime.Redis.Address,
				Username:     a.cfg.Runtime.Redis.Username,
				Password:     a.cfg.Runtime.Redis.Password,
				Database:     a.cfg.Runtime.Redis.Database,
				Namespace:    a.cfg.Runtime.Redis.Namespace,
				DefaultTTL:   a.cfg.Runtime.Redis.DefaultTTL,
				DialTimeout:  a.cfg.Runtime.Redis.DialTimeout,
				ReadTimeout:  a.cfg.Runtime.Redis.ReadTimeout,
				WriteTimeout: a.cfg.Runtime.Redis.WriteTimeout,
				PoolSize:     a.cfg.Runtime.Redis.PoolSize,
				MinIdleConns: a.cfg.Runtime.Redis.MinIdleConns,
				MaxRetries:   a.cfg.Runtime.Redis.MaxRetries,
				PingTimeout:  a.cfg.Runtime.Redis.PingTimeout,
			},
		}
	}

	cacheClient, err := platformcache.New(cacheBackend)
	if err != nil {
		return fmt.Errorf("init cache backend %q: %w", cacheBackend.Backend, err)
	}
	a.cache = cacheClient
	return nil
}

func (a *App) initRepositories() {
	a.repositories = Repositories{
		Database:       postgresdatabase.NewDatabaseRepository(a.postgresPool),
		BackupPlan:     postgresbackup.NewBackupPlanRepository(a.postgresPool),
		BackupJob:      postgresbackup.NewBackupJobRepository(a.postgresPool),
		BackupArtifact: postgresbackup.NewBackupArtifactRepository(a.postgresPool),
		RestoreJob:     postgresrestore.NewRestoreJobRepository(a.postgresPool),
	}
}

func (a *App) initHealthChecks() {
	a.health = health.NewRegistry(500 * time.Millisecond)
	a.health.AddLiveness("service", func(context.Context) error { return nil })
	a.health.AddReadiness("postgres", func(ctx context.Context) error {
		return a.postgresPool.Ping(ctx)
	})

	if a.cfg.Runtime.Redis.Enabled {
		// TODO: Replace with an active ping/probe once cache key strategy and failure semantics are finalized.
		a.health.AddReadiness("redis", func(context.Context) error { return nil })
	}
}
