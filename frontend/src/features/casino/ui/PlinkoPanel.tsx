import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import {
  type PlinkoDropResponse,
  type PlinkoRisk,
  useDropPlinkoMutation,
  useGetPlinkoConfigQuery,
  useGetPlinkoStateQuery,
} from '@/features/casino/api/casinoApi'
import { plinkoQuickStakeValues, plinkoRiskOptions } from '@/features/casino/lib/plinko'

import './PlinkoPanel.css'

interface PlinkoMainAction {
  label: string
  busy: boolean
  disabled: boolean
  onTrigger: () => void
}

interface PlinkoPanelProps {
  isAuthenticated: boolean
  isActive: boolean
  balance: number
  userName?: string | null
  onMainActionChange?: (action: PlinkoMainAction | null) => void
}

interface PlinkoVisualRound extends PlinkoDropResponse {
  lanePath: number[]
}

const defaultRisk: PlinkoRisk = 'medium'
const defaultStake = 25

const formatCompactNumber = (value: number) => new Intl.NumberFormat('ru-RU').format(value)

const buildLanePath = (path: number[]) => {
  const lanePath = [0]
  let lane = 0
  for (const step of path) {
    lane += step
    lanePath.push(lane)
  }
  return lanePath
}

export const PlinkoPanel = ({ isAuthenticated, isActive, balance, userName, onMainActionChange }: PlinkoPanelProps) => {
   const [wallet, setWallet] = useState(balance)
   const [stake, setStake] = useState(String(defaultStake))
   const [risk, setRisk] = useState<PlinkoRisk>(defaultRisk)
   const [status, setStatus] = useState('Plinko подключён к casino service. Выберите риск и нажмите drop.')
   const [history, setHistory] = useState<PlinkoVisualRound[]>([])
   const [activeRound, setActiveRound] = useState<PlinkoVisualRound | null>(null)
   const [activeStep, setActiveStep] = useState(0)
   const [isDropping, setIsDropping] = useState(false)
   const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

   const shouldQuery = isAuthenticated && isActive
   const { data: plinkoConfig } = useGetPlinkoConfigQuery(undefined, {
     skip: !shouldQuery,
   })
   const { data: plinkoState } = useGetPlinkoStateQuery(undefined, {
     skip: !shouldQuery,
     pollingInterval: shouldQuery ? 12000 : 0,
     refetchOnFocus: true,
   })
   const [dropPlinko, { isLoading: isSubmitting }] = useDropPlinkoMutation()

   // Sync wallet with prop balance when it changes externally
   useEffect(() => {
     setWallet(balance)
   }, [balance])

   // Sync wallet with plinko state when it updates from the server (e.g. polling or other games)
   useEffect(() => {
     if (plinkoState?.balance !== undefined && !isDropping) {
       setWallet(plinkoState.balance)
     }
   }, [plinkoState?.balance, isDropping])

  const clearTimer = useCallback(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current)
      timerRef.current = null
    }
  }, [])

  useEffect(() => clearTimer, [clearTimer])

  const rows = plinkoState?.config.rows ?? plinkoConfig?.rows ?? 12
  const multipliers = plinkoState?.config.multipliers?.[risk] ?? plinkoConfig?.multipliers?.[risk] ?? []
  const stakeValue = Number.parseInt(stake, 10)
  const canDrop = isAuthenticated && !isSubmitting && !isDropping && Number.isFinite(stakeValue) && stakeValue > 0 && stakeValue <= wallet
  const currentPathLabel = activeRound?.path?.length ? activeRound.path.join(' → ') : '—'
  const currentMultiplier = activeRound?.multiplier ?? multipliers[Math.floor(multipliers.length / 2)] ?? 1
  const currentWin = activeRound?.payout ?? 0

  const dropBall = useCallback(async () => {
    if (!canDrop) {
      setStatus(isAuthenticated ? 'Недостаточно bankroll для выбранной ставки.' : 'Нужно войти в Telegram mini app.')
      return
    }

    try {
      const response = await dropPlinko({
        bet: stakeValue,
        risk,
      }).unwrap()

      const visualRound: PlinkoVisualRound = {
        ...response,
        lanePath: buildLanePath(response.path),
      }

      clearTimer()
      setActiveRound(visualRound)
      setActiveStep(0)
      setIsDropping(true)
      setWallet(response.balance)
      setStatus(`Ball released · ${visualRound.risk.toUpperCase()} risk · ${visualRound.rows} rows`)

      let step = 0
      timerRef.current = window.setInterval(() => {
        step += 1
        if (step >= visualRound.lanePath.length) {
          clearTimer()
          setIsDropping(false)
          setHistory((current) => [visualRound, ...current].slice(0, 8))
          setStatus(
            `Path ${visualRound.path.join(' → ')} · x${visualRound.multiplier.toFixed(2)} · payout ${formatCompactNumber(visualRound.payout)}`,
          )
          return
        }

        setActiveStep(step)
      }, 120)
    } catch {
      setStatus('Не удалось выполнить Plinko drop через casino service.')
    }
  }, [canDrop, clearTimer, dropPlinko, isAuthenticated, risk, stakeValue])

  useEffect(() => {
    if (!isActive) {
      onMainActionChange?.(null)
      return
    }

    onMainActionChange?.({
      label: isSubmitting || isDropping ? 'Dropping...' : `Drop ball ${formatCompactNumber(stakeValue || defaultStake)}`,
      busy: isSubmitting || isDropping,
      disabled: !canDrop,
      onTrigger: () => {
        if (canDrop) {
          void dropBall()
        }
      },
    })
  }, [canDrop, dropBall, isActive, isDropping, isSubmitting, onMainActionChange, stakeValue])

  const boardRows = useMemo(
    () =>
      Array.from({ length: rows + 1 }, (_, rowIndex) => ({
        rowIndex,
        laneIndex: activeRound?.lanePath[rowIndex] ?? -1,
      })),
    [activeRound?.lanePath, rows],
  )

  return (
    <section className="plinko-panel">
      <header className="plinko-panel__header">
        <div>
          <span className="plinko-panel__eyebrow">Plinko</span>
          <h2>Service-backed board</h2>
          <p>
            {userName ? `@${userName}` : 'guest'} · {rows} rows · {risk} risk
          </p>
        </div>

        <div className="plinko-panel__wallet">
          <span>Bankroll</span>
          <strong>{formatCompactNumber(wallet)}</strong>
          <small>{isDropping ? 'Ball in motion' : 'Ready to drop'}</small>
        </div>
      </header>

      <div className="plinko-panel__board" aria-label="Plinko board">
        {boardRows.map(({ rowIndex, laneIndex }) => (
          <div key={rowIndex} className="plinko-panel__row">
            {Array.from({ length: rowIndex + 1 }, (_, lane) => {
              const isCurrent = rowIndex === activeStep && lane === laneIndex
              const isVisited = lane === laneIndex && rowIndex <= activeStep && laneIndex >= 0

              return (
                <div
                  key={`${rowIndex}-${lane}`}
                  className={[
                    'plinko-panel__peg',
                    isVisited ? 'plinko-panel__peg_visited' : '',
                    isCurrent ? 'plinko-panel__peg_current' : '',
                  ]
                    .filter(Boolean)
                    .join(' ')}
                >
                  {isCurrent ? <span className="plinko-panel__ball" /> : null}
                </div>
              )
            })}
          </div>
        ))}
      </div>

      <div className="plinko-panel__controls">
        <section className="plinko-panel__surface">
          <div className="plinko-panel__surface-head">
            <div>
              <span className="plinko-panel__kicker">Controls</span>
              <h3>Risk, stake, live payout</h3>
            </div>
            <div className="plinko-panel__result-pill">
              <span>x{currentMultiplier.toFixed(2)}</span>
              <small>{currentWin > 0 ? `+${formatCompactNumber(currentWin)}` : 'no result yet'}</small>
            </div>
          </div>

          <div className="plinko-panel__row-controls">
            {plinkoRiskOptions.map((option) => (
              <button
                key={option.value}
                type="button"
                className={risk === option.value ? 'plinko-panel__chip plinko-panel__chip_active' : 'plinko-panel__chip'}
                onClick={() => setRisk(option.value)}
              >
                <strong>{option.label}</strong>
                <small>{option.description}</small>
              </button>
            ))}
          </div>

          <div className="plinko-panel__row-controls">
            <button type="button" className="plinko-panel__chip plinko-panel__chip_active" disabled>
              <strong>{rows}</strong>
              <small>server rows</small>
            </button>
          </div>

          <div className="plinko-panel__form-grid">
            <label className="plinko-panel__field">
              <span>Stake</span>
              <input
                type="number"
                min={plinkoState?.config.min_bet ?? plinkoConfig?.min_bet ?? 1}
                step="1"
                value={stake}
                onChange={(event) => setStake(event.target.value)}
                disabled={!isAuthenticated || isSubmitting || isDropping}
              />
            </label>

            <div className="plinko-panel__quick-stakes">
              {plinkoQuickStakeValues.map((value) => (
                <button
                  key={value}
                  type="button"
                  className={stake === String(value) ? 'plinko-panel__quick-stake plinko-panel__quick-stake_active' : 'plinko-panel__quick-stake'}
                  onClick={() => setStake(String(value))}
                >
                  {value}
                </button>
              ))}
            </div>
          </div>

          <div className="plinko-panel__actions">
            <button type="button" className="plinko-panel__primary" onClick={() => void dropBall()} disabled={!canDrop}>
              {isSubmitting || isDropping ? 'Dropping...' : 'Drop ball'}
            </button>
            <button
              type="button"
              className="plinko-panel__secondary"
              onClick={() => {
                clearTimer()
                setActiveRound(null)
                setActiveStep(0)
                setIsDropping(false)
                setStatus('Plinko reset.')
              }}
              disabled={!isDropping && !activeRound}
            >
              Reset
            </button>
          </div>

          <div className="plinko-panel__result">
            <div>
              <span>Path</span>
              <strong>{currentPathLabel}</strong>
            </div>
            <div>
              <span>Multiplier</span>
              <strong>{activeRound ? `x${activeRound.multiplier.toFixed(2)}` : 'x0.00'}</strong>
            </div>
            <div>
              <span>Payout</span>
              <strong>{activeRound ? formatCompactNumber(activeRound.payout) : '0'}</strong>
            </div>
          </div>

          <p className="plinko-panel__status" role="status" aria-live="polite">
            {status}
          </p>
        </section>

        <aside className="plinko-panel__surface plinko-panel__surface_side">
          <div className="plinko-panel__surface-head">
            <div>
              <span className="plinko-panel__kicker">History</span>
              <h3>Recent drops</h3>
            </div>
            <span className="plinko-panel__surface-badge">{history.length} entries</span>
          </div>

          <div className="plinko-panel__history">
            {history.length === 0 ? (
              <p className="plinko-panel__empty">No drops yet.</p>
            ) : (
              history.map((item) => {
                const positive = item.net_result >= 0

                return (
                  <article key={`${item.created_at}-${item.slot_index}`} className={positive ? 'plinko-panel__history-item plinko-panel__history-item_win' : 'plinko-panel__history-item'}>
                    <div className="plinko-panel__history-copy">
                      <strong>
                        x{item.multiplier.toFixed(2)} · {formatCompactNumber(item.payout)}
                      </strong>
                      <span>
                        {item.risk.toUpperCase()} · {item.rows} rows · stake {formatCompactNumber(item.bet)}
                      </span>
                      <small>{item.path.join(' → ')}</small>
                    </div>
                    <div className="plinko-panel__history-result">
                      <strong>{positive ? `+${formatCompactNumber(item.net_result)}` : formatCompactNumber(item.net_result)}</strong>
                      <span>{new Date(item.created_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}</span>
                    </div>
                  </article>
                )
              })
            )}
          </div>
        </aside>
      </div>
    </section>
  )
}
