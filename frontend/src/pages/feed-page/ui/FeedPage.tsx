import { FeedWidget } from '@/widgets/feed/ui/FeedWidget'
import { Badge } from '@/shared/ui'

import './FeedPage.css'

export const FeedPage = () => {
  return (
    <main className="mudro-shell feed-page mudro-fade-up">
      <header className="feed-page-header">
        <div className="feed-page-header__left">
          <h1 className="feed-page-header__title">Лента</h1>
          <p className="feed-page-header__sub">Посты из ВКонтакте и Telegram в одном месте</p>
        </div>
        <div className="feed-page-header__sources">
          <Badge variant="vk">VK</Badge>
          <Badge variant="tg">TG</Badge>
        </div>
      </header>

      <section id="feed" aria-label="Лента">
        <FeedWidget />
      </section>
    </main>
  )
}
