---
description: Деплой фронтенда на VPS (сборка + копирование static файлов)
---

# Деплой Frontend на VPS

Собирает production-бандл и деплоит через rsync/scp на VPS nginx.

## Предварительные условия
- SSH-доступ к VPS `91.218.113.247`
- Nginx настроен на раздачу из `/var/www/mudro/frontend`

## Шаги

1. Установить зависимости и собрать бандл:
```powershell
cd d:\mudr\mudro11-main\frontend
npm.cmd ci
npm.cmd run build
```

2. Залить на сервер (через scp из PowerShell):
```powershell
scp -r d:\mudr\mudro11-main\frontend\dist\* admin@91.218.113.247:/var/www/mudro/frontend/
```

3. Проверить, что сайт отдается:
```powershell
curl http://91.218.113.247/healthz
```

## Критерии успеха
- `npm run build` завершается без ошибок
- Файлы появились в `/var/www/mudro/frontend/` на сервере
- `curl /healthz` возвращает `{"status":"ok"}`
