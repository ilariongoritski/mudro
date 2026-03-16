import './FeedPage.css'
import { FeedWidget } from '@/widgets/feed/ui/FeedWidget'

const overviewItems = [
  {
    label: 'Unified stream',
    title: 'VK и Telegram в одном браузерном контуре',
    description:
      'Лента уже живет поверх реального API, а не поверх статичного лендинга. Это важнее декоративной витрины.',
  },
  {
    label: 'Media-first',
    title: 'Карточки собраны вокруг контента и вложений',
    description:
      'Фокус на тексте, превью и метриках. Пользователь считывает пост до клика, а не после перехода.',
  },
  {
    label: 'Vercel-ready',
    title: 'Публичная страница и живая лента не конфликтуют',
    description:
      'Один экран уже можно показывать как draft продукта, пока ниже работает реальная подгрузка данных.',
  },
]

const workflowItems = [
  {
    step: '01',
    title: 'Открываешь страницу',
    description: 'Сразу читаешь, что такое Mudro и зачем ему объединенная лента.',
  },
  {
    step: '02',
    title: 'Смотришь состояние архива',
    description: 'Toolbar показывает source, сортировку, размер выборки и общий объем текущего потока.',
  },
  {
    step: '03',
    title: 'Переходишь к живому feed',
    description: 'Ниже идет не mockup, а реальный поток из `/api/front` с дозагрузкой и media-карточками.',
  },
  {
    step: '04',
    title: 'Дальше усиливаем detail layer',
    description: 'Следующий слой MVP — detail drawer и polished states, а не очередной декоративный блок.',
  },
]

const roadmapItems = [
  'Вынести отдельный landing route поверх той же визуальной системы.',
  'Добавить detail drawer для поста вместо простого чтения в grid.',
  'Закрыть loading, empty и error состояния как полноценные product surfaces.',
]

export const FeedPage = () => {
  return (
    <main className="mudro-shell">
      <section className="feed-page-hero mudro-fade-up">
        <div className="feed-page-hero__copy">
          <div className="feed-page-hero__eyebrow">
            <span className="feed-page-hero__badge">Magic draft</span>
            <span>Frontend Mudro11 · React + TS + Redux Toolkit + RTK Query + FSD</span>
          </div>
          <h1 className="mudro-title feed-page-hero__title">Mudro Feed Console</h1>
          <p className="mudro-lead feed-page-hero__lead">
            Быстрый браузерный слой для объединенной ленты VK и Telegram. Фокус на читаемость,
            дозагрузку, фильтры и подготовку к публичному веб-контуру.
          </p>
          <div className="feed-page-hero__actions">
            <a className="feed-page-hero__action feed-page-hero__action_primary" href="#feed">
              Открыть ленту
            </a>
            <a className="feed-page-hero__action" href="#focus">
              Смотреть фокус
            </a>
          </div>
        </div>

        <div className="feed-page-hero__panel" aria-label="Mudro focus panel">
          <div className="feed-page-hero__panel-line">
            <span className="feed-page-hero__panel-label">focus</span>
            <span>web-first content archive</span>
          </div>
          <div className="feed-page-hero__panel-grid">
            <article>
              <strong>01</strong>
              <span>Общая лента с cursor pagination</span>
            </article>
            <article>
              <strong>02</strong>
              <span>Media-first карточки и source badges</span>
            </article>
            <article>
              <strong>03</strong>
              <span>Подготовка к Vercel и публичному API</span>
            </article>
          </div>
        </div>
      </section>

      <section className="feed-page-focus mudro-fade-up" id="focus">
        {overviewItems.map((item) => (
          <article key={item.label} className="feed-page-focus__card">
            <span>{item.label}</span>
            <strong>{item.title}</strong>
            <p>{item.description}</p>
          </article>
        ))}
      </section>

      <section className="feed-page-journey mudro-fade-up" aria-labelledby="journey-title">
        <div className="feed-page-section-head">
          <span className="feed-page-section-head__eyebrow">UI MVP surface</span>
          <h2 id="journey-title" className="feed-page-section-head__title">
            Страница уже выглядит как продукт, а не как технический стенд
          </h2>
          <p className="feed-page-section-head__lead">
            Верх экрана продает идею Mudro, средний слой объясняет сценарий использования, а
            нижний блок показывает уже живой архивный поток.
          </p>
        </div>

        <div className="feed-page-journey__grid">
          {workflowItems.map((item) => (
            <article key={item.step} className="feed-page-journey__card">
              <strong>{item.step}</strong>
              <h3>{item.title}</h3>
              <p>{item.description}</p>
            </article>
          ))}
        </div>
      </section>

      <section className="feed-page-live mudro-fade-up" id="feed" aria-labelledby="live-feed-title">
        <div className="feed-page-live__intro">
          <span className="feed-page-live__eyebrow">Live archive</span>
          <h2 id="live-feed-title" className="feed-page-live__title">
            Живая лента уже встроена в страницу как основной продуктовый surface
          </h2>
          <p className="feed-page-live__lead">
            Ниже нет отдельного демо-режима. Здесь тот же контур, который дальше будет жить в
            публичном вебе: фильтры, mixed feed, подгрузка и media-first карточки.
          </p>
        </div>

        <div className="feed-page-live__layout">
          <aside className="feed-page-live__rail" aria-label="Что уже закрыто">
            <article className="feed-page-live__rail-card">
              <span>Сейчас</span>
              <strong>Можно читать mixed feed и быстро переключать источники</strong>
              <p>Toolbar уже собирает реальные метрики по ленте и не притворяется статичной витриной.</p>
            </article>
            <article className="feed-page-live__rail-card">
              <span>Дальше</span>
              <strong>Осталось закрыть detail layer и polished states</strong>
              <p>Именно это отделяет текущий draft от уверенного MVP интерфейса.</p>
            </article>
          </aside>

          <div className="feed-page-live__feed">
            <FeedWidget />
          </div>
        </div>
      </section>

      <section className="feed-page-roadmap mudro-fade-up" aria-labelledby="roadmap-title">
        <div className="feed-page-section-head">
          <span className="feed-page-section-head__eyebrow">Next UI steps</span>
          <h2 id="roadmap-title" className="feed-page-section-head__title">
            Минимальный визуальный план после этой итерации
          </h2>
        </div>

        <div className="feed-page-roadmap__grid">
          {roadmapItems.map((item, index) => (
            <article key={item} className="feed-page-roadmap__card">
              <strong>0{index + 1}</strong>
              <p>{item}</p>
            </article>
          ))}
        </div>
      </section>
    </main>
  )
}
