# ARCHITECTURE.md — Arsitektur GoCommerce

> Light Clean Architecture: dependency satu arah `cmd/api → internal → infrastructure`.
> Kode adalah sumber kebenaran; bila dokumen dan kode berbeda, perbarui dokumen ini.
>
> Referensi: [AGENTS.md](../AGENTS.md) (kontrak kerja) · [API.md](./API.md) (endpoint) ·
> [DATABASE.md](./DATABASE.md) (skema & Redis) · [PROGRESS.md](../PROGRESS.md) (status).

---

## 1. Ringkasan

Backend API e-commerce multi-vendor, Go 1.27.1 + Fiber v3.5.0. Hanya slice autentikasi yang selesai;
seluruh domain e-commerce lainnya masih backlog. Entity dibuat **per fitur** (ADR-003), bukan dikumpulkan
di awal.

---

## 2. Layer Map

```
cmd/api ──▶ internal/router ──▶ controller ──▶ service ──▶ repository ──▶ entity
                                   │              │
                                   │              ▼
                                   │        domain · pkg (pure)
                                   ▼
                        infrastructure/*  (GORM · Redis · SMTP · JWT · Viper · Logrus · Validator)
```

| Layer | Path | Tanggung jawab |
|---|---|---|
| Composition root | `cmd/api/main.go` | Wiring DI, listen, graceful shutdown |
| Router | `internal/router/` | Registrasi route Fiber + grouping `/api/v1` |
| Controller | `internal/controller/` | Bind body → service → envelope |
| Service | `internal/service/` | Logika bisnis, orkestrasi Redis/DB/mail/token |
| Repository | `internal/repository/` | Akses data GORM, query/CRUD |
| Entity | `internal/entity/` | Model domain + tag GORM |
| Model | `internal/model/` | DTO request/response + tag `validate` |
| Domain | `internal/domain/` | `AppError` + sentinel error |
| Pure utils | `internal/pkg/{jwt,otp}` | Utilitas murni, tanpa I/O |
| Infrastructure | `infrastructure/*` | Fiber app, middleware, error handler, health, config, datastore, mailer, logger, validation |

