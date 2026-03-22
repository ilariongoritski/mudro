import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'

import './CasinoHeroStrip.css'

export const CasinoHeroStrip = () => {
  return (
    <motion.section
      className="casino-hero-strip mudro-fade-up"
      initial={{ opacity: 0, y: 18 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.45, ease: 'easeOut' }}
      aria-label="Casino feature"
    >
      <div className="casino-hero-strip__copy">
        <span className="casino-hero-strip__eyebrow">MUDRO Casino</span>
        <h2>Отдельный игровой контур с отдельной БД и Telegram Mini App интерфейсом</h2>
        <p>
          Базовый web-режим остается доступным, а быстрый вход для Telegram запускается в отдельном miniapp-маршруте с
          теми же API и бизнес-логикой.
        </p>
      </div>

      <div className="casino-hero-strip__actions">
        <Link to="/tma/casino" className="casino-hero-strip__primary">
          Открыть mini app
        </Link>
        <div className="casino-hero-strip__signals" aria-label="Краткая сводка casino">
          <span>Virtual credits</span>
          <span>Admin RTP</span>
          <span>Telegram-ready</span>
        </div>
      </div>
    </motion.section>
  )
}
