# GoCommerce — Progress Log

> Backend e-commerce API · Go 1.27.1 · Fiber v3.5.0 · GORM · PostgreSQL · Redis · SMTP.
> Arsitektur: Light Clean Architecture (`controller → service → repository → entity`).

---

## Project Info

| Item | Value |
|---|---|
| Module | `github.com/mrbayss/Go-Commerce` |
| Go | 1.27.1 |
| Fiber | v3.5.0 |
| GORM | v1.31.2 (postgres driver v1.6.3, pgx v5.10.0) |
| Redis | v9.23.0 (go-redis) |
| Viper | v1.21.0 |
| Logrus | v1.10.2 |
| Validator | v10.30.5 |
| go-mail | v0.8.1 |
| golang-jwt | v5.3.1 |
| File Go | 23 file, 1.871 baris, **0 test** |

Konsep kerja: entity **hanya dibuat saat fiturnya dikerjakan**. Setiap fitur = satu slice vertikal,
status dicatat di tabel di bawah.

Legenda tahap: **E** entity · **M** migration/AutoMigrate · **R** repository · **S** service ·
**H** handler · **RT** route · **D** API spec · **T** test.

---

## Struktur (aktual)

```
GoCommerce/
├── cmd/api/main.go
├── internal/{controller,service,repository,entity,model,domain,pkg,router}
├── infrastructure/{app,config,datastore,logger,validation,pkg/mailer}
├── docs/{API,ARCHITECTURE,DATABASE}.md + api/spec-api.yaml
├── config.yaml.example (template)   config.yaml (gitignored)
└── PROGRESS.md
```

