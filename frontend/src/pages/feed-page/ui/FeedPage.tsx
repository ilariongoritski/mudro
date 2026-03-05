import { FeedWidget } from '@/widgets/feed/ui/FeedWidget'

export const FeedPage = () => {
  return (
    <main className="mudro-shell">
      <section className="mudro-hero mudro-fade-up">
        <p>Frontend Mudro11 · React + TS + Redux Toolkit + RTK Query + FSD</p>
        <h1 className="mudro-title">Mudro Feed Console</h1>
        <p className="mudro-lead">
          Быстрый интерфейс для просмотра объединенной ленты VK/TG с фильтрами, сортировкой и дозагрузкой.
        </p>
      </section>

      <FeedWidget />
    </main>
  )
}
