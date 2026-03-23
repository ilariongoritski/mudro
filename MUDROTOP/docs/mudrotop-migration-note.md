# MudroTop migration note

## Что сохранено концептуально

- фильтр по минимальному году
- фильтр по минимальной длительности
- include genre
- exclude genres
- карточка фильма
- кнопочная модель `Применить` / `Сбросить` / `Назад` / `Вперёд` / `Открыть карточку`

## Что переписано заново

- весь server/data access контур
- staging DB architecture
- frontend shell и визуальная иерархия
- data pipeline `raw -> slim -> postgres -> http -> ui`

## Что не считается активной базой

- старый `CRA` root
- giant JSON в `src`
- любая клиентская фильтрация полной выборки
- смешение SQL, parsing и transport logic в одном месте
