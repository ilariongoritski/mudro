import { useMemo, useState } from 'react'
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
import { ErrorBoundary } from '@/shared/ui/ErrorBoundary'

import './CasinoPage.css'

const reelFallback = ['🎰', '🍒', '🍋']
const betOptions = [10, 25, 50, 100]

const formatCasinoTimestamp = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return 'Только что'
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
  const [rtpDraft, setRtpDraft] = useState<string | null>(null)
  const [initialBalanceDraft, setInitialBalanceDraft] = useState<string | null>(null)

  const { isAuthenticated, user } = useAppSelector((state) => state.session)
  const shouldSkipCasinoQueries = !isAuthenticated
  const isAdmin = user?.role === 'admin'

  const {
    data: balanceData,
    isFetching: isBalanceFetching,
    isError: isBalanceError,
  } = useGetCasinoBalanceQuery(undefined, { skip: shouldSkipCasinoQueries })

  const { data: historyData, isFetching: isHistoryFetching } = useGetCasinoHistoryQuery(8, {
    skip: shouldSkipCasinoQueries,
  })

  const { data: configData, isFetching: isConfigFetching } = useGetCasinoConfigQuery(undefined, {
    skip: shouldSkipCasinoQueries || !isAdmin,
  })

  const [spinCasino, { isLoading: isSpinning }] = useSpinCasinoMutation()
  const [updateCasinoConfig, { isLoading: isSavingConfig }] = useUpdateCasinoConfigMutation()

  const balance = balanceData?.balance ?? 0
  const rtp = balanceData?.rtp ?? configData?.rtp_percent
  const history = historyData?.items ?? []
  const resolvedRtpDraft = rtpDraft ?? (configData ? String(configData.rtp_percent) : '')
  const resolvedInitialBalanceDraft = initialBalanceDraft ?? (configData ? String(configData.initial_balance) : '')
  const canSpin = isAuthenticated && !isSpinning && bet > 0 && balance >= bet

  const topSignal = useMemo(() => {
    if (!isAuthenticated) {
      return 'Войдите, чтобы открыть баланс, историю spins и полный игровой контур.'
    }

    if (isBalanceError) {
      return 'Casino backend временно недоступен. Повторите запрос позже или проверьте сервис.'
    }

    return 'Виртуальная экономика внутри MUDRO работает через отдельный casino-сервис и отдельную БД.'
  }, [isAuthenticated, isBalanceError])

  const backendStatus = useMemo(() => {
    if (!isAuthenticated) {
      return 'auth required'
    }
    if (isBalanceFetching || isHistoryFetching) {
      return 'syncing'
    }
    if (isBalanceError) {
      return 'degraded'
    }
    return 'online'
  }, [isAuthenticated, isBalanceError, isBalanceFetching, isHistoryFetching])

  const statusLine = useMemo(() => {
    if (spinFeedback) {
      return spinFeedback
    }
    if (!isAuthenticated) {
      return 'После входа страница покажет живой баланс, историю spins и управление ставкой.'
    }
    if (isBalanceFetching || isHistoryFetching) {
      return 'Подключаем игровой контур и синхронизируем историю.'
    }
    if (isBalanceError) {
      return 'Casino API не ответил. Проверьте backend или повторите spin позже.'
    }
    return 'Игровой контур подключён. Можно запускать spin и менять ставку.'
  }, [isAuthenticated, isBalanceError, isBalanceFetching, isHistoryFetching, spinFeedback])

  const lastResultLabel = useMemo(() => {
    if (isSpinning) {
      return 'Идёт spin...'
    }
    if (spinFeedback) {
      return 'См. статус'
    }
    return 'Готов к spin'
  }, [isSpinning, spinFeedback])

  const handleSpin = async () => {
    if (!isAuthenticated || isSpinning) {
      return
    }

    if (balance < bet) {
      setSpinFeedback('Недостаточно credits для выбранной ставки.')
      return
    }

    setSpinFeedback(null)
    setCelebrateSpin(false)
    setReels(['🎰', '🎰', '🎰'])

    try {
      const response = await spinCasino({ bet }).unwrap()
      const nextReels = response.symbols?.length === 3 ? response.symbols : reelFallback
      setReels(nextReels)
      setCelebrateSpin(response.win > 0)
      setSpinFeedback(
        response.win > 0
          ? `Выигрыш ${response.win} credits. Новый баланс ${response.balance}.`
          : `Spin завершён. Баланс ${response.balance}.`,
      )
    } catch {
      setReels(reelFallback)
      setSpinFeedback('Casino API временно недоступен. Проверьте backend и повторите spin.')
    }
  }

  const handleSaveConfig = async () => {
    if (!isAdmin || !configData) {
      return
    }

    const nextRtp = Number.parseFloat(normalizeDraftNumber(resolvedRtpDraft))
    const nextInitialBalance = Number.parseInt(resolvedInitialBalanceDraft, 10)

    if (!Number.isFinite(nextRtp) || nextRtp <= 0) {
      setSpinFeedback('RTP должен быть положительным числом.')
      return
    }

    if (!Number.isFinite(nextInitialBalance) || nextInitialBalance <= 0) {
      setSpinFeedback('Стартовый баланс должен быть положительным числом.')
      return
    }

    try {
      await updateCasinoConfig({
        ...configData,
        rtp_percent: nextRtp,
        initial_balance: nextInitialBalance,
      }).unwrap()
      setSpinFeedback(`Casino config сохранён: RTP ${nextRtp}%, стартовый баланс ${nextInitialBalance}.`)
    } catch {
      setSpinFeedback('Не удалось сохранить casino config. Проверьте admin-доступ и состояние backend-а.')
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
              <button type="button" className="casino-page__primary-action" onClick={handleSpin} disabled={!canSpin}>
                {isSpinning ? 'Идёт spin...' : 'Запустить spin'}
              </button>
            )}
            <Link to="/" className="casino-page__secondary-action">
              Вернуться в ленту
            </Link>
            <Link to="/tma/casino" className="casino-page__secondary-action">
              Mini App
            </Link>
          </div>

          <nav className="casino-page__mode-bar" aria-label="Режимы casino">
            <a href="#play">Играть</a>
            <a href="#history">История</a>
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
            <span>Баланс</span>
              <strong>{isBalanceFetching ? '...' : balance}</strong>
            </article>
            <article>
              <span>RTP</span>
              <strong>{typeof rtp === 'number' ? `${rtp}%` : 'Admin'}</strong>
            </article>
            <article>
            <span>Доступ</span>
              <strong>{user?.role?.toUpperCase() ?? (isAuthenticated ? 'USER' : 'GUEST')}</strong>
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

      <ErrorBoundary
        fallback={
          <section className="casino-page__surface casino-page__surface_fallback" role="alert">
            <span className="casino-page__kicker">Игровой runtime</span>
            <h2>Временный сбой игрового интерфейса</h2>
            <p>Обновите страницу или откройте Telegram Mini App. Основная лента продолжает работать.</p>
          </section>
        }
      >
        <>
          <section className="casino-page__workspace">
            <div className="casino-page__surface casino-page__surface_main">
              <div className="casino-page__surface-head">
                <div>
                  <span className="casino-page__kicker">Игровой контроль</span>
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
                  <strong>{lastResultLabel}</strong>
                </div>
                <div className="casino-page__summary-tile">
                  <span>Backend статус</span>
                  <strong>{backendStatus}</strong>
                </div>
              </div>

              <div className="casino-page__status-line" role="status" aria-live="polite">
                {statusLine}
              </div>
            </div>

            <aside className="casino-page__surface casino-page__surface_side" id="history">
              <div className="casino-page__side-block">
                <span className="casino-page__kicker">История</span>
                <h3>Последние spins</h3>
                <div className="casino-page__history-list">
                  {history.length === 0 && !isHistoryFetching ? (
                    <p className="casino-page__empty-copy">
                      {isAuthenticated
                        ? 'Первые spins появятся здесь автоматически после игры.'
                        : 'История станет доступна после входа в аккаунт.'}
                    </p>
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
                <span className="casino-page__kicker">Почему так</span>
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
                  <span className="casino-page__kicker">Панель администратора</span>
                  <h2>Тонкая настройка RTP и стартового баланса участников.</h2>
                </div>
                <p className="casino-page__admin-note">
                  Настройки сохраняются в отдельной casino БД и сразу влияют на новых участников и новые spins.
                </p>
              </div>

              <div className="casino-page__admin-grid">
                <label className="casino-page__field">
                  <span>RTP, %</span>
                  <input
                    type="number"
                    min="1"
                    step="0.01"
                    value={resolvedRtpDraft}
                    onChange={(event) => setRtpDraft(event.target.value)}
                    disabled={isConfigFetching || isSavingConfig}
                  />
                </label>
                <label className="casino-page__field">
                  <span>Стартовый баланс</span>
                  <input
                    type="number"
                    min="1"
                    step="1"
                    value={resolvedInitialBalanceDraft}
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
                  {isSavingConfig ? 'Сохраняем...' : 'Сохранить config'}
                </button>
                <span className="casino-page__admin-status">
                  {configData ? `Конфиг загружен · обновлён ${formatCasinoTimestamp(configData.updated_at)}` : 'Загружаем config...'}
                </span>
              </div>
            </motion.section>
          ) : null}
        </>
      </ErrorBoundary>
    </main>
  )
}
