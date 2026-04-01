package database

type DatabaseType string

const (
	DataSourceTypePostgres DatabaseType = "postgres"
	DataSourceTypeMongo    DatabaseType = "mongo"
)
