import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'

import {
  useGetCasinoBalanceQuery,
  useGetCasinoConfigQuery,
  useGetCasinoHistoryQuery,
  useSpinCasinoMutation,
  useUpdateCasinoConfigMutation,
} from '@/features/casino/api/casinoApi'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'

import './CasinoPage.css'

const reelFallback = ['7', 'BAR', 'STAR']
const betOptions = [10, 25, 50, 100]

const formatCasinoTimestamp = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return 'Сейчас'
  }
  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

const normalizeDraftNumber = (value: string) => value.replace(',', '.').trim()

export const CasinoPage = () => {
  const [bet, setBet] = useState(25)
  const [reels, setReels] = useState<string[]>(reelFallback)
  const [spinFeedback, setSpinFeedback] = useState<string | null>(null)
  const [celebrateSpin, setCelebrateSpin] = useState(false)
  const [rtpDraft, setRtpDraft] = useState('95')
  const [initialBalanceDraft, setInitialBalanceDraft] = useState('2500')

  const { isAuthenticated, user } = useAppSelector((state) => state.session)
  const shouldSkipCasinoQueries = !isAuthenticated
  const isAdmin = user?.role === 'admin'

  const {
    data: balanceData,
    isFetching: isBalanceFetching,
    isError: isBalanceError,
  } = useGetCasinoBalanceQuery(undefined, { skip: shouldSkipCasinoQueries })

  const {
    data: historyData,
    isFetching: isHistoryFetching,
  } = useGetCasinoHistoryQuery(8, { skip: shouldSkipCasinoQueries })

  const {
    data: configData,
    isFetching: isConfigFetching,
  } = useGetCasinoConfigQuery(undefined, { skip: shouldSkipCasinoQueries || !isAdmin })

  const [spinCasino, { isLoading: isSpinning }] = useSpinCasinoMutation()
  const [updateCasinoConfig, { isLoading: isSavingConfig }] = useUpdateCasinoConfigMutation()

  const balance = balanceData?.balance ?? 0
  const rtp = balanceData?.rtp ?? configData?.rtp_percent
  const history = historyData?.items ?? []

  useEffect(() => {
    if (!configData) return
    setRtpDraft(String(configData.rtp_percent))
    setInitialBalanceDraft(String(configData.initial_balance))
  }, [configData])

  const topSignal = useMemo(() => {
    if (!isAuthenticated) {
      return 'Авторизуйся, чтобы открыть баланс, spins и history.'
    }
    if (isBalanceError) {
      return 'Backend casino еще не отвечает. UI уже готов к подключению.'
    }
    return 'Виртуальная экономика внутри MUDRO с отдельным сервисным контуром.'
  }, [isAuthenticated, isBalanceError])

  const handleSpin = async () => {
    if (!isAuthenticated || isSpinning) {
      return
    }

    setSpinFeedback(null)
    setCelebrateSpin(false)
    setReels(['SPIN', 'SPIN', 'SPIN'])

    try {
      const response = await spinCasino({ bet }).unwrap()
      const nextReels = response.symbols?.length === 3 ? response.symbols : reelFallback
      setReels(nextReels)
      setCelebrateSpin(response.win > 0)
      setSpinFeedback(
        response.win > 0
          ? `Выигрыш ${response.win} credits. Новый баланс ${response.balance}.`
          : `Спин завершен. Баланс ${response.balance}.`
      )
    } catch {
      setReels(reelFallback)
      setSpinFeedback('Casino API недоступен. Как только backend подключится, эта витрина начнет работать без переделки.')
    }
  }

  const handleSaveConfig = async () => {
    if (!isAdmin || !configData) {
      return
    }

    const nextRtp = Number.parseFloat(normalizeDraftNumber(rtpDraft))
    const nextInitialBalance = Number.parseInt(initialBalanceDraft, 10)

    if (!Number.isFinite(nextRtp) || nextRtp <= 0) {
      setSpinFeedback('RTP должен быть положительным числом.')
      return
    }

    if (!Number.isFinite(nextInitialBalance) || nextInitialBalance <= 0) {
      setSpinFeedback('Seed balance должен быть положительным числом.')
      return
    }

    try {
      await updateCasinoConfig({
        ...configData,
        rtp_percent: nextRtp,
        initial_balance: nextInitialBalance,
      }).unwrap()
      setSpinFeedback(`Casino config сохранён: RTP ${nextRtp}%, seed ${nextInitialBalance}.`)
    } catch {
      setSpinFeedback('Не удалось сохранить casino config. Проверь backend/admin proxy.')
    }
  }

  return (
    <main className="casino-page">
      <section className="casino-page__hero" id="play">
        <div className="casino-page__hero-copy">
          <span className="casino-page__eyebrow">MUDRO Casino</span>
          <h1>Отдельный игровой слой внутри ленты, а не чужой модуль поверх неё.</h1>
          <p>{topSignal}</p>

          <div className="casino-page__hero-actions">
            {!isAuthenticated ? (
              <Link to="/login" className="casino-page__primary-action">
                Войти и открыть баланс
              </Link>
            ) : (
              <button type="button" className="casino-page__primary-action" onClick={handleSpin} disabled={isSpinning}>
                {isSpinning ? 'Идет spin...' : 'Запустить spin'}
              </button>
            )}
            <Link to="/" className="casino-page__secondary-action">
              Вернуться в ленту
            </Link>
          </div>

          <nav className="casino-page__mode-bar" aria-label="Режимы casino">
            <a href="#play">Play</a>
            <a href="#history">History</a>
            {isAdmin ? <a href="#admin">Admin</a> : null}
          </nav>
        </div>

        <motion.div
          className={`casino-page__hero-stage ${celebrateSpin ? 'casino-page__hero-stage_win' : ''}`}
          initial={{ opacity: 0, scale: 0.96 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.4, ease: 'easeOut' }}
        >
          <div className="casino-page__stage-ring" />
          <div className="casino-page__stage-metrics">
            <article>
              <span>Balance</span>
              <strong>{isBalanceFetching ? '...' : balance}</strong>
            </article>
            <article>
              <span>RTP</span>
              <strong>{typeof rtp === 'number' ? `${rtp}%` : 'Admin'}</strong>
            </article>
            <article>
              <span>Access</span>
              <strong>{user?.role?.toUpperCase() ?? 'GUEST'}</strong>
            </article>
          </div>
          <div className="casino-page__stage-reels" aria-label="Slot machine reels">
            {reels.map((symbol, index) => (
              <div
                key={`${symbol}-${index}`}
                className={`casino-page__reel ${isSpinning ? 'casino-page__reel_spinning' : ''}`}
              >
                <span>{symbol}</span>
              </div>
            ))}
          </div>
        </motion.div>
      </section>

      <section className="casino-page__workspace">
        <div className="casino-page__surface casino-page__surface_main">
          <div className="casino-page__surface-head">
            <div>
              <span className="casino-page__kicker">Game control</span>
              <h2>Один слот, быстрый цикл, управляемый RTP и чистая интеграция в основной UI.</h2>
            </div>
            <div className="casino-page__bet-row" aria-label="Bet controls">
              {betOptions.map((option) => (
                <button
                  key={option}
                  type="button"
                  onClick={() => setBet(option)}
                  className={`casino-page__bet-chip ${bet === option ? 'casino-page__bet-chip_active' : ''}`}
                >
                  {option}
                </button>
              ))}
            </div>
          </div>

          <div className="casino-page__action-row">
            <div className="casino-page__summary-tile">
              <span>Текущая ставка</span>
              <strong>{bet} credits</strong>
            </div>
            <div className="casino-page__summary-tile">
              <span>Последний результат</span>
              <strong>{spinFeedback ? 'См. статус' : 'Ожидает spin'}</strong>
            </div>
            <div className="casino-page__summary-tile">
              <span>Backend статус</span>
              <strong>{isBalanceError ? 'stub mode' : 'linked'}</strong>
            </div>
          </div>

          <div className="casino-page__status-line" role="status">
            {spinFeedback ?? 'Пока backend не подключен, UI работает как интеграционная витрина и готов к реальным ответам API.'}
          </div>
        </div>

        <aside className="casino-page__surface casino-page__surface_side" id="history">
          <div className="casino-page__side-block">
            <span className="casino-page__kicker">History</span>
            <h3>Последние spins</h3>
            <div className="casino-page__history-list">
              {history.length === 0 && !isHistoryFetching ? (
                <p className="casino-page__empty-copy">История появится после первых spins и подключения casino-service.</p>
              ) : null}

              {history.map((item) => (
                <article key={item.id} className="casino-page__history-item">
                  <div>
                    <strong>{item.symbols.join(' · ')}</strong>
                    <span>{formatCasinoTimestamp(item.created_at)}</span>
                  </div>
                  <div className="casino-page__history-metrics">
                    <span>bet {item.bet}</span>
                    <strong>{item.win > 0 ? `+${item.win}` : item.win}</strong>
                  </div>
                </article>
              ))}
            </div>
          </div>

          <div className="casino-page__side-block">
            <span className="casino-page__kicker">Why this shape</span>
            <ul className="casino-page__signal-list">
              <li>Отдельный вход в продукт, без отдельного бренда.</li>
              <li>Glass + neon слой только там, где он усиливает сценарий.</li>
              <li>Прямой путь к admin RTP и отдельной business-логике.</li>
            </ul>
          </div>
        </aside>
      </section>

      {isAdmin ? (
        <motion.section
          className="casino-page__surface casino-page__surface_admin"
          id="admin"
          initial={{ opacity: 0, y: 18 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.35, ease: 'easeOut' }}
        >
          <div className="casino-page__surface-head">
            <div>
              <span className="casino-page__kicker">Admin panel</span>
              <h2>Тонкая настройка RTP и стартового баланса участников.</h2>
            </div>
            <p className="casino-page__admin-note">
              Настройки сохраняются в отдельной casino БД и сразу влияют на новые участники и новые spins.
            </p>
          </div>

          <div className="casino-page__admin-grid">
            <label className="casino-page__field">
              <span>RTP, %</span>
              <input
                type="number"
                min="1"
                step="0.01"
                value={rtpDraft}
                onChange={(event) => setRtpDraft(event.target.value)}
                disabled={isConfigFetching || isSavingConfig}
              />
            </label>
            <label className="casino-page__field">
              <span>Seed balance</span>
              <input
                type="number"
                min="1"
                step="1"
                value={initialBalanceDraft}
                onChange={(event) => setInitialBalanceDraft(event.target.value)}
                disabled={isConfigFetching || isSavingConfig}
              />
            </label>
          </div>

          <div className="casino-page__admin-actions">
            <button
              type="button"
              className="casino-page__primary-action"
              onClick={handleSaveConfig}
              disabled={isSavingConfig || isConfigFetching || !configData}
            >
              {isSavingConfig ? 'Сохранение...' : 'Сохранить config'}
            </button>
            <span className="casino-page__admin-status">
              {configData ? `Loaded config · updated ${formatCasinoTimestamp(configData.updated_at)}` : 'Config loading...'}
            </span>
          </div>
        </motion.section>
      ) : null}
    </main>
  )
}
