---
name: git-workflow
description: Git workflow проекта mudro — ветки, коммиты, PR, ограничения по Mission.md
---

# Skill: Git Workflow

## Правила (из Mission.md + AGENTS.md)
- **Никогда** не пушить в `main` напрямую — только через PR
- **Никогда** не делать force push
- **Никогда** не удалять ветки/теги без подтверждения
- Максимум **25 файлов** и **800 строк diff** за один цикл/PR
- Commit messages на **английском**, пояснения на **русском**

## Формат веток
```
agent/<YYYYMMDD>-<slug>
```
Примеры:
- `agent/20260317-error-boundary`
- `agent/20260317-ci-frontend`
- `agent/20260318-auth-jwt`

## Создание рабочей ветки
```bash
cd E:\mudr\mudro11-reference
git checkout -b agent/$(date +%Y%m%d)-<slug>
```
Или в PowerShell:
```powershell
$slug = "error-boundary"
$date = Get-Date -Format "yyyyMMdd"
git checkout -b "agent/$date-$slug"
```

## Pre-commit чеклист

### 1. Статус
```bash
git status --short
```

### 2. Проверить что нет sensitive файлов
```bash
git diff --name-only | grep -E '\.env|data/|out/|output/|token|secret|key'
```
Если что-то нашлось — **стоп**, убрать из staging.

### 3. Проверить размер diff
```bash
git diff --stat | tail -1
# Должно быть: N files changed, не больше 25 файлов
git diff --shortstat
# Должно быть: не больше 800 insertions+deletions
```

### 4. Тесты
```bash
# Backend (WSL)
go test ./...

# Frontend
cd frontend && npm run build && npm run lint
```

### 5. Коммит
```bash
git add -A
git commit -m "feat: краткое описание на английском"
```

### 6. Push
```bash
git push -u origin agent/<дата>-<slug>
```

## PR Description Template
```markdown
## Что
Краткое описание изменений.

## Зачем
Почему это нужно.

## Как проверить
1. Шаг 1
2. Шаг 2

## Риски и откат
- Риск: ...
- Откат: `git revert <sha>`

## Checklist
- [ ] go test ./... green
- [ ] npm run build green
- [ ] diff < 25 files / 800 lines
- [ ] no sensitive files
```

## Worktrees (для параллельной работы)

| Директория | Ветка | Назначение |
|-----------|-------|-----------|
| `mudro11-reference` | `codex/mainline-base` | Чистая base, чтение, MCP |
| `mudro11-automation` | `codex/automation-track` | Агентский контур, автоматизация |
| `mudro11-bugs` | `codex/bugs-repo-db` | Багфиксы, тесты, БД, API |
| `mudro11-devops` | `codex/devops-vps` | VPS, deploy, nginx, infra |

## Полезные команды
```bash
git log --oneline -10                    # последние 10 коммитов
git log --oneline --since="1 day ago"    # за сутки
git diff --name-only HEAD~1              # файлы в последнем коммите
git stash && git stash pop               # временно убрать изменения
```
