package postgres

import (
	outbound_ports "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/ports/outbound"
	"github.com/amirhossein-shakeri/zhinux-platform/cache"
)

type databaseCacheRepositoryImpl struct {
	repo  outbound_ports.DatabaseRepository
	cache cache.Cache
}

func NewDatabaseCacheRepository(
	repo outbound_ports.DatabaseRepository, cache cache.Cache,
) outbound_ports.DatabaseRepository {
	return &databaseCacheRepositoryImpl{
		repo:  repo,
		cache: cache,
	}
}
