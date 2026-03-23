import { Link } from 'react-router-dom'

import { CasinoHeroStrip } from '@/features/casino/ui/CasinoHeroStrip'
import { CasinoNavButton } from '@/features/casino/ui/CasinoNavButton'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'
import { FeedWidget } from '@/widgets/feed/ui/FeedWidget'

import './FeedPage.css'

const sectionLinks = [
  { href: '#feed', label: 'Лента' },
  { href: '#accounts', label: 'Аккаунты' },
  { href: '#social', label: 'Соцслой' },
]

export const FeedPage = () => {
  const isAuthenticated = useAppSelector((state) => state.session.isAuthenticated)

  return (
    <main className="mudro-shell">
      <header className="feed-page-header mudro-fade-up" aria-label="Навигация Mudro">
        <a className="feed-page-header__logo" href="#feed" aria-label="Mudro home">
          <span className="feed-page-header__logo-mark">M</span>
          <span className="feed-page-header__logo-text">
            <strong>Mudro</strong>
            <small>личный архив</small>
          </span>
        </a>

        <nav className="feed-page-header__nav" aria-label="Разделы">
          {sectionLinks.map((section) => (
            <a key={section.href} className="feed-page-header__link" href={section.href}>
              {section.label}
            </a>
          ))}
          <CasinoNavButton />
          <Link to="/orchestration#bridge" className="feed-page-header__link">
            Карта bridge
          </Link>
          {!isAuthenticated && (
            <div className="feed-page-header__auth">
              <Link to="/login" className="feed-page-header__link feed-page-header__link_auth">
                Вход
              </Link>
              <Link to="/register" className="feed-page-header__link feed-page-header__link_auth feed-page-header__link_register">
                Регистрация
              </Link>
            </div>
          )}
          {isAuthenticated && (
            <div className="feed-page-header__auth">
              <Link to="/admin" className="feed-page-header__link feed-page-header__link_auth">
                Админ
              </Link>
            </div>
          )}
        </nav>
      </header>

      <section className="feed-page-hero mudro-fade-up" aria-label="MUDRO overview">
        <div className="feed-page-hero__copy">
          <span className="feed-page-hero__eyebrow">MUDRO workspace</span>
          <h1 className="feed-page-hero__title">Единая лента, control plane и изолированный casino-слой.</h1>
          <p className="feed-page-hero__lead">
            Это домашняя поверхность проекта: архив, фильтры, переход в orchestration и отдельный игровой экран без
            смешивания runtime-слоёв. Локальный Claude Opus bridge и Magic MCP описаны в control plane.
          </p>

          <div className="feed-page-hero__actions">
            <Link to="/orchestration#bridge" className="feed-page-hero__action">
              Открыть bridge map
            </Link>
            <Link to="/casino" className="feed-page-hero__action feed-page-hero__action_secondary">
              Открыть casino
            </Link>
          </div>
        </div>

        <div className="feed-page-hero__rail" aria-label="Платформенные сигналы">
          <article className="feed-page-hero__tile">
            <span>Feed API</span>
            <strong>Live archive</strong>
            <p>Показывает поток постов и медиа без потери контекста.</p>
          </article>
          <article className="feed-page-hero__tile">
            <span>Opus bridge</span>
            <strong>Local reasoning</strong>
            <p>Планирование, ревью и большие изменения идут локально через ключ.</p>
          </article>
          <article className="feed-page-hero__tile">
            <span>Magic MCP</span>
            <strong>Context & UI</strong>
            <p>Используется для визуальных и контекстных ориентиров, а не как runtime.</p>
          </article>
        </div>
      </section>

      <section className="feed-page-live mudro-fade-up" id="feed" aria-label="Живая лента Mudro">
        <FeedWidget />
      </section>

      <CasinoHeroStrip />

      <section className="feed-page-next mudro-fade-up" id="accounts" aria-labelledby="next-title">
        <div className="feed-page-next__panel">
          <span className="feed-page-next__eyebrow">Следующий слой</span>
          <h2 id="next-title">Поверх ленты дальше добавляются аккаунты, подписки, лайки и телеграмный social layer</h2>
          <p>
            Текущий MVP уже уверенно показывает архив. Следующий шаг — дать пользователю свои аккаунты, реакции и более
            живое поведение обсуждений, не ломая основной сценарий чтения.
          </p>
        </div>
      </section>

      <section className="feed-page-next feed-page-next_social mudro-fade-up" id="social" aria-label="Дальнейшее развитие интерфейса">
        <div className="feed-page-next__grid">
          <article className="feed-page-next__card">
            <span>Аккаунты</span>
            <strong>Несколько профилей и списков источников в одном интерфейсе</strong>
            <p>Лента уже готова расти в сторону персональных наборов аккаунтов без переделки основного чтения.</p>
          </article>
          <article className="feed-page-next__card">
            <span>Реакции</span>
            <strong>Живые лайки, статусы чтения и действия на уровне Telegram-паттернов</strong>
            <p>После стабилизации экрана сюда естественно ложатся пользовательские действия и более живой social-функционал.</p>
          </article>
          <article className="feed-page-next__card">
            <span>Комментарии</span>
            <strong>Треды, реплаи и media внутри обсуждений</strong>
            <p>Основа уже есть: комментарии, реакции и вложения. Дальше остаётся только расширять удобство работы с ними.</p>
          </article>
        </div>
      </section>
    </main>
  )
}
