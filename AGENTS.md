# AGENTS.md — Kontrak Kerja GoCommerce

> Backend API e-commerce multi-vendor · Go 1.27.1 · Fiber v3.5.0 · GORM · PostgreSQL · Redis · SMTP
>
> Dokumen ini adalah kontrak kerja untuk AI agent **dan** contributor manusia. Setiap aturan bisa
> diverifikasi secara objektif. Bila dokumen dan kode berbeda, **kode adalah sumber kebenaran** —
> perbarui dokumen ini, bukan sebaliknya.
>
> Struktur & keputusan desain: [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) · Status fitur &
> technical debt: [PROGRESS.md](./PROGRESS.md).

---

## 1. Peta Repo

| Path | Tanggung jawab |
|---|---|
| `cmd/api/main.go` | Composition root: wiring DI, listen, graceful shutdown |
| `internal/router/` | Registrasi route Fiber + grouping prefix |
| `internal/controller/` | Bind body → service → envelope |
| `internal/service/` | Logika bisnis, orkestrasi Redis/DB/SMTP/token |
| `internal/repository/` | Akses data GORM, query/CRUD |
| `internal/entity/` | Model domain + tag GORM |
| `internal/model/` | DTO request/response + tag `validate` |
| `internal/model/response/` | Envelope `success/message/data` |
| `internal/domain/` | `AppError` + sentinel error |
| `internal/pkg/{jwt,otp}` | Utilitas **murni**, tanpa I/O |
| `infrastructure/app/` | Fiber app, middleware, global error handler, health |
| `infrastructure/config/` | Typed `Config` + Viper loader |
| `infrastructure/datastore/` | Klien GORM & Redis |
| `infrastructure/pkg/mailer/` | Klien SMTP + template embedded |
| `infrastructure/{logger,validation}/` | Setup Logrus JSON, custom validation rule |
| `docs/api/spec-api.yaml` | Kontrak OpenAPI 3.1 |
| `config.yaml.example` | Template konfigurasi (satu sumber key config) |
| `PROGRESS.md` | Status fitur, known issues, backlog |

Dua direktori `pkg` berbeda peran: `internal/pkg/` = utilitas **murni** (hanya stdlib, tanpa
I/O); `infrastructure/pkg/` = utilitas yang **menyentuh dunia luar** (SMTP, embedded template).

---

## 2. Aturan Layering

```
cmd/api ──▶ internal/router ──▶ controller ──▶ service ──▶ repository ──▶ entity
                                   │              │
                                   │              ▼
                                   │        domain · pkg (pure)
                                   ▼
                        infrastructure/*  (GORM · Redis · SMTP · JWT · Viper · Logrus · Validator)
```

| # | Aturan |
|---|---|
| 1 | Dependency satu arah `cmd/api → internal → infrastructure`. Tidak ada panah balik. |
| 2 | `internal/entity/*` hanya import stdlib + GORM tag + UUID. |
| 3 | `internal/domain/*` tidak import layer lain. |
| 4 | Service tidak query DB langsung — semua akses lewat `internal/repository`. |
| 5 | Controller tanpa logika bisnis: bind DTO → panggil service → bungkus respons. |
| 6 | Repository tanpa aturan bisnis: query + mapping error saja. |
| 7 | DTO di `internal/model`, entity di `internal/entity`. Entity tidak pernah jadi response. |
| 8 | `internal/*` tidak import `infrastructure/*` — kontrak di `internal/*`, implementasi di `infrastructure/*`. |

**Pelanggaran yang diketahui:** aturan 8 belum terpenuhi. `internal/service/auth_service.go`
meng-import `*redis.Client` dan `*mailer.Client` secara konkret. Jangan memperburuknya: saat
menambah dependency eksternal ke service, **ekstrak interface kecil** di
`internal/service/ports.go` lebih dulu.

```go
type MailSender interface {
    SendTemplate(ctx context.Context, to []string, subject, templateName string, data any) error
}
```

`infrastructure/pkg/mailer.Client` sudah memenuhi interface ini secara struktural — tidak perlu
mengubah mailer.

---

## 3. Konvensi Kode

| Hal | Konvensi | Contoh |
|---|---|---|
| File | snake_case | `auth_service.go` |
| Package | snake_case, tanpa prefix `go` | `repository`, `entity` |
| Interface | noun + suffix | `AuthService`, `Tokenizer` |
| Implementasi | interface + `Impl` atau lowercase | `authRepository`, `authService` |
| Constructor | `NewXxx` | `NewAuthService` |
| DTO | `<Endpoint><Req\|Res>` | `SignUpReq`, `TokenPairRes` |
| Sentinel error | `ErrXxx` | `ErrNotFound` |
| Context key | typed key, bukan `string` | `type ctxKey string` |

**Error handling**

- Selalu `apperrors.Wrap(err, httpStatus, "pesan untuk client")`. Global error handler hanya
  mengenali `*apperrors.AppError` untuk menentukan status code — `fmt.Errorf("%w")` biasa akan
  jatuh ke 500.
- Detail internal hanya masuk log, tidak pernah ke response.
- `AppError.Code` adalah **HTTP status**, bukan kode error bisnis.

