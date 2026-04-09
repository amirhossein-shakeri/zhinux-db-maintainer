package database

import (
	"time"

	"github.com/amirhossein-shakeri/zhinux-platform/types"
)

type Database struct {
	ID types.ID // Internal ID(Fast joins)
	// PublicID string   // Exposed UUID at public APIs(Safe public identifiers)

	Title string
	Typ   DatabaseType

	Host     string
	Port     uint
	User     string
	Password string // todo: add support for other authentication methods later

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
