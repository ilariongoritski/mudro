# План интеграции MUDROTOP в Скаро

**Дата**: 2026-03-23  
**Статус**: Готов к реализации  
**Контекст**: Полная интеграция микросервиса movie-catalog в основной проект MUDRO с подключением ко всем БД и API

---

## Исполнительное резюме

MUDROTOP ([`services/movie-catalog`](../services/movie-catalog/README.md)) уже реализован как standalone микросервис и готов к интеграции. Этот план описывает пошаговую интеграцию в основной runtime Скаро с подключением к единой БД, API Gateway, BFF-web и Telegram bot.

### Текущий статус

| Компонент | Статус | Действие |
|-----------|--------|----------|
| Backend [`movie-catalog`](../services/movie-catalog/README.md) | ✅ Реализован | Добавить в core runtime |
| Frontend MUDROTOP | ✅ Реализован | Интегрировать в основной frontend |
| HTTP контракт | ✅ Определен | Использовать как есть |
| Docker compose | ⚠️ Изолирован | Добавить в [`docker-compose.core.yml`](../ops/compose/docker-compose.core.yml) |
| БД интеграция | ❌ Отдельная БД | Мигрировать на единую БД `gallery` |
| API проксирование | ❌ Отсутствует | Добавить в BFF-web |
| Telegram интеграция | ❌ Отсутствует | Добавить команду `/movies` |

---

## 1. Архитектура интеграции

### 1.1 Целевая архитектура

```mermaid
graph TB
    User[Пользователь Web]
    TG[Telegram Bot]
    
    User --> FE[Frontend :5173]
    TG --> TGMiniApp[Telegram Mini App]
    TGMiniApp --> FE
    
    FE --> BFF[bff-web :8086]
    
    BFF --> FeedAPI[feed-api :8080]
    BFF --> MovieCatalog[movie-catalog :8091]
    
    FeedAPI --> DB[(gallery DB :5433)]
    MovieCatalog --> DB
    
    style MovieCatalog fill:#e1f5ff
    style BFF fill:#ffe1e1
    style FE fill:#fff4e1
```

### 1.2 Ключевые решения

1. **Единая БД**: Использовать существующую БД `gallery` на порту 5433
2. **Проксирование через BFF**: Все запросы к movie-catalog идут через [`bff-web`](../services/bff-web/app/run.go)
3. **Прямая интеграция frontend**: Копировать компоненты из [`MUDROTOP/frontend`](../MUDROTOP/frontend/README.md) в основной проект
4. **Telegram Mini App**: Использовать существующую инфраструктуру [`telegram-miniapp`](../frontend/src/features/telegram-miniapp/hooks/useTelegramWebApp.ts)

---

## 2. Интеграция базы данных

### 2.1 Текущее состояние

**MUDROTOP использует:**
- Отдельный Postgres на порту 5434
- БД `movie_catalog`
- DSN: `postgres://postgres:postgres@localhost:5434/movie_catalog?sslmode=disable`

**Основной проект использует:**
- Postgres на порту 5433
- БД `gallery`
- DSN: `postgres://postgres:postgres@db:5432/gallery?sslmode=disable` (внутри Docker)

### 2.2 План миграции

**Вариант A: Отдельная схема в единой БД** (рекомендуется)

```sql
-- Создать схему movie_catalog в БД gallery
CREATE SCHEMA IF NOT EXISTS movie_catalog;

-- Все таблицы будут в схеме movie_catalog
CREATE TABLE movie_catalog.movies (...);
CREATE TABLE movie_catalog.genres (...);
```

**Преимущества:**
- Логическая изоляция данных
- Единая БД для всех сервисов
- Простая миграция существующих таблиц

**Вариант B: Таблицы в основной схеме public**

```sql
-- Префикс mc_ для таблиц
CREATE TABLE mc_movies (...);
CREATE TABLE mc_genres (...);
```

**Рекомендация**: Использовать **Вариант A** (отдельная схема)

### 2.3 Шаги миграции

1. Обновить миграции в [`migrations/movie_catalog/`](../migrations/movie_catalog/)
2. Изменить DSN в [`services/movie-catalog/cmd/main.go`](../services/movie-catalog/cmd/main.go)
3. Обновить search_path в коде репозитория
4. Запустить миграции на основной БД

---

## 3. Интеграция в Docker Compose

### 3.1 Добавление в core runtime

