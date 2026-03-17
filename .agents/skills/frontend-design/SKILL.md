---
name: frontend-design
description: Принципы premium UI/UX для React+TS frontend mudro — компоненты, анимации, мобильная адаптация
---

# Skill: Frontend Design Patterns

Адаптирован под стек mudro: React 19 + TypeScript + RTK + Vite + FSD.
Дизайн-система: розовый градиент (`#ff69b4`), стеклянные карточки, Space Grotesk.

## Принципы дизайна mudro

1. **Реальные данные, никаких placeholder** — если данных нет, показывать `FeedLoadingSkeleton`
2. **Mobile-first** — drawer, spacing, font-size проверять на 375px
3. **Micro-анимации** обязательны — `mudro-fade-up`, hover transitions
4. **Glass morphism** — `backdrop-filter: blur()`, полупрозрачные карточки
5. **Semantic HTML** — `<article>`, `<section>`, `<header>`, правильные `aria-label`

## Дизайн-токены (из variables.css)

```css
--mudro-bg-base: #ff69b4;      /* основной фон */
--mudro-card: rgba(255, 255, 255, 0.94);  /* карточка */
--mudro-text: #2a0f2d;          /* основной текст */
--mudro-muted: #735977;         /* приглушённый */
--mudro-vk: #3a78ff;            /* VK accent */
--mudro-tg: #2aabee;            /* Telegram accent */
--mudro-radius-lg: 24px;
--mudro-shadow: 0 18px 55px rgba(19, 10, 23, 0.2);
```

## Паттерны компонентов FSD

### Структура нового компонента
```
src/
  entities/<name>/
    ui/<Name>/<Name>.tsx
    ui/<Name>/<Name>.css
    model/types.ts    (если нужно)
```

### Базовый шаблон карточки
```tsx
interface <Name>CardProps {
  data: <Type>
  onOpen?: (item: <Type>) => void
}

export const <Name>Card = ({ data, onOpen }: <Name>CardProps) => {
  return (
    <article
      className="<name>-card"
      onClick={() => onOpen?.(data)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => e.key === 'Enter' && onOpen?.(data)}
    >
      {/* content */}
    </article>
  )
}
```

## Mobile Polish чеклист
- [ ] Проверить на 375px ширине (iPhone SE)
- [ ] `drawer` открывается снизу на мобильных (не сбоку)
- [ ] `touch-action: manipulation` на кликабельных элементах
- [ ] `font-size` минимум `14px` для body текста
- [ ] `padding` не меньше `1rem` по краям

```css
/* Mobile-first пример */
.post-card {
  padding: 1rem;
  border-radius: var(--mudro-radius-md);
}

@media (min-width: 768px) {
  .post-card {
    padding: 1.4rem;
    border-radius: var(--mudro-radius-lg);
  }
}
```

## Анимации (уже в global.css)
```css
/* Использовать класс для появления элементов */
.mudro-fade-up {
  animation: mudroFadeUp 420ms ease-out both;
}

/* Hover на кнопках */
.mudro-btn {
  transition: background 140ms, transform 80ms, box-shadow 140ms;
}
.mudro-btn:hover { transform: translateY(-1px); }
.mudro-btn:active { transform: translateY(0); }
```

## SEO (для index.html)
- `<title>` — конкретный и описательный
- `<meta name="description">` — первые 160 символов убедительны
- OG-теги обновлены: title, description, locale, site_name
- `lang="ru"` на `<html>`

## Частые ошибки
| Плохо | Хорошо |
|-------|--------|
| `<div onClick>` | `<button type="button">` или `role="button"` + `tabIndex` |
| Прямой `style={{color: 'red'}}` | CSS-переменная `var(--mudro-accent)` |
| `!important` в CSS | Правильная специфичность |
| `z-index: 9999` | Понятные слои: `--z-drawer: 100` |
