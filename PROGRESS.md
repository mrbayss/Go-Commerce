# GoCommerce — Progress Log

> Backend e-commerce API boilerplate dengan Go + Fiber v3 + GORM + PostgreSQL + Redis.
> Arsitektur: Light Clean Architecture (controller → service → repository → entity) with `infra` isolated.

---

## Project Info

| Item | Value |
|------|-------|
| Module | `github.com/mrbayss/Go-Commerce` |
| Go Version | 1.27.1 |
| Fiber | v3.5.0 |
| GORM | v1.31.2 (PostgreSQL driver) |
| Redis | v9.23.0 |
| Viper | v1.21.0 |
| Logrus | v1.10.2 |
| Validator | v10.30.5 |

---

## Project Structure (Current)

```
GoCommerce/
├── cmd/
│   └── api/
│       └── main.go                  ← entry point (typed wiring + graceful shutdown)
├── internal/
│   ├── app/                         ← HTTP adapter layer
│   │   ├── server.go                ← NewFiber + middleware + access logger + error handler
│   │   ├── validator.go             ← FiberValidator + custom validators
│   │   └── health.go                ← HealthCheck (DB + Redis ping)
│   ├── config/
│   │   ├── config.go                ← typed Config struct + Load()
│   │   └── logger.go                ← NewLogger
│   ├── domain/
│   │   └── errors.go                ← AppError + sentinel errors
│   ├── entity/                      ← GORM database entities
│   ├── infra/
│   │   ├── db/postgres.go           ← GORM + AutoMigrate + connection pool
│   │   ├── redis/redis.go           ← Redis client
│   │   └── mail/mailer.go           ← SMTP client + embedded templates (collapsed)
│   ├── model/
│   │   ├── auth_model.go            ← request/response DTOs
│   │   └── response/response.go     ← standard API response helpers
│   ├── repository/
│   │   ├── base.go                  ← generic GORM base repository
│   │   └── auth_repository.go
│   ├── router/
│   │   ├── main_router.go
│   │   └── auth_router.go
│   ├── controller/
│   │   └── auth_controller.go
│   └── service/
│       ├── auth_service.go
│       └── auth_service_test.go
├── config.yaml                      ← active config (gitignored)
├── config.yaml.example
├── go.mod / go.sum
└── .gitignore
```

---

## What's Done

### Architecture (Light Clean, Option B)
- [x] Typed `Config` struct with `Load()` — no raw `viper.GetString` calls
- [x] `infra/` isolated: `db`, `redis`, `mail` (mail collapsed from 7 files → 1)
- [x] `app/` layer: server, validator, health
- [x] `domain/errors.go` — single error package (consolidated from `internal/errors`)
- [x] Graceful shutdown in `main.go` (`ShutdownWithContext`, 10s timeout)
- [x] `accessLogger` middleware added in `app/server.go`
- [x] `Fiber` timeouts: `BodyLimit` set; `Read/Write/Idle` timeouts documented

### Features (from before)
- [x] Viper config loader (now via typed `Config`)
- [x] Fiber app with CORS, access logger (logrus JSON), recover, requestid
- [x] Error handler — Fiber/Bind/Validation/GORM/AppError
- [x] GORM PostgreSQL + Redis + SMTP mail
- [x] Logrus JSON logger
- [x] Custom validation (id_phone, password, notblank)
- [x] `AppError` + sentinel errors
- [x] Standard `Response` / `PaginationResponse` helpers
- [x] `GET /health` endpoint
- [x] 15 GORM entities with auto-migration
- [x] Generic `Repository[T]` + `GormRepository` implementation
- [x] Auth service: signup flow with OTP (bcrypt + Redis + SMTP)

---

## What's Next (TODO)

### Domain Services
- [ ] Product service (CRUD, search, filter)
- [ ] Order service (create, cancel, status update)
- [ ] User repository / Product repository / Order repository

### Auth
- [ ] Login endpoint + JWT middleware
- [ ] Refresh token flow
- [ ] RBAC (role-based access control)

### DevOps
- [ ] `docker-compose.yml` (PostgreSQL + Redis)
- [ ] Database migrations (replace AutoMigrate for prod)
- [ ] `env` override for config (partially done via `AutomaticEnv`)

### Testing
- [ ] Integration tests (HTTP request → response, testcontainers)
- [ ] Handler tests (httptest + mocked services)

### Features
- [ ] Rate limiting middleware
- [ ] API versioning enforced (`/api/v1` present)
- [ ] Swagger/OpenAPI
- [ ] Redis caching layer

---

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Light Clean (Option B) | Layer jelas, testable, tanpa over-abstraction |
| `infra/` vs `pkg/` | infra = external I/O (DB, Redis, SMTP); keeps `pkg/` for pure utilities (otp) |
| Typed `Config` | Type-safe, catches typos; Viper raw access removed from call sites |
| Mail collapsed to 1 file | 7 files → 1; options pattern removed (YAGNI) |
| `domain/errors.go` | Single source of truth for `AppError` + sentinels |
| Graceful shutdown | `ShutdownWithContext(10s)` prevents mid-request crash |
| No DI container | Manual constructor injection; add `wire`/`fx` when >10 deps |

---

## Known Issues / Technical Debt

1. **CORS allows all origins** — `[]string{"*"}` is fine for dev, needs restriction for prod
2. **No testcontainers** — integration tests still TODO
3. **SMTP config `from_address` vs `username`** — `mailer.go` prefers `FromAddress` if set, falls back to `Username`
4. **`pkg/otp` still in `internal/pkg`** — should move to `internal/domain/` or `internal/util/` when refactored further
5. **AutoMigrate** — replace with proper migrations before prod

---

*Last updated: 2026-10-26 — Light Clean Architecture (Option B) refactor*