**Файл**: [`ops/compose/docker-compose.core.yml`](../ops/compose/docker-compose.core.yml)

```yaml
services:
  # ... существующие сервисы (db, redis, kafka, api, agent)
  
  movie-catalog:
    image: golang:1.24
    restart: unless-stopped
    working_dir: /app
    command: sh -lc "/usr/local/go/bin/go run ./services/movie-catalog/cmd"
    environment:
      MOVIE_CATALOG_ADDR: ":8091"
      MOVIE_CATALOG_DB_DSN: "postgres://postgres:${POSTGRES_PASSWORD:-postgres}@db:5432/gallery?sslmode=disable&search_path=movie_catalog,public"
    depends_on:
      db:
        condition: service_healthy
    ports:
      - "127.0.0.1:8091:8091"
    volumes:
      - ../../:/app
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O- http://127.0.0.1:8091/healthz | grep -q '\"status\":\"ok\"'"]
      interval: 10s
      timeout: 5s
      retries: 12
```

### 3.2 Обновление Makefile

**Файл**: [`Makefile`](../Makefile)

```makefile
# Добавить в существующие targets

.PHONY: movie-catalog-migrate
movie-catalog-migrate:
	@echo "Running movie-catalog migrations..."
	go run ./tools/migrations/moviecatalogmigrate/cmd

.PHONY: movie-catalog-import
movie-catalog-import:
	@echo "Importing movie catalog data..."
	go run ./tools/importers/moviecatalogimport/cmd

.PHONY: movie-catalog-run
movie-catalog-run:
	@echo "Starting movie-catalog service..."
	go run ./services/movie-catalog/cmd
```

---

## 4. Интеграция API через BFF-web

### 4.1 Добавление проксирования

**Файл**: [`services/bff-web/app/run.go`](../services/bff-web/app/run.go)

```go
package app

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goritskimihail/mudro/internal/config"
	"github.com/goritskimihail/mudro/internal/posts"
	"github.com/goritskimihail/mudro/internal/tgexport"
	"github.com/goritskimihail/mudro/services/bff-web/internal/bffweb"
)

func Run() {
	addr := envOr("BFF_WEB_ADDR", ":8086")
	dsn := config.DSN()
	movieCatalogURL := envOr("MOVIE_CATALOG_URL", "http://movie-catalog:8091")
	
	if err := config.ValidateRuntime("bff-web"); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	var tgVisiblePostIDs []string
	if ids, _, err := tgexport.LoadVisibleSourcePostIDsFromRepo(config.RepoRoot()); err == nil && len(ids) > 0 {
		tgVisiblePostIDs = ids
	}

	// Создать reverse proxy для movie-catalog
	movieCatalogTarget, err := url.Parse(movieCatalogURL)
	if err != nil {
		log.Fatalf("invalid MOVIE_CATALOG_URL: %v", err)
	}
	movieCatalogProxy := httputil.NewSingleHostReverseProxy(movieCatalogTarget)

	// Создать основной handler
	mux := http.NewServeMux()
	
	// BFF endpoints
	bffHandler := bffweb.NewHandler(posts.NewService(pool, tgVisiblePostIDs), envOr("BFF_WEB_API_BASE_URL", config.APIBaseURL()))
	mux.Handle("/api/bff/web/v1/", bffHandler)
	
	// Movie catalog proxy
	mux.Handle("/api/movie-catalog/", http.StripPrefix("/api/movie-catalog", movieCatalogProxy))
	mux.Handle("/healthz/movie-catalog", http.StripPrefix("/healthz/movie-catalog", movieCatalogProxy))

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Printf("bff-web listening on %s", addr)
		log.Printf("proxying /api/movie-catalog/* to %s", movieCatalogURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
```

### 4.2 Добавление BFF-web в docker-compose

**Файл**: [`ops/compose/docker-compose.core.yml`](../ops/compose/docker-compose.core.yml)

```yaml
  bff-web:
    image: golang:1.24
    restart: unless-stopped
    working_dir: /app
    command: sh -lc "/usr/local/go/bin/go run ./services/bff-web/cmd"
    environment:
      BFF_WEB_ADDR: ":8086"
      DSN: "${DSN:-postgres://postgres:${POSTGRES_PASSWORD:-postgres}@db:5432/gallery?sslmode=disable}"
      MUDRO_ROOT: /app
      BFF_WEB_API_BASE_URL: "${BFF_WEB_API_BASE_URL:-http://api:8080}"
      MOVIE_CATALOG_URL: "http://movie-catalog:8091"
    depends_on:
      db:
        condition: service_healthy
      api:
        condition: service_healthy
      movie-catalog:
        condition: service_healthy
    ports:
      - "127.0.0.1:8086:8086"
    volumes:
      - ../../:/app
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O- http://127.0.0.1:8086/api/bff/web/v1/healthz || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 12
```

