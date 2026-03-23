# Исправление проблемы с E2E тестами в CI

**Дата**: 2026-03-23  
**Статус**: Готово к реализации  
**Приоритет**: Критический

---

## Проблема

E2E тест [`TestCmdAPISmokeHealthz`](../e2e/cmd_smoke_test.go) падает в CI job `test-backend`, потому что:

1. Job `test-backend` запускает `make test-active` → `go test ./...`
2. E2E тест требует PostgreSQL на `localhost:5433` (строка 77)
3. В job `test-backend` нет PostgreSQL service
4. PostgreSQL есть только в отдельном job `smoke-e2e`

**Текущий код в Makefile:**

```makefile
test-active:
	$(GO) test ./...
```

**Проблема:** Команда `go test ./...` включает все пакеты, включая `e2e/`.

---

## Решение

### Вариант 1: Исключить e2e из test-active (рекомендуется)

Изменить [`Makefile`](../Makefile) строки 180-181:

```makefile
test-active:
	$(GO) test $(shell $(GO) list ./... | grep -v /e2e)
```

**Преимущества:**
- ✅ E2E тесты запускаются только в job `smoke-e2e`, где есть PostgreSQL
- ✅ Job `test-backend` становится быстрым (только unit/integration тесты)
- ✅ Четкое разделение ответственности между jobs
- ✅ Не требует изменений в CI конфигурации

**Недостатки:**
- Нет

---

### Вариант 2: Добавить PostgreSQL в test-backend (не рекомендуется)

Дублировать секцию `services:` из job `smoke-e2e` в `test-backend`.

**Преимущества:**
- E2E тесты запускаются в обоих jobs

**Недостатки:**
- ❌ Избыточно — уже есть отдельный job для e2e
- ❌ Замедляет job `test-backend`
- ❌ Дублирование конфигурации
- ❌ Нарушает принцип единственной ответственности

---

## Корневая причина

Команда `make test-active` запускает все тесты включая e2e, но job `test-backend` предназначен только для unit/integration тестов без внешних зависимостей.

**Название job вводит в заблуждение:**
- Текущее: `test-backend`
- Лучше: `test-units` или `test-fast`

Но переименование job не обязательно, если исправить `test-active`.

---

## Реализация

### Шаг 1: Обновить Makefile

**Файл**: [`Makefile`](../Makefile)

**Было:**
```makefile
test-active:
	$(GO) test ./...
```

**Стало:**
```makefile
test-active:
	$(GO) test $(shell $(GO) list ./... | grep -v /e2e)
```

### Шаг 2: Проверить локально

```bash
# Проверить, что e2e исключены
go list ./... | grep -v /e2e

# Запустить test-active
make test-active

# Убедиться, что e2e тесты не запускаются
# (не должно быть ошибок про PostgreSQL)
```

### Шаг 3: Проверить в CI

После коммита изменений:

1. Job `test-backend` должен пройти успешно (без e2e)
2. Job `smoke-e2e` должен запустить e2e тесты с PostgreSQL

---

## Альтернативные подходы

### Подход A: Build tags

Добавить build tag в e2e тесты:

```go
//go:build e2e

package e2e
```

Тогда:
- `go test ./...` — пропускает e2e
- `go test -tags=e2e ./e2e` — запускает e2e

**Недостаток:** Требует изменения всех e2e файлов.

### Подход B: Отдельная директория вне модуля

Переместить e2e в `test/e2e/` с отдельным `go.mod`.

**Недостаток:** Усложняет структуру проекта.

### Подход C: Short mode

Использовать `testing.Short()` в e2e тестах:

```go
func TestCmdAPISmokeHealthz(t *testing.T) {
	if testing.Short() {
		t.Skip("skip smoke test in short mode")
	}
	// ...
}
```

Тогда:
- `go test -short ./...` — пропускает e2e
- `go test ./e2e` — запускает e2e

**Недостаток:** E2E тесты уже используют `testing.Short()`, но `make test-active` не передает `-short`.

---

## Рекомендация

**Использовать Вариант 1**: Исключить e2e из test-active через `grep -v /e2e`.

Это самое простое, чистое и эффективное решение, которое:
- Не требует изменений в e2e тестах
- Не требует изменений в CI конфигурации
- Четко разделяет unit/integration и e2e тесты
- Соответствует текущей архитектуре CI (отдельный job для e2e)

---

## Проверка

После внедрения проверить:

1. ✅ `make test-active` не запускает e2e тесты
2. ✅ Job `test-backend` проходит успешно
3. ✅ Job `smoke-e2e` запускает e2e тесты
4. ✅ E2E тесты проходят в job `smoke-e2e`

---

## Связанные файлы

- [`Makefile`](../Makefile) — строки 180-181
- [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) — jobs `test-backend` и `smoke-e2e`
- [`e2e/cmd_smoke_test.go`](../e2e/cmd_smoke_test.go) — E2E тесты

---

## Заключение

Проблема с E2E тестами в CI решается одной строкой в Makefile. Это критическое исправление, которое нужно внедрить немедленно, чтобы CI проходил успешно.
