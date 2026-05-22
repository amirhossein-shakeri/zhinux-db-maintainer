# Minimal Backup App

This app creates a PostgreSQL backup by streaming plain SQL from `pg_dump` directly into `zstd`.

Pipeline:

`pg_dump --format=plain | zstd -<level> -T... -o <file.sql.zst>`

This avoids loading the dump into memory and is suitable for large databases.

## Build

```bash
go build ./cmd/minimal
```

## Usage

```bash
go run ./cmd/minimal \
  -host localhost \
  -port 5432 \
  -username postgres \
  -password postgres \
  -db mydb
```

## Flags

- `-host` (required): PostgreSQL host.
- `-port` (default `5432`): PostgreSQL port.
- `-username` (required): PostgreSQL user.
- `-password`: PostgreSQL password (sets `PGPASSWORD` for `pg_dump`).
- `-db` (required): Database name.
- `-out-dir` (default `backups`): Output directory.
- `-out-name`: Output file name. If omitted, auto-generated as:
  - `<db>_<host>_<UTC timestamp>.sql.zst`
- `-pg-dump-bin` (default `pg_dump`): Path to `pg_dump`.
- `-zstd-bin` (default `zstd`): Path to `zstd`.
- `-zstd-level` (default `19`, range `1..22`): zstd compression level.
- `-zstd-all-cpus` (default `true`): Use `-T0` in zstd.
- `-zstd-threads` (default `0`): Used only when `-zstd-all-cpus=false`; must be `>=1`.

## Example equivalent to your shell command

```bash
go run ./cmd/minimal \
  -host localhost \
  -port 5432 \
  -username postgres \
  -password postgres \
  -db ns \
  -out-name ns-local.sql.zst \
  -pg-dump-bin /opt/homebrew/opt/postgresql@18/bin/pg_dump \
  -zstd-level 19 \
  -zstd-all-cpus=true
```

## Resource behavior

- Data is streamed from `pg_dump` to `zstd` via OS pipes.
- No full dump buffering in app memory.
- Compression CPU usage is controlled by `-zstd-all-cpus` / `-zstd-threads`.