---

## 5. Интеграция Frontend

### 5.1 Копирование компонентов

**Структура для копирования:**

```bash
# Entities
cp -r MUDROTOP/frontend/src/entities/movie frontend/src/entities/

# Features
cp -r MUDROTOP/frontend/src/features/movie-filters frontend/src/features/

# Widgets
cp -r MUDROTOP/frontend/src/widgets/movie-catalog frontend/src/widgets/

# Pages
cp -r MUDROTOP/frontend/src/pages/movie-catalog-page frontend/src/pages/
```

### 5.2 Обновление MoviesPage

**Файл**: [`frontend/src/pages/movies-page/ui/MoviesPage.tsx`](../frontend/src/pages/movies-page/ui/MoviesPage.tsx)

```typescript
import { MovieCatalogPage } from '@/pages/movie-catalog-page/ui/MovieCatalogPage'
import { useTelegramWebApp } from '@/features/telegram-miniapp/hooks/useTelegramWebApp'

export const MoviesPage = () => {
  const { isTelegram, themeParams } = useTelegramWebApp()
  
  return (
    <div 
      style={isTelegram ? { 
        backgroundColor: themeParams?.bg_color,
        color: themeParams?.text_color,
        minHeight: '100vh'
      } : undefined}
    >
      <MovieCatalogPage />
    </div>
  )
}
```

### 5.3 Обновление API конфигурации

**Файл**: [`frontend/src/entities/movie/api/movieCatalogApi.ts`](../frontend/src/entities/movie/api/movieCatalogApi.ts)

```typescript
import { http } from '@/shared/api/http'
import type { GenreListResponse, MoviePage, MovieFilters } from '../model/types'

const BASE_URL = '/api/movie-catalog'

export const movieCatalogApi = {
  async getGenres(): Promise<GenreListResponse> {
    const response = await http.get<GenreListResponse>(`${BASE_URL}/genres`)
    return response
  },

  async getMovies(filters: MovieFilters): Promise<MoviePage> {
    const params = new URLSearchParams()
    
    if (filters.yearMin) params.append('year_min', filters.yearMin.toString())
    if (filters.durationMin) params.append('duration_min', filters.durationMin.toString())
    if (filters.includeGenre) params.append('include_genre', filters.includeGenre)
    if (filters.excludeGenres?.length) {
      filters.excludeGenres.forEach(g => params.append('exclude_genres', g))
    }
    params.append('page', filters.page.toString())
    params.append('page_size', filters.pageSize.toString())
    
    const response = await http.get<MoviePage>(`${BASE_URL}/movies?${params}`)
    return response
  }
}
```

### 5.4 Обновление Vite proxy

**Файл**: [`frontend/vite.config.ts`](../frontend/vite.config.ts)

```typescript
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'
import { defineConfig } from 'vite'

const apiProxyTarget = process.env.MUDRO_API_PROXY_TARGET ?? 'http://127.0.0.1:8080'
const bffProxyTarget = process.env.MUDRO_BFF_PROXY_TARGET ?? 'http://127.0.0.1:8086'

export default defineConfig({
  plugins: [tailwindcss(), react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'state-vendor': ['@reduxjs/toolkit', 'react-redux'],
          'motion-vendor': ['framer-motion'],
        },
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api/movie-catalog': {
        target: bffProxyTarget,
        changeOrigin: true,
      },
      '/api': apiProxyTarget,
      '/healthz': apiProxyTarget,
      '/feed': apiProxyTarget,
      '/media': apiProxyTarget,
    },
  },
})
```

---

## 6. Интеграция в Telegram Bot

### 6.1 Добавление команды /movies

**Файл**: [`internal/bot/handler.go`](../internal/bot/handler.go)

```go
// В функции RegisterBotCommands добавить:
{Command: "movies", Description: "Каталог фильмов 🎬"},

// В функции HandleCommands добавить case:
case "movies":
    handleMovies(bot, update)
```

### 6.2 Реализация handler

**Новый файл**: `internal/bot/movies.go`

