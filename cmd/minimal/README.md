# Minimal Backup App

A maintainable, modular CLI that streams PostgreSQL plain SQL dumps into `zstd`.

Pipeline per DB:

`pg_dump --format=plain | zstd -<level> -T... -o <file.sql.zst>`

Compressed mode keeps memory usage low for large databases because data is streamed through pipes.
Non-compressed mode writes plain SQL directly from `pg_dump --file`.

## Features

- Single quick-run mode from CLI args (`-db ...`).
- Multi-source mode from `cmd/minimal/db-sources/*.json`.
- Fallback `pg_dump` binary detection (first match by priority).
- Sync/async execution with configurable concurrency.
- Non-fail-fast behavior: continue and aggregate per-db errors.
- Per-database report files next to backup output (enabled by default).
- Aggregate run report in output directory (enabled by default).

## Source modes

- `-source-mode auto` (default):
  - Uses single mode when `-db` is set.
  - Uses file mode when `-db` is empty.
- `-source-mode single`: uses CLI database connection flags.
- `-source-mode files`: parses `-source-files` (glob or CSV list).

`.json` and `.jsonc` are supported.

## Supported source JSON shapes

Array/object with per-entry fields:

- `host`, `port`, `username`, `password`
- `database` (single DB)
- `databases` (multiple DBs)
- `disabled` (skip entry)
- `reportDisabled` (disable per-db report for that entry)
- `compress` (optional bool, defaults `true`; if `false`, writes plain `.sql`)

## Build

```bash
go build ./cmd/minimal
```

## Single DB quick run

```bash
go run ./cmd/minimal \
  -source-mode single \
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

## Multi DB from sources

```bash
go run ./cmd/minimal \
  -source-mode files \
  -source-files "cmd/minimal/db-sources/*.json" \
  -process-mode async \
  -concurrency 4
```

## Key flags

- Connection defaults:
  - `-host=localhost`
  - `-port=5432`
  - `-username=postgres`
- Compression:
  - `-zstd-level` (`1..22`, default `19`)
  - `-zstd-all-cpus` (default `true`)
  - `-zstd-threads` (used when all-cpus is false)
  - Source-level `compress=false` skips zstd for that task.
- Timestamp:
  - `-jalali` uses Jalali timestamp in `Asia/Tehran` format like `1405-03-01--13-40-00-000`.
- Processing:
  - `-process-mode=sync|async`
  - `-concurrency` (for async mode)
- Reports:
  - `-report-db` (default `true`)
  - `-report-run` (default `true`)
  - `-report-ext` (default `.report.json`)
  - `-profile-summary` (default `true`)
- Logging:
  - `-verbose` or `-v` enables extra debug logs.
  - Info logs remain detailed by default for long-running visibility.

## pg_dump fallback priority

When `-pg-dump-bin` is empty, first available path is used:

1. `pg_dump` from `PATH`
2. `/opt/homebrew/opt/postgresql@18/bin/pg_dump`
3. `/opt/homebrew/opt/postgresql@17/bin/pg_dump`
4. `/opt/homebrew/opt/postgresql@16/bin/pg_dump`
5. `/opt/homebrew/bin/pg_dump`
6. `/usr/local/bin/pg_dump`
7. `/usr/bin/pg_dump`

## Reports

Per-db report (next to backup):

`<backup-file>.report.json` (configurable ext)

Includes:

- uncompressed/compressed size
- human-readable sizes (`KiB/MiB/GiB...`)
- compression ratio and percent saved
- dump time, compression time, total time
- stderr and error details
- optional runtime profile summary

Run report (`out-dir`):

- total/succeeded/failed
- total size and overall compression stats
- total dump duration and total compression duration
- end-to-end duration and run config
- all per-db result entries
