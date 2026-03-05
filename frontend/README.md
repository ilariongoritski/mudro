# Mudro11 Frontend

Frontend-часть Mudro11 на стеке:
- React
- TypeScript
- Redux Toolkit
- RTK Query
- Vite
- FSD архитектура

## Запуск
Из папки `frontend/`:

```bash
npm.cmd install
npm.cmd run dev
```

По умолчанию Vite поднимается на `http://127.0.0.1:5173` и проксирует API-запросы на backend `http://127.0.0.1:8080`.

## Переменные окружения
- `VITE_API_BASE_URL` — базовый URL API.
  - пусто/не задано: используется `/` (proxy в dev, same-origin в prod).

## Проверка
```bash
npm.cmd run lint
npm.cmd run build
```

## Структура (FSD)
- `src/app` — инициализация приложения, store, глобальные стили.
- `src/pages` — страницы.
- `src/widgets` — крупные композиционные блоки UI.
- `src/features` — пользовательские сценарии (фильтры/сортировка).
- `src/entities` — доменные сущности (post).
- `src/shared` — переиспользуемые утилиты, API-base, конфиг.
