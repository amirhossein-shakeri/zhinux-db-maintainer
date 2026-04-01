package database

import "time"

type Database struct {
	ID    string
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
