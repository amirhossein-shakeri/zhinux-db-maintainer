# Zhinux DB Maintainer

A service to maintain and manage databases and data sources.

## Core Functionality

- **Backup**: Backup/Export Postgres, Mongo, or other supported data sources with different supported formats(dump, plain sql, etc.) to a target storage(provided by storage service).
- **Restore**: Resote/Import Postgres, Mongo, or other supported data sources by different supported formats(dump, plain sql, etc.) from a storage(provided by storage service).
- **Run Migration**: Run a SQL(Or others that will be supported in the future) migration on a specifi data source.
