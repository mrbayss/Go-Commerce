# GoCommerce — Progress Log

> Backend e-commerce API dengan Go + Fiber v3 + GORM + PostgreSQL + Redis + SMTP.
> Arsitektur: Light Clean Architecture (`controller → service → repository → entity`), dengan `infrastructure/` diisolasi.

---

## Project Info

| Item | Value |
|------|-------|
| Module | `github.com/mrbayss/Go-Commerce` |
| Go Version | 1.27.1 |
| Fiber | v3.5.0 |
| GORM | v1.31.2 (PostgreSQL driver v1.6.3) |
| PostgreSQL driver | pgx v5.10.0 |
| Redis | v9.23.0 (go-redis) |
| Viper | v1.21.0 |
| Logrus | v1.10.2 |
| Validator | v10.30.5 |
| Mail | wneessen/go-mail v0.8.1 |
| JWT | golang-jwt/jwt v5.3.1 |

### Konsep kerja

Entity **hanya dibuat saat fiturnya dikerjakan**, bukan dikumpulkan di awal.
Setiap fitur diselesaikan sebagai satu slice vertikal, dan statusnya dicatat di tabel di bawah.

Legenda tahap: **E** entity · **M** migration/AutoMigrate · **R** repository · **S** service · **H** handler · **RT** route · **D** API spec · **T** test

---

## Struktur (aktual)

```
GoCommerce/
├── cmd/api/main.go                      ← entry point + wiring + graceful shutdown
├── infrastructure/
│   ├── app/
│   │   ├── fiber.go                     ← Fiber app, middleware, error handler
│   │   └── health.go                    ← GET /health (DB + Redis ping)
│   ├── config/
│   │   ├── configuration.go             ← typed Config struct
│   │   └── viper.go                     ← Load() + env overrides
│   ├── datastore/
│   │   ├── gorm.go                      ← GORM + AutoMigrate (hanya entity aktif)
│   │   └── redis.go                     ← Redis client
│   ├── logger/logrus.go                 ← JSON logger
│   ├── pkg/mailer/mailer.go             ← SMTP client + embedded templates
│   └── validation/validator.go          ← custom validator (id_phone, password, notblank)
├── internal/
│   ├── controller/auth_controller.go
│   ├── domain/errors.go                 ← AppError + sentinels
│   ├── entity/
│   │   ├── entity.go                    ← BaseModel (UUID PK, timestamps, soft delete)
│   │   └── user.go                      ← User + UserRole
│   ├── model/
│   │   ├── auth_model.go                ← request/response DTO auth
│   │   └── response/response.go
│   ├── pkg/
│   │   ├── jwt/jwt.go                   ← token pair + hashing
│   │   └── otp/otp.go                   ← OTP generate + HMAC hash
│   ├── repository/
│   │   ├── base.go                      ← generic Repository[T]
│   │   └── auth_repository.go
│   ├── router/
│   │   ├── main_router.go               ← /api/v1
│   │   └── auth_router.go
│   └── service/auth_service.go
├── docs/api/spec-api.yaml               ← OpenAPI (baru: Auth)
├── config.yaml (gitignored) / config.yaml.example
├── PROGRESS.md
└── go.mod / go.sum
```

---

## Status Fitur

| Fitur | E | M | R | S | H | RT | D | T | Status |
|-------|---|---|---|---|---|---|---|---|--------|
| Infrastructure (config, logger, error, health) | – | – | – | – | – | ✓ | – | ✗ | selesai |
| Auth — SignUp + OTP | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | selesai |
| Auth — Verify OTP | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | selesai |
| Auth — Resend OTP | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | selesai |
| Auth — Login | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | selesai |
| Auth — Refresh token | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Auth — Logout / revoke session | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Auth — JWT middleware + RBAC | – | – | – | – | ✗ | ✗ | ✗ | ✗ | belum |
| Auth — Forgot/reset password | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Profile + Address | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Store | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Category + Product | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Cart | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Order + StoreOrder | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Payment | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Shipment | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Voucher | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Wallet | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Campaign / Flash Sale | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Review | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Chat | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Notification | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |

### Endpoint aktif

| Method | Path | Handler |
|--------|------|---------|
| GET | `/health` | `app.HealthCheck` |
| POST | `/api/v1/auth/signup` | `AuthController.SignUp` |
| POST | `/api/v1/auth/verify-otp` | `AuthController.VerifyOTP` |
| POST | `/api/v1/auth/resend-otp` | `AuthController.ResendOTP` |
| POST | `/api/v1/auth/login` | `AuthController.Login` |

---

## Keputusan Desain

| Decision | Rationale |
|----------|-----------|
| Light Clean (tanpa use-case layer) | Layer jelas, testable, tanpa over-abstraction |
| Entity per fitur | Entity yang belum dipakai tidak bisa direview dan jadi "progress palsu" |
| `infrastructure/` vs `internal/pkg/` | infrastructure = I/O eksternal; `internal/pkg` = utilitas murni (jwt, otp) |
| Typed `Config` | Type-safe, typo terdeteksi, tidak ada `viper.GetString` di call site |
| OTP disimpan di Redis | Ephemeral; row user baru dibuat setelah OTP terverifikasi |
| Rate limit di Redis | Distributed, tahan restart, per-email |
| Generic `Repository[T]` | CRUD reusable; repo spesifik embed untuk method domain |
| Manual DI | Constructor injection; pertimbangkan `wire`/`fx` kalau dependency > 10 |
| AutoMigrate terbatas entity aktif | Mencegah tabel kosong dan schema drift |

---

## Known Issues / Technical Debt

1. **`ctx.Value("request_id").(string)`** di `auth_service.go` bisa panic bila key tidak ada — pakai comma-ok assertion dan typed context key.
2. **AutoMigrate** belum diganti versioned migration (golang-migrate/Atlas) — wajib sebelum production.
3. **Belum ada test** (unit/service, handler, integration).
4. **Belum ada JWT middleware, RBAC, refresh/logout** — token dibuat tapi belum bisa dirotasi/dicabut lewat endpoint.
5. **CORS `*`** masih untuk dev, perlu dibatasi.
6. **`RepositoryImpl.UpdateIsEnable`** tidak ada di interface `Repository[T]` dan tidak dipakai — kandidat dihapus.
7. **Belum ada pagination helper** di repository base (hanya `GetAll`).
8. **Mailer SMTP** memakai `FromAddress` lalu fallback ke `Username` — perlu ditegaskan di config.
9. **Tidak ada access logger terstruktur** (method, path, status, latency) — baru `requestid`.

> Riwayat desain entity lama (store, product, order, wallet, dll.) disimpan di git tag `wip-domain-sketch` untuk referensi saat fiturnya dikerjakan.

---

*Last updated: 2026-11-13 — pruning entity ke `User` saja + AutoMigrate terbatas entity aktif*
