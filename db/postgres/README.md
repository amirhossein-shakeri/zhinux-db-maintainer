# SQL Layout for sqlc

This service uses a domain-first SQL layout so each bounded context can evolve independently.

## Directory Tree

```text
internal/adapters/db/postgres/sql/
├── schema/
│   ├── shared/               # Shared primitives loaded first
│   ├── database/             # Database registry tables/indexes
│   ├── backup/               # Backup plans, jobs, artifacts
│   └── restore/              # Restore jobs
└── queries/
    ├── database/             # Queries used by DatabaseRepository
    ├── backup/               # Queries used by backup repositories
    └── restore/              # Queries used by RestoreJobRepository
```

## Working Agreement

- Keep each `.sql` file focused on one use case (`find_*`, `list_*`, `upsert_*`, `mark_*`).
- Keep schema and query files under ~150 lines unless there is a strong reason.
- Use one `sqlc` block per domain package (`database`, `backup`, `restore`).
- Include only the schema folders needed by each domain block.
- Prefer additive schema changes with migration files rather than rewriting large SQL files.


_Updates_
## Schema 
Schema is derived from applying all migrations on a fresh DB and then dumping from the DB using `pg_dump --schema-only > schema.sql` so that `sqlc` can generate type-safe queries and schemas won't drift from migrations.
