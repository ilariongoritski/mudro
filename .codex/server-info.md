# VPS Server Info

- **Хостинг**: Hostkey.ru
- **Имя хоста**: LARR1I
- **IP**: 91.218.113.247
- **Конфигурация**: 4 vCore / 8 GB RAM / 120 GB NVMe
- **API-ключ панели**: название `OCMUDROSR (5df9...)`, хранится в `.env` как `HOSTKEY_API_KEY`
- **Панель управления**: `https://panel.hostkey.ru/controlpanel.html?key=<HOSTKEY_API_KEY>`
- **Проект на сервере**: `/root/projects/mudro`
- **Сервисы**: Docker Compose (API :8080, Postgres :5432), nginx (:80)
- **SSH юзер**: `root` (доступ есть через ключ и пароль)
- **SSH пароль**: `If%2zvElra` (для root)
- **SSH ключи** (локальные, авторизованы на сервере):
  - Windows: `~/.ssh/codex_mudro_vps`, `~/.ssh/codex_mudro_vps2`
  - WSL: `~/.ssh/codex_mudro_vps2`, `~/.ssh/id_ed25519`

## Vercel

- **CLI**: установлен, авторизован как `goritskimihail-2652`
- **Проект `mudro`** (Go serverless backend):
  - URL: `https://mudro.vercel.app`
  - Project ID: `prj_1KfDU2JpyyLqrXLM0jkQmFRcAjN1`
  - Source: ветка `main`
  - Error Rate: **42.9%** (причина — скорее всего нет DSN к БД)
- **Проект `frontend`** (React, оригинальный сайт):
  - URL: `https://frontend-psi-ten-33.vercel.app`
  - Проксирует `/api/*` и `/media/*` на VPS `91.218.113.247:8080`
- **Backend (serverless)**: `vercel.json` → `pkg/vercelapi/handler.go`