**Jangan pernah**

- `panic()` di luar `cmd/`; `log.Fatal` di luar `cmd/api/main.go` (`Fatal` memanggil `os.Exit(1)`
  di tengah proses).
- `fmt.Println` / `print` — pakai `logrus` yang sudah di-inject.
- Membuang error dengan `_ =` tanpa komentar alasan.
- `map[string]any` sebagai "DTO dinamis" — selalu struct bertipe.
- Menulis secret, credential, atau PII ke kode, log, atau dokumen.

---

## 4. Perintah Verifikasi WAJIB

```bash
gofmt -l .                          # harus kosong
go build ./...                      # harus sukses
go vet ./...                        # harus bersih
go test ./... -race                 # harus lulus
go mod tidy && git diff --exit-code go.mod go.sum
govulncheck ./...                   # tidak boleh ada CVE reachable
```

Status saat ini:

- **Test:** repo belum punya satu pun `*_test.go`. Setiap fitur baru wajib menambahkan test.
- **`go mod tidy` mengubah `go.mod`:** `github.com/golang-jwt/jwt/v5` ada di blok *indirect*
  padahal di-import langsung, dan `gopkg.in/yaml.v3` ada di blok *direct* padahal tidak dipakai.
  Jalankan sekali, commit, lalu jadikan `git diff --exit-code` gate CI.