Detail: [README.md §Struktur Proyek](./README.md#struktur-proyek).

---

## Status Fitur

| Fitur | E | M | R | S | H | RT | D | T | Status |
|---|---|---|---|---|---|---|---|---|---|
| Infrastructure (config, logger, error, health) | – | – | – | – | – | ✓ | – | ✗ | selesai |
| Auth (SignUp + Verify OTP + Resend OTP + Login) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | selesai |
| Auth — Refresh token | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | selesai |
| Auth — Logout / revoke session | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | selesai (kodim, belum test) |
| Auth — JWT middleware + RBAC | – | – | – | – | ✗ | ✗ | ✗ | ✗ | belum |
| Auth — Forgot/reset password | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |
| Profile + Address, Store, Category + Product, Cart, Order, Payment, Shipment, Voucher, Wallet, Campaign, Review, Chat, Notification | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | belum |

### Endpoint aktif

| Method | Path | Handler |
|---|---|---|
| GET | `/api/v1/health` | `app.HealthCheck` |
| POST | `/api/v1/auth/signup` | `AuthController.SignUp` |
| POST | `/api/v1/auth/verify-otp` | `AuthController.VerifyOTP` |
| POST | `/api/v1/auth/resend-otp` | `AuthController.ResendOTP` |
| POST | `/api/v1/auth/login` | `AuthController.Login` |
| POST | `/api/v1/auth/refresh` | `AuthController.RefreshToken` |

---

## Keputusan Desain

| Decision | Rationale |
|---|---|
| Light Clean (tanpa use-case layer) | Layer jelas, testable, tanpa over-abstraction |
| Entity per fitur | Entity yang belum dipakai tidak bisa direview dan jadi "progress palsu" |
| `infrastructure/` vs `internal/pkg/` | infrastructure = I/O eksternal; `internal/pkg` = utilitas murni (jwt, otp) |
| Typed `Config` | Type-safe, typo terdeteksi, tidak ada `viper.GetString` di call site |
| OTP disimpan di Redis | Ephemeral; row user baru dibuat setelah OTP terverifikasi |
| Generic `Repository[T]` | CRUD reusable; repo spesifik embed untuk method domain |
| Manual DI | Constructor injection; pertimbangkan `wire`/`fx` kalau dependency > 10 |
| AutoMigrate terbatas entity aktif | Mencegah tabel kosong dan schema drift |

> **Catatan:** rate limit di Redis **belum diimplementasi** (config dead). Jangan anggap
> keputusan ini aktif.

---

## Known Issues

Severity: **K** kritis · **H** tinggi · **M** sedang · **L** rendah. Lokasi relatif ke repo root.
Daftar lengkap tersimpan di `git show HEAD~:docs/SECURITY.md`.

| ID | Severity | Lokasi | Masalah | Fix singkat |
|---|---|---|---|---|
| K1 | K | `config.yaml.example:41` | JWT secret placeholder | Rotasi + pisahkan OTP secret |
| K2 | K | `auth_service.go:364-374` | Timing attack + user enumeration di login | Dummy bcrypt compare |
| K3 | K | `configuration.go:49-63` | Rate limit signup/login dead config | Implementasi counter Redis |
| K4 | K | `gorm.go:60-64` | AutoMigrate jalan di production | `auto_migrate: false` di prod |
| K5 | K | `entity/user.go:16` | `password_hash` NULL → panic | `NOT NULL` + guard `nil` |
| K6 | K | `go.mod` | 2 CVE reachable + 1 module retracted | `go get` + tidy |
| H1 | H | `auth_service.go:436` | `ctx.Value("request_id").(string)` tanpa comma-ok → panic | Typed context key |
| H2 | H | `fiber.go:33` | `context.WithValue` dengan key `string` (SA1029) | Typed key `ctxKey string` |
| H3 | H | `fiber.go:40-43` | CORS wildcard `*` | Whitelist origin per env |
| H4 | H | `fiber.go:23-29` | Fiber tanpa BodyLimit/timeout | Set timeout |
| M1 | M | `base.go:63-73` | `GetAll` tanpa pagination → OOM | Tambah `LIMIT`/cursor |
| L3 | L | `auth_service.go:204-206` | TTL `signup:attempts` hanya saat counter == 1 → yatim | `INCR` + `EXPIRE` dalam pipeline |

Tambahan: H5, H6, H7, H9–H16, M2–M10, L1–L2, L4–L12 (semua di `git show HEAD~:docs/SECURITY.md`).

---

## Backlog & Open Questions

### Backlog teknis (prioritas tinggi)

- **P0** Refresh token + rotasi
- **P1** Logout / revoke session + blacklist
- **P2** JWT middleware + RBAC
- **P2** Forgot/reset password
- **P1** Versioned migration (Atlas) + `auto_migrate: false` di prod
- **P0** Test P0 (`internal/pkg/{jwt,otp}`, `internal/domain`), lalu P1–P4
- **P2** Port interface Redis/Mailer ke `internal/service/ports.go`
- **P2** Perbaikan DSN, typed context key, bcrypt 72-byte, dummy compare

### Open Questions (dari PRD, belum diputuskan)
- Payment gateway mana? (Stripe/Midtrans/Xendit?)
- Model split order & refund (1 order, 3 toko?)
- Siapa penanggung pajak (seller/platform) & PPh?
- Cara verifikasi seller (badge/dokumen)?
- Kategori produk & skema price?
- Cache strategy katalog (cache-aside + invalidation)?
- Seberapa penting timezone & format waktu?
- Whitelist origin CORS di semua environment?
- Redis (OTP & session) shared antar node/cluster/persistent?
- Webhook harus di-signature (HMAC) & di-resend?

---

## Known Drift vs spec-api.yaml

Ringkas: lihat [docs/API.md §8](./docs/API.md#8-drift-vs-spec-apiyaml) (13 item).

---

> Riwayat desain entity lama (store, product, order, wallet, dll.) disimpan di git tag
> `wip-domain-sketch` untuk referensi saat fiturnya dikerjakan.

*Last updated: 2026-11-16 — konsolidasi dokumen + simplify tables (Status Fitur & Known Issues).*

**Simpanan:** 
- Status Fitur ringkas (gabung Auth)
- Known Issues ringkas
- Backlog diurutkan prioritas (P0/P1/P2)
