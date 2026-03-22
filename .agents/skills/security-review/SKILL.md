---
name: security-review
description: Чеклист безопасности для кода, API и инфраструктуры проекта mudro
---

# Skill: Security Review

Адаптирован под стек mudro и правила `Mission.md`.

## Чеклист перед каждым PR

### Секреты и credentials
- [ ] Нет токенов, паролей, ключей в коде или diff
- [ ] `.env` в `.gitignore` и не в staging
- [ ] Нет хардкода DSN, API ключей в исходниках
- [ ] Секреты передаются через env vars, не через аргументы командной строки

```bash
# Проверить что нет утечек секретов в diff
git diff --cached | grep -iE '(password|secret|token|api_key|dsn|Bearer)\s*=\s*["\047][^"\047]{6,}'
```

### Go backend (cmd/api, cmd/bot, cmd/agent)
- [ ] SQL-запросы используют `$1, $2` плейсхолдеры (не `fmt.Sprintf`)
- [ ] Входные данные валидируются перед записью в БД
- [ ] HTTP-хендлеры не раскрывают внутренние ошибки пользователю
- [ ] Rate limiter активен (`API_RATE_LIMIT_RPS`)
- [ ] `/api/*` не возвращает stack trace в production

```go
// Плохо — SQL injection
db.Exec(fmt.Sprintf("SELECT * FROM posts WHERE source = '%s'", source))

// Хорошо — параметризованный запрос
db.Exec("SELECT * FROM posts WHERE source = $1", source)
```

### Frontend (React/TS)
- [ ] Нет `dangerouslySetInnerHTML` без санитизации
- [ ] API-ключи не в клиентском коде (только backend)
- [ ] CSP headers через nginx (\`Content-Security-Policy\`)
- [ ] XSS: пользовательский текст всегда через React (не \`innerHTML\`)

### API endpoints
- [ ] `GET /api/front` и `GET /api/posts` — публичные, но с Rate Limiter
- [ ] Нет эндпоинтов, возвращающих внутренние данные без проверки
- [ ] `POST` запросы (если добавляются) требуют CSRF защиту

### Инфраструктура (VPS)
- [ ] Postgres привязан только к loopback (127.0.0.1)
- [ ] Порт 5433 закрыт от внешнего интернета (UFW)
- [ ] SSH: только ключи, password auth выключен
- [ ] fail2ban активен
- [ ] Docker daemon недоступен по сети

## Быстрый security scan Go кода

```bash
# Установить gosec (один раз)
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Запустить
gosec ./...
```

## Критические запреты по Mission.md
| Операция | Статус |
|----------|--------|
| Менять SSH config/UFW/iptables | ⛔ без двухфазного подтверждения |
| Удалять данные из `/srv`, `/etc`, `/root` | ⛔ без явного подтверждения |
| Публиковать токены в логах | ⛔ никогда |
| Пушить в main напрямую | ⛔ только через PR |
