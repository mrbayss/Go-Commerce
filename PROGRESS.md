# GoCommerce — Progress

> Backend e-commerce API (alpha) · Go · Fiber v3 · GORM · PostgreSQL · Redis · SMTP.

## Status saat ini
- **Selesai/aktif:** Health check; Auth (signup dengan proteksi pending session, verify-otp, resend-otp, login, refresh token dengan rotasi, logout/revoke session via blacklist); middleware dasar JWT stateless + RBAC (`RequireAuth` aktif di logout dan endpoint user); modul user dengan `GET /users/me` (baca profil dari DB, tanpa field sensitif), `PATCH /users/me` (ubah nama/telepon), serta `PUT`/`DELETE /users/me/avatar` (upload/ganti/hapus foto profil, penyimpanan lokal `storage/uploads/avatars/` disajikan publik via `/uploads/...`, validasi ukuran 2 MB & magic bytes JPEG/PNG/WebP, invalidasi cache profil).
- **Belum:** Ganti password; forgot/reset password; domain lain (address, store, product, cart, order, payment, shipment, voucher, wallet, campaign, review, chat, notification).

## Endpoint aktif (`/api/v1`)
- `GET /health`
- `POST /auth/signup` · `POST /auth/verify-otp` · `POST /auth/resend-otp`
- `POST /auth/login` · `POST /auth/refresh` · `POST /auth/logout` (Bearer Auth)
- `GET /users/me` (Bearer Auth)
- `PATCH /users/me` (Bearer Auth) — ubah nama/telepon
- `PUT /users/me/avatar` (Bearer Auth, multipart field `avatar`) — upload/ganti foto profil
- `DELETE /users/me/avatar` (Bearer Auth) — hapus foto profil
- `GET /uploads/avatars/<file>` (Public) — file avatar (static serving, prefix `/uploads`)

## Berikutnya (arah besar)
- Ganti password, lalu forgot/reset password.
- Alamat user (address) dan domain e-commerce berikutnya (store, product, cart, order, payment, shipment, voucher, wallet, campaign, review, chat, notification).
- Pasang guard auth ke endpoint saat domain dikerjakan; test & perapihan migrasi seperlunya.

*Last updated: 2026-11-20*