```go
package bot

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleMovies(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	webAppURL := getMoviesWebAppURL()
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonWebApp("🎬 Открыть каталог фильмов", webAppURL),
		),
	)
	
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"🎬 *Каталог фильмов MUDROTOP*\n\n"+
			"Откройте каталог для просмотра фильмов с фильтрами:\n"+
			"• По году выпуска\n"+
			"• По длительности\n"+
			"• По жанрам\n"+
			"• С рейтингом и описанием",
	)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	if _, err := bot.Send(msg); err != nil {
		log.Printf("handleMovies send error: %v", err)
	}
}

func getMoviesWebAppURL() string {
	// В production это будет реальный URL
	baseURL := getEnvOr("MUDRO_WEB_URL", "https://mudro.example.com")
	return fmt.Sprintf("%s/movies", baseURL)
}

func getEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
```

### 6.3 Настройка Web App в BotFather

```
1. Открыть @BotFather
2. /mybots → выбрать бота
3. Bot Settings → Menu Button → Configure Menu Button
4. Указать URL: https://mudro.example.com/movies
5. Текст кнопки: "Фильмы 🎬"
```

---

## 7. Обновление документации

### 7.1 README.md

**Файл**: [`README.md`](../README.md)

Добавить секцию:

```markdown
## Микросервисы

### movie-catalog

Каталог фильмов с фильтрацией и пагинацией.

- **Порт**: 8091
- **API**: `/api/movie-catalog/movies`, `/api/movie-catalog/genres`
- **Документация**: [services/movie-catalog/README.md](services/movie-catalog/README.md)
- **Контракт**: [contracts/http/movie-catalog-v1.yaml](contracts/http/movie-catalog-v1.yaml)

**Запуск:**
```bash
make movie-catalog-migrate  # Миграции
make movie-catalog-import   # Импорт данных
make movie-catalog-run      # Запуск сервиса
```
```

### 7.2 Обновление service-catalog.md

**Файл**: [`docs/service-catalog.md`](../docs/service-catalog.md)

