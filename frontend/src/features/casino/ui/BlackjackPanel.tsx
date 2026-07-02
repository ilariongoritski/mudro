import { useCallback, useEffect, useMemo, useState } from 'react'

import {
  type BlackjackStateResponse,
  type BlackjackStatus,
  useActionBlackjackMutation,
  useGetBlackjackStateQuery,
  useStartBlackjackMutation,
} from '@/features/casino/api/casinoApi'

import './BlackjackPanel.css'

interface BlackjackMainAction {
  label: string
  busy: boolean
  disabled: boolean
  onTrigger: () => void
}

interface BlackjackPanelProps {
  isAuthenticated: boolean
  isActive: boolean
  balance: number
  userName?: string | null
  onMainActionChange?: (action: BlackjackMainAction | null) => void
}

const SuitIcons: Record<string, string> = {
  hearts: '♥', diamonds: '♦', clubs: '♣', spades: '♠',
}

const winnerLabel: Record<string, string> = {
  player: 'Выигрыш!',
  dealer: 'Проигрыш',
  push: 'Ничья',
}

const betOptions = [25, 50, 100, 200]
const defaultBet = 50

const formatCompact = (v: number) => new Intl.NumberFormat('ru-RU').format(v)

interface CardProps {
  card: { suit: string; rank: string; value: number }
  hidden?: boolean
}

const Card = ({ card, hidden }: CardProps) => {
  if (hidden) return <div className="bjp-card bjp-card-hidden" />
  return (
    <div className={`bjp-card bjp-card-${card.suit}`}>
      <div className="bjp-card-rank">{card.rank}</div>
      <div className="bjp-card-suit">{SuitIcons[card.suit] ?? card.suit}</div>
    </div>
  )
}

interface HandProps {
  hand: { cards: { suit: string; rank: string; value: number }[]; score: number; is_bust: boolean }
  title: string
  hideSecondDealerCard?: boolean
}

const Hand = ({ hand, title, hideSecondDealerCard }: HandProps) => (
  <div className="bjp-hand-area">
    <div className="bjp-hand-header">
      <span className="bjp-hand-title">{title}</span>
      <span className="bjp-score-badge">
        {hideSecondDealerCard ? '??' : hand.score}
        {hand.is_bust && ' BUST'}
      </span>
    </div>
    <div className="bjp-hand-container">
      {hand.cards.map((card, idx) => (
        <Card key={`card-${idx}`} card={card} hidden={hideSecondDealerCard && idx === 1} />
      ))}
    </div>
  </div>
)