- **CVE reachable:** `brotli` GO-2026-6888 (via `viper.ReadInConfig`) dan
  `klauspost/compress` GO-2026-6865 (via `gorm.Open`).
  `github.com/gabriel-vasile/mimetype v1.4.15` juga *retracted* upstream.
  Detail dan remediasi: [PROGRESS.md](./PROGRESS.md#known-issues).

---

## 5. Gotcha yang Wajib Diketahui

| Gotcha | Konsekuensi |
|---|---|
| `ctx.Value("request_id").(string)` tanpa comma-ok (`auth_service.go`) | **Panic** di luar HTTP flow (test, worker, goroutine) |
| `context.WithValue` dengan key `string` (`infrastructure/app/fiber.go`) | staticcheck **SA1029**, rawan collision |
| Fiber v3 membungkus bind error jadi `*fiber.Error` | Jalur 422 praktis tidak terpakai di HTTP; banyak input cacat berakhir 400 |
| `Updates` hanya menulis field non-zero (`repository/base.go`) | `IsActive = false` dan `avatar_url = nil` **tidak bisa** di-set |
| `GetAll` tanpa pagination (`repository/base.go`) | Memuat seluruh tabel → OOM. Jangan dipakai di endpoint list |
| `GetByColumn`/`Exists` menerima nama kolom bebas | Bukan SQL injection, tapi typo = runtime error. Whitelist |
| `UpdateIsEnable` dan `repositoryImpl[T]` | Dead code; kolom `is_enable` tidak ada (yang ada `is_active`) |
| `AuthRepository` | Persis `Repository[entity.User]`, tanpa method domain tambahan |
| 7 sentinel error di `internal/domain` | Tidak dipakai & tidak dipetakan → jatuh ke 500 |
| Package `internal/domain` bernama `errors` | Bentrok dengan stdlib; selalu alias `apperrors` |
| Dead config `signup_rate_limit`, `login_rate_limit`, `resend_cooldown` | Tidak dibaca kode mana pun — jangan dianggap aktif |
| TTL `signup:attempts` hanya di-set saat counter == 1 | Counter bisa yatim tanpa TTL → user terkunci permanen |
| Tidak ada `db.Transaction` | Verify-OTP → create user tidak atomik |
| `model.VerifyOTP` tanpa suffix `Req` | Inkonsisten dengan DTO lain |
| `auth.otp_ttl` tidak punya default di kode | Jangan hardcode asumsi 5 menit |

---

## 6. Alur Kerja Menambah Fitur (Vertical Slice)

1. **Baca [PROGRESS.md](./PROGRESS.md)** — cek tabel status, jangan duplikat pekerjaan.
2. **Entity hanya jika fiturnya dikerjakan.** Jangan membuat entity "siap pakai" di muka.
   Bila perlu, tambahkan ke `AutoMigrate` di `infrastructure/datastore/gorm.go`.
3. **DTO** di `internal/model` — struct + tag `validate`.
4. **Repository** — embed `RepositoryImpl[T]`, tambahkan method domain di interface spesifik.
   Jangan menambah query domain di generic.
5. **Service** — validasi ulang (defense-in-depth), normalisasi input, orkestrasi, `apperrors.Wrap`.
6. **Controller** — bind → service → `response.SuccessResponse`. Tanpa logika bisnis.
7. **Route** di `internal/router/` pada grup `/api/v1/<domain>`.
8. **AutoMigrate** — update bila ada entity baru.
9. **Dokumentasi** — `docs/API.md` (endpoint/error), `docs/DATABASE.md` (skema), `docs/ARCHITECTURE.md`
   (bila ada keputusan baru), `docs/api/spec-api.yaml`, `config.yaml.example` (key baru).
10. **PROGRESS.md** — update kolom E/M/R/S/H/RT/D/T dan tabel endpoint.
11. **Test** — `internal/pkg/*` dan `internal/domain/*` wajib; service & handler menyusul.
12. **Verifikasi** — jalankan §4.

### Tier test

| Tier | Cakupan | Wajib |
|---|---|---|
| P0 | `internal/pkg/{jwt,otp}`, `internal/domain` | setiap perubahan |
| P1 | Service dengan dependency di-mock (butuh port interface) | setiap fitur |
| P2 | Handler + router (fiber `app.Test`) | setiap endpoint baru |
| P3 | Integrasi DB/Redis (testcontainer) | sebelum rilis |
| P4 | E2E alur auth | milestone |

---

## 7. Konvensi Error

| Kondisi | HTTP | Cara |
|---|---|---|
| Input tidak valid | 422 | Kembalikan error `Validation.Struct(req)` apa adanya |
| Sudah terdaftar | 409 | `apperrors.Wrap(err, 409, "email already registered")` |
| Kredensial salah | 400 | `apperrors.Wrap(err, 400, "email or password is wrong")` |
| Akun nonaktif / belum verifikasi | 403 | `apperrors.Wrap(err, 403, "account is inactive")` |
| Rate limit / abuse | 429 | `apperrors.Wrap(err, 429, "too many invalid OTP attempts")` |
| Dependency luar gagal | 502 | `apperrors.Wrap(err, 502, "failed to store otp")` |
| Bug internal | 500 | `apperrors.Wrap(err, 500, "failed to lookup user")` |

Urutan branch global error handler (yang pertama cocok menang):

1. `*fiber.Error` → status dari error
2. `*fiber.BindError` → 400
3. `validator.ValidationErrors` → 422 + `data[]`
4. `*apperrors.AppError` → `appErr.Code`
5. `gorm.ErrRecordNotFound` → 404
6. `gorm.ErrDuplicatedKey` → 409 (berdasarkan nama index)
7. `*pgconn.PgError` SQLSTATE `23503` → 409
8. lainnya → 500

`AppError` dicek **setelah** `*fiber.Error`: bila `AppError` membungkus `*fiber.Error`, kode
`AppError` hilang.

---

## 8. Konvensi Redis

| Aturan | Detail |
|---|---|
| Prefix | `signup:` untuk registrasi, `session:` untuk sesi |
| Email di key | Selalu lowercase + trim |
| TTL | Selalu dari config (`auth.otp_ttl`, `jwt.refresh_expiry`) |
| Pipeline | Operasi multi-key atomik pakai `TxPipeline` |
| Konsumsi OTP | Wajib `GetDel`, jangan `Get` lalu `Del` |
| PII | Email jangan dicetak ke log |

| Key | TTL | Isi |
|---|---|---|
| `signup:pending:<email>` | `auth.otp_ttl` | JSON draft registrasi (hash bcrypt) |
| `signup:otp:<email>` | `auth.otp_ttl` | HMAC-SHA256 hex OTP |
| `signup:attempts:<email>` | `auth.otp_ttl` | Counter OTP gagal |
| `signup:resend:<email>` | `auth.otp_ttl` | Counter resend |
| `session:<userID>:<sessionID>` | `jwt.refresh_expiry` | SHA-256 refresh token |

Detail dan temuan: [docs/DATABASE.md §5](./docs/DATABASE.md#5-redis).

---

## 9. Aturan Commit & PR

**Conventional Commits**, scope = area repo:

```
feat(auth): implement refresh token endpoint
fix(auth): guard nil password hash before bcrypt compare
docs(api): add login endpoint to openapi spec
chore(deps): bump fiber to v3.7.0
```

Subject imperatif, ≤ 72 karakter, tanpa titik. Body menjelaskan **kenapa**, bukan **apa**.
Jangan campur refactor dan fitur dalam satu commit.

**Checklist PR**

- [ ] `gofmt -l .` kosong; `go build ./...` & `go vet ./...` bersih
- [ ] `go test ./... -race` lulus, dengan test baru untuk fitur
- [ ] `docs/api/spec-api.yaml` dan `docs/API.md` sinkron bila endpoint berubah
- [ ] `docs/DATABASE.md` diperbarui bila skema berubah
- [ ] `PROGRESS.md` diperbarui
- [ ] Tidak ada secret, credential, atau PII di kode/log/dokumen
- [ ] Schema change → file migration baru, bukan AutoMigrate
- [ ] `go mod tidy` dijalankan dan di-commit

---

## 10. Checklist Sebelum Menutup Task

- [ ] Semua perintah §4 dijalankan dan bersih
- [ ] Tidak ada file `.go` berubah tanpa alasan jelas
- [ ] Tidak ada `TODO`/`FIXME` tanpa komentar alasan
- [ ] Error baru di-wrap `apperrors.Wrap` dengan status code tepat
- [ ] Tidak ada secret, credential, atau API key di kode, log, dokumen
- [ ] `config.yaml` tidak pernah masuk staging area
- [ ] Dokumentasi relevan diperbarui
- [ ] `PROGRESS.md` mencerminkan kondisi sebenarnya, bukan rencana
