# GoCommerce — Progress

> Backend e-commerce API (alpha) · Go · Fiber v3 · GORM · PostgreSQL · Redis · SMTP.

## Status saat ini
- **Selesai/aktif:** Health check; Auth (signup dengan proteksi pending session, verify-otp, resend-otp, login, refresh token dengan rotasi, logout/revoke session via blacklist); middleware dasar JWT stateless + RBAC (`RequireAuth` aktif di logout dan endpoint user); modul user dengan `GET /users/me` (baca profil dari DB, tanpa field sensitif), `PATCH /users/me` (ubah nama/telepon), serta `PUT`/`DELETE /users/me/avatar` (upload/ganti/hapus foto profil, penyimpanan lokal `storage/uploads/avatars/` disajikan publik via `/uploads/...`, validasi ukuran 2 MB & magic bytes JPEG/PNG/WebP, invalidasi cache profil); **Store** (daftar toko, kelola toko sendiri, upload/hapus logo & banner, approve/reject oleh admin — role buyer↔seller naik-turun otomatis); **Category** (CRUD admin hierarkis dengan proteksi cycle & guard anak/produk, baca kategori aktif untuk seller); **Product** (buat produk dengan varian, CRUD milik sendiri dengan guard ownership, manajemen varian — termasuk recompute `min_price`/`max_price` — serta upload/ganti/hapus foto produk dengan satu primary yang otomatis dipindah, maks. 10 foto & 3 MB per foto).
- **Belum:** Ganti password; forgot/reset password; alamat user (address), cart, order, payment, shipment, voucher, wallet, campaign, review, chat, notification.

## Endpoint aktif (`/api/v1`)
- `GET /health`
- `POST /auth/signup` · `POST /auth/verify-otp` · `POST /auth/resend-otp`
- `POST /auth/login` · `POST /auth/refresh` · `POST /auth/logout` (Bearer Auth)
- `GET /users/me` (Bearer Auth)
- `PATCH /users/me` (Bearer Auth) — ubah nama/telepon
- `PUT /users/me/avatar` (Bearer Auth, multipart field `avatar`) — upload/ganti foto profil
- `DELETE /users/me/avatar` (Bearer Auth) — hapus foto profil
- `GET /uploads/avatars/<file>` (Public) — file avatar (static serving, prefix `/uploads`)

Store:
- `POST /stores` (Bearer Auth) — daftar toko (status `pending` sampai disetujui)
- `GET /stores/me` · `PATCH /stores/me` (Bearer Auth) — baca/ubah toko sendiri
- `PUT /stores/me/logo` · `DELETE /stores/me/logo` (Bearer Auth, multipart field `logo`)
- `PUT /stores/me/banner` · `DELETE /stores/me/banner` (Bearer Auth, multipart field `banner`)
- `GET /admin/stores` · `PATCH /admin/stores/:id/approve` · `PATCH /admin/stores/:id/reject` (Admin)

Category:
- `GET /categories` (Bearer Auth) — hanya kategori aktif
- `POST /categories` · `PATCH /categories/:id` · `DELETE /categories/:id` (Admin)
- `GET /categories/:id` (Bearer Auth)
- `GET /admin/categories` (Admin) — termasuk kategori nonaktif

Product:
- `POST /products` (Bearer Auth, toko active) — produk + minimal 1 varian
- `GET /products/me` (Bearer Auth) — produk milik seller
- `GET /products/:id` (Bearer Auth; pemilik atau admin)
- `PATCH /products/:id` · `DELETE /products/:id` (Bearer Auth, pemilik)
- `POST /products/:id/variants` · `PATCH /products/:id/variants/:variantId` · `DELETE /products/:id/variants/:variantId` (Bearer Auth, pemilik)
- `POST /products/:id/images` (Bearer Auth, pemilik, multipart field `image`)
- `PATCH /products/:id/images/:imageId` · `DELETE /products/:id/images/:imageId` (Bearer Auth, pemilik)
- `GET /uploads/products/<file>` (Public) — foto produk

## Berikutnya (arah besar)
- Ganti password, lalu forgot/reset password.
- Alamat user (address) lalu cart, order, payment, shipment, voucher, wallet, campaign, review, chat, notification.
- Pasang guard auth ke endpoint saat domain dikerjakan; test & perapihan migrasi seperlunya.

*Last updated: 2026-11-21*
