# Frontend

Новый активный UI-контур для staging-подготовки `MudroTop` к будущему merge в `mudro`.

## Стек

- Vite
- React
- TypeScript
- ESLint flat config
- FSD layout

## Структура

- `src/pages/movie-catalog-page` - orchestration page
- `src/widgets/movie-catalog` - grid, states, pagination
- `src/features/movie-filters` - draft/apply/reset
- `src/entities/movie` - typed contract and API client
- `src/shared` - env, HTTP helper, primitives, CSS tokens

## Локальный API

По умолчанию frontend ходит в `http://127.0.0.1:8091/api`.

Для override используйте `VITE_MOVIE_CATALOG_API_BASE_URL`.
