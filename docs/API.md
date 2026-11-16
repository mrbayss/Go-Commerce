# API.md — Referensi API GoCommerce

> Referensi endpoint, DTO, dan katalog error. Kontrak mesin ada di [`api/spec-api.yaml`](./api/spec-api.yaml) (OpenAPI 3.1); bila berbeda, **dokumen ini mencerminkan implementasi**.
>
> Arsitektur & alur request: [ARCHITECTURE.md](./ARCHITECTURE.md) · Skema & Redis key: [DATABASE.md](./DATABASE.md).

---

## 1. Dasar

| Aspek | Konvensi |
|---|---|
| Base URL | `http://localhost:8080` (`app.port`) |
| Path prefix | `/api/v1` |
| Content-Type | `application/json` |
| Format tanggal | RFC3339 UTC |
| Envelope sukses | `{ "success": true, "message": "…", "data": { … } }` |
| Envelope error | `{ "success": false, "message": "…", "data"?: [ … ] }` |
| Authorization | `Authorization: Bearer <access_token>` (belum ada middleware JWT — semua publik) |
| Pagination | **Tidak ada** |

Tidak ada field `code` numerik di body. HTTP status adalah satu-satunya diskriminasi; `AppError.Code` menyimpan HTTP status dan tidak diserialisasi.

---

## 2. Daftar Endpoint

| Method | Path | Handler | Auth | Sukses | Error |
|---|---|---|---|---|---|
| GET | `/api/v1/health` | `app.HealthCheck` | publik | 200 | — |
| POST | `/api/v1/auth/signup` | `AuthController.SignUp` | publik | 200 | 409, 422, 500, 502 |
| POST | `/api/v1/auth/verify-otp` | `AuthController.VerifyOTP` | publik | 200 | 400, 409, 429, 500, 502 |
| POST | `/api/v1/auth/resend-otp` | `AuthController.ResendOTP` | publik | 200 | 400, 422, 429, 500, 502 |
| POST | `/api/v1/auth/login` | `AuthController.Login` | publik | 200 | 400, 403, 422, 500, 502 |

Registrasi route: `internal/router/auth_router.go`. Route health didaftarkan langsung di `cmd/api/main.go`.

