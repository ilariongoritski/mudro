# MUDRO MVP Code Review — 2026-03-23

**Дата**: 2026-03-23  
**Контекст**: Финальное кодревью перед MVP, анализ MUDROTOP и готовности проекта

---

## Исполнительное резюме

### Статус MVP: ⚠️ Частично готов

**Что работает:**
- ✅ Core backend (feed-api, auth, agent queue, casino)
- ✅ Frontend UI (лента, комментарии, реакции, casino)
- ✅ VPS deployment (systemd services активны)
- ✅ WebSocket инфраструктура (casino)

**Критические находки:**
- ❌ **MUDROTOP не существует в кодовой базе**
- ❌ WebSocket для feed/chat не реализован
- ❌ Агенты в "комнатах" — концепция отсутствует
- ⚠️ Несколько измененных файлов не закоммичены

---

## 1. MUDROTOP: Критическая находка

### Что показано на скриншоте

Скриншот демонстрирует виртуальное пространство с двумя комнатами:
- Кафе (слева)
- Гостиная (справа)
- Панель управления "OpenClaw Control" справа
- URL: `http://127.0.0.1:4173/#feed`

### Что найдено в коде

**Результат поиска**: 0 упоминаний `mudrotop` или `MUDROTOP` в кодовой базе.

**Анализ:**

1. **Это НЕ часть текущего репозитория `mudro11`**
   - Нет файлов с виртуальными комнатами
   - Нет 2D-рендеринга пространств
   - Нет системы "агентов в комнатах"

2. **Возможные объяснения:**
   - Это отдельный проект/прототип
   - Это mock-up или концепт
   - Это внешний сервис на порту 4173
   - Это старая версия, не попавшая в текущий репозиторий

3. **Связь с текущим проектом:**
   - URL содержит `#feed` — возможно, это альтернативный frontend
   - Панель "OpenClaw Control" совпадает с worker-plane из текущего проекта
   - Порт 4173 — стандартный Vite preview port

### Рекомендация

**Необходимо выяснить:**
- Где находится код MUDROTOP?
- Это отдельный репозиторий?
- Это планируемая фича или существующий прототип?
- Как это связано с текущим MVP?

---

## 2. Архитектура проекта

### Структура сервисов

```
services/
├── feed-api/          ✅ Работает (HTTP API для постов)
├── agent/             ✅ Работает (agent-worker, agent-planner)
├── bot/               ✅ Работает (Telegram bot)
├── casino/            ✅ Работает (изолированный игровой контур)
├── auth-api/          ⚠️ Частично (в миграции)
├── orchestration-api/ ⚠️ Частично (в миграции)
├── bff-web/           ⚠️ Частично (в миграции)
└── api-gateway/       ⚠️ Частично (в миграции)
```

### Текущее состояние

**Работающий монолит:**
- [`internal/api/server.go`](internal/api/server.go:1) — основной HTTP сервер
- [`internal/posts/posts.go`](internal/posts/posts.go:1) — бизнес-логика постов
- [`internal/auth/service.go`](internal/auth/service.go:1) — JWT auth
- [`internal/agent/repo.go`](internal/agent/repo.go:1) — очередь задач агентов

**Микросервисы в процессе выделения:**
- Контракты готовы: [`contracts/http/`](contracts/http/)
- Сервисы частично реализованы
- Миграция не завершена

---

## 3. WebSocket и Real-time

### Что реализовано

**Casino WebSocket** ([`internal/casino/wshandler.go`](internal/casino/wshandler.go:1)):
```go
type WSHub struct {
    mu    sync.RWMutex
    conns map[string][]*websocket.Conn
}

func (h *WSHub) HandleUpgrade(w http.ResponseWriter, r *http.Request)
func (h *WSHub) Emit(userID, eventType string, payload any)
```

- ✅ Работает для casino
- ✅ Per-user connections
- ✅ Event broadcasting

### Что НЕ реализовано

**Feed/Chat WebSocket:**
- ❌ Нет `/api/ws` для ленты
- ❌ Нет `/api/firehose` (SSE)
- ❌ Нет live updates для новых постов/комментариев
- ❌ Нет presence API (кто онлайн)

**Frontend WebSocket:**
- ❌ Нет WS клиента в [`frontend/src/`](frontend/src/)
- ❌ RTK Query не подключен к WS
- ❌ Нет real-time подписок

### Из roadmap (не реализовано)

[`docs/practical-plan-march-2026.md`](docs/practical-plan-march-2026.md:162):
```markdown
**P1 (важно)**:
- [ ] WebSocket `/api/ws` + SSE `/api/firehose`
- [ ] Kafka consumer → WS broadcast (новые посты/комментарии)
- [ ] Frontend: RTK Query + WS subscription
```

---

## 4. Агенты и "комнаты"

### Текущая реализация агентов

