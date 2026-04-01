package database_events

import (
	"time"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/database"
)

// DatabaseRegistered
type DatabaseRegistered struct {
	DatabaseID string
	Title      string
	Typ        database.DatabaseType
	CreatedAt  time.Time
}

func (e DatabaseRegistered) Name() string {
	return "database.completed"
}

func (e DatabaseRegistered) Payload() string {
	panic("not implemented")
}

func (e DatabaseRegistered) OccuredAt() time.Time {
	return e.CreatedAt
}
