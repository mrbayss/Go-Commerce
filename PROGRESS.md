# GoCommerce — Progress Log

> Backend e-commerce API boilerplate dengan Go + Fiber v3 + GORM + PostgreSQL + Redis.
> Arsitektur: Clean Architecture (handler → service → repository → entity).

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
│       └── main.go                  ← Entry point (thin, only wiring)
├── internal/
│   ├── config/
│   │   ├── app.go                   ← BootStrapConfig + HealthCheck
│   │   ├── fiber.go                 ← NewFiber + errorHandler
│   │   ├── gorm.go                  ← NewDatabase + connection pool
│   │   ├── redis.go                 ← NewRedis + ping check
│   │   ├── viper.go                 ← NewViper (config loader)
│   │   ├── validation.go            ← FiberValidator adapter + formatValidationError
│   │   └── logrus.go                ← NewLogger
│   ├── domain/
│   │   └── errors.go                ← AppError struct
│   └── model/
│       └── response/
│           └── response.go          ← Standard API response format
├── config.yaml                      ← Active config (gitignored)
├── config.yaml.example              ← Config template
├── go.mod / go.sum
└── .gitignore
```

---

## What's Done

### 1. Configuration Layer
- [x] Viper config loader (`config.yaml` based)
- [x] Fiber app with CORS, logrus access logger, recover, requestid middleware
- [x] Error handler + access log keduanya memakai logrus (structured JSON)
- [x] GORM PostgreSQL connection with configurable pool settings
- [x] Redis client with ping verification
- [x] Logrus logger with configurable level
- [x] FiberValidator adapter wrapping go-playground/validator
- [x] Custom validation error messages (required, email, min, max, gte, lte, eq, ne, oneof, numeric, url, uuid)
- [x] `config.yaml.example` template
- [x] `.gitignore` (secrets, IDE, logs)

### 2. Error Handling
- [x] `AppError` struct in `internal/domain/errors.go`
- [x] Global error handler in `internal/config/fiber.go`
- [x] Handles: Fiber errors, Bind errors, Validation errors, GORM errors (ErrRecordNotFound, ErrDuplicatedKey), PostgreSQL FK violation (23503), AppError
- [x] Unhandled errors logged with logrus (structured JSON), returns generic 500

### 3. Response Format
- [x] Standard `Response` struct (success, message, data)
- [x] `PaginationResponse` struct (with pagination metadata)
- [x] Helper functions: `SuccessResponse`, `CreatedResponse`, `PaginatedResponse`, `ErrorResponse`, `ErrorResponseWithDetails`

### 4. Health Check
- [x] `GET /health` endpoint
- [x] Pings database and Redis with 2s timeout
- [x] Returns `{"status": "ok|degraded", "components": {"database": "ok|error", "redis": "ok|error"}}`

### 5. Middleware
- [x] Access log via logrus (custom middleware, structured JSON) — fields: method, path, status, latency_ms, rid
- [x] Recover (panic recovery)
- [x] RequestID — didaftarkan **sebelum** access logger
- [x] CORS (allow all origins, standard methods + Authorization header)

### 6. Entry Point
- [x] `cmd/api/main.go` — thin main, only wiring
- [x] `config.Bootstrap()` function for DI container
- [x] Server listens on configurable port from `config.yaml`

### 7. Tests
- (none)

---

## Fix Log

### 2026-10-12 — Migrasi access log dari Fiber built-in logger ke logrus

**Gejala:** Fiber built-in logger (plain text) dan Logrus (JSON) tercampur di stdout — format tidak konsisten untuk log aggregation.

**Perbaikan** (`internal/config/fiber.go`):
- Fiber built-in `logger.New()` middleware dihapus.
- Diganti custom `accessLogger` middleware yang menulis JSON via logrus.
- Fields: `method`, `path`, `status`, `latency_ms`, `rid`.
- Request ID dibaca dari response header `X-Request-Id` (diset oleh `requestid.New()`).
- Status code: untuk `*fiber.Error` pakai `.Code`, untuk `*domain.AppError` pakai `.Code`, sisanya fallback ke `c.Response().StatusCode()`.

**Output nyata:**

```json
{"latency_ms":1.237,"level":"info","method":"GET","msg":"request completed","path":"/health","rid":"qVwqJqJyaS3KHyVrYu0d5VoVKC3of0jCr_0QfYwmssw","status":200,"time":"2026-10-12T10:59:18.514+07:00"}
{"error":"Not Found","latency_ms":0.004,"level":"warning","method":"GET","msg":"request completed with error","path":"/notfound","rid":"jJbb2vcLMApTYDWiAfIRFANcbF42y9SVot5bueMTbxo","status":404,"time":"2026-10-12T10:59:18.523+07:00"}
```

---

## What's Next (TODO)

### Phase 1: Domain & Entity Layer
- [ ] Create `internal/entity/` — GORM database models (User, Product, Order, etc.)
- [ ] Create `internal/domain/` — core business models + repository interfaces
- [ ] Create `internal/model/` — request/response DTOs

### Phase 2: Repository Layer
- [ ] Create `internal/repository/` — implements domain interfaces
- [ ] User repository (CRUD)
- [ ] Product repository (CRUD)
- [ ] Order repository (CRUD)

### Phase 3: Service Layer
- [ ] Create `internal/service/` — business logic
- [ ] User service (register, login, profile)
- [ ] Product service (CRUD, search, filter)
- [ ] Order service (create, cancel, status update)

### Phase 4: Handler Layer
- [ ] Create `internal/handler/` — HTTP handlers
- [ ] User handler
- [ ] Product handler
- [ ] Order handler

### Phase 5: Routes & DI
- [ ] Create `internal/routes/` — route definitions + DI wiring
- [ ] `routes.RegisterRoutes(app, db, validate)` composition root
- [ ] Update `main.go` to use routes package

### Phase 6: Auth & Security
- [ ] JWT authentication middleware
- [ ] Password hashing (bcrypt)
- [ ] Auth handler (login, register, refresh)
- [ ] Role-based access control (RBAC)

### Phase 7: Infrastructure & DevOps
- [ ] `docker-compose.yml` (PostgreSQL + Redis)
- [ ] `.env.example` file
- [ ] Database migrations (`migrations/`)
- [ ] Graceful shutdown (signal handling)
- [ ] Struct-based config (replace Viper with typed config struct)
- [ ] Env var override support (12-factor app)

### Phase 8: Testing
- [ ] Unit tests (mock interfaces)
- [ ] Integration tests (HTTP request → response)
- [ ] Handler tests (httptest + mocked services)

### Phase 9: Additional Features
- [ ] Rate limiting middleware
- [ ] Request validation middleware
- [ ] API versioning (`/api/v1/...`)
- [ ] Swagger/OpenAPI documentation
- [ ] Structured logging with request ID correlation
- [ ] Caching layer (Redis integration for hot data)

---

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Clean Architecture | Separation of concerns, testability, maintainability |
| Constructor injection | No DI container needed, explicit dependencies |
| Domain interfaces in `internal/domain/` | Enables mocking, decouples layers |
| `internal/entity/` vs `internal/model/` | DB models vs DTOs separation |
| `AppError` in domain layer | Framework-agnostic business errors |
| `FiberValidator` adapter | go-playground/validator doesn't satisfy `fiber.StructValidator` |
| Config-driven pool/timeout | Production needs different values than dev |
| `config.yaml` + Viper | Simple, file-based config with env override potential |

---

## Known Issues / Technical Debt

1. **`Bootstrap()` is empty** — currently does nothing, needs AutoMigrate + route registration
2. **No graceful shutdown** — `app.Listen()` dies mid-request on SIGTERM
3. **No env var override** — Viper doesn't have `AutomaticEnv()` configured
4. **No struct-based config** — Viper is used directly, not type-safe
5. **CORS allows all origins** — `[]string{"*"}` is fine for dev, needs restriction for prod
6. **No request timeouts** — Fiber `ReadTimeout`, `WriteTimeout`, `IdleTimeout` not set
7. **No body limit** — Fiber `BodyLimit` not configured
8. **No API versioning** — routes not prefixed with `/api/v1/`
9. **Test coverage masih minimal** — belum ada unit/integration test untuk layer manapun
10. **`ServerHeader` not set** — defaults to empty, which is actually fine (omits `Server: fasthttp`)
11. ~~**Logger order in `NewFiber`**~~ — **FIXED** 2026-10-12. `requestid.New()` sekarang didaftarkan sebelum access logger. Lihat Fix Log di atas.
12. ~~**Log syntax: `${requestid}`, NOT `${locals:requestid}`**~~ — **FIXED** 2026-10-12. Custom access log via logrus menggantikan Fiber built-in logger.
13. ~~**`log.Printf` untuk unhandled error**~~ — **FIXED** 2026-10-12. Error handler memakai logrus dengan fields (`error`, `path`). `NewFiber` menerima `*logrus.Logger` dan error handler di-bind via closure.
14. ~~**Fiber built-in logger (plain text) vs Logrus (JSON)**~~ — **FIXED** 2026-10-12. Fiber built-in logger dihapus. Access log sekarang memakai custom middleware yang menulis JSON via logrus — semua log konsisten dalam satu format.
15. ~~**Access log status code salah untuk error**~~ — **FIXED** 2026-10-12. `c.Response().StatusCode()` dibaca sebelum error handler set status, jadi 404 tercatat 200. Fix: untuk `*fiber.Error` pakai `.Code`, untuk `*domain.AppError` pakai `.Code`, sisanya fallback ke response status.

---

## References

- [Fiber v3 Docs](https://docs.gofiber.io/)
- [GORM Docs](https://gorm.io/docs/)
- [Viper Docs](https://github.com/spf13/viper)
- [Logrus Docs](https://github.com/sirupsen/logrus)
- [go-playground/validator](https://github.com/go-playground/validator)
- [go-redis](https://github.com/redis/go-redis)

---

*Last updated: 2026-10-12*
