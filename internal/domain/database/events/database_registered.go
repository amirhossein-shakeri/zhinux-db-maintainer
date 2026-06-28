package database_events

import (
	"time"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/database"
	"github.com/amirhossein-shakeri/zhinux-platform/types"
)

// DatabaseRegistered. The business meaning is "this database is now
// known to our system and can participate in workflows" then the
// domain event should express that business fact: registered.
// DatabaseCreated means a new database instance has actually been
// created/provisioned(Infrastructure event).
type DatabaseRegistered struct {
	DatabaseID types.ID
	PublicID   string
	Title      string
	Typ        database.DatabaseType
	Host       string
	Port       string
	User       string
	CreatedAt  time.Time
}

func (e DatabaseRegistered) EventName() string {
	return "database.registered"
}

func (e DatabaseRegistered) AggregateID() types.ID {
	return e.DatabaseID
}
