import './FeedPage.css'
import { FeedWidget } from '@/widgets/feed/ui/FeedWidget'
import { Link } from 'react-router-dom'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'

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
          {!isAuthenticated && (
            <div className="feed-page-header__auth">
              <Link to="/login" className="feed-page-header__link feed-page-header__link_auth">Вход</Link>
              <Link to="/register" className="feed-page-header__link feed-page-header__link_auth feed-page-header__link_register">Регистрация</Link>
            </div>
          )}
          {isAuthenticated && (
             <div className="feed-page-header__auth">
                <Link to="/admin" className="feed-page-header__link feed-page-header__link_auth">Админ</Link>
             </div>
          )}
        </nav>
      </header>

      <section className="feed-page-live mudro-fade-up" id="feed" aria-label="Живая лента Mudro">
        <FeedWidget />
      </section>

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
            <p>После стабилизации экрана сюда естественно ложатся пользовательские лайки и более живой social-функционал.</p>
          </article>
          <article className="feed-page-next__card">
            <span>Комментарии</span>
            <strong>Треды, реплаи и media внутри обсуждений</strong>
            <p>Основа уже есть: комментарии, реакции и вложения. Дальше остается только расширять удобство работы с ними.</p>
          </article>
        </div>
      </section>
    </main>
  )
}
