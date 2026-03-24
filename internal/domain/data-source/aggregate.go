package datasource

import "time"

type DataSource struct {
	ID       string
	Title    string
	Typ      DataSourceType
	Host     string
	Port     uint
	User     string
	Password string // todo: add support for other authentication methods later

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
