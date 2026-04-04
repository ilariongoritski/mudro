import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import {
  type RouletteBetType,
  type RoulettePhase,
  type RouletteStateResponse,
  useGetRouletteHistoryQuery,
  useGetRouletteStateQuery,
  usePlaceRouletteBetsMutation,
} from '@/features/casino/api/casinoApi'
import {
  formatRouletteClock,
  formatRouletteNumber,
  getRouletteColor,
  rouletteBetTypeLabels,
  rouletteQuickStakeValues,
  rouletteStakeTypeOptions,
  rouletteWheelOrder,
} from '@/features/casino/lib/roulette'
import { useRouletteStream } from '@/features/casino/lib/useRouletteStream'
import { InstantRoulette } from '../roulette/ui/InstantRoulette'

import './RoulettePanel.css'

interface RouletteMainAction {
  label: string
  busy: boolean
  disabled: boolean
  onTrigger: () => void
}

interface RoulettePanelProps {
  isAuthenticated: boolean
  isActive: boolean
  userName?: string | null
  onMainActionChange?: (action: RouletteMainAction | null) => void
}

interface DraftBet {
  id: string
  bet_type: RouletteBetType
  bet_value?: number | null
  stake: number
}

const defaultStake = 25
const trackNumbers = Array.from({ length: rouletteWheelOrder.length * 3 }, (_, index) => rouletteWheelOrder[index % rouletteWheelOrder.length])

const makeDraftBetId = () => `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`

const mergeRouletteState = (previous: RouletteStateResponse | null, incoming: Partial<RouletteStateResponse>) => {
  if (!incoming || typeof incoming !== 'object') {
    return previous
  }

  return {
    ...(previous ?? {}),
    ...incoming,
    history: incoming.history ?? previous?.history,
    my_bets: incoming.my_bets ?? previous?.my_bets,
    display_sequence: incoming.display_sequence ?? previous?.display_sequence,
    result_sequence: incoming.result_sequence ?? previous?.result_sequence,
  }
}

const formatCompactNumber = (value: number) => new Intl.NumberFormat('ru-RU').format(value)

const phaseLabel: Record<RoulettePhase, string> = {
  idle: 'Ожидание раунда',
  betting: 'Ставки открыты',
  locking: 'Ставки закрываются',
  spinning: 'Спин',
  result: 'Результат',
}

const connectionLabel: Record<'idle' | 'connecting' | 'connected' | 'disconnected' | 'error', string> = {
  idle: 'sync idle',
  connecting: 'syncing',
  connected: 'live',
  disconnected: 'reconnecting',
  error: 'error',
}

