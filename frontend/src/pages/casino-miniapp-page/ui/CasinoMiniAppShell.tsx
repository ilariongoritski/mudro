import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useDispatch } from 'react-redux'

import { useTelegramBootstrapMutation } from '@/entities/session/api/authApi'
import { setCredentials } from '@/entities/session/model/sessionSlice'
import { useTelegramWebApp } from '@/features/telegram-miniapp/hooks/useTelegramWebApp'
import {
  useGetCasinoActivityQuery,
  useGetCasinoBalanceQuery,
  useGetCasinoBonusStateQuery,
  useGetCasinoHistoryQuery,
  useGetCasinoProfileQuery,
  useSpinCasinoMutation,
} from '@/features/casino/api/casinoApi'
import { BonusPanel } from '@/features/casino/bonus/ui/BonusPanel'
import { CasinoLiveSidebar } from '@/features/casino/live-sidebar/ui/CasinoLiveSidebar'
import { buildCasinoActivityFromApi, buildCasinoProfileSummary } from '@/features/casino/model/profile'
import { PlinkoPanel } from '@/features/casino/plinko/ui/PlinkoPanel'
import { CasinoProfileModal } from '@/features/casino/profile/ui/CasinoProfileModal'
import { RoulettePanel } from '@/features/casino/roulette/ui/RoulettePanel'
import { SlotsPanel } from '@/features/casino/slots/ui/SlotsPanel'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'
import { ErrorBoundary } from '@/shared/ui/ErrorBoundary'

import './CasinoMiniAppPage.css'

const reelFallback = ['🎰', '🍒', '🍋']
const betOptions = [10, 25, 50, 100]

type GameTab = 'slots' | 'roulette' | 'bonus' | 'plinko' | 'crash' | 'coinflip'
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

const buildInitials = (label: string) =>
  label
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')