```markdown
## movie-catalog

**Порт**: 8091  
**Тип**: Read-only HTTP API  
**Статус**: ✅ Production Ready

### Назначение
Каталог фильмов с серверной фильтрацией и пагинацией.

### Endpoints
- `GET /healthz` - Health check
- `GET /api/movie-catalog/genres` - Список жанров
- `GET /api/movie-catalog/movies` - Фильтрованный список фильмов

### Зависимости
- PostgreSQL (схема `movie_catalog` в БД `gallery`)

### Интеграция
- Frontend: `/movies` страница
- BFF-web: проксирование через `/api/movie-catalog/*`
- Telegram: команда `/movies` → Mini App
```

---

## 8. Пошаговый план реализации

### Фаза 1: Подготовка БД (1-2 часа)

1. ✅ Обновить миграции для использования схемы `movie_catalog`
2. ✅ Изменить DSN в коде сервиса
3. ✅ Запустить миграции на основной БД
4. ✅ Импортировать данные

### Фаза 2: Docker Compose интеграция (30 минут)

1. ✅ Добавить `movie-catalog` в [`docker-compose.core.yml`](../ops/compose/docker-compose.core.yml)
2. ✅ Добавить `bff-web` в [`docker-compose.core.yml`](../ops/compose/docker-compose.core.yml)
3. ✅ Обновить [`Makefile`](../Makefile)
4. ✅ Протестировать запуск через `make core-up`

### Фаза 3: BFF-web проксирование (1 час)

1. ✅ Обновить [`services/bff-web/app/run.go`](../services/bff-web/app/run.go)
2. ✅ Добавить reverse proxy для movie-catalog
3. ✅ Протестировать endpoints через BFF

### Фаза 4: Frontend интеграция (2-3 часа)

1. ✅ Скопировать компоненты из MUDROTOP
2. ✅ Обновить [`MoviesPage`](../frontend/src/pages/movies-page/ui/MoviesPage.tsx)
3. ✅ Настроить API proxy в [`vite.config.ts`](../frontend/vite.config.ts)
4. ✅ Добавить Telegram theme support
5. ✅ Протестировать UI

### Фаза 5: Telegram интеграция (1 час)

1. ✅ Добавить команду `/movies` в [`internal/bot/handler.go`](../internal/bot/handler.go)
2. ✅ Реализовать `handleMovies` в `internal/bot/movies.go`
3. ✅ Настроить Web App в BotFather
4. ✅ Протестировать в Telegram

### Фаза 6: Документация (30 минут)

1. ✅ Обновить [`README.md`](../README.md)
2. ✅ Обновить [`docs/service-catalog.md`](../docs/service-catalog.md)
3. ✅ Создать migration guide

---

## 9. Тестирование

### 9.1 Чеклист тестирования

**Backend:**
- [ ] `curl http://127.0.0.1:8091/healthz` возвращает `{"status":"ok"}`
- [ ] `curl http://127.0.0.1:8091/api/movie-catalog/genres` возвращает список жанров
- [ ] `curl http://127.0.0.1:8091/api/movie-catalog/movies?page=1&page_size=12` возвращает фильмы

**BFF-web:**
- [ ] `curl http://127.0.0.1:8086/api/movie-catalog/genres` проксирует запрос
- [ ] `curl http://127.0.0.1:8086/api/movie-catalog/movies?page=1` проксирует запрос

**Frontend:**
- [ ] Страница `/movies` загружается без ошибок
- [ ] Фильтры работают корректно
- [ ] Пагинация работает
- [ ] Карточки фильмов отображаются

**Telegram:**
- [ ] Команда `/movies` отправляет сообщение с кнопкой
- [ ] Кнопка открывает Mini App
- [ ] Mini App применяет Telegram theme
- [ ] Навигация работает внутри Mini App

### 9.2 Smoke test команды

```bash
# Backend health
curl http://127.0.0.1:8091/healthz

# Genres через BFF
curl http://127.0.0.1:8086/api/movie-catalog/genres

# Movies через BFF
curl "http://127.0.0.1:8086/api/movie-catalog/movies?page=1&page_size=5"

# Frontend
curl http://127.0.0.1:5173/movies
```

---

## 10. Риски и митигация

| Риск | Вероятность | Влияние | Митигация |
|------|-------------|---------|-----------|
| Конфликт портов | Низкая | Низкое | Использовать стандартные порты из документации |
| Проблемы с миграцией БД | Средняя | Среднее | Тестировать на копии БД, использовать транзакции |
| CORS ошибки в production | Средняя | Среднее | Настроить CORS middleware в BFF-web |
| Конфликт зависимостей frontend | Низкая | Низкое | MUDROTOP использует те же версии React/Vite |
| Telegram Web App ограничения | Низкая | Низкое | Fallback на обычную веб-версию |

---

## 11. Rollback план

Если интеграция вызывает проблемы:

1. **Откатить docker-compose**: Удалить `movie-catalog` и `bff-web` из [`docker-compose.core.yml`](../ops/compose/docker-compose.core.yml)
2. **Откатить frontend**: Восстановить заглушку в [`MoviesPage`](../frontend/src/pages/movies-page/ui/MoviesPage.tsx)
3. **Откатить Telegram**: Удалить команду `/movies` из [`handler.go`](../internal/bot/handler.go)
4. **Откатить БД**: Удалить схему `movie_catalog` (если данные не критичны)

```sql
DROP SCHEMA IF EXISTS movie_catalog CASCADE;
```

---

## 12. Production checklist

Перед деплоем в production:

- [ ] Добавить метрики (Prometheus)
- [ ] Настроить structured logging
- [ ] Добавить rate limiting в BFF-web
- [ ] Настроить CORS для production домена
- [ ] Написать integration тесты
- [ ] Настроить мониторинг и алерты
- [ ] Обновить CI/CD pipeline
- [ ] Создать backup БД перед миграцией
- [ ] Документировать production URLs
- [ ] Настроить SSL для Telegram Web App

---

## 13. Следующие шаги после интеграции

1. **Оптимизация производительности**
   - Добавить индексы в БД
   - Настроить кеширование в Redis
   - Оптимизировать SQL запросы

2. **Расширение функциональности**
   - Добавить поиск по названию
   - Добавить избранное
   - Добавить рекомендации

3. **Улучшение UX**
   - Добавить skeleton loaders
   - Оптимизировать загрузку изображений
   - Добавить infinite scroll

---

## Заключение

План интеграции MUDROTOP в Скаро готов к реализации. Все компоненты уже существуют и требуют только конфигурационных изменений и копирования кода. Интеграция займет примерно 6-8 часов работы и обеспечит полное подключение к единой БД, API Gateway, BFF-web и Telegram bot.

**Готовность к старту**: ✅ Все зависимости проверены, план детализирован, риски идентифицированы.
