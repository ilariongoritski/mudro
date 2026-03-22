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
        <h2>Отдельный игровой слой внутри ленты с виртуальным балансом и быстрым входом</h2>
        <p>
          Премиальная витрина уже встроена в интерфейс. Дальше сюда подключается отдельный Go-сервис со своим
          PostgreSQL-контуром, admin-настройкой RTP и историей spins.
        </p>
      </div>

      <div className="casino-hero-strip__actions">
        <Link to="/casino" className="casino-hero-strip__primary">
          Открыть казино
        </Link>
        <div className="casino-hero-strip__signals" aria-label="Краткая сводка casino">
          <span>Virtual credits</span>
          <span>Admin RTP</span>
          <span>Живая история</span>
        </div>
      </div>
    </motion.section>
  )
}