**Belum ada:** refresh-token, logout/revoke, RBAC, `me`, forgot/reset password, pagination. Backlog lengkap: [../PROGRESS.md](../PROGRESS.md#backlog--open-questions).

---

## 3. Validasi & Format Error

Pemetaan tag → pesan via `FormatValidationError` di `infrastructure/validation/validator.go`.

| Tag | Aturan |
|---|---|
| `required` | wajib diisi |
| `email` | format valid, `max=255` |
| `min` / `max` | panjang minimum / maksimum |
| `notblank` | tidak boleh hanya whitespace |
| `password` | upper + lower + digit + simbol, tanpa whitespace |
| `id_phone` | `^628[1-9][0-9]{7,11}$` (10–13 digit, awalan `628`) |
| `omitempty` | boleh kosong; aturan lain dilewati |

**Gotcha:** Fiber v3 membungkus body-bind error sebagai `*fiber.Error`, bukan `validator.ValidationErrors`. Jalur 422 praktis mati di HTTP; 422 hanya lewat `Validation.Struct` di service.

---

## 4. Detail Endpoint

### 4.1 POST /api/v1/auth/signup

Membuat sesi signup di Redis, mengirim OTP 6 digit ke email, dan **belum** membuat row `users`.

| Field | Tipe | Wajib | Validasi | Catatan |
|---|---|---|---|---|
| `email` | string | ya | `required,email,max=255` | lowercase + trim |
| `password` | string | ya | `required,min=8,max=72,password` | `max=72` rune, bcrypt batasi 72 byte |
| `first_name` | string | ya | `required,notblank,min=1,max=100` | — |
| `last_name` | string | ya | `required,notblank,min=1,max=100` | nullable di DB |
| `phone` | string | tidak | `omitempty,id_phone` | `628…` |

Response: `{ "success": true, "message": "success signup, verify otp send to your mail", "data": { "email": "...", "expired_at": "…" } }`

| HTTP | Pesan | Pemicu |
|---|---|---|
| 409 | `email already registered` | email sudah ada |
| 409 | `phone number already registered` | phone sudah ada |
| 422 | `validation failed` + `data[]` | input tidak valid |
| 500 | `failed to lookup user`, `failed to hashing password`, `failed generated otp` | bug internal |
| 502 | `failed to store signup session`, `failed to store otp` | Redis error |

**Efek samping:** `signup:pending:<email>` + `signup:otp:<email>` dibuat (TTL `auth.otp_ttl`); `signup:attempts`, `signup:resend` dihapus. Email dikirim async.

### 4.2 POST /api/v1/auth/verify-otp

| Field | Tipe | Wajib | Validasi |
|---|---|---|---|
| `code_otp` | string | ya | `required,max=6` |
| `email` | string | ya | `required,email,max=255` |

Response: `{ "success": true, "message": "OTP verified successfully", "data": { "access_token": "...", "refresh_token": "...", "token_type": "Bearer", "expires_in": 900 } }`

| HTTP | Pesan | Pemicu |
|---|---|---|
| 400 | `OTP expired or invalid` | kode salah / expired / key tidak ada |
| 400 | `invalid OTP code` | HMAC mismatch |
| 400 | `signup session expired` | `signup:pending:` hilang |
| 409 | `email already registered` | INSERT duplicate key |
| 429 | `too many invalid OTP attempts` | counter ≥ 5 → semua key signup dihapus |
| 500 | `failed to create user`, `failed to process signup` | bug internal |
| 502 | `failed to verify otp`, `failed to store session` | Redis / DB error |

**Efek samping:** OTP via `GETDEL` (atomik, single-use); 1 row `users` di-INSERT; 4 key signup dihapus; session `session:<userID>:<sessionID>` dibuat (SHA-256 refresh, TTL `jwt.refresh_expiry`).

### 4.3 POST /api/v1/auth/resend-otp

| Field | Tipe | Wajib | Validasi |
|---|---|---|---|
| `email` | string | ya | `required,email,max=255` |

Response: `{ "success": true, "message": "OTP resent successfully", "data": { "email": "...", "expired_at": "…" } }`

| HTTP | Pesan | Pemicu |
|---|---|---|
| 400 | `no pending signup found, please signup first` | `signup:pending:` tidak ada |
| 422 | `validation failed` | input tidak valid |
| 429 | `maximum resend attempts reached` | `signup:resend` ≥ `auth.max_resend_per_session` (3) |
| 500 | `failed generated otp` | bug internal |
| 502 | `failed to check signup session`, `failed to check resend limit`, `failed to store otp` | Redis error |

**Efek samping (TxPipeline):** `SET signup:otp` + `EXPIRE pending` + `INCR resend` + `EXPIRE resend` + `DEL attempts`. Email dikirim async.

**Tidak ada cooldown.** `auth.resend_cooldown` ada di config tetapi tidak dibaca kode mana pun.

### 4.4 POST /api/v1/auth/login

| Field | Tipe | Wajib | Validasi |
|---|---|---|---|
| `email` | string | ya | `required,email,max=255` |
| `password` | string | ya | `required` |

Response: `{ "success": true, "message": "success login", "data": { "access_token": "...", "refresh_token": "...", "token_type": "Bearer", "expires_in": 900 } }`

| HTTP | Pesan | Pemicu |
|---|---|---|
| 400 | `email or password is wrong` | email tidak ditemukan atau password salah |
| 403 | `account is inactive` | `is_active = false` |
| 403 | `account is not verified` | `is_verified = false` |
| 422 | `validation failed` | input tidak valid |
| 502 | `failed to store session` | Redis error |

**Efek samping:** read-only pada DB; membuat session baru di Redis.

**Celah:** tanpa dummy bcrypt compare → user enumeration via timing. `password_hash` nullable → 500. Lihat [../PROGRESS.md](../PROGRESS.md#known-issues).

### 4.5 GET /api/v1/health

Response: `{ "status": "ok", "components": { "database": "ok", "redis": "ok" } }`

- `status = "degraded"` bila DB atau Redis gagal, tapi HTTP tetap 200. Tidak layak sebagai readiness/liveness probe.
- Timeout 2 s via `context.Background()` (bukan request context).
- Shape tidak mengikuti envelope standar.

---

## 5. JWT

`internal/pkg/jwt/jwt.go` — HS256.

| Parameter | Nilai |
|---|---|
| Algoritma | HS256, `exp` wajib |
| Secret | `jwt.secret`, minimal 32 byte (divalidasi saat boot) |
| Access TTL | `jwt.access_expiry` → `expires_in: 900` |
| Refresh TTL | `jwt.refresh_expiry` (168 h) |
| Issuer | `jwt.issuer` (`gocommerce`), opsional |

Access token **stateless**. Refresh token disimpan sebagai SHA-256 hex di `session:<userID>:<sessionID>`. Tidak ada blacklist, tidak ada revoke.

---

## 6. DTO

| Struct | Field | Arah |
|---|---|---|
| `SignUpReq` | email, password, first_name, last_name, phone | request |
| `SignUpRes` | email, expired_at | response |
| `VerifyOTP` | code_otp, email | request |
| `ResendOTPReq` | email | request |
| `ResendOTPRes` | email, expired_at | response |
| `LoginReq` | email, password | request |
| `TokenPairRes` | access_token, refresh_token, token_type, expires_in | response |

Semua di `internal/model/auth_model.go`. `VerifyOTP` tidak memakai suffix `Req` seperti DTO lain.

Envelope: `internal/model/response/response.go`

---

## 7. Alur End-to-End

```bash
# 1. Signup — OTP dikirim ke email
curl -s -X POST http://localhost:8080/api/v1/auth/signup -H 'Content-Type: application/json' \
  -d '{"email":"budi@example.com","password":"RahasiaKuat1!","first_name":"Budi","last_name":"Santoso"}'

# 2. Verify OTP — dapat token pair
curl -s -X POST http://localhost:8080/api/v1/auth/verify-otp -H 'Content-Type: application/json' \
  -d '{"email":"budi@example.com","code_otp":"123456"}'

# 3. Login
curl -s -X POST http://localhost:8080/api/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"budi@example.com","password":"RahasiaKuat1!"}'
```

Logout belum tersedia — token berlaku sampai TTL habis.

---

## 8. Drift vs `spec-api.yaml`

Implementasi adalah kebenaran; spec belum disinkronkan.

| # | Item | Spec | Implementasi |
|---|---|---|---|
| 1 | `POST /auth/login` | tidak ada | ada |
| 2 | Path health | `/health` | `/api/v1/health` |
| 3 | 429 signup rate limit | ada | tidak ada |
| 4 | 429 resend cooldown (60 s) | ada | tidak ada |
| 5 | 429 verify-otp (5 percobaan) | tidak ada | ada |
| 6 | Pesan error | versi spec | berbeda di kode |
| 7 | 502 `failed to send otp to email` | ada | tidak reachable (email async) |
| 8 | `components.securitySchemes` | tidak ada | API menerbitkan token |
| 9 | `operationId` | tidak ada | — |
| 10 | TTL OTP "5 minutes" | 5 m | dari config, tidak ada default |
| 11 | Aturan password | boleh whitespace | tanpa whitespace |
| 12 | Komponen `RateLimited` | didefinisikan | tidak dipakai (`$ref` 0) |
| 13 | `StandardResponse.required` | tidak ada | `success` & `message` selalu ada |

Rencana: sinkronkan spec sekali jalan (path health, tambah login, pesan error, hapus 429 yang tidak ada, tambah `securitySchemes` + `operationId`).