export const RoulettePanel = ({ isAuthenticated, isActive, userName, onMainActionChange }: RoulettePanelProps) => {
  const [gameMode, setGameMode] = useState<'live' | 'instant'>('live')
  const [liveState, setLiveState] = useState<RouletteStateResponse | null>(null)
  const [draftBetType, setDraftBetType] = useState<RouletteBetType>('straight')
  const [draftBetValue, setDraftBetValue] = useState('0')
  const [draftStake, setDraftStake] = useState(String(defaultStake))
  const [basket, setBasket] = useState<DraftBet[]>([])
  const [feedback, setFeedback] = useState('Соберите купон и дождитесь открытия ставок.')
  const [visualNumber, setVisualNumber] = useState<number>(0)
  const [isSpinningPreview, setIsSpinningPreview] = useState(false)
  const [serverClock, setServerClock] = useState(() => Date.now())
  const trackRefs = useRef<Array<HTMLButtonElement | null>>([])
  const previousRoundId = useRef<string | null>(null)

  const shouldQuery = isAuthenticated && isActive

  const {
    data: rouletteSnapshot,
    isFetching: isStateFetching,
    isError: isStateError,
  } = useGetRouletteStateQuery(undefined, {
    skip: !shouldQuery,
    pollingInterval: shouldQuery ? 5000 : 0,
    refetchOnFocus: true,
  })

  const { data: historySnapshot, isFetching: isHistoryFetching } = useGetRouletteHistoryQuery(12, {
    skip: !shouldQuery,
    pollingInterval: shouldQuery ? 10000 : 0,
    refetchOnFocus: true,
  })

  const [placeRouletteBets, { isLoading: isSubmitting }] = usePlaceRouletteBetsMutation()
  const [streamState, setStreamState] = useState<RouletteStateResponse | null>(null)

  useEffect(() => {
    if (rouletteSnapshot) {
      setLiveState(rouletteSnapshot)
      setServerClock(Date.now())
    }
  }, [rouletteSnapshot])

  useEffect(() => {
    if (!shouldQuery) {
      setStreamState(null)
      return
    }
  }, [shouldQuery])

  const stream = useRouletteStream({
    enabled: shouldQuery,
    onMessage: (message) => {
      const nextState = (message.state ?? message.round ?? message) as Partial<RouletteStateResponse>
      if (!nextState || typeof nextState !== 'object') {
        return
      }

      setStreamState((previous) => mergeRouletteState(previous, nextState))
      if (typeof nextState.server_time === 'string' || typeof nextState.betting_closes_at === 'string') {
        setServerClock(Date.now())
      }
    },
  })

  useEffect(() => {
    if (!isActive) {
      onMainActionChange?.(null)
      return
    }

    const mergedState = streamState ?? liveState ?? rouletteSnapshot ?? null
    const canSubmit = isAuthenticated && basket.length > 0 && mergedState?.phase === 'betting' && !isSubmitting
    const totalStake = basket.reduce((sum, item) => sum + item.stake, 0)

    onMainActionChange?.({
      label: isSubmitting
        ? 'Отправляем купон...'
        : basket.length > 0
          ? `Поставить ${formatCompactNumber(totalStake)}`
          : 'Соберите ставку',
      busy: isSubmitting,
      disabled: !canSubmit,
      onTrigger: () => {
        if (canSubmit) {
          void submitBets()
        }
      },
    })

  }, [basket, isActive, isAuthenticated, isSubmitting, liveState, onMainActionChange, rouletteSnapshot, streamState])

  const rouletteState = streamState ?? liveState ?? rouletteSnapshot ?? null
  const rouletteHistory = rouletteState?.history ?? historySnapshot?.items ?? []
  const myBets = rouletteState?.my_bets ?? []

  const selectedStake = Number.parseInt(draftStake, 10)
  const selectedStraightValue = Number.parseInt(draftBetValue, 10)
  const canDraftBet =
    isAuthenticated &&
    rouletteState?.phase === 'betting' &&
    Number.isFinite(selectedStake) &&
    selectedStake > 0 &&
    (!isNaN(selectedStraightValue) || draftBetType !== 'straight')

  const bettingDeadline = rouletteState?.betting_closes_at ? Date.parse(rouletteState.betting_closes_at) : null
  const timeLeft = bettingDeadline ? Math.max(bettingDeadline - serverClock, 0) : 0
  const liveBadge = connectionLabel[stream.status]
  const roundLabel = rouletteState?.round_id ? rouletteState.round_id.slice(-8).toUpperCase() : 'WAITING'

  const trackIndex = useMemo(() => {
    const searchValue = typeof rouletteState?.winning_number === 'number' ? rouletteState.winning_number : visualNumber
    const middleIndex = Math.floor(trackNumbers.length / 2)
    for (let index = middleIndex; index < trackNumbers.length; index += 1) {
      if (trackNumbers[index] === searchValue) {
        return index
      }
    }
    return trackNumbers.findIndex((item) => item === searchValue)
  }, [rouletteState?.winning_number, visualNumber])

  useEffect(() => {
    if (rouletteState?.phase === 'spinning') {
      setIsSpinningPreview(true)
      const timer = window.setInterval(() => {
        setVisualNumber(rouletteWheelOrder[Math.floor(Math.random() * rouletteWheelOrder.length)])
      }, 95)
      return () => {
        window.clearInterval(timer)
      }
    }

    setIsSpinningPreview(false)
    if (typeof rouletteState?.winning_number === 'number') {
      const timer = window.setTimeout(() => {
        setVisualNumber(rouletteState.winning_number ?? 0)
      }, 180)
      return () => window.clearTimeout(timer)
    }

    return undefined
  }, [rouletteState?.phase, rouletteState?.winning_number])

  useEffect(() => {
    if (trackIndex < 0) {
      return
    }

    const target = trackRefs.current[trackIndex]
    target?.scrollIntoView({
      behavior: isSpinningPreview ? 'smooth' : 'smooth',
      inline: 'center',
      block: 'nearest',
    })
  }, [isSpinningPreview, trackIndex, visualNumber])

  useEffect(() => {
    const timer = window.setInterval(() => {
      setServerClock(Date.now())
    }, 1000)

    return () => window.clearInterval(timer)
  }, [])

  useEffect(() => {
    const currentRoundId = rouletteState?.round_id ?? null
    if (!currentRoundId || previousRoundId.current === currentRoundId) {
      return
    }

    previousRoundId.current = currentRoundId
    setBasket([])
    setFeedback('Новый раунд открыт. Соберите свежий купон.')
  }, [rouletteState?.round_id])

  const addBet = useCallback(() => {
    if (!canDraftBet) {
      setFeedback('Ставки сейчас закрыты.')
      return
    }

    const stake = Number.parseInt(draftStake, 10)
    if (!Number.isFinite(stake) || stake <= 0) {
      setFeedback('Введите корректную сумму ставки.')
      return
    }

    if (draftBetType === 'straight') {
      const value = Number.parseInt(draftBetValue, 10)
      if (!Number.isFinite(value) || value < 0 || value > 36) {
        setFeedback('Для straight нужен номер от 0 до 36.')
        return
      }

      setBasket((current) => [
        ...current,
        {
          id: makeDraftBetId(),
          bet_type: draftBetType,
          bet_value: value,
          stake,
        },
      ])
      setFeedback(`Добавлена ставка straight на ${value}.`)
      return
    }

    setBasket((current) => [
      ...current,
      {
        id: makeDraftBetId(),
        bet_type: draftBetType,
        bet_value: null,
        stake,
      },
    ])
    setFeedback(`Добавлена ставка ${rouletteBetTypeLabels[draftBetType]}.`)
  }, [canDraftBet, draftBetType, draftBetValue, draftStake])

  const resetBasket = useCallback(() => {
    setBasket([])
    setDraftStake(String(defaultStake))
    setDraftBetType('straight')
    setDraftBetValue('0')
    setFeedback('Купон сброшен.')
  }, [])

  const submitBets = useCallback(async () => {
    if (!isAuthenticated) {
      setFeedback('Нужно войти в Telegram mini app.')
      return
    }

    const currentState = rouletteState ?? liveState ?? rouletteSnapshot ?? null
    if (!currentState?.round_id) {
      setFeedback('Раунд ещё не готов.')
      return
    }

    if (basket.length === 0) {
      setFeedback('Сначала соберите купон.')
      return
    }

    if (currentState.phase !== 'betting') {
      setFeedback('Ставки закрыты, дождитесь следующего раунда.')
      return
    }

    try {
      const response = await placeRouletteBets({
        round_id: currentState.round_id,
        bets: basket.map(({ bet_type, bet_value, stake }) => ({
          bet_type,
          bet_value,
          stake,
        })),
      }).unwrap()

      if (response.state) {
        setStreamState((previous) => mergeRouletteState(previous, response.state ?? {}))
      }

      setBasket([])
      setFeedback(
        response.status === 'accepted'
          ? `Ставки приняты. Раунд ${response.round_id ?? currentState.round_id.slice(-6)}.`
          : 'Купон отправлен.',
      )
    } catch {
      setFeedback('Не удалось отправить ставки. Проверьте соединение с casino proxy.')
    }
  }, [basket, isAuthenticated, liveState, placeRouletteBets, rouletteSnapshot, rouletteState])

  const totalStake = basket.reduce((sum, item) => sum + item.stake, 0)
  const currentWinningNumber = typeof rouletteState?.winning_number === 'number' ? rouletteState.winning_number : null
  const currentWinningColor = currentWinningNumber != null ? getRouletteColor(currentWinningNumber) : rouletteState?.winning_color
  const phase = rouletteState?.phase ?? 'idle'

  const activeBetSummary = myBets.length
    ? myBets.map((item) => `${rouletteBetTypeLabels[item.bet_type]} ${item.bet_value ?? ''}`.trim()).join(' · ')
    : 'Ставок в раунде пока нет.'

  const historyLabel = isHistoryFetching ? 'Обновляем историю...' : `${rouletteHistory.length} rounds`

  return (
    <section className={`roulette-panel ${isActive ? 'roulette-panel_active' : ''}`}>
      <div 
        className="flex mb-4 p-1 bg-white/5 rounded-2xl w-fit border border-white/10"
        style={{ margin: '0 1rem 1rem' }}
      >
        <button
          onClick={() => setGameMode('live')}
          className={`px-4 py-2 rounded-xl text-xs font-bold transition-all ${
            gameMode === 'live' ? 'bg-[#f5c842] text-[#0d0d1a]' : 'text-white/40'
          }`}
        >
          LIVE (SSE)
        </button>
        <button
          onClick={() => setGameMode('instant')}
          className={`px-4 py-2 rounded-xl text-xs font-bold transition-all ${
            gameMode === 'instant' ? 'bg-[#f5c842] text-[#0d0d1a]' : 'text-white/40'
          }`}
        >
          МГНОВЕННАЯ
        </button>
      </div>

      {gameMode === 'instant' ? (
        <InstantRoulette
          isAuthenticated={isAuthenticated}
          userName={userName}
          onMainActionChange={onMainActionChange}
        />
      ) : (
        <>
          <header className="roulette-panel__header">
            <div>
              <span className="roulette-panel__eyebrow">Roulette 0-36</span>
              <h2>Live round</h2>
              <p>
                {phaseLabel[phase]} · {userName ? `@${userName}` : 'guest'}
              </p>
            </div>

            <div className="roulette-panel__status">
              <span className={`roulette-panel__live roulette-panel__live_${stream.status}`}>{liveBadge}</span>
              <strong>{formatRouletteClock(timeLeft)}</strong>
              <small>#{roundLabel}</small>
            </div>
          </header>

          <div className="roulette-panel__wheel">
            <div className="roulette-panel__pointer" aria-hidden="true" />
            <div className="roulette-panel__track" aria-label="Roulette number strip">
              {trackNumbers.map((value, index) => {
                const active = trackIndex === index
                const color = getRouletteColor(value)

                return (
                  <button
                    key={`${index}-${value}`}
                    ref={(element) => {
                      trackRefs.current[index] = element
                    }}
                    type="button"
                    className={`roulette-panel__tile roulette-panel__tile_${color} ${active ? 'roulette-panel__tile_active' : ''}`}
                  >
                    {formatRouletteNumber(value)}
                  </button>
                )
              })}
            </div>
          </div>

          <div className="roulette-panel__summary">
            <article>
              <span>Последний результат</span>
              <strong>
                {currentWinningNumber != null ? formatRouletteNumber(currentWinningNumber) : '—'}
                {currentWinningColor ? <em>{currentWinningColor}</em> : null}
              </strong>
            </article>
            <article>
              <span>Истекает через</span>
              <strong>{formatRouletteClock(timeLeft)}</strong>
            </article>
            <article>
              <span>Мои ставки</span>
              <strong>{myBets.length}</strong>
            </article>
          </div>

          <div className="roulette-panel__grid">
            <section className="roulette-panel__surface roulette-panel__surface_main">
              <div className="roulette-panel__surface-head">
                <div>
                  <span className="roulette-panel__kicker">Ставки</span>
                  <h3>Соберите купон на текущий раунд.</h3>
                </div>
              </div>

              <div className="roulette-panel__type-row" role="tablist" aria-label="Roulette bet types">
                {rouletteStakeTypeOptions.map((option) => (
                  <button
                    key={option.value}
                    type="button"
                    className={`roulette-panel__type ${draftBetType === option.value ? 'roulette-panel__type_active' : ''}`}
                    onClick={() => setDraftBetType(option.value)}
                  >
                    <strong>{option.label}</strong>
                    <small>{option.description}</small>
                  </button>
                ))}
              </div>

              <div className="roulette-panel__form-grid">
                {draftBetType === 'straight' ? (
                  <label className="roulette-panel__field">
                    <span>Number</span>
                    <input
                      type="number"
                      min="0"
                      max="36"
                      value={draftBetValue}
                      onChange={(event) => setDraftBetValue(event.target.value)}
                      disabled={!isAuthenticated || rouletteState?.phase !== 'betting'}
                    />
                  </label>
                ) : (
                  <div className="roulette-panel__field roulette-panel__field_readonly">
                    <span>Number</span>
                    <strong>{rouletteStakeTypeOptions.find((item) => item.value === draftBetType)?.label ?? '—'}</strong>
                  </div>
                )}

                <label className="roulette-panel__field">
                  <span>Stake</span>
                  <input
                    type="number"
                    min="1"
                    step="1"
                    value={draftStake}
                    onChange={(event) => setDraftStake(event.target.value)}
                    disabled={!isAuthenticated || rouletteState?.phase !== 'betting'}
                  />
                </label>
              </div>

              <div className="roulette-panel__quick-stakes" aria-label="Quick stake values">
                {rouletteQuickStakeValues.map((value) => (
                  <button
                    key={value}
                    type="button"
                    className={draftStake === String(value) ? 'roulette-panel__quick-stake roulette-panel__quick-stake_active' : 'roulette-panel__quick-stake'}
                    onClick={() => setDraftStake(String(value))}
                  >
                    {value}
                  </button>
                ))}
              </div>

              <div className="roulette-panel__actions">
                <button
                  type="button"
                  className="roulette-panel__primary"
                  onClick={addBet}
                  disabled={!canDraftBet}
                >
                  Добавить ставку
                </button>
                <button
                  type="button"
                  className="roulette-panel__secondary"
                  onClick={resetBasket}
                  disabled={basket.length === 0 && draftBetType === 'straight' && draftStake === String(defaultStake) && draftBetValue === '0'}
                >
                  Сброс
                </button>
              </div>

              <div className="roulette-panel__basket">
                <div className="roulette-panel__basket-head">
                  <strong>Купон</strong>
                  <span>{formatCompactNumber(totalStake)} credits</span>
                </div>

                {basket.length === 0 ? (
                  <p className="roulette-panel__empty">Ставок в купоне пока нет.</p>
                ) : (
                  <div className="roulette-panel__basket-list">
                    {basket.map((item) => (
                      <button
                        key={item.id}
                        type="button"
                        className="roulette-panel__basket-item"
                        onClick={() => setBasket((current) => current.filter((entry) => entry.id !== item.id))}
                      >
                        <span>
                          {rouletteBetTypeLabels[item.bet_type]}
                          {item.bet_type === 'straight' && typeof item.bet_value === 'number' ? ` ${item.bet_value}` : ''}
                        </span>
                        <strong>{formatCompactNumber(item.stake)}</strong>
                      </button>
                    ))}
                  </div>
                )}
              </div>

              <div className="roulette-panel__bets-note" role="status" aria-live="polite">
                {feedback}
              </div>
            </section>

            <aside className="roulette-panel__surface roulette-panel__surface_side">
              <div className="roulette-panel__surface-head">
                <div>
                  <span className="roulette-panel__kicker">Раунд</span>
                  <h3>История и состояние</h3>
                </div>
                <span className="roulette-panel__surface-badge">{historyLabel}</span>
              </div>

              <div className="roulette-panel__history">
                {rouletteHistory.length === 0 && !isStateFetching && !isStateError ? (
                  <p className="roulette-panel__empty">
                    История появится после первых результатов.
                  </p>
                ) : null}

                {rouletteHistory.map((item) => {
                  const color = item.winning_color
                  const tone = color === 'green' ? 'roulette-panel__history-item_green' : color === 'red' ? 'roulette-panel__history-item_red' : 'roulette-panel__history-item_black'

                  return (
                    <article key={item.round_id} className={`roulette-panel__history-item ${tone}`}>
                      <div className="roulette-panel__history-copy">
                        <strong>{formatRouletteNumber(item.winning_number)}</strong>
                        <span>{phaseLabel[item.phase] ?? item.phase}</span>
                        <small>{item.round_id.slice(-8).toUpperCase()}</small>
                      </div>
                      <span className={`roulette-panel__history-color roulette-panel__history-color_${color}`}>
                        {color}
                      </span>
                    </article>
                  )
                })}
              </div>

              <div className="roulette-panel__side-block">
                <span className="roulette-panel__kicker">Active bets</span>
                <p className="roulette-panel__active-bets">{activeBetSummary}</p>
              </div>

              <div className="roulette-panel__side-block roulette-panel__side-block_subtle">
                <span className="roulette-panel__kicker">State</span>
                <p className="roulette-panel__active-bets">
                  {stream.status === 'connected'
                    ? 'SSE connected'
                    : stream.status === 'error'
                      ? `SSE error${stream.error ? `: ${stream.error}` : ''}`
                      : 'Polling fallback'}
                </p>
              </div>
            </aside>
          </div>
        </>
      )}
    </section>
  )
}
