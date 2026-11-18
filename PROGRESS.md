# GoCommerce — Progress

> Backend e-commerce API (alpha) · Go · Fiber v3 · GORM · PostgreSQL · Redis · SMTP.

## Status saat ini
- **Selesai/aktif:** Health check; Auth (signup dengan proteksi pending session, verify-otp, resend-otp, login, refresh token dengan rotasi, logout/revoke session via blacklist); middleware dasar JWT stateless + RBAC (`RequireAuth` sudah aktif di logout; endpoint domain belum ada).
- **Belum:** Forgot/reset password; domain lain (profile/address, store, product, cart, order, payment, shipment, voucher, wallet, campaign, review, chat, notification).

## Endpoint aktif (`/api/v1`)
- `GET /health`
- `POST /auth/signup` · `POST /auth/verify-otp` · `POST /auth/resend-otp`
- `POST /auth/login` · `POST /auth/refresh` · `POST /auth/logout` (Bearer Auth)

## Berikutnya (arah besar)
- Forgot/reset password, pasang guard auth ke endpoint saat domain dikerjakan.
- Pengerjaan domain e-commerce berikutnya + test & perapihan migrasi seperlunya.

*Last updated: 2026-11-18*