export const BlackjackPanel = ({ isAuthenticated, isActive, balance, userName, onMainActionChange }: BlackjackPanelProps) => {
  const [bet, setBet] = useState(defaultBet)
  const [status, setStatus] = useState('Blackjack стол готов. Выберите ставку и нажмите «Раздать».')

  const shouldQuery = isAuthenticated && isActive
  const { data: stateData, isLoading } = useGetBlackjackStateQuery(undefined, { skip: !shouldQuery })
  const [startGame, { isLoading: isStarting }] = useStartBlackjackMutation()
  const [actGame, { isLoading: isActing }] = useActionBlackjackMutation()

  const game = stateData && 'id' in stateData ? (stateData as BlackjackStateResponse) : null
  const isResolved = game?.status === ('resolved' as BlackjackStatus)
  const isBusy = isStarting || isActing
  const canStart = isAuthenticated && !isBusy && bet > 0 && bet <= balance

  const handleStart = useCallback(async () => {
    if (!canStart) {
      setStatus(isAuthenticated ? 'Недостаточно credits для ставки.' : 'Нужно войти.')
      return
    }
    try {
      await startGame({ bet }).unwrap()
      setStatus('Карты розданы. Ваш ход — Hit или Stand.')
    } catch {
      setStatus('Не удалось начать игру. Проверьте casino backend.')
    }
  }, [canStart, isAuthenticated, bet, startGame])

  const handleAction = useCallback(async (action: 'hit' | 'stand') => {
    try {
      const result = await actGame({ action }).unwrap()
      if (result.status === ('resolved' as BlackjackStatus)) {
        const winLabel = winnerLabel[result.winner ?? ''] ?? 'Раунд завершён'
        setStatus(`${winLabel} · выплата ${formatCompact(result.payout)}`)
      } else if (action === 'hit') {
        setStatus('Карта выдана. Score обновлён.')
      } else {
        setStatus('Дилер играет...')
      }
    } catch {
      setStatus('Ошибка действия. Повторите.')
    }
  }, [actGame])

  useEffect(() => {
    if (!isActive) {
      onMainActionChange?.(null)
      return
    }

    if (isResolved) {
      onMainActionChange?.({
        label: 'Новая игра',
        busy: isStarting,
        disabled: !canStart,
        onTrigger: () => void handleStart(),
      })
    } else if (game) {
      onMainActionChange?.({
        label: 'Stand',
        busy: isActing,
        disabled: false,
        onTrigger: () => void handleAction('stand'),
      })
    } else {
      onMainActionChange?.({
        label: 'Раздать',
        busy: isStarting,
        disabled: !canStart,
        onTrigger: () => void handleStart(),
      })
    }
  }, [isActive, isResolved, game, isStarting, isActing, canStart, handleStart, handleAction, onMainActionChange])

  const displayName = useMemo(() => {
    const name = userName?.toString().trim()
    return name || 'Player'
  }, [userName])

  if (!isAuthenticated) {
    return (
      <section className="bjp-panel">
        <div className="bjp-panel__not-auth">
          <h2>Blackjack</h2>
          <p>Войдите через Telegram mini app чтобы начать игру.</p>
        </div>
      </section>
    )
  }

  if (isLoading) {
    return (
      <section className="bjp-panel">
        <div className="bjp-panel__loading">Загружаем стол...</div>
      </section>
    )
  }

  return (
    <section className="bjp-panel">
      <div className="bjp-panel__head">
        <div>
          <span className="bjp-panel__kicker">Blackjack 21</span>
          <h2>Beat the dealer</h2>
        </div>
        <span className="bjp-panel__balance">Balance: {formatCompact(balance)}</span>
      </div>

      <div className="bjp-panel__table">
        {game ? (
          <div className="bjp-panel__game">
            <Hand hand={game.dealer_hand} title="Дилер" hideSecondDealerCard={!isResolved} />
            <Hand hand={game.player_hand} title={displayName} />

            {isResolved && (
              <div className="bjp-panel__result-overlay">
                <div className="bjp-panel__result-title">
                  {winnerLabel[game.winner ?? ''] ?? 'Раунд завершён'}
                </div>
                <div className="bjp-panel__result-subtitle">
                  Ставка: {formatCompact(game.bet)} · Выплата: {formatCompact(game.payout)}
                </div>
                <button
                  type="button"
                  className="bjp-btn bjp-btn-primary"
                  onClick={() => void handleStart()}
                  disabled={!canStart || isStarting}
                >
                  {isStarting ? 'Раздаём...' : 'Новая игра'}
                </button>
              </div>
            )}

            {!isResolved && (
              <div className="bjp-panel__controls">
                <button
                  type="button"
                  className="bjp-btn bjp-btn-primary"
                  onClick={() => void handleAction('hit')}
                  disabled={isBusy}
                >
                  Hit
                </button>
                <button
                  type="button"
                  className="bjp-btn bjp-btn-secondary"
                  onClick={() => void handleAction('stand')}
                  disabled={isBusy}
                >
                  Stand
                </button>
              </div>
            )}
          </div>
        ) : (
          <div className="bjp-panel__start">
            <h3>Blackjack</h3>
            <div className="bjp-panel__bet-row">
              <span>Ставка:</span>
              <div className="bjp-panel__bet-chips">
                {betOptions.map((v) => (
                  <button
                    key={v}
                    type="button"
                    className={bet === v ? 'bjp-chip bjp-chip-active' : 'bjp-chip'}
                    onClick={() => setBet(v)}
                  >
                    {v}
                  </button>
                ))}
              </div>
            </div>
            <button
              type="button"
              className="bjp-btn bjp-btn-primary"
              onClick={() => void handleStart()}
              disabled={!canStart || isStarting}
            >
              {isStarting ? 'Раздаём...' : 'Раздать'}
            </button>
          </div>
        )}
      </div>

      <p className="bjp-panel__status" role="status" aria-live="polite">
        {status}
      </p>
    </section>
  )
}
