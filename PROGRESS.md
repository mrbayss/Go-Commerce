# GoCommerce — Progress

> Backend e-commerce API (alpha) · Go · Fiber v3 · GORM · PostgreSQL · Redis · SMTP.

## Status saat ini
- **Selesai/aktif:** Health check; Auth (signup dengan proteksi pending session, verify-otp, resend-otp, login, refresh token dengan rotasi, logout/revoke session via blacklist); middleware dasar JWT stateless + RBAC (`RequireAuth` aktif di logout dan endpoint user); modul user dengan `GET /users/me` (baca profil dari DB, tanpa field sensitif) dan `PATCH /users/me` (ubah nama/telepon).
- **Belum:** Avatar; ganti password; forgot/reset password; domain lain (address, store, product, cart, order, payment, shipment, voucher, wallet, campaign, review, chat, notification).

## Endpoint aktif (`/api/v1`)
- `GET /health`
- `POST /auth/signup` · `POST /auth/verify-otp` · `POST /auth/resend-otp`
- `POST /auth/login` · `POST /auth/refresh` · `POST /auth/logout` (Bearer Auth)
- `GET /users/me` (Bearer Auth)
- `PATCH /users/me` (Bearer Auth) — ubah nama/telepon

## Berikutnya (arah besar)
- Avatar (upload/hapus foto profil).
- Ganti password, lalu forgot/reset password.
- Alamat user (address) dan domain e-commerce berikutnya (store, product, cart, order, payment, shipment, voucher, wallet, campaign, review, chat, notification).
- Pasang guard auth ke endpoint saat domain dikerjakan; test & perapihan migrasi seperlunya.

*Last updated: 2026-11-19*
