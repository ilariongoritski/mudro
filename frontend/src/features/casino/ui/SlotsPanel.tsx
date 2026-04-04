import { Link } from 'react-router-dom'
import { Minus, Plus, Zap } from 'lucide-react'

import type { CasinoHistoryItem } from '@/features/casino/api/casinoApi'
import { SlotMachine } from './SlotMachine'

const BET_OPTIONS_FALLBACK = [10, 25, 50, 100]

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
}: SlotsPanelProps) => {
  const lastWin = isSpinning ? 0 : (history[0]?.win ?? 0)
  const opts    = betOptions.length > 0 ? betOptions : BET_OPTIONS_FALLBACK

  const adjustBet = (delta: number) => {
    const idx    = opts.indexOf(bet)
    const newIdx = Math.max(0, Math.min(opts.length - 1, idx + delta))
    onBetChange(opts[newIdx])
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>

      {/* ── Slot machine ─────────────────────────────────────────────── */}
      <SlotMachine
        symbols={reels}
        isSpinning={isSpinning}
        lastWin={lastWin}
      />

      {/* ── Bet controls ─────────────────────────────────────────────── */}
      <div style={{
        borderRadius: 16,
        padding: 16,
        background: 'rgba(255,255,255,0.03)',
        border: '1px solid rgba(255,255,255,0.07)',
      }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
          <span style={{ fontSize: 11, textTransform: 'uppercase', letterSpacing: '0.2em', opacity: 0.5, color: '#fff' }}>
            Ставка
          </span>
          <span style={{ fontSize: 11, fontFamily: 'monospace', opacity: 0.5, color: '#fff' }}>
            {statusText}
          </span>
        </div>

        {/* Bet amount with +/- */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
          <button
            type="button"
            onClick={() => adjustBet(-1)}
            disabled={isSpinning || bet === opts[0]}
            style={{
              width: 44, height: 44, borderRadius: 12,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              background: 'rgba(255,255,255,0.07)',
              border: '1px solid rgba(255,255,255,0.12)',
              color: '#fff', cursor: isSpinning || bet === opts[0] ? 'not-allowed' : 'pointer',
              opacity: isSpinning || bet === opts[0] ? 0.4 : 1,
              transition: 'opacity 0.15s',
            }}
          >
            <Minus size={16} />
          </button>

          <div style={{
            flex: 1, height: 56, borderRadius: 12,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            background: 'linear-gradient(135deg, rgba(245,200,66,0.12) 0%, rgba(245,200,66,0.06) 100%)',
            border: '2px solid rgba(245,200,66,0.35)',
          }}>
            <span style={{
              fontSize: 24, fontFamily: "'Exo 2', sans-serif", fontWeight: 900, color: '#f5c842',
            }}>
              {bet.toLocaleString()}
            </span>
            <span style={{ fontSize: 12, marginLeft: 6, opacity: 0.5, color: '#fff' }}>кр</span>
          </div>

          <button
            type="button"
            onClick={() => adjustBet(1)}
            disabled={isSpinning || bet === opts[opts.length - 1]}
            style={{
              width: 44, height: 44, borderRadius: 12,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              background: 'rgba(255,255,255,0.07)',
              border: '1px solid rgba(255,255,255,0.12)',
              color: '#fff', cursor: isSpinning || bet === opts[opts.length - 1] ? 'not-allowed' : 'pointer',
              opacity: isSpinning || bet === opts[opts.length - 1] ? 0.4 : 1,
              transition: 'opacity 0.15s',
            }}
          >
            <Plus size={16} />
          </button>
        </div>

        {/* Preset chips */}
        <div style={{ display: 'grid', gridTemplateColumns: `repeat(${opts.length}, 1fr)`, gap: 6 }}>
          {opts.map((amount) => {
            const isActive = bet === amount
            return (
              <button
                key={amount}
                type="button"
                onClick={() => onBetChange(amount)}
                disabled={isSpinning}
                style={{
                  height: 36, borderRadius: 10,
                  fontSize: 12, fontFamily: 'monospace', fontWeight: 700,
                  cursor: isSpinning ? 'not-allowed' : 'pointer',
                  transition: 'all 0.15s',
                  background: isActive
                    ? 'linear-gradient(135deg, #f5c842, #c8a000)'
                    : 'rgba(255,255,255,0.05)',
                  color: isActive ? '#0d0d1a' : 'rgba(255,255,255,0.7)',
                  border: isActive
                    ? '1px solid #f5c842'
                    : '1px solid rgba(255,255,255,0.08)',
                  boxShadow: isActive ? '0 0 12px rgba(245,200,66,0.4)' : undefined,
                }}
              >
                {amount >= 1000 ? `${amount / 1000}K` : amount}
              </button>
            )
          })}
        </div>
      </div>

      {/* ── Spin button ───────────────────────────────────────────────── */}
      <button
        type="button"
        onClick={onSpin}
        disabled={!canSpin}
        className="casino-spin-btn"
        style={{ opacity: canSpin ? 1 : 0.5 }}
      >
        <Zap size={24} style={{ animation: isSpinning ? 'spin 0.5s linear infinite' : undefined }} />
        {isSpinning ? 'КРУТИМ...' : 'СПИН'}
      </button>

      {/* ── History ───────────────────────────────────────────────────── */}
      {(history.length > 0 || isHistoryFetching) && (
        <div style={{
          borderRadius: 16,
          background: 'rgba(255,255,255,0.02)',
          border: '1px solid rgba(255,255,255,0.06)',
          overflow: 'hidden',
        }}>
          <div style={{
            padding: '10px 16px',
            display: 'flex', justifyContent: 'space-between', alignItems: 'center',
            borderBottom: '1px solid rgba(255,255,255,0.05)',
          }}>
            <span style={{ fontSize: 11, textTransform: 'uppercase', letterSpacing: '0.2em', opacity: 0.5, color: '#fff' }}>
              История
            </span>
            <span style={{ fontSize: 10, opacity: 0.35, color: '#fff' }}>
              {isHistoryFetching ? 'Обновляем...' : `${history.length} записей`}
            </span>
          </div>
          <div>
            {history.map((item) => {
              const isWin = item.win > 0
              return (
                <div
                  key={item.id}
                  style={{
                    display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                    padding: '8px 16px',
                    borderBottom: '1px solid rgba(255,255,255,0.03)',
                  }}
                >
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                    <span style={{ fontSize: 15, letterSpacing: '0.08em' }}>
                      {item.symbols.join(' ')}
                    </span>
                    <span style={{ fontSize: 10, opacity: 0.35, color: '#fff' }}>
                      {formatCasinoTimestamp(item.created_at)}
                    </span>
                  </div>
                  <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 2 }}>
                    <span style={{ fontSize: 11, opacity: 0.4, color: '#fff' }}>bet {item.bet}</span>
                    <span style={{
                      fontSize: 15, fontWeight: 700, fontFamily: "'Exo 2', sans-serif",
                      color: isWin ? '#00ff88' : '#ff4466',
                    }}>
                      {isWin ? `+${item.win}` : `-${item.bet}`}
                    </span>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* ── Auth link ─────────────────────────────────────────────────── */}
      <div style={{ textAlign: 'center' }}>
        {!isAuthenticated ? (
          <Link
            to="/login"
            style={{ fontSize: 12, opacity: 0.4, color: '#f5c842', textDecoration: 'none', letterSpacing: '0.1em' }}
          >
            Войти для полного доступа →
          </Link>
        ) : (
          <Link
            to="/casino"
            style={{ fontSize: 12, opacity: 0.4, color: '#f5c842', textDecoration: 'none', letterSpacing: '0.1em' }}
          >
            Полный режим →
          </Link>
        )}
      </div>
    </div>
  )
}
