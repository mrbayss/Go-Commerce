# GoCommerce — Progress

> Backend e-commerce API (alpha) · Go · Fiber v3 · GORM · PostgreSQL · Redis · SMTP.

## Status saat ini
- **Selesai/aktif:** Health check; Auth (signup dengan proteksi pending session, verify-otp, resend-otp, login, refresh token dengan rotasi, logout/revoke session via blacklist); middleware dasar JWT stateless + RBAC (`RequireAuth` aktif di logout dan `GET /users/me`); modul user dengan endpoint `GET /users/me` (baca profil dari DB, tanpa field sensitif).
- **Belum:** Forgot/reset password; `PATCH /users/me`, avatar, ganti password; domain lain (address, store, product, cart, order, payment, shipment, voucher, wallet, campaign, review, chat, notification).

## Endpoint aktif (`/api/v1`)
- `GET /health`
- `POST /auth/signup` · `POST /auth/verify-otp` · `POST /auth/resend-otp`
- `POST /auth/login` · `POST /auth/refresh` · `POST /auth/logout` (Bearer Auth)
- `GET /users/me` (Bearer Auth)

## Berikutnya (arah besar)
- `PATCH /users/me` (ubah nama/telepon/avatar), ganti password, lalu alamat user.
- Forgot/reset password, pasang guard auth ke endpoint saat domain dikerjakan.
- Pengerjaan domain e-commerce berikutnya + test & perapihan migrasi seperlunya.

*Last updated: 2026-11-18*
