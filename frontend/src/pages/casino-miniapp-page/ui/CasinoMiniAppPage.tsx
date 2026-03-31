import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useDispatch } from 'react-redux'

import { useTelegramWebApp } from '@/features/telegram-miniapp/hooks/useTelegramWebApp'
import { useTelegramBootstrapMutation } from '@/entities/session/api/authApi'
import { setCredentials } from '@/entities/session/model/sessionSlice'
import {
  useGetCasinoBalanceQuery,
  useGetCasinoHistoryQuery,
  useSpinCasinoMutation,
} from '@/features/casino/api/casinoApi'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'
import { ErrorBoundary } from '@/shared/ui/ErrorBoundary'

import './CasinoMiniAppPage.css'

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

export const CasinoMiniAppPage = () => {
  const dispatch = useDispatch()
  const navigate = useNavigate()

  const [bet, setBet] = useState(25)
  const [reels, setReels] = useState<string[]>(reelFallback)
  const [status, setStatus] = useState<string | null>(null)
  const [winPulse, setWinPulse] = useState(false)
  const bootstrapStarted = useRef(false)

  const { initData, isTelegram, webApp } = useTelegramWebApp()
  const { isAuthenticated, user } = useAppSelector((state) => state.session)

  const [telegramBootstrap, { isLoading: isBootstrapping }] = useTelegramBootstrapMutation()
  const [spinCasino, { isLoading: isSpinning }] = useSpinCasinoMutation()

  const { data: balanceData, isFetching: isBalanceFetching, isError: isBalanceError } = useGetCasinoBalanceQuery(undefined, {
    skip: !isAuthenticated,
  })

  const { data: historyData, isFetching: isHistoryFetching } = useGetCasinoHistoryQuery(6, {
    skip: !isAuthenticated,
  })

  const balance = balanceData?.balance ?? 0
  const rtp = balanceData?.rtp
  const history = historyData?.items ?? []
  const canSpin = isAuthenticated && !isSpinning && bet > 0 && balance >= bet

  useEffect(() => {
    if (!isTelegram || isAuthenticated || bootstrapStarted.current || !initData) {
      return
    }

    bootstrapStarted.current = true
    telegramBootstrap({ initData })
      .unwrap()
      .then((result) => {
        dispatch(setCredentials(result))
        setStatus('Telegram-профиль подключён.')
      })
      .catch(() => {
        setStatus('Не удалось проверить Telegram-сессию.')
      })
  }, [dispatch, initData, isAuthenticated, isTelegram, telegramBootstrap])

  const onSpin = useCallback(async () => {
    if (!isAuthenticated || isSpinning) {
      return
    }

    if (balance < bet) {
      setStatus('Недостаточно credits для выбранной ставки.')
      return
    }

    setStatus(null)
    setWinPulse(false)
    setReels(['🎰', '🎰', '🎰'])
    webApp?.HapticFeedback?.impactOccurred('medium')

    try {
      const response = await spinCasino({ bet }).unwrap()
      setReels(response.symbols?.length === 3 ? response.symbols : reelFallback)

      if (response.win > 0) {
        setWinPulse(true)
        setStatus(`Выигрыш +${response.win}. Баланс ${response.balance}.`)
        webApp?.HapticFeedback?.notificationOccurred('success')
      } else {
        setStatus(`Spin завершён. Баланс ${response.balance}.`)
      }
    } catch {
      setReels(reelFallback)
      setStatus('Casino API сейчас недоступен.')
      webApp?.HapticFeedback?.notificationOccurred('warning')
    }
  }, [balance, bet, isAuthenticated, isSpinning, spinCasino, webApp])

  useEffect(() => {
    if (!webApp) {
      return
    }

    const onBack = () => navigate('/')
    webApp.BackButton.show()
    webApp.BackButton.onClick(onBack)

    return () => {
      webApp.BackButton.offClick(onBack)
      webApp.BackButton.hide()
    }
  }, [navigate, webApp])

  useEffect(() => {
    if (!webApp || !isAuthenticated) {
      webApp?.MainButton.hide()
      return
    }

    const onMainButton = () => void onSpin()
    const buttonText = canSpin ? `Spin ${bet}` : `Нужно ${bet} credits`

    webApp.MainButton.setText(isSpinning ? 'Spinning...' : buttonText)
    webApp.MainButton.show()
    webApp.MainButton.onClick(onMainButton)

    return () => {
      webApp.MainButton.offClick(onMainButton)
      webApp.MainButton.hide()
    }
  }, [bet, canSpin, isAuthenticated, isSpinning, onSpin, webApp])

  const topMessage = useMemo(() => {
    if (isBootstrapping) {
      return 'Подключаем Telegram-профиль...'
    }
    if (!isAuthenticated && isTelegram) {
      return 'Проверяем mini app сессию.'
    }
    if (!isAuthenticated) {
      return 'Откройте страницу в Telegram Mini App или войдите вручную.'
    }
    if (isBalanceError) {
      return 'Доступ есть, но casino backend временно недоступен.'
    }
    return 'Игровой контур подключён к отдельному casino сервису.'
  }, [isAuthenticated, isBalanceError, isBootstrapping, isTelegram])

  const statusText = status ?? (user ? `Игрок: ${user.username}` : 'Ожидание авторизации')

  return (
    <main className={`casino-miniapp ${winPulse ? 'casino-miniapp_win' : ''}`}>
      <header className="casino-miniapp__top">
        <div>
          <span className="casino-miniapp__eyebrow">MUDRO Telegram Mini App</span>
          <h1>Casino</h1>
          <p>{topMessage}</p>
        </div>
        <div className="casino-miniapp__balance">
          <span>Balance</span>
          <strong>{isBalanceFetching ? '...' : balance}</strong>
          <small>{typeof rtp === 'number' ? `RTP ${rtp}%` : 'RTP: admin config'}</small>
        </div>
      </header>

      <ErrorBoundary
        fallback={
          <section className="casino-miniapp__stage casino-miniapp__stage_fallback" role="alert">
            <h2>Сбой mini app слоя</h2>
            <p>Обновите страницу. Если ошибка повторяется, откройте полный режим.</p>
            <Link className="casino-miniapp__secondary" to="/casino">
              Открыть полный режим
            </Link>
          </section>
        }
      >
        <>
          <section className="casino-miniapp__stage">
            <div className="casino-miniapp__reels">
              {reels.map((symbol, index) => (
                <article
                  key={`${symbol}-${index}`}
                  className={`casino-miniapp__reel ${isSpinning ? 'casino-miniapp__reel_spinning' : ''}`}
                >
                  <span>{symbol}</span>
                </article>
              ))}
            </div>

            <div className="casino-miniapp__bets">
              {betOptions.map((option) => (
                <button
                  key={option}
                  type="button"
                  onClick={() => setBet(option)}
                  className={bet === option ? 'casino-miniapp__chip casino-miniapp__chip_active' : 'casino-miniapp__chip'}
                >
                  {option}
                </button>
              ))}
            </div>

            <div className="casino-miniapp__actions">
              <button type="button" onClick={onSpin} disabled={!canSpin} className="casino-miniapp__primary">
                {isSpinning ? 'Идёт spin...' : `Spin ${bet}`}
              </button>
              {!isAuthenticated ? (
                <Link className="casino-miniapp__secondary" to="/login">
                  Войти вручную
                </Link>
              ) : (
                <Link className="casino-miniapp__secondary" to="/casino">
                  Полный режим
                </Link>
              )}
            </div>

            <p className="casino-miniapp__status" role="status" aria-live="polite">
              {statusText}
            </p>
          </section>

          <section className="casino-miniapp__history">
            <div className="casino-miniapp__history-head">
              <h2>Последние spins</h2>
              <small>{isHistoryFetching ? 'Обновляем...' : `${history.length} записей`}</small>
            </div>
            <div className="casino-miniapp__history-list">
              {history.length === 0 && !isHistoryFetching ? (
                <p className="casino-miniapp__empty">
                  {isAuthenticated ? 'История появится после первых spins.' : 'История станет доступна после входа.'}
                </p>
              ) : null}

              {history.map((item) => (
                <article key={item.id} className="casino-miniapp__history-item">
                  <div>
                    <strong>{item.symbols.join(' · ')}</strong>
                    <span>{formatCasinoTimestamp(item.created_at)}</span>
                  </div>
                  <div>
                    <span>bet {item.bet}</span>
                    <strong>{item.win > 0 ? `+${item.win}` : item.win}</strong>
                  </div>
                </article>
              ))}
            </div>
          </section>
        </>
      </ErrorBoundary>
    </main>
  )
}
