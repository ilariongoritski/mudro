---
description: Код-ревью изменений перед коммитом в main
---

# Код-ревью (Pre-commit Review)

Стандартный процесс проверки перед коммитом/PR.

## Шаги

1. Проверить текущий статус git:
```powershell
cd E:\mudr\mudro11-reference && git status --short
```

2. Посмотреть diff изменений:
```powershell
cd E:\mudr\mudro11-reference && git diff --stat
```

3. Убедиться что в diff нет чувствительных файлов:
   - Не должно быть: `.env`, `data/`, `out/`, `output/`, токенов, ключей
   - Проверить: `git diff --name-only | Select-String -Pattern '\.env|data/|out/|output/|token|secret|key'`

4. Прогнать backend тесты (если менялся Go-код):
```powershell
wsl -d Ubuntu -- bash -c "cd ~/projects/mudro && go test ./..."
```

5. Прогнать frontend проверки (если менялся frontend/):
```powershell
cd E:\mudr\mudro11-reference\frontend && npm.cmd run build && npm.cmd run lint
```

6. Обновить `.codex/done.md` с кратким результатом.

## Критерии успеха
- В diff нет sensitive файлов
- Backend тесты проходят
- Frontend build + lint без ошибок
- Максимум 25 файлов и 800 строк diff (по правилам Mission.md)
