# Entrypoint Bootstrap Plan

## Goal

Build a reusable bootstrap flow that uses `zhinux-platform` cross-cutting primitives while allowing each service to extend configuration safely.

## Runtime Composition Order

1. Load base config from platform config loader.
2. Load service config extension (Postgres, Redis, HTTP/gRPC toggles, limits).
3. Initialize structured logger.
4. Initialize adapters/resources (Postgres pool, cache backend).
5. Initialize repositories and inject into app container.
6. Initialize HTTP server and health probes.
7. Initialize gRPC runtime + interceptors and register handlers.
8. Start servers and wait for stop signal/runtime failure.
9. Execute coordinated graceful shutdown hooks in reverse order.

## Configuration Extension Strategy

- Keep `platform/config.Base` as the required shared contract for every service.
- Add a local `internal/config.Config` with `Base` + service sections.
- Reuse the same Viper options for both common and service-specific keys (`APP_CONFIG_*`), preserving env override behavior.
- Fallback to env-only mode if platform is built without Viper support.
- Naming rule:
  - shared: `service.*`, `app.*`, `server.*` (platform)
  - service: `db.postgres.*`, `db.redis.*`, `server.http.*`, `server.grpc.*` (service extension)

## Pending Scaffolds (Ready to Fill)

- gRPC server handler registration once contracts/use-cases are finalized.
- Redis active readiness probe once cache keys/probe policy is defined.
- Inbound adapters (HTTP routes, gRPC services) and use-case wiring.
- Metrics/tracing middleware once telemetry conventions are finalized.