export const CasinoMiniAppShell = () => {
  const dispatch = useDispatch()
  const navigate = useNavigate()

  const [activeTab, setActiveTab] = useState<GameTab>('slots')
  const [bet, setBet] = useState(25)
  const [reels, setReels] = useState<string[]>(reelFallback)
  const [status, setStatus] = useState<string | null>(null)
  const [winPulse, setWinPulse] = useState(false)
  const [profileOpen, setProfileOpen] = useState(false)
  const [rouletteMainAction, setRouletteMainAction] = useState<GameMainAction | null>(null)
  const [bonusMainAction, setBonusMainAction] = useState<GameMainAction | null>(null)
  const [plinkoMainAction, setPlinkoMainAction] = useState<GameMainAction | null>(null)
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
  const { data: activityData } = useGetCasinoActivityQuery(20, {
    skip: !isAuthenticated,
  })
  const { data: bonusData } = useGetCasinoBonusStateQuery(undefined, {
    skip: !isAuthenticated,
  })

  const balance = balanceData?.balance ?? 0
  const freeSpinsBalance = balanceData?.free_spins_balance ?? bonusData?.free_spins_available ?? 0
  const rtp = balanceData?.rtp
  const history = historyData?.items ?? []
  const activity = useMemo(
    () => buildCasinoActivityFromApi(activityData?.items ?? profileData?.recent_activity ?? []),
    [activityData?.items, profileData?.recent_activity],
  )
  const profileSummary = useMemo(
    () =>
      buildCasinoProfileSummary({
        profile: profileData,
        user,
        balance,
        currency: balanceData?.currency,
        rtp,
        history,
        activity,
      }),
    [activity, balance, balanceData?.currency, history, profileData, rtp, user],
  )

  const displayName = profileData?.display_name ?? profileSummary.displayName
  const resolvedLevel = profileData?.level ?? profileSummary.level
  const resolvedUsername = profileData?.username ?? profileSummary.username
  const resolvedUserId = profileData?.user_id ?? profileSummary.userId
  const displayLabel = displayName || 'Гость'
  const usernameLabel = resolvedUsername || 'guest'
  const playerMonogram = buildInitials(displayName || resolvedUsername || 'M')
  const profileStatus = useMemo(() => {
    if (!isAuthenticated) {
      return 'Войдите, чтобы открыть casino profile.'
    }

    if (profileData?.last_game_at) {
      return `Последняя игра ${formatCasinoTimestamp(profileData.last_game_at)}`
    }

    return 'Profile ready.'
  }, [isAuthenticated, profileData?.last_game_at])

  const canSpin = isAuthenticated && !isSpinning && bet > 0 && (balance >= bet || freeSpinsBalance > 0)

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

    if (balance < bet && freeSpinsBalance <= 0) {
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
        setStatus(
          response.free_spin_used
            ? `Free spin сыграл на +${response.win}. Баланс ${response.balance}. Осталось ${response.free_spins_balance ?? 0}.`
            : `Выигрыш +${response.win}. Баланс ${response.balance}.`,
        )
        webApp?.HapticFeedback?.notificationOccurred('success')
      } else {
        setStatus(
          response.free_spin_used
            ? `Free spin использован. Баланс ${response.balance}. Осталось ${response.free_spins_balance ?? 0}.`
            : `Spin завершён. Баланс ${response.balance}.`,
        )
      }
    } catch {
      setReels(reelFallback)
      setStatus('Casino API сейчас недоступен.')
      webApp?.HapticFeedback?.notificationOccurred('warning')
    }
  }, [balance, bet, freeSpinsBalance, isAuthenticated, isSpinning, spinCasino, webApp])

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

    if (activeTab === 'bonus') {
      const currentAction = bonusMainAction
      const onMainButton = () => {
        if (!currentAction?.disabled) {
          currentAction?.onTrigger()
        }
      }

      webApp.MainButton.setText(currentAction?.busy ? 'Проверяем...' : currentAction?.label ?? 'Открыть bonus')
      webApp.MainButton.show()
      webApp.MainButton.onClick(onMainButton)

      return () => {
        webApp.MainButton.offClick(onMainButton)
        webApp.MainButton.hide()
      }
    }

    if (activeTab === 'crash' || activeTab === 'coinflip') {
      const onMainButton = () => setStatus('Эта игра готовится. Каркас shell уже подключён.')
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

    webApp.MainButton.setText(isSpinning ? 'Spinning...' : buttonText)
    webApp.MainButton.show()
    webApp.MainButton.onClick(onMainButton)

    return () => {
      webApp.MainButton.offClick(onMainButton)
      webApp.MainButton.hide()
    }
  }, [activeTab, bet, bonusMainAction, canSpin, isAuthenticated, isSpinning, onSpin, plinkoMainAction, rouletteMainAction, webApp])

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
    if (activeTab === 'roulette') {
      return 'Live roulette идёт через `/casino/roulette/*` proxy и SSE state.'
    }
    if (activeTab === 'plinko') {
      return 'Plinko живёт отдельным casino service screen с backend drop и общим балансом.'
    }
    if (activeTab === 'bonus') {
      return 'Bonus screen завязан на подписку, free spins и отдельный bonus state.'
    }
    if (activeTab === 'crash' || activeTab === 'coinflip') {
      return 'Shell уже готов под будущую игру.'
    }
    return 'Игровой контур подключён к отдельному casino сервису.'
  }, [activeTab, isAuthenticated, isBalanceError, isBootstrapping, isTelegram])

  const statusText = status ?? (user ? `Игрок: ${user.username}` : 'Ожидание авторизации')
  const bonusAvailable = bonusData?.free_spins_available ?? 0
  const bonusSubscribed = bonusData?.subscribed ?? false
  const bonusCtaLabel = bonusAvailable > 0 ? `+${bonusAvailable}` : '+10'
  const centerChipLabel =
    activeTab === 'roulette'
      ? 'Рулетка: live'
      : activeTab === 'bonus'
        ? 'Bonus: subscription'
        : activeTab === 'plinko'
          ? 'Plinko: ready'
          : activeTab === 'slots'
            ? 'Slots: primary'
            : 'Coming soon'

  return (
    <main className={`casino-miniapp casino-miniapp_theme_${theme} ${winPulse ? 'casino-miniapp_win' : ''}`}>
      <header className="casino-miniapp__top casino-miniapp__top_shell">
        <div className="casino-miniapp__brand">
          <span className="casino-miniapp__eyebrow">MUDRO Telegram Mini App</span>
          <h1>Telegram Web App MVP</h1>
          <p>{topMessage}</p>
        </div>

        <div className="casino-miniapp__top-actions">
          <button
            type="button"
            className="casino-miniapp__bonus-cta"
            onClick={() => setActiveTab('bonus')}
          >
            <strong>{bonusCtaLabel}</strong>
            <span>{bonusSubscribed ? 'free spins' : 'за подписку'}</span>
          </button>

          <div className="casino-miniapp__player-badge" aria-label="Текущий игрок">
            <div className="casino-miniapp__player-avatar" aria-hidden="true">
              {playerMonogram}
            </div>
            <div className="casino-miniapp__player-copy">
              <strong>{displayLabel}</strong>
              <span>@{usernameLabel}</span>
            </div>
          </div>

          <div className="casino-miniapp__balance">
            <span>Баланс</span>
            <strong>{isBalanceFetching ? '...' : balance}</strong>
            <small>
              {displayLabel} · @{usernameLabel} · LVL {resolvedLevel}
            </small>
            <small>User #{resolvedUserId || '—'}</small>
            <small>{profileStatus}</small>
            <button type="button" className="casino-miniapp__profile-button" onClick={() => setProfileOpen(true)}>
              Casino profile
            </button>
            <Link to="/profile" className="casino-miniapp__profile-button">
              Mudro profile
            </Link>
          </div>
        </div>
      </header>

      <div className="casino-miniapp__layout">
        <aside className="casino-miniapp__menu" aria-label="Casino games">
          <button
            type="button"
            className={activeTab === 'slots' ? 'casino-miniapp__menu-item casino-miniapp__menu-item_active' : 'casino-miniapp__menu-item'}
            onClick={() => setActiveTab('slots')}
          >
            <span>🎰</span> Slots
          </button>
          <button
            type="button"
            className={activeTab === 'roulette' ? 'casino-miniapp__menu-item casino-miniapp__menu-item_active' : 'casino-miniapp__menu-item'}
            onClick={() => setActiveTab('roulette')}
          >
            <span>🎯</span> Рулетка
          </button>
          <button
            type="button"
            className={activeTab === 'bonus' ? 'casino-miniapp__menu-item casino-miniapp__menu-item_active' : 'casino-miniapp__menu-item'}
            onClick={() => setActiveTab('bonus')}
          >
            <span>🎁</span> Bonus
          </button>
          <button
            type="button"
            className={activeTab === 'plinko' ? 'casino-miniapp__menu-item casino-miniapp__menu-item_active' : 'casino-miniapp__menu-item'}
            onClick={() => setActiveTab('plinko')}
          >
            <span>🔻</span> Plinko
          </button>
          <button
            type="button"
            className={
              activeTab === 'crash'
                ? 'casino-miniapp__menu-item casino-miniapp__menu-item_active casino-miniapp__menu-item_coming'
                : 'casino-miniapp__menu-item casino-miniapp__menu-item_coming'
            }
            onClick={() => {
              setActiveTab('crash')
              setStatus('Crash is coming soon. Shell is ready.')
            }}
          >
            <span>📈</span> Crash
          </button>
          <button
            type="button"
            className={
              activeTab === 'coinflip'
                ? 'casino-miniapp__menu-item casino-miniapp__menu-item_active casino-miniapp__menu-item_coming'
                : 'casino-miniapp__menu-item casino-miniapp__menu-item_coming'
            }
            onClick={() => {
              setActiveTab('coinflip')
              setStatus('Coin Flip is coming soon. Shell is ready.')
            }}
          >
            <span>🪙</span> Coinflip
          </button>

          <div className="casino-miniapp__menu-tools">
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
              Fairness
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
          </div>
        </aside>

        <section className="casino-miniapp__center">
          <div className="casino-miniapp__center-bar">
            <div className="casino-miniapp__center-status">
              <span className="casino-miniapp__center-chip">{centerChipLabel}</span>
              <span className="casino-miniapp__center-user">Пользователь: {displayLabel}</span>
            </div>
            <p className="casino-miniapp__center-note">
              {activeTab === 'bonus' && bonusAvailable > 0 ? `Доступно ${bonusAvailable} free spins.` : statusText}
            </p>
          </div>

          {fairnessOpen ? (
            <section className="casino-miniapp__fairness">
              <div>
                <span>Fairness</span>
                <strong>Provably fair surfaces ready</strong>
              </div>
              <p>Seed review, nonce strip и round audit уже предусмотрены в layout и activity metadata.</p>
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
              <SlotsPanel
                isAuthenticated={isAuthenticated}
                isSpinning={isSpinning}
                canSpin={canSpin}
                bet={bet}
                betOptions={betOptions}
                reels={reels}
                history={history}
                isHistoryFetching={isHistoryFetching}
                statusText={statusText}
                onBetChange={setBet}
                onSpin={() => void onSpin()}
                formatCasinoTimestamp={formatCasinoTimestamp}
              />
            ) : activeTab === 'roulette' ? (
              <RoulettePanel
                isAuthenticated={isAuthenticated}
                isActive={activeTab === 'roulette'}
                userName={usernameLabel}
                onMainActionChange={setRouletteMainAction}
              />
            ) : activeTab === 'bonus' ? (
              <BonusPanel
                isAuthenticated={isAuthenticated}
                isActive={activeTab === 'bonus'}
                userName={usernameLabel}
                telegramInitData={initData}
                onMainActionChange={setBonusMainAction}
              />
            ) : activeTab === 'plinko' ? (
              <PlinkoPanel
                isAuthenticated={isAuthenticated}
                isActive={activeTab === 'plinko'}
                balance={balance}
                userName={usernameLabel}
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
        </section>

        <CasinoLiveSidebar isAuthenticated={isAuthenticated} activeGame={activeTab} bonusAvailable={bonusAvailable} />
      </div>
      <CasinoProfileModal open={profileOpen} summary={profileSummary} activity={activity} onClose={() => setProfileOpen(false)} />
    </main>
  )
}

export { CasinoMiniAppShell as CasinoMiniAppPage }
