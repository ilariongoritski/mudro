import { createPortal } from 'react-dom'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { ChevronDown, Music2, Palette, ShieldCheck, X } from 'lucide-react'

import type { CasinoActivityItem, CasinoProfileSummary } from '@/features/casino/model/profile'
import { formatCasinoDateTime } from '@/features/casino/model/profile'

import './CasinoProfileModal.css'

interface CasinoProfileModalProps {
  open: boolean
  summary: CasinoProfileSummary
  activity: CasinoActivityItem[]
  onClose: () => void
}

const statusLabel: Record<CasinoActivityItem['status'], string> = {
  WIN: 'WIN',
  LOST: 'LOST',
  CASHOUT: 'CASHOUT',
}

const statusTone: Record<CasinoActivityItem['status'], string> = {
  WIN: 'casino-profile-modal__activity-status_win',
  LOST: 'casino-profile-modal__activity-status_lost',
  CASHOUT: 'casino-profile-modal__activity-status_cashout',
}

const gameToneClass = (gameType: string) => {
  const normalized = gameType.trim().toLowerCase()

  switch (normalized) {
    case 'slots':
      return 'casino-profile-modal__activity-game_slots'
    case 'roulette':
      return 'casino-profile-modal__activity-game_roulette'
    case 'plinko':
      return 'casino-profile-modal__activity-game_plinko'
    case 'crash':
      return 'casino-profile-modal__activity-game_crash'
    case 'coinflip':
      return 'casino-profile-modal__activity-game_coinflip'
    case 'bonus':
      return 'casino-profile-modal__activity-game_bonus'
    default:
      return 'casino-profile-modal__activity-game_default'
  }
}

const resolveGameBucket = (gameType: string) => {
  const normalized = gameType.trim().toLowerCase()

  if (normalized === 'slots') return 'slots'
  if (normalized === 'roulette') return 'roulette'
  return 'other'
}

const formatCompact = (value: number) => new Intl.NumberFormat('ru-RU').format(value)

