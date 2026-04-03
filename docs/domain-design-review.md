# DB Maintainer Domain Review

## Current strengths

- Clear bounded contexts already exist: `database`, `backup`, `restore`.
- Domain models are simple and readable.
- Repository abstraction exists and aligns with hexagonal architecture.

## Gaps and future-proof opportunities

1. **Aggregate boundaries are not explicit**
   - Treat `Database`, `BackupPlan`, `BackupJob`, `BackupArtifact`, and `RestoreJob` as separate aggregates with independent repositories.
2. **Lifecycle state transitions are permissive**
   - Add application-level guards to enforce legal transitions only (e.g. `pending -> in_progress -> success|failed|canceled`).
3. **Operational metadata is limited**
   - Add failure reason, retry count, execution actor, and correlation ID in future iterations.
4. **Scheduling and retention are basic**
   - Add richer retention config (`keep_last_n`, `max_age_days`) and schedule timezone support.
5. **Secrets handling needs hardening**
   - Move plain credentials out of row-level storage and use a secret reference (vault key/handle).
6. **Idempotency and concurrency controls**
   - Add idempotency keys for create/start operations and optimistic locking/version fields where race risk exists.

## Suggested roadmap

- **V1 (implemented in this change)**
  - Domain-scoped schema and query structure.
  - Dedicated outbound repositories for database/backup/restore aggregates.
  - SQL and adapter support for plans, jobs, artifacts, and restore jobs.
- **V2**
  - Add app-service orchestration and explicit transition guards.
  - Add retry/backoff fields and failure diagnostics.
- **V3**
  - Add multi-tenant support (tenant_id partitioning strategy).
  - Add audit/event outbox for integration with observability and workflow engines.

## Feature ideas

- Backup verification jobs (checksum + restore dry-run).
- Restore preview and safety checks (target compatibility + conflict detection).
- Point-in-time restore support for engines that provide WAL/binlog.
- Policy engine for backup cadence by data class.
- Storage lifecycle policies (hot/cold/archive) and cost optimization hooks.
