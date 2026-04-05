import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useDispatch } from 'react-redux'

import { useTelegramBootstrapMutation } from '@/entities/session/api/authApi'
import { setCredentials } from '@/entities/session/model/sessionSlice'
import { useTelegramWebApp } from '@/features/telegram-miniapp/hooks/useTelegramWebApp'
import {
  useGetCasinoBalanceQuery,
  useGetCasinoActivityQuery,
  useGetCasinoHistoryQuery,
  useGetCasinoProfileQuery,
  useSpinCasinoMutation,
} from '@/features/casino/api/casinoApi'
import {
  buildCasinoActivityFromApi,
  buildCasinoActivityFromHistory,
  buildCasinoProfileSummary,
} from '@/features/casino/model/profile'
import { CasinoProfileModal } from '@/features/casino/ui/CasinoProfileModal'
import { PlinkoPanel } from '@/features/casino/ui/PlinkoPanel'
import { RoulettePanel } from '@/features/casino/ui/RoulettePanel'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'
import { ErrorBoundary } from '@/shared/ui/ErrorBoundary'

import './CasinoMiniAppPage.css'

const reelFallback = ['🎰', '🍒', '🍋']
const betOptions = [10, 25, 50, 100]
const slotSpinSymbols = ['🎰', '🍒', '🍋', '🍊', '🍇', '🔔', '⭐', '💎', '7️⃣', '🃏']

type GameTab = 'slots' | 'roulette' | 'plinko' | 'crash' | 'coinflip'
type CasinoTheme = 'midnight' | 'aurora' | 'ember'

const themeOptions: Array<{ value: CasinoTheme; label: string }> = [
  { value: 'midnight', label: 'Midnight' },
  { value: 'aurora', label: 'Aurora' },
  { value: 'ember', label: 'Ember' },
]

interface GameMainAction {
  label: string
  busy: boolean
  disabled: boolean
  onTrigger: () => void
}

interface FairnessSummary {
  total: number
  verified: number
  pending: number
  latestTitle: string
  latestAt: string | null
}

type ReelVisualState = 'idle' | 'spinning' | 'stopping' | 'winning'

