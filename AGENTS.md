# AGENTS.md — GoCommerce

> Backend API e-commerce multi-vendor · Go 1.27.1 · Fiber v3.5.0 · GORM · PostgreSQL · Redis · SMTP.
> Kode adalah sumber kebenaran. Ikuti pola yang sudah ada di repo; kalau ragu, cek implementasi.

## Struktur
- `cmd/api/main.go` — entrypoint & wiring
- `internal/` — controller, service, repository, entity, model, domain, middleware, router, pkg (util murni)
- `infrastructure/` — app, config, datastore, logger, validation, mailer
- `config.yaml.example` — template config (jangan commit `config.yaml`)

## Cara kerja (panduan longgar)
- Baca kode/pola terkait sebelum mengubah; jaga perubahan tetap fokus ke yang diminta.
- Alur umum: controller → service → repository → entity (boleh menyesuaikan bila memang perlu).
- Utamakan kejelasan & konsistensi dengan sekitar; hindari over-engineering.

## Batasan penting
- Jangan tulis/commit secret, credential, atau PII (password, token, API key) ke kode/log/dokumen.
- Jangan tindakan destruktif (hapus data, reset, force push, drop db) tanpa konfirmasi eksplisit.
- Jangan commit/push kecuali diminta.

## Commit
- Gunakan [Conventional Commits](https://www.conventionalcommits.org/): `<type>(<scope>): <deskripsi singkat>`.
- Contoh type: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`. Contoh: `feat(auth): add refresh token`.
