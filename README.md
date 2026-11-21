# GoCommerce

[![Go](https://img.shields.io/badge/Go-1.27.1-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-v3.5.0-00ADD8?logo=go&logoColor=white)](https://gofiber.io/)
[![GORM](https://img.shields.io/badge/GORM-v1.31.2-00ADD8)](https://gorm.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13%2B-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-v9-DC382D?logo=redis&logoColor=white)](https://redis.io/)
[![Status](https://img.shields.io/badge/status-alpha-yellow)](./PROGRESS.md)

> Backend API e-commerce **multi-vendor** (Go + Fiber v3). Status: **alpha** — slice autentikasi dasar + domain katalog (store, category, product) sudah ada, domain lain masih backlog.

## Deskripsi

GoCommerce adalah backend API e-commerce multi-vendor (pola Shopee/Tokopedia). Dibangun dengan Go + Fiber v3 di atas PostgreSQL, Redis, dan SMTP, dengan alur umum `controller → service → repository → entity`.

## Status singkat

- **Auth:** signup → verify OTP → resend OTP → login → refresh token → logout/revoke session.
- **User:** `GET /users/me` (baca profil), `PATCH /users/me` (ubah nama/telepon), `PUT /users/me/avatar` (upload/ganti foto profil), `DELETE /users/me/avatar` (hapus foto profil) — semuanya Bearer Token. File avatar tersaji publik di `/uploads/avatars/...`.
- **Store:** daftar toko + kelola toko sendiri (nama, deskripsi, logo, banner), disetujui admin (`approve`/`reject`) yang otomatis menaikkan/menurunkan role `buyer`↔`seller`.
- **Category:** CRUD admin hierarkis (proteksi cycle, guard bila punya anak/produk) + baca kategori aktif.
- **Product:** produk dengan varian (SKU, harga, stok), CRUD milik sendiri, foto produk (maks. 10 foto, satu primary otomatis pindah saat dihapus). Semua endpoint domain ini Bearer Auth.
- **Middleware:** JWT stateless + RBAC dasar (`RequireAuth`, `RequireRoles`).
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
| PUT | `/api/v1/users/me/avatar` | Bearer Token |
| DELETE | `/api/v1/users/me/avatar` | Bearer Token |
| POST | `/api/v1/stores` | Bearer Token |
| GET | `/api/v1/stores/me` | Bearer Token |
| PATCH | `/api/v1/stores/me` | Bearer Token |
| PUT | `/api/v1/stores/me/logo` | Bearer Token |
| DELETE | `/api/v1/stores/me/logo` | Bearer Token |
| PUT | `/api/v1/stores/me/banner` | Bearer Token |
| DELETE | `/api/v1/stores/me/banner` | Bearer Token |
| GET | `/api/v1/admin/stores` | Admin |
| PATCH | `/api/v1/admin/stores/:id/approve` | Admin |
| PATCH | `/api/v1/admin/stores/:id/reject` | Admin |
| GET | `/api/v1/categories` | Bearer Token (hanya kategori aktif) |
| GET | `/api/v1/categories/:id` | Bearer Token |
| POST | `/api/v1/categories` | Admin |
| PATCH | `/api/v1/categories/:id` | Admin |
| DELETE | `/api/v1/categories/:id` | Admin |
| GET | `/api/v1/admin/categories` | Admin (termasuk nonaktif) |
| POST | `/api/v1/products` | Bearer Token (toko `active`) |
| GET | `/api/v1/products/me` | Bearer Token |
| GET | `/api/v1/products/:id` | Bearer Token (pemilik atau admin) |
| PATCH | `/api/v1/products/:id` | Bearer Token (pemilik) |
| DELETE | `/api/v1/products/:id` | Bearer Token (pemilik) |
| POST | `/api/v1/products/:id/variants` | Bearer Token (pemilik) |
| PATCH | `/api/v1/products/:id/variants/:variantId` | Bearer Token (pemilik) |
| DELETE | `/api/v1/products/:id/variants/:variantId` | Bearer Token (pemilik) |
| POST | `/api/v1/products/:id/images` | Bearer Token (pemilik) |
| PATCH | `/api/v1/products/:id/images/:imageId` | Bearer Token (pemilik) |
| DELETE | `/api/v1/products/:id/images/:imageId` | Bearer Token (pemilik) |
| GET | `/uploads/avatars/<file>` · `/uploads/stores/<file>` · `/uploads/products/<file>` | Public (file statis) |

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

### `PUT /api/v1/users/me/avatar`

Upload atau ganti foto profil user yang sedang login. Request `multipart/form-data` dengan satu file di field `avatar`. Batas ukuran 2 MB (config `upload.max_avatar_size`); format diizinkan JPEG, PNG, WebP — divalidasi dari isi file (magic bytes), bukan ekstensi/Content-Type kiriman klien. File disimpan sebagai `storage/uploads/avatars/<uuid>.<ext>` dan `avatar_url` berisi path publik relatif (`/uploads/avatars/<uuid>.<ext>`); file lama otomatis dihapus. Cache profil di-invalidasi sehingga `GET /users/me` langsung menampilkan avatar baru. Respons berupa `UserProfileRes` terbaru (bentuk sama seperti `GET /users/me`).

```bash
curl -X PUT http://localhost:8080/api/v1/users/me/avatar \
  -H "Authorization: Bearer <access_token>" \
  -F "avatar=@foto.jpg"
```

Respons sukses:

```json
{
  "success": true,
  "message": "avatar updated successfully",
  "data": {
    "id": "9f9b1f7e-0d49-4f0e-9cc2-6f2d4f6f0b11",
    "email": "user@mail.com",
    "first_name": "Budi",
    "last_name": "Santoso",
    "phone": "6281234567890",
    "avatar_url": "/uploads/avatars/9e4b05bf-0c56-497a-bf96-a7f76e1b16d2.jpg",
    "role": "buyer",
    "is_active": true,
    "created_at": "2026-11-18T08:00:00Z",
    "updated_at": "2026-11-20T08:00:00Z"
  }
}
```

| Status | Penyebab |
|---|---|
| 401 | Token absen, invalid, expired, revoked, atau `user_id` bukan UUID valid |
| 403 | Akun tidak aktif |
| 404 | User tidak ditemukan |
| 413 | Body request melebihi batas Fiber (default 4 MB) |
| 422 | Field `avatar` absen, file > 2 MB, atau format bukan JPEG/PNG/WebP |
| 500 | Gagal menyimpan file atau mengakses database |

### `DELETE /api/v1/users/me/avatar`

Menghapus foto profil user yang sedang login: `avatar_url` di-set `null`, file di disk dihapus, dan cache profil di-invalidasi. Respons berupa `UserProfileRes` terbaru dengan `avatar_url: null` (pesan `"avatar deleted successfully"`).

| Status | Penyebab |
|---|---|
| 401 | Token absen, invalid, expired, revoked, atau `user_id` bukan UUID valid |
| 403 | Akun tidak aktif |
| 404 | User tidak ditemukan, atau user tidak punya avatar (`"avatar not found"`) |
| 500 | Gagal mengakses database |

### Store (`/api/v1/stores`, `/api/v1/admin/stores`)

Alur seller: daftar via `POST /stores` (status `pending`) → admin `PATCH /admin/stores/:id/approve` → role user naik `buyer` → `seller` dan toko menjadi `active`. `PATCH /admin/stores/:id/reject` menurunkan role `seller` → `buyer` dan menolak toko yang belum pernah aktif (`rejected`) atau men-suspend yang sudah aktif (`suspended`). Admin tidak pernah turun role. Approve/reject idempoten untuk status yang sama.

Catatan penting:

- **Role diambil dari claims JWT** — user yang baru di-approve harus **login ulang** agar token membawa role `seller`. Begitu pula admin yang dibuat/diubah langsung via SQL: **login ulang** setelah perubahan role.
- Slug toko dibuat dari nama saat create dan **tidak berubah** saat rename (URL stabil); bentrok slug otomatis diberi sufiks acak.
- Satu user maksimal satu toko (`409` bila daftar lagi).
- `PUT /stores/me/logo` (maks. 1 MB) dan `PUT /stores/me/banner` (maks. 2 MB) menerima multipart field `logo`/`banner`; format divalidasi dari magic bytes (JPEG/PNG/WebP), file lama otomatis dihapus. Upload boleh dilakukan walau toko belum `active`.
- `GET /admin/stores` mendukung pagination (`page`, `limit`) dan filter `status`.

### Category (`/api/v1/categories`, `/api/v1/admin/categories`)

Kategori dikelola admin (create/update/delete) dan dibaca user login. `GET /categories` hanya menampilkan kategori aktif; `GET /admin/categories` termasuk nonaktif. Hierarki via `parent_id` dengan pengecekan cycle saat re-parent; `parent_id: ""` mengembalikan kategori ke root. Level subtree dihitung ulang otomatis saat pindah parent. Delete ditolak `409` bila kategori masih punya subkategori atau produk aktif.

### Product (`/api/v1/products`)

Produk dibuat beserta minimal satu varian (`variants`, maks. 100). Semua operasi tulis hanya untuk **pemilik** (seller pemilik toko); akses seller lain dijawab `404` (bukan `403`, agar tidak membocorkan keberadaan data). Admin bisa membaca produk apa pun.

Aturan bisnis:

- Toko harus `active` untuk create produk (`403` bila `pending`/`suspended`); kategori harus aktif (`422` bila tidak).
- **SKU unik per toko**, bukan global — toko berbeda boleh memakai SKU yang sama (`409` `"SKU already used in this store"` hanya untuk duplikat dalam satu toko).
- Harga dalam rupiah utuh (`int64`, `1` = Rp 1). `min_price`/`max_price` produk dihitung otomatis dari varian aktif dan di-recompute setiap varian ditambah/diubah/dihapus.
- Varian aktif terakhir tidak bisa dihapus (`409`); nonaktifkan dulu atau tambah varian lain.
- `metadata` (produk) dan `attributes` (varian) berupa objek JSON maks. 4 KB; kirim `null` untuk meng-clear.
- Slug produk stabil sejak create (tidak berubah saat rename).
- Foto produk: `POST /products/:id/images` multipart field `image`, maks. 3 MB & 10 foto per produk (config `product.*`), format JPEG/PNG/WebP via magic bytes. Foto pertama otomatis menjadi primary; `PATCH {"is_primary": true}` memindahkan primary (hanya satu per produk); menghapus foto primary otomatis mempromosikan foto dengan `sort_order` terkecil. Menghapus produk menghapus semua varian + foto beserta file di disk.

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
│   ├── pkg/{jwt,otp,cache,storage,slug} # utilitas murni
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
