---
name: go-testing-patterns
description: Идиоматические Go-паттерны для тестирования и TDD workflow в проекте mudro
---

# Skill: Go Testing Patterns

Адаптирован для проекта `mudro` (Go 1.22+, Postgres, Docker).

## Базовые правила тестирования в mudro

- Тесты запускать через: `go test ./...` (из корня через WSL или Makefile)
- Тест-файлы: `<package>_test.go` рядом с пакетом
- Benchmark и integration тесты — с build-тегами
- **Никаких** тестов с реальными credentials в коде

## Структура теста (table-driven)

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   SomeType
        want    SomeResult
        wantErr bool
    }{
        {
            name:  "happy path",
            input: validInput,
            want:  expectedOutput,
        },
        {
            name:    "invalid input",
            input:   invalidInput,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && got != tt.want {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Integration тесты (с БД)

```go
//go:build integration

func TestSomeDBOperation(t *testing.T) {
    dsn := os.Getenv("TEST_DSN")
    if dsn == "" {
        t.Skip("TEST_DSN not set, skipping integration test")
    }
    // ... тест
}
```

Запуск: `go test -tags=integration ./...`

## Тестирование HTTP endpoints (cmd/api)

```go
func TestHandlerGetFront(t *testing.T) {
    // Использовать httptest.NewRecorder + httptest.NewRequest
    w := httptest.NewRecorder()
    r := httptest.NewRequest(http.MethodGet, "/api/front?limit=5", nil)
    
    handler.ServeHTTP(w, r)
    
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
    }
}
```

## Анти-паттерны (избегать в mudro)

| Плохо | Хорошо |
|-------|--------|
| `t.Fatal("error")` без контекста | `t.Fatalf("ParseFeed() error: %v", err)` |
| Хардкод DSN в тестах | `os.Getenv("TEST_DSN")` или `t.Skip()` |
| Global state в тестах | `t.Cleanup(func() { /* teardown */ })` |
| Тесты зависят от порядка | Каждый тест самодостаточен |

## Запуск тестов локально

```powershell
# Через Makefile (предпочтительно)
wsl -d Ubuntu -- bash -c "cd ~/projects/mudro && make test"

# Или напрямую через WSL
wsl -d Ubuntu -- bash -c "cd ~/projects/mudro && go test ./... -v -count=1"

# Конкретный пакет
wsl -d Ubuntu -- bash -c "cd ~/projects/mudro && go test ./internal/... -v"
```

## Покрытие кода

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # визуализация
```
