import { FeedWidget } from '@/widgets/feed/ui/FeedWidget'

import './FeedPage.css'

export const FeedPage = () => {
  return (
    <main className="mudro-shell feed-page mudro-fade-up">
      <section className="feed-page-hero" aria-label="MUDRO overview">
        <div className="feed-page-hero__copy">
          <span className="feed-page-hero__eyebrow">Лента</span>
          <h1 className="feed-page-hero__title">Ваш поток постов, медиа и обновлений.</h1>
          <p className="feed-page-hero__lead">
            Единый архив из всех источников: ВКонтакте, Telegram и других платформ. Читайте, фильтруйте и следите за обновлениями в реальном времени.
          </p>
        </div>

        <div className="feed-page-hero__rail" aria-label="Статистика">
          <article className="feed-page-hero__tile">
            <span>Источники</span>
            <strong>VK + Telegram</strong>
            <p>Посты из нескольких платформ в одной ленте.</p>
          </article>
          <article className="feed-page-hero__tile">
            <span>Архив</span>
            <strong>Live archive</strong>
            <p>Данные хранятся локально и обновляются автоматически.</p>
          </article>
          <article className="feed-page-hero__tile">
            <span>Доступ</span>
            <strong>Без регистрации</strong>
            <p>Просматривайте ленту без создания аккаунта.</p>
          </article>
        </div>
      </section>

      <section className="feed-page-live mudro-fade-up" id="feed" aria-label="Лента">
        <FeedWidget />
      </section>
    </main>
  )
}
