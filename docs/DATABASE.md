# DATABASE.md — Database & Redis GoCommerce

> Kondisi nyata: PostgreSQL dengan satu tabel (`users`) dan Redis sebagai state store OTP/session.
> Desain domain e-commerce yang belum dikerjakan disimpan di git tag `wip-domain-sketch` (lihat
> [ADR-007](./ARCHITECTURE.md#adr-007--pengetahuan-domain-disimpan-di-git-tag-wip-domain-sketch)).
>
> Arsitektur: [ARCHITECTURE.md](./ARCHITECTURE.md) · Endpoint & error: [API.md](./API.md).

---

## 1. Ringkasan

| Aspek | Nilai | Sumber |
|---|---|---|
| Database | PostgreSQL 13+ | — |
| Driver | `gorm.io/driver/postgres` v1.6.3 (transport `pgx/v5` v5.10.0) | `go.mod` |
| ORM | GORM v1.31.2, `TranslateError: true` | `infrastructure/datastore/gorm.go` |
| State store | go-redis/v9 v9.23.0 | `infrastructure/datastore/redis.go` |
| Tabel aktif | `users` saja | `internal/entity/user.go` |
| Skema | `AutoMigrate` saat boot, bila `postgres.auto_migrate = true` | `infrastructure/datastore/gorm.go` |
| Versioned migration | belum ada | — |
| Transaksi | belum ada `db.Transaction` di seluruh repo | — |

Seluruh tabel e-commerce (store, product, cart, order, payment, shipment, voucher, wallet,
review, chat, campaign, notification) tidak pernah dibuat. Entity dipangkas di commit
`a850dce`; desain via `git show wip-domain-sketch:<path>`.

---

## 2. Koneksi

### 2.1 Config

Bentuk lengkap ada di [`config.yaml.example`](../config.yaml.example) dan
`infrastructure/config/configuration.go` — sengaja tidak diduplikasi di sini.

| Key | Default example | Catatan |
|---|---|---|
| `postgres.host` / `port` / `user` / `db` | `localhost` / `5432` / `postgres` / `gocommerce` | — |
| `postgres.pass` | `your_password_here` | rahasia, isi via env |
| `postgres.sslmode` | `disable` | hanya dev lokal |
| `postgres.auto_migrate` | `true` | **wajib `false` di production** |
| `postgres.max_open_conns` | `25` | `0` = unlimited (`database/sql`) |
| `postgres.max_idle_conns` | `10` | jaga `≤ max_open_conns` |
| `postgres.conn_max_lifetime` | `5m` | kalibrasi load balancer |
| `postgres.conn_max_idle_time` | `1m` | — |

Pool diterapkan di `gorm.go` via `sqlDB.SetMaxOpenConns` dan kawan-kawan. Kegagalan koneksi saat
boot memicu `panic` (proses mati).

### 2.2 DSN — perlu perbaikan

DSN dirakit manual sebagai key/value:

```go
dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", ...)
```

Format ini **tidak melakukan escaping**. Password yang memuat spasi, `'`, `"`, `\`, atau `=`
membuat parsing rusak. Perbaikan: pakai URL (`postgres://user:pass@host:port/db?sslmode=…`
dengan `url.QueryEscape`), atau `pgxpool.ParseConfig` lalu
`gorm.Open(postgres.New(postgres.Config{Conn: pool}))`.

Nilai default `postgres.pass` di config contoh **bukan** credential; jangan pernah menulis
nilai nyata ke dokumen.

---

## 3. Skema `users`

DDL setara dengan hasil `AutoMigrate` terhadap `entity.User`:

```sql
CREATE TABLE "users" (
    "id"            uuid         DEFAULT gen_random_uuid(),
    "created_at"    timestamptz,
    "updated_at"    timestamptz,
    "deleted_at"    timestamptz,
    "email"         varchar(255) NOT NULL,
    "password_hash" varchar(255),
    "first_name"    varchar(100) NOT NULL,
    "last_name"     varchar(100),
    "phone"         varchar(20),
    "avatar_url"    varchar(500),
    "role"          varchar(20)  NOT NULL DEFAULT 'buyer',
    "is_active"     boolean      NOT NULL DEFAULT false,
    "is_verified"   boolean      NOT NULL DEFAULT false,
    PRIMARY KEY ("id")
);
CREATE INDEX        IF NOT EXISTS "idx_users_deleted_at" ON "users" ("deleted_at");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_email"      ON "users" ("email") WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_phone"      ON "users" ("phone") WHERE deleted_at IS NULL;
```

| Kolom | Go | SQL | Null | Default | Index |
|---|---|---|---|---|---|
| `id` | `uuid.UUID` | `uuid` | PK | `gen_random_uuid()` | `users_pkey` |
| `created_at` / `updated_at` | `time.Time` | `timestamptz` | ya | diisi GORM | — |
| `deleted_at` | `gorm.DeletedAt` | `timestamptz` | ya | `NULL` | `idx_users_deleted_at` |
| `email` | `string` | `varchar(255)` | tidak | — | `idx_users_email` (partial unique) |
| `password_hash` | `*string` | `varchar(255)` | ya | `NULL` | — |
| `first_name` | `string` | `varchar(100)` | tidak | — | — |
| `last_name` | `string` | `varchar(100)` | ya | `NULL` | — |
| `phone` | `*string` | `varchar(20)` | ya | `NULL` | `idx_users_phone` (partial unique) |
| `avatar_url` | `*string` | `varchar(500)` | ya | `NULL` | — |
| `role` | `UserRole` | `varchar(20)` | tidak | `'buyer'` | — |
| `is_active` | `bool` | `boolean` | tidak | `false` | — |
| `is_verified` | `bool` | `boolean` | tidak | `false` | — |

Hal yang perlu diketahui:

1. `created_at`/`updated_at` tanpa `DEFAULT` di DB; diisi GORM (`autoCreateTime`). Insert lewat
   psql/loader wajib mengisi eksplisit.
2. `last_name` & `password_hash` nullable. `password_hash` nullable adalah celah: baris tanpa
   password memicu 500 saat login.
3. `role` bukan enum DB; hanya konstanta Go, tidak ada `CHECK` constraint.
4. Tidak ada audit trail (`updated_by`, `version`, `deleted_by`).
5. Email/phone unik secara case-sensitive di DB. Normalisasi lowercase + trim hanya di service,
   sehingga jalur non-service bisa menyisipkan duplikat logis.

### 3.1 Mengapa unique index harus *partial*

Tanpa predikat `WHERE deleted_at IS NULL`, baris soft-deleted memonopoli email/phone — akun yang
dihapus tidak bisa didaftarkan ulang. Dengan partial index, hanya baris aktif yang masuk index.

Konsekuensi:
- Query wajib menyertakan `deleted_at IS NULL` agar index terpakai. GORM melakukannya otomatis
  untuk model dengan `gorm.DeletedAt`; raw SQL tanpa filter akan *seq scan*.
- Di PostgreSQL `NULL != NULL`, jadi banyak user dengan `phone = NULL` tetap koeksis — alasan
  `phone` boleh `UNIQUE` tanpa `NOT NULL`.

### 3.2 Nama index di-hardcode

Error duplikat dipetakan ke pesan HTTP dengan mencocokkan nama index:

- `infrastructure/app/fiber.go` — `idx_users_email` → `email already registered`,
  `idx_users_phone` → `phone number already registered`.
- `internal/service/auth_service.go` — pencocokan serupa memakai `strings.Contains`.

`gorm.ErrDuplicatedKey` aktif karena `TranslateError: true` (SQLSTATE `23505`).

> **Aturan migrasi:** pertahankan nama `idx_users_email` dan `idx_users_phone`. Jika skema
> diubah menjadi named constraint (`users_email_key`), pencocokan gagal diam-diam dan API
> kembali mengembalikan `duplicate entry`.

---

## 4. Entity & AutoMigrate

`internal/entity/entity.go` — `BaseModel`:

| Field | Tag GORM | Efek |
|---|---|---|
| `ID uuid.UUID` | `type:uuid;primaryKey;default:gen_random_uuid()` | PK di-generate DB |
| `CreatedAt` / `UpdatedAt` | `autoCreateTime` / `autoUpdateTime` | diisi aplikasi |
| `DeletedAt gorm.DeletedAt` | `index` | soft delete + index |

`AutoMigrate` hanya memigrasikan `&entity.User{}`:

```go
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(&entity.User{})
}
```

Kebijakan **entity per fitur** (ADR-003): menambah baris di `AutoMigrate` langsung menciptakan
tabel. AutoMigrate berjalan setiap boot sebelum Fiber listen; kegagalan memicu `log.Fatal`.

Yang perlu dipahami tentang AutoMigrate:

- **Additive-only** — menambah kolom/index, tidak pernah menghapus atau mengubah tipe.
- **Rename struct = tabel baru.** Snake_case + pluralisasi (`User` → `users`, `StoreOrder` → `store_orders`).
- **Index dibandingkan berdasarkan nama saja** — perubahan definisi (predicate, kolom) tidak terdeteksi.
- **Tidak transaksional** — DDL satu per satu; kegagalan di tengah meninggalkan skema separuh.
- **Race multi-instance** — dua instance boot bersamaan bisa gagal `already exists` → crash loop.
- **Tidak ada audit skema** — tidak ada `schema_migrations`, tidak ada file migration.

---

## 5. Redis

Redis adalah **state store**, bukan cache data domain. Klien tanpa TLS, tanpa namespace prefix,
DB index dari `redis.db`.

### 5.1 Key aktif

| Key | Tipe | TTL | Isi |
|---|---|---|---|
| `signup:pending:<email>` | string (JSON) | `auth.otp_ttl` | draft registrasi: email, hash bcrypt, nama, phone |
| `signup:otp:<email>` | string (hex) | `auth.otp_ttl` | HMAC-SHA256 dari `email + ":" + code` — bukan OTP polos |
| `signup:attempts:<email>` | counter | `auth.otp_ttl` | jumlah percobaan OTP gagal (batas 5) |
| `signup:resend:<email>` | counter | `auth.otp_ttl` | jumlah resend (batas `auth.max_resend_per_session`, 3) |
| `session:<userID>:<sessionID>` | string (hex) | `jwt.refresh_expiry` (168 h) | SHA-256 hex refresh token |

### 5.2 Aturan

| Aturan | Detail |
|---|---|
| Normalisasi | Email selalu `strings.ToLower(strings.TrimSpace(...))`. Redis case-sensitive; tanpa ini `John@x.com` dan `john@x.com` jadi dua key dan bisa bypass counter |
| Konsumsi OTP | Wajib `GetDel` — atomik dan single-use |
| Multi-key | `TxPipeline` untuk update counter + TTL saat resend |
| TTL | Selalu dari config (`auth.otp_ttl`, `jwt.refresh_expiry`), tidak di-hardcode |
| PII | Email adalah PII dan menjadi bagian nama key — jangan cetak key penuh ke log |

### 5.3 Temuan

1. **TTL `signup:attempts` hanya di-set saat counter == 1.** Key bisa yatim tanpa TTL → user
   terkunci permanen setelah 5 percobaan. Perbaikan: `INCR` + `EXPIRE` tanpa syarat dalam
   pipeline (EXPIRE idempotent).
2. **`session:*` tidak pernah dihapus.** Belum ada logout/revoke; expiry satu-satunya TTL 168 jam.
   Refresh token yang bocor tetap valid 7 hari.
3. **Tidak ada namespace per environment.** Dua environment yang berbagi instance Redis akan
   saling menimpa key signup. Rekomendasi: prefix `redis.key_prefix`.
4. **Tidak ada TLS.** Aman untuk jaringan privat; wajib `TLSConfig` untuk managed Redis.
5. **Tidak ada rate limit Redis.** `auth.signup_rate_limit`, `auth.login_rate_limit`,
   `auth.resend_cooldown` ada di config tetapi tidak dibaca kode mana pun. Batas aktif hanya
   `signup:attempts` (5×) dan `signup:resend` (3×) per email.
6. **Tidak ada lock/idempotency key** saat verify-otp.
7. **`maxmemory-policy` dikonfigurasi di server, bukan aplikasi.** `allkeys-lru` bisa evict OTP/
   session → logout misterius atau OTP invalid.

---

## 6. Query

Semua query berasal dari `Repository[T]` (`internal/repository/base.go`).

| Method | SQL representatif | Index |
|---|---|---|
| `Exists("email", v)` | `SELECT count(*) … WHERE email = $1 AND deleted_at IS NULL` | `idx_users_email` |
| `Exists("phone", v)` | idem untuk `phone` | `idx_users_phone` |
| `GetByColumn("email", v)` | `SELECT * … WHERE email = $1 AND deleted_at IS NULL ORDER BY id LIMIT 1` | `idx_users_email` |
| `GetByID(id)` | `SELECT * … WHERE id = $1 AND deleted_at IS NULL` | `users_pkey` |
| `Create(u)` | `INSERT … RETURNING "id"` | unique check email/phone |
| `Delete(id)` | `UPDATE … SET deleted_at = $1 WHERE id = $2` | `users_pkey` |
| `Updates(u, id)` | `UPDATE … SET <field non-zero> WHERE id = $1` | `users_pkey` |
| `GetAll()` | `SELECT * … WHERE deleted_at IS NULL` | **tidak ada** — seq scan |

Batasan yang wajib diingat:

- **`GetAll` tanpa `LIMIT`.** Jangan dipakai di jalur request; tabel besar bisa OOM.
- **`Updates` hanya menulis field non-zero.** GORM melewati zero value: `IsActive = false` dan
  `AvatarURL = nil` tidak bisa di-set. Pakai `UpdateColumns` atau `map[string]any`. Risiko mass
  assignment (`Role`, `IsActive`, `PasswordHash`).
- **`GetByColumn`/`Exists` menerima nama kolom bebas.** Aman selama literal dari kode, tapi harus
  di-whitelist sebelum menerima input dinamis.
- **Tidak ada `Preload`/`Joins`/`Raw`.** Belum jadi masalah karena belum ada relasi; tambahkan
  API preload sebelum fitur order agar tidak N+1.
- **`RepositoryImpl.UpdateIsEnable`** adalah dead code — kolom `is_enable` tidak ada (yang ada
  `is_active`); method tidak ada di interface. Kandidat dihapus.

---

## 7. Konsistensi & Transaksi

Tidak ada `db.Transaction` di repo. Alur verify-otp menyentuh dua sistem:

```
1. GET    signup:otp:<email>        baca hash
2. verify HMAC (constant-time)
3. GETDEL signup:otp:<email>        konsumsi OTP (atomik)
4. GET    signup:pending:<email>    baca draft
5. INSERT users                     buat user
6. DEL    signup:pending|otp|attempts|resend
7. SET    session:<userID>:<sessionID>
```

| Risiko | Dampak | Status |
|---|---|---|
| INSERT gagal setelah `GETDEL` | OTP hilang, user harus mengulang signup | belum tertangani |
| `DEL` Redis gagal setelah user dibuat | `resend-otp` masih bisa membuat OTP baru untuk email yang sudah terdaftar | error diabaikan |
| `SET session` gagal | akun ada, client tanpa token — bisa login ulang | recoverable |
| Double verify bersamaan | `GETDEL` memastikan hanya satu menang | tertangani atomik |
| TOCTOU `Exists` → `Create` | unique index partial menjadi arbiter akhir → 409 | tertangani |

Pola yang disarankan: **consume-last** — konsumsi OTP setelah INSERT berhasil, plus guard
idempotensi (`SET NX EX` atau kolom `signup_token uuid` unik). Untuk multi-write ke depan,
bungkus dalam `db.Transaction` dan tambahkan `WithTx(tx)` di repository agar `ctx` tidak hilang.

---

## 8. Strategi Migrasi

`AutoMigrate` bukan strategi jangka panjang (ADR-006). Rencana:

1. **`auto_migrate: false` di production**, dengan guard saat boot.
2. **Migration sebagai release step**, bukan saat start — `cmd/migrate` menjalankan tool
   versioned (Atlas sebagai provider resmi GORM) sebelum deployment.
3. **Satu file per versi, immutable.** File yang sudah dipakai tidak diedit; koreksi lewat file baru.
4. **Expand / contract** untuk perubahan destruktif: tambah kolom nullable → dual write →
   backfill batch → pasang constraint → hapus kolom lama pada rilis berikutnya.
5. **`CREATE INDEX CONCURRENTLY`** untuk tabel besar, di file non-transaksional
   (`-- atlas:txmode none`).
6. **Backup + timeout** sebelum migrasi produksi: `SET lock_timeout = '3s'`, `SET statement_timeout = '60s'`.
7. **Drift detector di CI** — `atlas schema diff` antara entity Go dan database container, fail
   bila berbeda.

Detail: <https://atlasgo.io/guides/orms/gorm>.

---

## 9. Retensi & Cleanup

| Data | Batas sekarang | Masalah | Aksi |
|---|---|---|---|
| `session:*` | TTL 168 h | tidak pernah dihapus saat logout (belum ada logout); key menumpuk | endpoint logout + `SET sessions:user:<id>` untuk revoke massal, batasi sesi aktif per user |
| `signup:*` | TTL `auth.otp_ttl` | ok; residual hanya bila `DEL` gagal | pertahankan TTL |
| `signup:attempts` | TTL hanya saat counter == 1 | bisa yatim tanpa TTL | `INCR` + `EXPIRE` dalam pipeline |
| `users` soft-deleted | tidak ada batas | baris menumpuk selamanya | purge berkala (`deleted_at < now() - 90 days`) setelah kebutuhan audit jelas |
| Audit log | tidak ada | perubahan `role`/`is_active` tanpa jejak | tabel `audit_logs` (actor, action, entity, before/after JSONB) |

Pisahkan instance Redis untuk OTP (TTL pendek, volatil) dan session (TTL panjang, kritikal)
agar eviction agresif di satu sisi tidak merusak sisi lain.

---

## 10. Referensi

- `internal/entity/{entity,user}.go` — `BaseModel`, `User`, `UserRole`
- `internal/repository/{base,auth_repository}.go` — generic repository
- `internal/service/auth_service.go` — alur signup/verify/resend/login + key Redis
- `internal/pkg/{otp,jwt}` — generate/HMAC OTP, token pair, SHA-256 refresh
- `infrastructure/datastore/{gorm,redis}.go` — koneksi, pool, AutoMigrate, klien Redis
- `infrastructure/app/fiber.go` — pemetaan error duplikat per nama index
- [`config.yaml.example`](../config.yaml.example) — template konfigurasi
- `git show wip-domain-sketch:<path>` — sketsa 13 entity e-commerce (belum ada di DB)