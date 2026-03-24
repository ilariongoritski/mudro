# VPS Server Info

- **Хостинг**: Hostkey.ru
- **Имя хоста**: LARR1I
- **IP**: 91.218.113.247
- **Конфигурация**: 4 vCore / 8 GB RAM / 120 GB NVMe
- **Канонический путь проекта на сервере**: `/srv/mudro`
- **Публичный runtime**: `nginx` на `:80/:443`, frontend в `/var/www/mudro/frontend`, API на `127.0.0.1:8080`
- **Основной runtime-файл**: `docker-compose.prod.yml`
- **SSH доступ**: целевое состояние — `admin` + SSH key only
- **Секреты**: пароль root и API key панели в репозитории не хранятся; при любой утечке их нужно ротировать вне git
- **SSH ключи** (локальные, авторизованные на сервере):
  - Windows: `~/.ssh/codex_mudro_vps`, `~/.ssh/codex_mudro_vps2`
  - WSL: `~/.ssh/codex_mudro_vps2`, `~/.ssh/id_ed25519`

## Vercel
- **CLI**: установлен, авторизован как `goritskimihail-2652`
- **Проект `mudro`**: `https://mudro.vercel.app`
- **Проект `frontend`**: `https://frontend-psi-ten-33.vercel.app`
- **Статус**: Vercel допускается как preview/fallback, но основным публичным контуром считается VPS-first runtime