export const CasinoProfileModal = ({ open, summary, activity, onClose }: CasinoProfileModalProps) => {
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [themeMode, setThemeMode] = useState<'auto' | 'dark' | 'neon'>('auto')
  const [soundEnabled, setSoundEnabled] = useState(true)

  useEffect(() => {
    if (!open) return

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [onClose, open])

  useEffect(() => {
    if (!open) return

    const previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      document.body.style.overflow = previousOverflow
    }
  }, [open])

  if (!open) return null

  const progressPercent = Math.max(4, Math.min(100, Math.round(summary.currentLevelProgress * 100)))
  const remainingToNextLevel = Math.max(summary.nextLevelTarget - summary.totalWagered, 0)
  const levelLabel = `LVL ${summary.level}`
  const fairnessVerified = activity.filter((item) => item.fairnessStatus === 'verified').length
  const slotsActivity = activity.filter((item) => resolveGameBucket(item.gameType) === 'slots')
  const rouletteActivity = activity.filter((item) => resolveGameBucket(item.gameType) === 'roulette')
  const secondaryActivity = activity.filter((item) => resolveGameBucket(item.gameType) === 'other')

  const renderActivityItem = (item: CasinoActivityItem) => (
    <article key={`${item.gameType}-${item.id}`} className="casino-profile-modal__activity-item">
      <div className="casino-profile-modal__activity-copy">
        <span className={`casino-profile-modal__activity-game ${gameToneClass(item.gameType)}`}>{item.gameType}</span>
        <strong>{item.title}</strong>
        <small>{formatCasinoDateTime(item.createdAt)}</small>
        {item.details ? <p>{item.details}</p> : null}
        <div className="casino-profile-modal__activity-markers">
          {item.markers?.map((marker) => (
            <span key={marker} className="casino-profile-modal__activity-marker">
              {marker}
            </span>
          ))}
          {item.fairnessStatus ? (
            <span className="casino-profile-modal__activity-marker casino-profile-modal__activity-marker_fairness">
              fairness {item.fairnessStatus}
            </span>
          ) : null}
        </div>
      </div>

      <div className="casino-profile-modal__activity-result">
        <strong className={item.amount >= 0 ? 'casino-profile-modal__activity-amount_positive' : 'casino-profile-modal__activity-amount_negative'}>
          {item.amount >= 0 ? `+${formatCompact(item.amount)}` : formatCompact(item.amount)}
        </strong>
        <span className={`casino-profile-modal__activity-status ${statusTone[item.status]}`}>
          {statusLabel[item.status]}
        </span>
        <small>bet {formatCompact(item.bet)}</small>
      </div>
    </article>
  )

  return createPortal(
    <div className="casino-profile-modal" role="dialog" aria-modal="true" aria-label="Casino profile">
      <button type="button" className="casino-profile-modal__backdrop" aria-label="Закрыть профиль" onClick={onClose} />

      <section className="casino-profile-modal__panel" data-theme={themeMode}>
        <header className="casino-profile-modal__head">
          <div>
            <span className="casino-profile-modal__eyebrow">Профиль</span>
            <h2>{summary.displayName}</h2>
            <p>
              User #{summary.userId || '—'} · @{summary.username}
            </p>
          </div>

          <div className="casino-profile-modal__actions">
            <Link to="/profile" className="casino-profile-modal__ghost-action">
              Mudro profile
            </Link>
            <button
              type="button"
              className="casino-profile-modal__ghost-action"
              onClick={() => setSettingsOpen((value) => !value)}
              aria-expanded={settingsOpen}
            >
              <Palette size={16} />
              <span>Theme</span>
              <ChevronDown size={14} />
            </button>
            <button type="button" className="casino-profile-modal__close" onClick={onClose}>
              <X size={16} />
            </button>
          </div>
        </header>

        <div className="casino-profile-modal__summary-grid">
          <article className="casino-profile-modal__summary-card">
            <span>User</span>
            <strong>{summary.userId || '—'}</strong>
            <small>@{summary.username}</small>
          </article>
          <article className="casino-profile-modal__summary-card">
            <span>Баланс</span>
            <strong>{formatCompact(summary.balance)}</strong>
            <small>{summary.currency}</small>
          </article>
          <article className="casino-profile-modal__summary-card">
            <span>Уровень</span>
            <strong>{levelLabel}</strong>
            <small>{summary.gamesPlayed} игр</small>
          </article>
          <article className="casino-profile-modal__summary-card">
            <span>Отыграл</span>
            <strong>{formatCompact(summary.totalWagered)}</strong>
            <small>{summary.rtp ? `RTP ${summary.rtp}%` : 'RTP по config'}</small>
          </article>
        </div>

        <div className="casino-profile-modal__progress">
          <div className="casino-profile-modal__progress-head">
            <strong>{levelLabel}</strong>
            <span>
              {formatCompact(summary.totalWagered)} / {formatCompact(summary.nextLevelTarget)} credits
            </span>
          </div>
          <div className="casino-profile-modal__progress-bar" aria-hidden="true">
            <span style={{ width: `${progressPercent}%` }} />
          </div>
          <p>До следующего уровня осталось {formatCompact(remainingToNextLevel)} credits.</p>
        </div>

        <div className="casino-profile-modal__metrics">
          <article>
            <span>Выиграл</span>
            <strong>{formatCompact(summary.totalWon)}</strong>
          </article>
          <article>
            <span>Игры</span>
            <strong>{summary.gamesPlayed}</strong>
          </article>
          <article>
            <span>Slots rounds</span>
            <strong>{summary.slotsRoundsPlayed}</strong>
          </article>
          <article>
            <span>Roulette rounds</span>
            <strong>{summary.rouletteRoundsPlayed}</strong>
          </article>
          <article>
            <span>Secondary</span>
            <strong>{summary.otherGamesPlayed}</strong>
          </article>
        </div>

        <section className={`casino-profile-modal__settings ${settingsOpen ? 'casino-profile-modal__settings_open' : ''}`}>
          <div className="casino-profile-modal__settings-head">
            <div>
              <span>Fairness / theme / sound</span>
              <p>Slots and roulette are primary. Other game types stay supported as secondary labels and future-friendly markers.</p>
            </div>
            <button type="button" className="casino-profile-modal__settings-toggle" onClick={() => setSettingsOpen((value) => !value)}>
              {settingsOpen ? 'Скрыть' : 'Показать'}
            </button>
          </div>

          {settingsOpen ? (
            <div className="casino-profile-modal__settings-body">
              <div className="casino-profile-modal__settings-group">
                <span className="casino-profile-modal__settings-label">
                  <ShieldCheck size={14} />
                  Fairness
                </span>
                <div className="casino-profile-modal__settings-row">
                  <span className="casino-profile-modal__settings-pill">verified {fairnessVerified}</span>
                  <span className="casino-profile-modal__settings-pill">pending {Math.max(activity.length - fairnessVerified, 0)}</span>
                </div>
              </div>

              <div className="casino-profile-modal__settings-group">
                <span className="casino-profile-modal__settings-label">
                  <Palette size={14} />
                  Theme
                </span>
                <div className="casino-profile-modal__settings-row">
                  {(['auto', 'dark', 'neon'] as const).map((mode) => (
                    <button
                      key={mode}
                      type="button"
                      className={themeMode === mode ? 'casino-profile-modal__settings-pill casino-profile-modal__settings-pill_active' : 'casino-profile-modal__settings-pill'}
                      onClick={() => setThemeMode(mode)}
                    >
                      {mode}
                    </button>
                  ))}
                </div>
              </div>

              <div className="casino-profile-modal__settings-group">
                <span className="casino-profile-modal__settings-label">
                  <Music2 size={14} />
                  Sound
                </span>
                <button
                  type="button"
                  className={soundEnabled ? 'casino-profile-modal__settings-pill casino-profile-modal__settings-pill_active' : 'casino-profile-modal__settings-pill'}
                  onClick={() => setSoundEnabled((value) => !value)}
                >
                  {soundEnabled ? 'on' : 'off'}
                </button>
              </div>
            </div>
          ) : null}
        </section>

        <section className="casino-profile-modal__activity">
          <div className="casino-profile-modal__activity-head">
            <span>История ставок</span>
            <small>{activity.length} записей</small>
          </div>

          <div className="casino-profile-modal__activity-list">
            {activity.length === 0 ? (
              <p className="casino-profile-modal__empty">
                История появится после первого spin, ставки или roulette-раунда.
              </p>
            ) : (
              <>
                {slotsActivity.length > 0 ? (
                  <section className="casino-profile-modal__activity-group">
                    <div className="casino-profile-modal__activity-group-head">
                      <span>Slots</span>
                      <small>{slotsActivity.length}</small>
                    </div>
                    <div className="casino-profile-modal__activity-group-list">
                      {slotsActivity.map(renderActivityItem)}
                    </div>
                  </section>
                ) : null}

                {rouletteActivity.length > 0 ? (
                  <section className="casino-profile-modal__activity-group">
                    <div className="casino-profile-modal__activity-group-head">
                      <span>Roulette</span>
                      <small>{rouletteActivity.length}</small>
                    </div>
                    <div className="casino-profile-modal__activity-group-list">
                      {rouletteActivity.map(renderActivityItem)}
                    </div>
                  </section>
                ) : null}

                {secondaryActivity.length > 0 ? (
                  <section className="casino-profile-modal__activity-group casino-profile-modal__activity-group_secondary">
                    <div className="casino-profile-modal__activity-group-head">
                      <span>Secondary / future labels</span>
                      <small>{secondaryActivity.length}</small>
                    </div>
                    <div className="casino-profile-modal__activity-group-list">
                      {secondaryActivity.map(renderActivityItem)}
                    </div>
                  </section>
                ) : null}
              </>
            )}
          </div>
        </section>
      </section>
    </div>,
    document.body,
  )
}
