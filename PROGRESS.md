# GoCommerce — Progress Log

> Backend e-commerce API boilerplate dengan Go + Fiber v3 + GORM + PostgreSQL + Redis + SMTP.
> Arsitektur: Light Clean Architecture (controller → service → repository → entity) with `infrastructure` isolated.

---

## Project Info

| Item | Value |
|------|-------|
| Module | `github.com/mrbayss/Go-Commerce` |
| Go Version | 1.27.1 |
| Fiber | v3.5.0 |
| GORM | v1.31.2 (PostgreSQL driver) |
| Redis | v9.23.0 (go-redis) |
| Viper | v1.21.0 |
| Logrus | v1.10.2 |
| Validator | v10.30.5 |
| Mail | wneessen/go-mail v0.23.0 |

---

## Project Structure (Current)

```
GoCommerce/
├── cmd/
│   └── api/
│       └── main.go                      ← entry point (typed wiring + graceful shutdown)
├── infrastructure/
│   ├── app/                             ← HTTP adapter layer
│   │   ├── fiber.go                     ← NewFiber + middleware + error handler
│   │   └── health.go                    ← HealthCheck (DB + Redis ping)
│   ├── config/
│   │   ├── configuration.go             ← typed Config struct
│   │   └── viper.go                     ← Load() with env overrides + defaults
│   ├── datastore/
│   │   ├── gorm.go                      ← GORM + AutoMigrate + connection pool
│   │   └── redis.go                     ← Redis client
│   ├── logger/
│   │   └── logrus.go                    ← NewLogger (JSON formatter)
│   ├── pkg/
│   │   └── mailer/
│   │       └── mailer.go                ← SMTP client + embedded templates
│   └── validation/
│       └── validator.go                 ← FiberValidator + custom validators (id_phone, password, notblank)
├── internal/
│   ├── controller/
│   │   └── auth_controller.go
│   ├── domain/
│   │   └── errors.go                    ← AppError + sentinel errors
│   ├── entity/                          ← GORM database entities (15+)
│   │   ├── entity.go                    ← BaseModel (UUID, timestamps, soft delete)
│   │   ├── user.go
│   │   ├── address.go
│   │   ├── store.go
│   │   ├── category.go
│   │   ├── product.go
│   │   ├── product_variant.go
│   │   ├── product_image.go
│   │   ├── cart.go / cart_item.go
│   │   ├── voucher.go / voucher_usage.go
│   │   ├── order.go / store_order.go / order_item.go
│   │   ├── payment.go
│   │   ├── shipment.go
│   │   ├── wallet.go / wallet_transaction.go
│   │   ├── campaign.go / flash_sale_item.go
│   │   ├── review.go
│   │   ├── conversation.go / participant.go / message.go
│   │   └── notification.go
│   ├── model/
│   │   ├── auth_model.go                ← request/response DTOs
│   │   └── response/
│   │       └── response.go              ← standard API response helpers
│   ├── pkg/
│   │   └── otp/
│   │       └── otp.go                   ← OTP generation + hashing
│   ├── repository/
│   │   ├── base.go                      ← generic GORM base repository (Repository[T])
│   │   └── auth_repository.go
│   ├── router/
│   │   ├── main_router.go
│   │   └── auth_router.go
│   └── service/
│       └── auth_service.go              ← signup flow with OTP (bcrypt + Redis + SMTP)
├── config.yaml                          ← active config (gitignored)
├── config.yaml.example
├── go.mod / go.sum
└── .gitignore
```

---

## What's Done

### Architecture (Light Clean, `infrastructure/` isolated)
- [x] Typed `Config` struct with `Load()` — no raw `viper.GetString` calls
- [x] `infrastructure/` isolated: `app`, `config`, `datastore` (db/redis), `logger`, `pkg/mailer`, `validation`
- [x] `internal/domain/errors.go` — single error package (`AppError` + sentinels)
- [x] Graceful shutdown in `main.go` (`ShutdownWithContext`, 10s timeout)
- [x] `requestid` + `recover` + `cors` + request-id context middleware in `infrastructure/app/fiber.go`
- [x] Centralized error handler: Fiber/Bind/Validation/GORM/PgError/AppError
- [x] `Fiber` config: `StrictRouting`, `CaseSensitive`, `StructValidator`, `ErrorHandler`
- [x] Connection pool configured via typed config (max open/idle, lifetime, idle time)

