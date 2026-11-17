# GoCommerce — Progress

> Backend e-commerce API (alpha) · Go · Fiber v3 · GORM · PostgreSQL · Redis · SMTP.

## Status saat ini
- **Selesai/aktif:** Health check; Auth (signup, verify-otp, resend-otp, login, refresh token, logout/revoke session); middleware dasar JWT stateless + RBAC (`RequireAuth`/`RequireRoles`, sudah ada/wired, belum dipasang ke endpoint domain).
- **Belum:** Forgot/reset password; domain lain (profile/address, store, product, cart, order, payment, shipment, voucher, wallet, campaign, review, chat, notification).

## Endpoint aktif (`/api/v1`)
- `GET /health`
- `POST /auth/signup` · `POST /auth/verify-otp` · `POST /auth/resend-otp`
- `POST /auth/login` · `POST /auth/refresh` · `POST /auth/logout`

## Berikutnya (arah besar)
- Forgot/reset password, pasang guard auth ke endpoint saat domain dikerjakan.
- Pengerjaan domain e-commerce berikutnya + test & perapihan migrasi seperlunya.

*Last updated: 2026-11-16*
