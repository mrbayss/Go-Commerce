# GoCommerce

[![Go](https://img.shields.io/badge/Go-1.27.1-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-v3.5.0-00ADD8?logo=go&logoColor=white)](https://gofiber.io/)
[![GORM](https://img.shields.io/badge/GORM-v1.31.2-00ADD8)](https://gorm.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13%2B-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-v9-DC382D?logo=redis&logoColor=white)](https://redis.io/)
[![Status](https://img.shields.io/badge/status-alpha-yellow)](./PROGRESS.md)

> Backend API e-commerce **multi-vendor**, Go + Fiber v3. Status: **alpha** — slice autentikasi selesai,
> fitur lain masih backlog.

---

## Deskripsi

GoCommerce adalah backend API e-commerce multi-vendor (pola Shopee/Tokopedia) untuk pasar Indonesia.
Dibangun dengan Go + Fiber v3 di atas PostgreSQL, Redis, dan SMTP, dengan arsitektur *Light Clean
Architecture* (`controller → service → repository → entity`). Saat ini hanya slice autentikasi yang
selesai; seluruh domain e-commerce lainnya masih backlog.

---

## Status

- **Alpha.** Tabel aktif: `users` saja. 23 file Go, 1.871 baris, **0 test**.
- Auth: signup → OTP email → verify-otp → token pair → login. Belum ada refresh/logout/RBAC.
- Status fitur, known issues, backlog: [PROGRESS.md](./PROGRESS.md).

---

## Tech Stack

| Komponen | Versi |
|---|---|
| Go | 1.27.1 |
| Fiber | v3.5.0 |
| GORM | v1.31.2 (postgres driver v1.6.3, pgx v5.10.0) |
| PostgreSQL | 13+ |
| Redis | v9.23.0 (go-redis) |
| Viper | v1.21.0 |
| Logrus | v1.10.2 |
| Validator | v10.30.5 |
| go-mail | v0.8.1 |
| golang-jwt | v5.3.1 |

---

## Arsitektur (ringkas)

Dependency satu arah: `cmd/api → internal → infrastructure`. Entity dibuat per fitur (bukan
dikumpulkan di awal). Detail layer map, request flow 4 endpoint, error handler, dan 7 ADR:
[docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md).

---

## Prasyarat

- Go 1.27.1+
- PostgreSQL 13+ (database `gocommerce`)
- Redis 6+
- Akun SMTP (untuk kirim OTP)

---

## Quick Start

```bash
# 1. Clone
git clone https://github.com/mrbayss/Go-Commerce.git
cd Go-Commerce

# 2. Buat config dari template (jangan commit)
cp config.yaml.example config.yaml
# edit: postgres.pass, jwt.secret (>=32 byte), smtp.*

# 3. Buat database
psql -U postgres -c "CREATE DATABASE gocommerce;"

# 4. Jalankan
go run ./cmd/api
# server di http://localhost:8080
```

Env override: Viper mengganti `.` → `_` (mis. `POSTGRES_PASS`, `JWT_SECRET`). Kunci config lengkap:
[`config.yaml.example`](./config.yaml.example).

---

## Endpoint (ringkas)

| Method | Path | Auth |
|---|---|---|
| GET | `/api/v1/health` | publik |
| POST | `/api/v1/auth/signup` | publik |
| POST | `/api/v1/auth/verify-otp` | publik |
| POST | `/api/v1/auth/resend-otp` | publik |
| POST | `/api/v1/auth/login` | publik |

Semua endpoint publik (belum ada middleware JWT). Detail body/DTO/error/alur e2e:
[docs/API.md](./docs/API.md).

---

## Struktur Proyek

```
GoCommerce/
├── cmd/api/                 # composition root
├── internal/
│   ├── controller/          # HTTP handler
│   ├── service/             # logika bisnis
│   ├── repository/          # akses data (GORM)
│   ├── entity/              # model domain
│   ├── model/               # DTO + response envelope
│   ├── domain/              # AppError + sentinel
│   ├── pkg/{jwt,otp}        # utilitas murni
│   └── router/              # registrasi route
├── infrastructure/
│   ├── app/                 # Fiber app + middleware + error handler
│   ├── config/              # Config typed + Viper
│   ├── datastore/           # GORM & Redis client
│   ├── pkg/mailer/          # SMTP + template
│   └── {logger,validation}/
├── docs/                    # API, ARCHITECTURE, DATABASE, spec OpenAPI
├── config.yaml.example
└── PROGRESS.md
```

---

## Konfigurasi

Semua key konfigurasi ada di [`config.yaml.example`](./config.yaml.example) — satu sumber.
Key wajib diisi: `postgres.pass`, `jwt.secret` (>=32 byte), `smtp.password`, `auth.otp_secret`.

Tabel config **tidak** diduplikasi di sini (hindari drift).

---

## Perintah Pengembangan

```bash
gofmt -l .            # harus kosong
go build ./...         # harus sukses
go vet ./...           # harus bersih
go test ./... -race    # harus lulus
go mod tidy            # jangan ubah go.mod/go.sum di CI
govulncheck ./...      # tidak boleh ada CVE reachable
```

Detail workflow, troubleshooting, tier test: [AGENTS.md](./AGENTS.md).

---

## Database

- Tabel aktif: `users` (partial unique index `email`/`phone`, soft delete).
- Skema dibuat via `AutoMigrate` saat boot (wajib `false` di production).
- Redis: state store OTP/session. Key & temuan: [docs/DATABASE.md](./docs/DATABASE.md).

---

## Keamanan (ringkas)

- Password di-hash bcrypt cost 10; OTP crypto/rand + HMAC-SHA256 + constant-time compare.
- **Celah aktif:** JWT secret placeholder, rate limit mati, CORS wildcard, tidak ada logout,
  `ctx.Value` tanpa comma-ok, 2 CVE reachable, module retracted. Daftar lengkap:
  [PROGRESS.md](./PROGRESS.md#known-issues).
- Credential bocor di config lokal (DB pass, SMTP app-password) — **rotate manual** bila repo
  tersebar; `config.yaml` tidak pernah ter-commit.

---

## Roadmap

Belum dikerjakan (lihat [PROGRESS.md](./PROGRESS.md#backlog--open-questions)):
refresh token, logout/revoke, RBAC, `me`, forgot/reset password, store, product, cart, order,
payment, shipment, voucher, wallet, review, chat, campaign, notification; versioned migration;
test P0–P4.

---

## Dokumentasi

| Dokumen | Isi |
|---|---|
| [AGENTS.md](./AGENTS.md) | Kontrak kerja, layering, konvensi, verifikasi, vertical slice |
| [docs/API.md](./docs/API.md) | Endpoint, DTO, error, JWT, alur e2e, drift vs spec |
| [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) | Layer map, request flow, error handler, ADR-001..007 |
| [docs/DATABASE.md](./docs/DATABASE.md) | Skema `users`, Redis key, query, migrasi, retensi |
| [docs/api/spec-api.yaml](./docs/api/spec-api.yaml) | Kontrak OpenAPI 3.1 |
| [PROGRESS.md](./PROGRESS.md) | Status fitur, known issues, backlog, open questions |

Desain domain e-commerce lama disimpan di git tag `wip-domain-sketch` (lihat ADR-007).

---

## Kontribusi

Ikuti [AGENTS.md](./AGENTS.md): Conventional Commits, scope = area repo, PR checklist §9.
Jangan commit `config.yaml`. Jangan tambah dependency tanpa alasan kuat.

---

## License

Belum ada LICENSE. Tambahkan sebelum rilis publik.