**Agent Queue** ([`internal/agent/repo.go`](internal/agent/repo.go:1)):
```go
type Task struct {
    ID          int64
    Kind        string
    Payload     json.RawMessage
    Status      string  // queued, locked, done, failed
    Priority    int
    Attempts    int
    LockedBy    *string
}
```

**Типы задач:**
- Обработка постов
- Модерация контента
- Импорт данных
- Бэкфилл медиа

**Что НЕТ:**
- ❌ Концепции "комнат"
- ❌ Виртуальных пространств
- ❌ Визуального представления агентов
- ❌ Системы присутствия агентов в локациях

### Почему "в комнате никого"

**Ответ:** Потому что концепция "комнат с агентами" не реализована в текущей кодовой базе.

**Что есть вместо этого:**
- Абстрактная очередь задач в PostgreSQL
- Worker-процессы (agent-worker, agent-planner)
- Нет UI для визуализации агентов

---

## 5. Frontend состояние

### Реализованные страницы

```
frontend/src/pages/
├── feed-page/           ✅ Лента постов
├── casino-page/         ✅ Casino (полный режим)
├── casino-miniapp-page/ ✅ Casino (Telegram Mini App)
├── orchestration-page/  ✅ Control plane
├── admin-page/          ✅ Админка
├── auth-page/           ✅ Вход/регистрация
├── chat-page/           ❌ Заглушка
├── movies-page/         ❌ Заглушка
└── profile-page/        ❌ Заглушка
```

### Chat Page ([`frontend/src/pages/chat-page/ui/ChatPage.tsx`](frontend/src/pages/chat-page/ui/ChatPage.tsx:1))

```tsx
const ChatPage = () => {
  return (
    <Card>
      <CardContent>
        <MessageCircle size={48} />
        <h2>Чат</h2>
        <p>Этот раздел скоро появится</p>
      </CardContent>
    </Card>
  )
}
```

**Вывод:** Чат не реализован, только placeholder.

---

## 6. Casino — единственный работающий real-time

### Backend

**Сервис:** [`services/casino/cmd/casino/main.go`](services/casino/cmd/casino/main.go:1)

**Фичи:**
- ✅ WebSocket `/api/casino/ws`
- ✅ Provably fair gaming
- ✅ RTP profiles
- ✅ Idempotency
- ✅ Double-entry ledger
- ✅ Telegram Mini App auth

### Frontend

**Страницы:**
- [`CasinoPage.tsx`](frontend/src/pages/casino-page/ui/CasinoPage.tsx:1) — полный режим
- [`CasinoMiniAppPage.tsx`](frontend/src/pages/casino-miniapp-page/ui/CasinoMiniAppPage.tsx:1) — Telegram Mini App

**API:** [`casinoApi.ts`](frontend/src/features/casino/api/casinoApi.ts:1)
```typescript
export const casinoApi = mudroApi.injectEndpoints({
  endpoints: (build) => ({
    getCasinoBalance: build.query<CasinoBalance, void>({ ... }),
    getCasinoHistory: build.query<CasinoHistory, number>({ ... }),
    spinCasino: build.mutation<SpinResult, SpinRequest>({ ... }),
  }),
})
```

---

## 7. Критические проблемы

### 7.1 Незакоммиченные изменения

```
git status:
 M README.md
 M tools/README.md
 M tools/opus-gateway/.env.example
 M tools/opus-gateway/README.md
 M tools/opus-gateway/package.json
 M tools/opus-gateway/src/paths.ts
 M tools/opus-gateway/src/runner.ts
 M tools/opus-gateway/src/validation.ts
 M tools/opus-gateway/test/server.test.ts
 M tools/opus-gateway/test/validation.test.ts
... 3 more files
```

**Риск:** Локальные изменения могут быть потеряны.

### 7.2 MUDROTOP неясность

**Критический вопрос:** Что такое MUDROTOP и где его код?

**Возможные сценарии:**
1. Это отдельный репозиторий — нужна ссылка
2. Это прототип вне Git — нужно решить, включать ли в MVP
3. Это концепт — нужно документировать как roadmap item
4. Это старая версия — нужно выяснить, актуально ли

### 7.3 WebSocket для feed не реализован

**Проблема:** План предполагает real-time feed, но реализован только casino WS.

**Последствия:**
- Нет live updates постов
- Нет онлайн-присутствия
- Нет чата

### 7.4 Микросервисная миграция не завершена

**Статус:**
- Контракты готовы
- Сервисы частично реализованы
- Роутинг не переключен
- Старый монолит всё ещё основной

---

## 8. Что готово для MVP

### ✅ Готово и работает

1. **Feed API**
   - Посты с пагинацией
   - Комментарии с threading
   - Реакции (emoji)
   - Медиа (фото, видео)
   - Фильтры по источникам

