package datasource

type DataSourceType string

const (
	DataSourceTypePostgres DataSourceType = "postgres"
	DataSourceTypeMongo    DataSourceType = "mongo"
)
