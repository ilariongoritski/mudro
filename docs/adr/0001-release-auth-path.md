# ADR 0001: Release Auth Path

## Status
Accepted

## Date
2026-04-11

## Context

Во фронтенде появился незавершённый поворот в сторону Supabase Auth, при этом backend и публичный API продолжают жить на собственном JWT и `users`/`auth` контуре.

Смешение двух auth-моделей ломает предсказуемость входа, усложняет демо и повышает риск регрессии в релизе.

## Decision

Для текущего релизного цикла используется только backend auth:

- `/api/auth/login`
- `/api/auth/register`
- `/api/auth/me`
- `/api/auth/refresh`
- `/api/auth/telegram`

Supabase Auth не входит в release path и не используется в пользовательском UI текущего выпуска.

## Consequences

Плюсы:

- один предсказуемый login flow;
- единая модель `session` во фронтенде;
- меньше скрытых отказов в демо и релизе.

Минусы:

- Supabase migration path откладывается;
- прямой client-side auth через Supabase не поддерживается в этом цикле.

## Follow-up

- Если Supabase понадобится позже, это должен быть отдельный архитектурный эпик с RLS, user model migration и новым ADR.
- Showcase/demo path должен опираться только на backend JWT flow и не смешивать его с Supabase traces.
