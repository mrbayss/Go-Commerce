# GoCommerce

[![Go](https://img.shields.io/badge/Go-1.27.1-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-v3.5.0-00ADD8?logo=go&logoColor=white)](https://gofiber.io/)
[![GORM](https://img.shields.io/badge/GORM-v1.31.2-00ADD8)](https://gorm.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13%2B-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-v9-DC382D?logo=redis&logoColor=white)](https://redis.io/)
[![Status](https://img.shields.io/badge/status-alpha-yellow)](./PROGRESS.md)

> Backend API e-commerce **multi-vendor** (Go + Fiber v3). Status: **alpha** — slice autentikasi dasar sudah ada, domain lain masih backlog.

## Deskripsi

GoCommerce adalah backend API e-commerce multi-vendor (pola Shopee/Tokopedia). Dibangun dengan Go + Fiber v3 di atas PostgreSQL, Redis, dan SMTP, dengan alur umum `controller → service → repository → entity`.

## Status singkat

- **Auth:** signup → verify OTP → resend OTP → login → refresh token → logout/revoke session.
- **User:** `GET /users/me` (baca profil) dan `PATCH /users/me` (ubah nama/telepon) — keduanya Bearer Token.
- **Middleware:** JWT stateless + RBAC dasar (`RequireAuth` aktif di logout dan endpoint user).
- **Domain lain:** belum dikerjakan (lihat [PROGRESS.md](./PROGRESS.md)).

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

## Prasyarat

- Go 1.27.1+
- PostgreSQL 13+ (database `gocommerce`)
- Redis 6+
- Akun SMTP (untuk kirim OTP)

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

Env override: Viper mengganti `.` → `_` (mis. `POSTGRES_PASS`, `JWT_SECRET`). Kunci config lengkap: [`config.yaml.example`](./config.yaml.example).

## Endpoint (ringkas)

| Method | Path | Auth |
|---|---|---|
| GET | `/api/v1/health` | Public |
| POST | `/api/v1/auth/signup` | Public |
| POST | `/api/v1/auth/verify-otp` | Public |
| POST | `/api/v1/auth/resend-otp` | Public |
| POST | `/api/v1/auth/login` | Public |
| POST | `/api/v1/auth/refresh` | Public |
| POST | `/api/v1/auth/logout` | Bearer Token |
| GET | `/api/v1/users/me` | Bearer Token |
| PATCH | `/api/v1/users/me` | Bearer Token |

### `GET /api/v1/users/me`

Mengembalikan profil user yang sedang login. Data profil dibaca dari database (bukan dari claims token) dan di-cache selama 15 menit. Field sensitif seperti `password_hash` tidak pernah dikirim.

```json
{
  "success": true,
  "message": "profile retrieved successfully",
  "data": {
    "id": "9f9b1f7e-0d49-4f0e-9cc2-6f2d4f6f0b11",
    "email": "user@mail.com",
    "first_name": "Budi",
    "last_name": "Santoso",
    "phone": "6281234567890",
    "avatar_url": null,
    "role": "buyer",
    "is_active": true,
    "created_at": "2026-11-18T08:00:00Z",
    "updated_at": "2026-11-18T08:00:00Z"
  }
}
```

Respons error:

| Status | Penyebab |
|---|---|
| 401 | Token absen, invalid, expired, revoked, atau `user_id` di token bukan UUID valid |
| 403 | Akun tidak aktif (`is_active = false`) |
| 404 | User tidak ditemukan (mis. sudah dihapus) |
| 500 | Gagal mengakses database |
| 502 | Redis gagal diverifikasi saat pengecekan revokasi token |

Status `is_active` juga diperiksa di `login` dan `refresh`: akun nonaktif tidak bisa mendapat token baru. Token access yang sudah terbit tetap berlaku sampai expired (maks. 15 menit).

### `PATCH /api/v1/users/me`

Mengubah sebagian field profil user yang sedang login. Hanya `first_name`, `last_name`, dan `phone` yang dapat diubah; field yang tidak dikirim dibiarkan. Whitespace di awal/akhir nilai di-trim otomatis. `first_name`/`last_name` wajib non-blank bila dikirim (maks. 100 karakter setelah trim), `phone` berformat `628xxxxxxxxx`, dan string kosong persis (`""`) menghapus nomor telepon — whitespace-only (mis. `" "`) ditolak 422. Respons berupa `UserProfileRes` terbaru yang dibaca langsung dari database (bentuk sama seperti `GET /users/me`).

```json
{
  "first_name": "Budi",
  "last_name": "Santoso",
  "phone": "6281234567890"
}
```

Respons sukses:

```json
{
  "success": true,
  "message": "profile updated successfully",
  "data": {
    "id": "9f9b1f7e-0d49-4f0e-9cc2-6f2d4f6f0b11",
    "email": "user@mail.com",
    "first_name": "Budi",
    "last_name": "Santoso",
    "phone": "6281234567890",
    "avatar_url": null,
    "role": "buyer",
    "is_active": true,
    "created_at": "2026-11-18T08:00:00Z",
    "updated_at": "2026-11-19T08:00:00Z"
  }
}
```

| Status | Penyebab |
|---|---|
| 400 | Payload tidak valid |
| 401 | Token absen, invalid, expired, revoked, atau `user_id` bukan UUID valid |
| 403 | Akun tidak aktif |
| 404 | User tidak ditemukan |
| 409 | Nomor telepon sudah dipakai user lain |
| 422 | Validasi gagal (nama blank, nama > 100 karakter, format telepon salah, atau tidak ada field dikirim) |
| 500 | Gagal mengakses database |

## Struktur Proyek

```
GoCommerce/
├── cmd/api/                 # entrypoint & wiring
├── internal/
│   ├── controller/          # HTTP handler
│   ├── service/             # logika bisnis
│   ├── repository/          # akses data (GORM)
│   ├── entity/              # model domain
│   ├── model/               # DTO + response envelope
│   ├── domain/              # AppError + sentinel
│   ├── middleware/          # guard (auth/RBAC)
│   ├── pkg/{jwt,otp}        # utilitas murni
│   └── router/              # registrasi route
├── infrastructure/
│   ├── app/                 # Fiber app + middleware + error handler
│   ├── config/              # Config typed + Viper
│   ├── datastore/           # GORM & Redis client
│   ├── pkg/mailer/          # SMTP + template
│   └── {logger,validation}/
├── config.yaml.example
└── PROGRESS.md
```

## Perintah pengembangan (opsional)

```bash
gofmt -w .
go build ./...
go vet ./...
go test ./... -race
```

## Dokumentasi

| Dokumen | Isi |
|---|---|
| [AGENTS.md](./AGENTS.md) | Panduan ringkas untuk agent/kontributor |
| [PROGRESS.md](./PROGRESS.md) | Status fitur & arah berikutnya |

## Catatan

- `config.yaml` berisi data sensitif lokal — jangan di-commit.
- Kode adalah sumber kebenaran; dokumen bisa tertinggal dari implementasi.

## License

Belum ada LICENSE.