interface SlotRoundSummary {
  outcome: 'WIN' | 'MISS'
  bet: number
  payout: number
  balance: number
  symbols: string[]
}

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

  const [activeTab, setActiveTab] = useState<GameTab>('slots')
  const [bet, setBet] = useState(25)
  const [reels, setReels] = useState<string[]>(reelFallback)
  const [reelStates, setReelStates] = useState<ReelVisualState[]>(['idle', 'idle', 'idle'])
  const [slotStatus, setSlotStatus] = useState<string | null>(null)
  const [serviceStatus, setServiceStatus] = useState<string | null>(null)
  const [slotRoundSummary, setSlotRoundSummary] = useState<SlotRoundSummary | null>(null)
  const [isResolvingSpin, setIsResolvingSpin] = useState(false)
  const [winPulse, setWinPulse] = useState(false)
  const [rouletteMainAction, setRouletteMainAction] = useState<GameMainAction | null>(null)
  const [plinkoMainAction, setPlinkoMainAction] = useState<GameMainAction | null>(null)
  const [profileOpen, setProfileOpen] = useState(false)
  const [soundEnabled, setSoundEnabled] = useState(() => {
    if (typeof window === 'undefined') {
      return true
    }

    const stored = window.localStorage.getItem('mudro.casino.sound')
    return stored === null ? true : stored === '1'
  })
  const [theme, setTheme] = useState<CasinoTheme>(() => {
    if (typeof window === 'undefined') {
      return 'midnight'
    }

    const stored = window.localStorage.getItem('mudro.casino.theme')
    return stored === 'aurora' || stored === 'ember' || stored === 'midnight' ? stored : 'midnight'
  })
  const [fairnessOpen, setFairnessOpen] = useState(false)
  const bootstrapStarted = useRef(false)
  const spinRequestId = useRef(0)

  const { initData, isTelegram, webApp } = useTelegramWebApp()
  const { isAuthenticated, user } = useAppSelector((state) => state.session)

  const [telegramBootstrap, { isLoading: isBootstrapping }] = useTelegramBootstrapMutation()
  const [spinCasino, { isLoading: isSpinning }] = useSpinCasinoMutation()

  const {
    data: balanceData,
    isFetching: isBalanceFetching,
    isError: isBalanceError,
  } = useGetCasinoBalanceQuery(undefined, {
    skip: !isAuthenticated,
  })

  const { data: historyData, isFetching: isHistoryFetching } = useGetCasinoHistoryQuery(6, {
    skip: !isAuthenticated,
  })

  const { data: profileData } = useGetCasinoProfileQuery(undefined, {
    skip: !isAuthenticated,
  })
  const { data: activityData } = useGetCasinoActivityQuery(12, {
    skip: !isAuthenticated,
  })

  const balance = balanceData?.balance ?? 0
  const rtp = balanceData?.rtp
  const history = useMemo(() => historyData?.items ?? [], [historyData?.items])
  const profileActivity = useMemo(() => {
    const remoteActivity = buildCasinoActivityFromApi(activityData?.items ?? profileData?.recent_activity ?? [])
    return remoteActivity.length > 0 ? remoteActivity : buildCasinoActivityFromHistory(history)
  }, [activityData?.items, history, profileData?.recent_activity])
  const profileSummary = useMemo(
    () =>
      buildCasinoProfileSummary({
        profile: profileData ?? null,
        user,
        balance,
        currency: balanceData?.currency,
        rtp,
        history,
        activity: profileActivity,
      }),
    [balance, balanceData?.currency, history, profileActivity, profileData, rtp, user],
  )
  const fairnessSummary = useMemo<FairnessSummary>(() => {
    const verified = profileActivity.filter((item) => item.fairnessStatus === 'verified').length
    const total = profileActivity.length
    const latestEntry = profileActivity[0]

    return {
      total,
      verified,
      pending: Math.max(total - verified, 0),
      latestTitle: latestEntry?.title ?? 'Сессия пока без истории',
      latestAt: latestEntry?.createdAt ?? null,
    }
  }, [profileActivity])

  const displayName = profileData?.display_name ?? profileSummary.displayName
  const resolvedLevel = profileData?.level ?? profileSummary.level
  const resolvedUsername = profileData?.username ?? profileSummary.username
  const resolvedUserId = profileData?.user_id ?? profileSummary.userId
  const profileStatus = !isAuthenticated
    ? 'Войдите, чтобы открыть casino profile.'
    : profileData?.last_game_at
      ? `Последняя игра ${formatCasinoTimestamp(profileData.last_game_at)}`
      : profileActivity.length > 0
        ? 'Profile собирается из live activity.'
        : 'Profile ready.'

  const canSpin = isAuthenticated && !isSpinning && !isResolvingSpin && bet > 0 && balance >= bet
  const spinDelay = useCallback((ms: number) => new Promise<void>((resolve) => window.setTimeout(resolve, ms)), [])
  const randomSpinSymbol = useCallback(
    () => slotSpinSymbols[Math.floor(Math.random() * slotSpinSymbols.length)] ?? '🎰',
    [],
  )

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }

    window.localStorage.setItem('mudro.casino.sound', soundEnabled ? '1' : '0')
  }, [soundEnabled])

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }

    window.localStorage.setItem('mudro.casino.theme', theme)
  }, [theme])

  useEffect(() => {
    if (!isTelegram || isAuthenticated || bootstrapStarted.current || !initData) {
      return
    }

    bootstrapStarted.current = true
    telegramBootstrap({ initData })
      .unwrap()
      .then((result) => {
        dispatch(setCredentials(result))
        setServiceStatus('Telegram-профиль подключён.')
      })
      .catch(() => {
        setServiceStatus('Не удалось проверить Telegram-сессию.')
      })
  }, [dispatch, initData, isAuthenticated, isTelegram, telegramBootstrap])

  useEffect(() => {
    if (!winPulse) {
      return
    }

    const timer = window.setTimeout(() => setWinPulse(false), 1400)
    return () => window.clearTimeout(timer)
  }, [winPulse])

  useEffect(() => {
    if (!serviceStatus) {
      return
    }

    const timer = window.setTimeout(() => setServiceStatus(null), 3600)
    return () => window.clearTimeout(timer)
  }, [serviceStatus])

  const onSpin = useCallback(async () => {
    if (!isAuthenticated || isSpinning) {
      return
    }

    if (balance < bet) {
      setSlotStatus('Недостаточно credits для выбранной ставки.')
      return
    }

    const requestId = ++spinRequestId.current
    setSlotStatus('Крутим барабаны...')
    setSlotRoundSummary(null)
    setWinPulse(false)
    setIsResolvingSpin(true)
    setReels(['🎰', '🎰', '🎰'])
    setReelStates(['spinning', 'spinning', 'spinning'])
    webApp?.HapticFeedback?.impactOccurred('medium')

    try {
      const response = await spinCasino({ bet }).unwrap()
      const finalReels = response.symbols?.length === 3 ? response.symbols : reelFallback
      const stagedReels = ['🎰', '🎰', '🎰']

      for (let reelIndex = 0; reelIndex < stagedReels.length; reelIndex += 1) {
        for (let tick = 0; tick < 4; tick += 1) {
          if (requestId !== spinRequestId.current) {
            return
          }
          stagedReels[reelIndex] = randomSpinSymbol()
          setReels([...stagedReels])
          setReelStates((current) =>
            current.map((state, index) => {
              if (index < reelIndex) {
                return state === 'winning' ? state : 'idle'
              }
              return 'spinning'
            }),
          )
          await spinDelay(75)
        }

        stagedReels[reelIndex] = finalReels[reelIndex] ?? reelFallback[reelIndex]
        setReels([...stagedReels])
        setReelStates((current) =>
          current.map((state, index) => {
            if (index < reelIndex) {
              return state === 'winning' ? state : 'idle'
            }
            if (index === reelIndex) {
              return 'stopping'
            }
            return 'spinning'
          }),
        )
        await spinDelay(90 + reelIndex * 40)
      }

      if (response.win > 0) {
        setWinPulse(true)
        setSlotStatus(`Выигрыш +${response.win}. Баланс ${response.balance}.`)
        setSlotRoundSummary({
          outcome: 'WIN',
          bet,
          payout: response.win,
          balance: response.balance,
          symbols: finalReels,
        })
        setReelStates(['winning', 'winning', 'winning'])
        setIsResolvingSpin(false)
        webApp?.HapticFeedback?.notificationOccurred('success')
      } else {
        setSlotStatus(`Spin завершён. Баланс ${response.balance}.`)
        setSlotRoundSummary({
          outcome: 'MISS',
          bet,
          payout: 0,
          balance: response.balance,
          symbols: finalReels,
        })
        setReelStates(['idle', 'idle', 'idle'])
        setIsResolvingSpin(false)
      }
    } catch {
      setReels(reelFallback)
      setReelStates(['idle', 'idle', 'idle'])
      setIsResolvingSpin(false)
      setSlotStatus('Casino API сейчас недоступен.')
      setSlotRoundSummary(null)
      webApp?.HapticFeedback?.notificationOccurred('warning')
    }
  }, [balance, bet, isAuthenticated, isSpinning, randomSpinSymbol, spinCasino, spinDelay, webApp])

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

    if (activeTab === 'roulette') {
      const currentAction = rouletteMainAction
      const onMainButton = () => {
        if (!currentAction?.disabled) {
          currentAction?.onTrigger()
        }
      }

      webApp.MainButton.setText(currentAction?.busy ? 'Отправляем...' : currentAction?.label ?? 'Соберите ставку')
      webApp.MainButton.show()
      webApp.MainButton.onClick(onMainButton)

      return () => {
        webApp.MainButton.offClick(onMainButton)
        webApp.MainButton.hide()
      }
    }

    if (activeTab === 'plinko') {
      const currentAction = plinkoMainAction
      const onMainButton = () => {
        if (!currentAction?.disabled) {
          currentAction?.onTrigger()
        }
      }

      webApp.MainButton.setText(currentAction?.busy ? 'Dropping...' : currentAction?.label ?? 'Drop ball')
      webApp.MainButton.show()
      webApp.MainButton.onClick(onMainButton)

      return () => {
        webApp.MainButton.offClick(onMainButton)
        webApp.MainButton.hide()
      }
    }

    if (activeTab === 'crash' || activeTab === 'coinflip') {
      const onMainButton = () => setServiceStatus('Эта игра готовится. Каркас shell уже подключён.')
      webApp.MainButton.setText('Coming soon')
      webApp.MainButton.show()
      webApp.MainButton.onClick(onMainButton)

      return () => {
        webApp.MainButton.offClick(onMainButton)
        webApp.MainButton.hide()
      }
    }

    const onMainButton = () => void onSpin()
    const buttonText = canSpin ? `Spin ${bet}` : `Нужно ${bet} credits`

    webApp.MainButton.setText(isSpinning || isResolvingSpin ? 'Spinning...' : buttonText)
    webApp.MainButton.show()
    webApp.MainButton.onClick(onMainButton)

    return () => {
      webApp.MainButton.offClick(onMainButton)
      webApp.MainButton.hide()
    }
  }, [activeTab, bet, canSpin, isAuthenticated, isResolvingSpin, isSpinning, onSpin, plinkoMainAction, rouletteMainAction, webApp])

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
    if (serviceStatus) {
      return serviceStatus
    }
    if (activeTab === 'roulette') {
      return 'Live roulette идёт через `/casino/roulette/*` proxy и SSE state.'
    }
    if (activeTab === 'plinko') {
      return 'Plinko живёт отдельным self-contained экраном с board, risk и history.'
    }
    if (activeTab === 'crash' || activeTab === 'coinflip') {
      return 'Shell уже готов под будущую игру.'
    }
    return 'Игровой контур подключён к отдельному casino сервису.'
  }, [activeTab, isAuthenticated, isBalanceError, isBootstrapping, isTelegram, serviceStatus])

  const statusText = slotStatus ?? (user ? `Игрок: ${user.username}` : 'Ожидание авторизации')
  const tabsNote = useMemo(() => {
    switch (activeTab) {
      case 'slots':
        return 'Быстрый слот-цикл для коротких сессий.'
      case 'roulette':
        return 'Живой раунд, таймер и купон ставок.'
      case 'plinko':
        return 'Отдельная доска с риском, сеткой и payout.'
      default:
        return 'Слоты под будущие игры уже готовы.'
    }
  }, [activeTab])

  return (
    <main className={`casino-miniapp casino-miniapp_theme_${theme} ${winPulse ? 'casino-miniapp_win' : ''}`}>
      <header className="casino-miniapp__top">
        <div className="casino-miniapp__brand">
          <span className="casino-miniapp__eyebrow">MUDRO Telegram Mini App</span>
          <h1>Casino shell</h1>
          <p>{topMessage}</p>
        </div>

        <div className="casino-miniapp__balance">
          <span>Balance</span>
          <strong>{isBalanceFetching ? '...' : balance}</strong>
          <small>
            {displayName} · @{resolvedUsername} · LVL {resolvedLevel}
          </small>
          <small>User #{resolvedUserId || '—'}</small>
          <small>{profileStatus}</small>
          {isAuthenticated ? (
            <button type="button" className="casino-miniapp__profile-button" onClick={() => setProfileOpen(true)}>
              Casino profile
            </button>
          ) : (
            <Link to="/login" className="casino-miniapp__profile-button">
              Войти
            </Link>
          )}
        </div>
      </header>

      <nav className="casino-miniapp__tabs" aria-label="Casino games">
        <button
          type="button"
          className={activeTab === 'slots' ? 'casino-miniapp__tab casino-miniapp__tab_active' : 'casino-miniapp__tab'}
          onClick={() => setActiveTab('slots')}
        >
          Slots
        </button>
        <button
          type="button"
          className={activeTab === 'roulette' ? 'casino-miniapp__tab casino-miniapp__tab_active' : 'casino-miniapp__tab'}
          onClick={() => setActiveTab('roulette')}
        >
          Roulette
        </button>
        <button
          type="button"
          className={activeTab === 'plinko' ? 'casino-miniapp__tab casino-miniapp__tab_active' : 'casino-miniapp__tab'}
          onClick={() => setActiveTab('plinko')}
        >
          Plinko
        </button>
        <button
          type="button"
          className={activeTab === 'crash' ? 'casino-miniapp__tab casino-miniapp__tab_coming' : 'casino-miniapp__tab'}
          onClick={() => {
            setActiveTab('crash')
            setServiceStatus('Crash is coming soon. Shell is ready.')
          }}
        >
          Crash <span>soon</span>
        </button>
        <button
          type="button"
          className={activeTab === 'coinflip' ? 'casino-miniapp__tab casino-miniapp__tab_coming' : 'casino-miniapp__tab'}
          onClick={() => {
            setActiveTab('coinflip')
            setServiceStatus('Coin Flip is coming soon. Shell is ready.')
          }}
        >
          Coin Flip <span>soon</span>
        </button>
        <span className="casino-miniapp__tabs-note">{tabsNote}</span>
      </nav>

      <div className="casino-miniapp__utility-bar">
        <button
          type="button"
          className={soundEnabled ? 'casino-miniapp__utility-chip casino-miniapp__utility-chip_active' : 'casino-miniapp__utility-chip'}
          onClick={() => setSoundEnabled((value) => !value)}
        >
          Sound {soundEnabled ? 'on' : 'off'}
        </button>
        <button
          type="button"
          className={fairnessOpen ? 'casino-miniapp__utility-chip casino-miniapp__utility-chip_active' : 'casino-miniapp__utility-chip'}
          onClick={() => setFairnessOpen((value) => !value)}
        >
          Fairness check
        </button>
        {themeOptions.map((option) => (
          <button
            key={option.value}
            type="button"
            className={theme === option.value ? 'casino-miniapp__utility-chip casino-miniapp__utility-chip_active' : 'casino-miniapp__utility-chip'}
            onClick={() => setTheme(option.value)}
          >
            {option.label}
          </button>
        ))}
        <span className="casino-miniapp__utility-note">
          {soundEnabled ? 'Предпочтение звука сохранено локально.' : 'Звук отключён только для этого устройства.'}
        </span>
      </div>

      {fairnessOpen ? (
        <section className="casino-miniapp__fairness">
          <div className="casino-miniapp__fairness-head">
            <span>Fairness</span>
            <strong>
              {fairnessSummary.verified}/{fairnessSummary.total} verified
            </strong>
            <small>
              {fairnessSummary.latestAt ? `Последняя активность ${formatCasinoTimestamp(fairnessSummary.latestAt)}` : 'История ещё не накопилась'}
            </small>
          </div>
          <div className="casino-miniapp__fairness-grid">
            <article className="casino-miniapp__fairness-card">
              <span>Проверено</span>
              <strong>{fairnessSummary.verified}</strong>
              <small>записей с fairness-proof</small>
            </article>
            <article className="casino-miniapp__fairness-card">
              <span>В очереди</span>
              <strong>{fairnessSummary.pending}</strong>
              <small>появятся после новых раундов</small>
            </article>
            <article className="casino-miniapp__fairness-card">
              <span>Последнее событие</span>
              <strong>{fairnessSummary.latestTitle}</strong>
              <small>{profileSummary.gamesPlayed} игр в профиле</small>
            </article>
          </div>
          <p>
            Mini app уже тянет live activity и профиль из casino API, поэтому fairness и прогресс видны здесь без перехода в полный экран.
          </p>
          {isAuthenticated ? (
            <button type="button" className="casino-miniapp__fairness-action" onClick={() => setProfileOpen(true)}>
              Открыть casino profile
            </button>
          ) : null}
        </section>
      ) : null}

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
        {activeTab === 'slots' ? (
          <>
            <section className="casino-miniapp__stage">
              <div className="casino-miniapp__reels">
                {reels.map((symbol, index) => (
                  <article
                    key={`${symbol}-${index}`}
                    className={[
                      'casino-miniapp__reel',
                      reelStates[index] === 'spinning' ? 'casino-miniapp__reel_spinning' : '',
                      reelStates[index] === 'stopping'
                        ? `casino-miniapp__reel_stopping casino-miniapp__reel_stopping_${index + 1}`
                        : '',
                      reelStates[index] === 'winning' ? 'casino-miniapp__reel_winning' : '',
                    ]
                      .filter(Boolean)
                      .join(' ')}
                    data-win={reelStates[index] === 'winning' ? 'true' : 'false'}
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
                  {isSpinning || isResolvingSpin ? 'Идёт spin...' : `Spin ${bet}`}
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
              {slotRoundSummary ? (
                <section className="casino-miniapp__round-summary" aria-label="Round summary">
                  <article className="casino-miniapp__round-stat">
                    <span>Round</span>
                    <strong>{slotRoundSummary.outcome}</strong>
                  </article>
                  <article className="casino-miniapp__round-stat">
                    <span>Bet</span>
                    <strong>{slotRoundSummary.bet}</strong>
                  </article>
                  <article className="casino-miniapp__round-stat">
                    <span>Payout</span>
                    <strong>{slotRoundSummary.payout}</strong>
                  </article>
                  <article className="casino-miniapp__round-stat">
                    <span>Combo</span>
                    <strong>{slotRoundSummary.symbols.join(' · ')}</strong>
                  </article>
                </section>
              ) : null}
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
        ) : activeTab === 'roulette' ? (
          <RoulettePanel
            isAuthenticated={isAuthenticated}
            isActive={activeTab === 'roulette'}
            userName={user?.username}
            onMainActionChange={setRouletteMainAction}
          />
        ) : activeTab === 'plinko' ? (
          <PlinkoPanel
            isAuthenticated={isAuthenticated}
            isActive={activeTab === 'plinko'}
            balance={balance}
            userName={user?.username}
            onMainActionChange={setPlinkoMainAction}
          />
        ) : (
          <section className="casino-miniapp__coming-soon">
            <span className="casino-miniapp__kicker">Coming soon</span>
            <h2>{activeTab === 'crash' ? 'Crash' : 'Coin Flip'}</h2>
            <p>
              {activeTab === 'crash'
                ? 'Optimized crash screen will land here with its own multiplier curve and cashout loop.'
                : 'Coin flip will reuse the same shell primitives, utilities, and balance state.'}
            </p>
            <div className="casino-miniapp__coming-grid">
              <article>
                <span>Navigation</span>
                <strong>Prepared</strong>
              </article>
              <article>
                <span>State</span>
                <strong>Ready</strong>
              </article>
              <article>
                <span>Styles</span>
                <strong>Hooked</strong>
              </article>
            </div>
          </section>
        )}
      </ErrorBoundary>

      <CasinoProfileModal
        open={profileOpen}
        summary={profileSummary}
        activity={profileActivity}
        onClose={() => setProfileOpen(false)}
      />

    </main>
  )
}
