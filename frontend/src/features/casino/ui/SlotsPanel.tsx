import { Link } from 'react-router-dom'

import type { CasinoHistoryItem } from '@/features/casino/api/casinoApi'

interface SlotsPanelProps {
  isAuthenticated: boolean
  isSpinning: boolean
  canSpin: boolean
  bet: number
  betOptions: number[]
  reels: string[]
  history: CasinoHistoryItem[]
  isHistoryFetching: boolean
  statusText: string
  onBetChange: (bet: number) => void
  onSpin: () => void
  formatCasinoTimestamp: (value: string) => string
}

export const SlotsPanel = ({
  isAuthenticated,
  isSpinning,
  canSpin,
  bet,
  betOptions,
  reels,
  history,
  isHistoryFetching,
  statusText,
  onBetChange,
  onSpin,
  formatCasinoTimestamp,
}: SlotsPanelProps) => (
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
            onClick={() => onBetChange(option)}
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
)