2. **Auth**
   - JWT authentication
   - Регистрация/вход
   - Session management
   - Telegram Mini App auth

3. **Casino**
   - Полнофункциональный игровой контур
   - WebSocket real-time
   - Provably fair
   - Telegram Mini App интеграция

4. **Agent Queue**
   - Очередь задач
   - Worker/planner архитектура
   - Retry logic
   - Priority handling

5. **Frontend**
   - Лента постов
   - Детальный просмотр
   - Комментарии
   - Casino UI
   - Orchestration dashboard

6. **VPS Deployment**
   - Systemd services
   - Tracked configuration
   - Health checks
   - Worker plane (OpenClaw, Skaro)

### ⚠️ Частично готово

1. **Микросервисы**
   - Контракты есть
   - Реализация неполная
   - Не переключено в production

2. **Real-time**
   - Casino WS работает
   - Feed WS отсутствует
   - SSE не реализован

3. **Monitoring**
   - Health endpoints есть
   - Metrics отсутствуют
   - Grafana не настроен

### ❌ Не готово

1. **MUDROTOP**
   - Код отсутствует
   - Концепция неясна
   - Интеграция неизвестна

2. **Chat**
   - Только placeholder
   - Backend отсутствует
   - WebSocket не подключен

3. **Presence API**
   - Нет отслеживания онлайн
   - Нет "кто в комнате"
   - Нет визуализации агентов

4. **RSS/Atom feeds**
   - Не реализовано (из roadmap)

5. **Export/Import API**
   - Не реализовано (из roadmap)

---

## 9. Рекомендации

### Немедленные действия (P0)

1. **Выяснить статус MUDROTOP**
   ```bash
   # Где код?
   # Это отдельный репозиторий?
   # Нужно ли включать в MVP?
   ```

2. **Закоммитить изменения**
   ```bash
   git add -A
   git commit -m "chore: sync local changes before MVP review"
   git push origin main
   ```

3. **Документировать MUDROTOP**
   - Если это roadmap item → добавить в [`docs/practical-plan-march-2026.md`](docs/practical-plan-march-2026.md:1)
   - Если это отдельный проект → добавить ссылку в [`README.md`](README.md:1)
   - Если это прототип → решить, включать ли в MVP

### Краткосрочные (P1) — 1-2 недели

1. **Завершить WebSocket для feed**
   - Реализовать `/api/ws` endpoint
   - Подключить Kafka consumer
   - Добавить frontend WS client
   - Реализовать live updates

2. **Реализовать базовый chat**
   - Backend: messages table
   - WebSocket broadcast
   - Frontend: chat UI
   - Интеграция с auth

3. **Завершить микросервисную миграцию**
   - Переключить роутинг на новые сервисы
   - Протестировать end-to-end
   - Обновить deployment

### Среднесрочные (P2) — 1 месяц

1. **Presence API**
   - User online/offline tracking
   - WebSocket heartbeat
   - Frontend presence indicators

2. **RSS/Atom feeds**
   - `/api/feed.rss`
   - `/api/feed.atom`
   - Per-user feeds

3. **Monitoring**
   - Prometheus metrics
   - Grafana dashboards
   - Alerting

---

## 10. Ответы на вопросы

### "це шо уже что-то готово?"

**Да, многое готово:**
- ✅ Лента постов работает
- ✅ Casino полностью функционален
- ✅ VPS deployment активен
- ✅ Frontend UI готов
- ⚠️ Но MUDROTOP отсутствует в коде

### "где агенты на сайте?"

**Агенты есть, но не визуализированы:**
- ✅ Agent queue работает (PostgreSQL)
- ✅ Worker/planner процессы активны
- ❌ UI для агентов отсутствует
- ❌ Концепция "комнат" не реализована

### "почему в комнате никого?"

**Потому что "комнат" нет в коде:**
- ❌ MUDROTOP не найден в репозитории
- ❌ Виртуальные пространства не реализованы
- ❌ Визуализация агентов отсутствует
- ✅ Но абстрактная agent queue работает

---

## 11. Заключение

### Статус MVP: 70% готов

**Сильные стороны:**
- Solid backend архитектура
- Работающий casino с real-time
- Качественный frontend
- Production deployment

**Критические пробелы:**
- MUDROTOP неясен
- WebSocket для feed отсутствует
- Chat не реализован
- Микросервисы не завершены

### Следующий шаг

**Необходимо принять решение по MUDROTOP:**
1. Если это отдельный проект → предоставить ссылку на репозиторий
2. Если это roadmap item → документировать и запланировать
3. Если это прототип → решить, включать ли в текущий MVP

**После этого можно:**
- Завершить WebSocket для feed
- Реализовать базовый chat
- Закрыть микросервисную миграцию
- Запустить полноценный MVP

---

**Автор**: Kiro Code Review Agent  
**Дата**: 2026-03-23  
**Версия**: 1.0