### Features
- [x] Viper config loader with `AutomaticEnv()` + `SetEnvKeyReplacer` (dot → underscore)
- [x] Fiber app with CORS (all origins for dev), access logger (logrus JSON), recover, requestid
- [x] Error handler — Fiber/Bind/Validation/GORM/PostgreSQL/AppError with details
- [x] GORM PostgreSQL + Redis + SMTP mail (wneessen/go-mail)
- [x] Logrus JSON logger with configurable level, RFC3339 timestamp
- [x] Custom validation tags: `id_phone` (Indonesian), `password` (complexity), `notblank`
- [x] `AppError` + sentinel errors (`ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, `ErrConflict`, `ErrValidation`, `ErrBadGateway`, `ErrServiceUnavailable`)
- [x] Standard `Response` / `PaginationResponse` helpers
- [x] `GET /health` endpoint (DB + Redis ping with 2s timeout)
- [x] 15+ GORM entities with auto-migration (UUID PK, soft delete, indexes)
- [x] Generic `Repository[T]` + `RepositoryImpl[T]` implementation
- [x] Auth service: signup flow with OTP (bcrypt + Redis rate limiting + SMTP template email)
- [x] Redis rate limiting: per-email signup attempts with window + resend cooldown
- [x] Embedded HTML email templates (`go:embed`)

---

## What's Next (TODO)

### Domain Services
- [ ] Product service (CRUD, search, filter)
- [ ] Order service (create, cancel, status update)
- [ ] User repository / Product repository / Order repository
- [ ] Cart service
- [ ] Payment integration

### Auth
- [ ] Login endpoint + JWT middleware
- [ ] Refresh token flow
- [ ] OTP verification endpoint (complete signup)
- [ ] Password reset flow
- [ ] RBAC (role-based access control: admin/seller/buyer)

### DevOps
- [ ] `docker-compose.yml` (PostgreSQL + Redis)
- [ ] Database migrations (replace AutoMigrate for prod — golang-migrate or Atlas)
- [ ] `env` override for config (done via `AutomaticEnv`)
- [ ] CI/CD pipeline

### Testing
- [ ] Integration tests (HTTP request → response, testcontainers)
- [ ] Handler tests (httptest + mocked services)
- [ ] Unit tests for services

### Features
- [ ] Rate limiting middleware (global, not just auth)
- [ ] API versioning enforced (`/api/v1` present)
- [ ] Swagger/OpenAPI (swaggo/swag)
- [ ] Redis caching layer (query caching)
- [ ] Request/Response logging middleware (structured)
- [ ] Idempotency keys for mutations

---

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Light Clean (Option B) | Layer jelas, testable, tanpa over-abstraction |
| `infrastructure/` vs `pkg/` | infrastructure = external I/O (DB, Redis, SMTP, HTTP); keeps `internal/pkg` for pure utilities (otp) |
| Typed `Config` | Type-safe, catches typos; Viper raw access removed from call sites |
| Mail collapsed to 1 file | Embedded templates; options pattern removed (YAGNI) |
| `internal/domain/errors.go` | Single source of truth for `AppError` + sentinels |
| Graceful shutdown | `ShutdownWithContext(10s)` prevents mid-request crash |
| No DI container | Manual constructor injection; add `wire`/`fx` when >10 deps |
| Generic `Repository[T]` | Reusable CRUD; specific repos embed for domain methods |
| OTP in Redis (not DB) | Ephemeral; account row only created after verification |
| Rate limit in Redis | Distributed, survives app restart; per-email window |

---

## Known Issues / Technical Debt

1. **CORS allows all origins** — `[]string{"*"}` is fine for dev, needs restriction for prod
2. **No testcontainers** — integration tests still TODO
3. **SMTP config `FromAddress` vs `Username`** — `mailer.go` prefers `FromAddress` if set, falls back to `Username`
4. **`internal/pkg/otp`** — should move to `internal/domain/` or `internal/util/` when refactored further
5. **AutoMigrate** — replace with proper migrations before prod
6. **No access logger middleware** — currently only requestid; add structured access log (method, path, status, latency)
7. **Auth service lacks OTP verify/resend** — only signup implemented
8. **No pagination helper** in repository base (only `GetAll`)

---

*Last updated: 2026-10-27 — infrastructure/ reorganization + Light Clean Architecture refactor*