**Aturan dependency** (lihat [AGENTS.md §2](../AGENTS.md#2-aturan-layering)):

1. Satu arah `cmd/api → internal → infrastructure`. Tidak ada panah balik.
2. `internal/entity/*` hanya import stdlib + GORM tag + UUID.
3. `internal/domain/*` tidak import layer lain.
4. Service tidak query DB langsung — semua lewat `internal/repository`.
5. Controller tanpa logika bisnis.
6. Repository tanpa aturan bisnis.
7. DTO di `internal/model`, entity di `internal/entity`. Entity tidak pernah jadi response.
8. `internal/*` tidak import `infrastructure/*` — kontrak di `internal/*`, implementasi di
   `infrastructure/*`.

**Pelanggaran yang diketahui:** aturan 8 belum terpenuhi. `internal/service/auth_service.go`
meng-import `*redis.Client` dan `*mailer.Client` secara konkret. Saat menambah dependency
eksternal ke service, ekstrak interface kecil di `internal/service/ports.go` lebih dulu.

---

## 3. Bootstrap & Konfigurasi

`cmd/api/main.go`:

1. Load config via Viper (`infrastructure/config`).
2. Validasi JWT secret ≥ 32 byte.
3. Buka koneksi GORM (PostgreSQL) + pool.
4. Buka klien Redis.
5. AutoMigrate (bila `postgres.auto_migrate = true`).
6. Wiring: repository → service → controller → router.
7. Fiber app listen (`app.port`).
8. Graceful shutdown (SIGINT/SIGTERM).

Kegagalan koneksi DB/Redis saat boot → `panic`/`log.Fatal` (proses mati). AutoMigrate berjalan
setiap boot sebelum Fiber listen; kegagalannya memicu `log.Fatal`.

---

## 4. Request Flow

Semua endpoint publik (belum ada middleware JWT). Envelope standar:
`{ "success": bool, "message": string, "data"?: any }`.

### 4.1 POST /api/v1/auth/signup

1. Controller bind `SignUpReq` → service `SignUp`.
2. Service normalisasi email (lowercase + trim), cek duplikat (email/phone) via repository.
3. Hash password bcrypt cost 10.
4. Generate OTP 6 digit (`crypto/rand`) → HMAC-SHA256 hex dengan binding email.
5. Simpan `signup:pending:<email>` + `signup:otp:<email>` (Redis, TTL `auth.otp_ttl`); hapus
   `signup:attempts` & `signup:resend`.
6. Kirim email OTP **async** (goroutine, tanpa retry).
7. Response 200 + `expired_at`.

### 4.2 POST /api/v1/auth/verify-otp

1. Bind `VerifyOTP` (`code_otp`, `email`).
2. `GETDEL signup:otp:<email>` (atomik, single-use) → verifikasi HMAC constant-time.
3. Counter `signup:attempts` (batas 5) → 429 + hapus semua key signup bila ≥ 5.
4. Baca `signup:pending:<email>`.
5. INSERT `users` (race → 409 via unique index partial).
6. Buat JWT access + refresh; simpan `session:<userID>:<sessionID>` (SHA-256 refresh, TTL
   `jwt.refresh_expiry`).
7. Hapus 4 key signup; response `TokenPairRes`.

**Tidak atomik:** verify-OTP → create user menyentuh dua sistem; lihat
[DATABASE.md §7](./DATABASE.md#7-konsistensi--transaksi).

### 4.3 POST /api/v1/auth/resend-otp

1. Bind `ResendOTPReq` (`email`).
2. Cek `signup:pending:<email>` ada.
3. Counter `signup:resend` (batas `auth.max_resend_per_session` = 3) → 429 bila ≥ 3.
4. Generate OTP baru → `SET signup:otp` + `EXPIRE pending` + `INCR resend` + `EXPIRE resend`
   + `DEL attempts` dalam satu `TxPipeline`.
5. Kirim email async. Response 200 + `expired_at`.

**Tidak ada cooldown.** `auth.resend_cooldown` dead config.

### 4.4 POST /api/v1/auth/login

1. Bind `LoginReq` (`email`, `password`).
2. `GetByColumn("email")` → 400 `email or password is wrong` bila tidak ada.
3. Cek `is_active` (403) dan `is_verified` (403).
4. `bcrypt.CompareHashAndPassword` → 400 bila salah.
5. Buat session baru di Redis; response `TokenPairRes`.

**Celah:** tanpa dummy bcrypt compare → user enumeration via timing. `password_hash` nullable →
500. Lihat [PROGRESS.md §Known Issues](../PROGRESS.md#known-issues).

### 4.5 GET /api/v1/health

Ping DB + Redis. `status = "degraded"` bila salah satu gagal, tapi HTTP tetap 200. Timeout 2 s via
`context.Background()`. Shape tidak mengikuti envelope standar.

---

## 5. Global Error Handler

Urutan branch (yang pertama cocok menang) — `infrastructure/app/fiber.go`:

1. `*fiber.Error` → status dari error
2. `*fiber.BindError` → 400
3. `validator.ValidationErrors` → 422 + `data[]`
4. `*apperrors.AppError` → `appErr.Code`
5. `gorm.ErrRecordNotFound` → 404
6. `gorm.ErrDuplicatedKey` → 409 (berdasarkan nama index)
7. `*pgconn.PgError` SQLSTATE `23503` → 409
8. lainnya → 500

`AppError` dicek **setelah** `*fiber.Error`: bila `AppError` membungkus `*fiber.Error`, kode
`AppError` hilang. Jalur 422 praktis mati di HTTP; 422 hanya lewat `Validation.Struct` di service.

Detail konvensi error: [AGENTS.md §7](../AGENTS.md#7-konvensi-error).

---

## 6. Konvensi & Utilitas

- **JWT** (`internal/pkg/jwt/jwt.go`): HS256, secret ≥ 32 byte, `exp` wajib, issuer opsional.
  Access token stateless; refresh token disimpan sebagai SHA-256 hex di
  `session:<userID>:<sessionID>`. Tidak ada blacklist/revoke.
- **OTP** (`internal/pkg/otp/otp.go`): `crypto/rand` 6 digit, HMAC-SHA256 dengan binding email,
  verifikasi constant-time.
- **Validation** (`infrastructure/validation/validator.go`): custom rule `id_phone`, `password`,
  `notblank`; pemetaan tag → pesan via `FormatValidationError`.
- **Logger** (`infrastructure/logger`): Logrus JSON.
- **Mailer** (`infrastructure/pkg/mailer`): SMTP + template embedded; `SendTemplate`.

---

## 7. Repository Generic

`Repository[T]` (`internal/repository/base.go`): `Create`, `GetByID`, `GetByColumn`, `Exists`,
`Update`, `Updates`, `Delete`, `GetAll`, `UpdateIsEnable` (dead code). `AuthRepository` = alias
`Repository[entity.User]` tanpa method domain tambahan.

Batasan:
- `Updates` melewati zero value (`IsActive = false` dan `AvatarURL = nil` tidak bisa di-set).
- `GetAll` tanpa `LIMIT` (OOM di tabel besar).
- `GetByColumn`/`Exists` menerima nama kolom bebas (typo = runtime error, whitelist).

---

## 8. Redis & Skema

- Skema DB: satu tabel `users` (partial unique index `email`/`phone`, soft delete).
  Detail: [DATABASE.md §3](./DATABASE.md#3-skema-users).
- Redis: state store OTP/session, key & temuan: [DATABASE.md §5](./DATABASE.md#5-redis).
- AutoMigrate terbatas entity aktif; tidak ada versioned migration
  ([ADR-006](#adr-006--versioned-migration-automigrate-dipangkas)).

---

## 9. Architecture Decision Records

### ADR-001 — Light Clean Architecture (tanpa use-case layer)

**Status:** diterima. Layer `controller → service → repository → entity`, tanpa use-case terpisah.
Konsekuensi: lebih sedikit boilerplate; bila domain membesar, service bisa pecah jadi use-case.

### ADR-002 — Typed Config (Viper + struct)

**Status:** diterima. `Config` struct typed, loader Viper, env override `.` → `_`.
Konsekuensi: typo terdeteksi compile-time; call site tidak memanggil `viper.GetString`.

### ADR-003 — Entity dibuat per fitur

**Status:** diterima. Entity hanya dibuat saat fiturnya dikerjakan; AutoMigrate hanya entity
aktif. Konsekuensi: repo tetap kecil; sketsa lama disimpan di git tag `wip-domain-sketch`.

### ADR-004 — OTP disimpan di Redis, bukan DB

**Status:** diterima. OTP + draft signup di Redis (`signup:*`); row `users` dibuat setelah OTP
terverifikasi. Konsekuensi: konsumsi atomik via `GetDel`; state hilang bila Redis restart.

### ADR-005 — Generic Repository[T]

**Status:** diterima. Generic `Repository[T]` + embed di repo spesifik untuk method domain.
Konsekuensi: reusable; batasan zero-value `Updates` dan mass-assignment harus dijaga.

### ADR-006 — Versioned migration (AutoMigrate dipangkas)

**Status:** direncanakan. `auto_migrate: false` di production + tool versioned (Atlas) sebagai
release step. Konsekuensi: butuh disiplin file migration; lihat [DATABASE.md §8](./DATABASE.md#8-strategi-migrasi).

### ADR-007 — Pengetahuan domain di git tag `wip-domain-sketch`

**Status:** diterima. Entity e-commerce lama dipangkas (commit `a850dce`) agar repo fokus. Desain
lama dapat diakses via `git show wip-domain-sketch:<path>`.

---

## 10. Celah Arsitektur (Known)

Ringkasan; detail lengkap di [PROGRESS.md §Known Issues](../PROGRESS.md#known-issues).

- `internal/*` import `infrastructure/*` (pelanggaran layering) via `*redis.Client`, `*mailer.Client` konkret.
- Tidak ada `db.Transaction` → verify-OTP → create user tidak atomik.
- Tidak ada port interface Redis/Mailer di service.
- `ctx.Value("request_id").(string)` tanpa comma-ok → panic di luar HTTP flow.
- `context.WithValue` dengan key `string` (SA1029).
- Email (PII) masuk nama key Redis & log.
- Tidak ada pagination helper di repository.
- `GetAll` tanpa `LIMIT`; `Updates` melewati zero value.
- Tidak ada access log terstruktur, timeout Fiber, CORS whitelist.
- JWT secret placeholder; rate limit dead config.
- 2 CVE reachable + 1 module retracted.

---

## 11. Referensi

- `cmd/api/main.go` — composition root
- `internal/router/` — route
- `internal/controller/auth_controller.go` — handler
- `internal/service/auth_service.go` — logika auth
- `internal/repository/{base,auth_repository}.go` — data access
- `internal/entity/{entity,user}.go` — model
- `internal/model/auth_model.go` — DTO
- `internal/domain/errors.go` — AppError
- `internal/pkg/{jwt,otp}` — utilitas murni
- `infrastructure/app/fiber.go` — app + error handler
- `infrastructure/config/` — config
- `infrastructure/datastore/{gorm,redis}.go` — koneksi
- `infrastructure/pkg/mailer/mailer.go` — SMTP
- `docs/api/spec-api.yaml` — OpenAPI 3.1
- `config.yaml.example` — template config
- [AGENTS.md](../AGENTS.md) · [API.md](./API.md) · [DATABASE.md](./DATABASE.md) ·
  [PROGRESS.md](../PROGRESS.md) · [README.md](../README.md)