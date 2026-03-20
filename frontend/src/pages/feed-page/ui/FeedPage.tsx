import './FeedPage.css'
import { FeedWidget } from '@/widgets/feed/ui/FeedWidget'

const sectionLinks = [
  { href: '#feed', label: 'Лента' },
  { href: '#accounts', label: 'Слои' },
  { href: '#social', label: 'Дальше' },
]

export const FeedPage = () => {
  return (
    <main className="mudro-shell">
      <header className="feed-page-header mudro-fade-up" aria-label="Навигация Mudro">
        <a className="feed-page-header__logo" href="#feed" aria-label="Mudro home">
          <span className="feed-page-header__logo-mark">M</span>
          <span className="feed-page-header__logo-text">
            <strong>Mudro</strong>
            <small>живой архив</small>
          </span>
        </a>

        <div className="feed-page-header__cluster">
          <nav className="feed-page-header__nav" aria-label="Разделы">
            {sectionLinks.map((section) => (
              <a key={section.href} className="feed-page-header__link" href={section.href}>
                {section.label}
              </a>
            ))}
          </nav>
          <span className="feed-page-header__status">MVP · server live</span>
        </div>
      </header>

      <section className="feed-page-live mudro-fade-up" id="feed" aria-label="Живая лента Mudro">
        <FeedWidget />
      </section>

      <section className="feed-page-next mudro-fade-up" id="accounts" aria-labelledby="next-title">
        <div className="feed-page-next__panel">
          <span className="feed-page-next__eyebrow">После ленты</span>
          <h2 id="next-title">Следующий слой уже понятен: аккаунты, личные лайки и более живые треды</h2>
          <p>
            Базовый MVP уже уверенно показывает архив. Дальше интерфейс растет не в ширину, а вглубь: персональные
            наборы источников, действия пользователя и телеграмный social layer поверх той же ленты.
          </p>
        </div>
      </section>

      <section className="feed-page-next feed-page-next_social mudro-fade-up" id="social" aria-label="Дальнейшее развитие интерфейса">
        <div className="feed-page-next__grid">
          <article className="feed-page-next__card">
            <span>Аккаунты</span>
            <strong>Несколько профилей и свои наборы источников без смены основного сценария</strong>
            <p>Поверх текущей ленты можно добавить персональные пространства без нового экрана и без отдельного продукта.</p>
          </article>
          <article className="feed-page-next__card">
            <span>Лайки</span>
            <strong>Живые действия пользователя: лайки, статусы чтения и быстрые паттерны как в Telegram</strong>
            <p>Когда лента стабильна, сюда естественно ложатся свои реакции и короткие действия без перегруза интерфейса.</p>
          </article>
          <article className="feed-page-next__card">
            <span>Комментарии</span>
            <strong>Треды, реплаи и media внутри обсуждений как отдельный рабочий слой</strong>
            <p>Основа уже есть: комментарии, реакции и вложения. Следующий шаг — сделать обсуждения быстрее и аккуратнее.</p>
          </article>
        </div>
      </section>
    </main>
  )
}
