package database

import "time"

type Filter struct {
	IDs       []string
	PublicIDs []string
	Title     *string
	Types     []DatabaseType
	Host      *string
	Port      *uint
	User      *string
	Password  *string

	Deleted *bool

	CreatedAtSince *time.Time
	CreatedAtUntil *time.Time
	UpdatedAtSince *time.Time
	UpdatedAtUntil *time.Time
	DeletedAtSince *time.Time
	DeletedAtUntil *time.Time
}
