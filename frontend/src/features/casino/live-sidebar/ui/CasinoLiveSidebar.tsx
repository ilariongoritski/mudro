import { memo } from 'react'
import {
  useAddCasinoReactionMutation,
  useGetCasinoLiveFeedQuery,
  useGetCasinoReactionsQuery,
  useGetCasinoTopWinsQuery,
} from '@/features/casino/api/casinoApi'
import { EmojiPickerTrigger } from '@/features/casino/reactions/ui/EmojiPicker'

import './CasinoLiveSidebar.css'

interface CasinoLiveSidebarProps {
  isAuthenticated: boolean
  activeGame: 'slots' | 'roulette' | 'bonus' | 'plinko' | 'crash' | 'coinflip'
  bonusAvailable: number
}

const formatCompactNumber = (value: number) => new Intl.NumberFormat('ru-RU').format(value)

const formatTime = (value: string) =>
  new Intl.DateTimeFormat('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date(value))

const buildInitials = (label: string) =>
  label
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')

export const CasinoLiveSidebar = memo(({ isAuthenticated, activeGame, bonusAvailable }: CasinoLiveSidebarProps) => {
  const { data: liveFeedData } = useGetCasinoLiveFeedQuery(8, {
    skip: !isAuthenticated,
    pollingInterval: isAuthenticated ? 5000 : 0,
    refetchOnFocus: true,
  })
  const { data: topWinsData } = useGetCasinoTopWinsQuery(4, {
    skip: !isAuthenticated,
    pollingInterval: isAuthenticated ? 10000 : 0,
    refetchOnFocus: true,
  })
  const { data: reactionsData } = useGetCasinoReactionsQuery(6, {
    skip: !isAuthenticated,
    pollingInterval: isAuthenticated ? 6000 : 0,
    refetchOnFocus: true,
  })
  const [addCasinoReaction] = useAddCasinoReactionMutation()

  const liveFeed = liveFeedData?.items ?? []
  const topWins = topWinsData?.items ?? []
  const reactions = reactionsData?.items ?? []
  const targetActivityID = typeof liveFeed[0]?.id === 'number' ? liveFeed[0].id : Number(liveFeed[0]?.id ?? 0)
  const fairStrip =
    activeGame === 'roulette'
      ? 'Provably fair: roulette-public · client public · nonce live'
      : activeGame === 'bonus'
        ? 'Provably fair: bonus-wheel-public · subscription nonce'
        : 'Provably fair: slots-public · server seed · nonce active'

  const onEmojiSelect = async (emoji: string) => {
    if (!targetActivityID) {
      return
    }
    try {
      await addCasinoReaction({
        activity_id: targetActivityID,
        emoji,
      }).unwrap()
    } catch {
      // keep sidebar non-blocking
    }
  }

  return (
    <aside className="casino-live-sidebar">
      <section className="casino-live-sidebar__card">
        <div className="casino-live-sidebar__head">
          <div>
            <span className="casino-live-sidebar__eyebrow">Live</span>
            <h3>Лента ставок</h3>
          </div>
          <span className="casino-live-sidebar__live-pill">LIVE</span>
        </div>

        <div className="casino-live-sidebar__list">
          {liveFeed.length === 0 ? (
            <p className="casino-live-sidebar__empty">Лента оживёт после первых ставок.</p>
          ) : (
            liveFeed.map((item) => {
              const playerLabel = item.player.display_name || item.player.username
              const positive = item.net_result >= 0

              return (
                <article key={`${item.id}-${item.created_at}`} className="casino-live-sidebar__row">
                  {item.player.avatar_url ? (
                    <img className="casino-live-sidebar__avatar casino-live-sidebar__avatar_image" src={item.player.avatar_url} alt="" />
                  ) : (
                    <div className="casino-live-sidebar__avatar" aria-hidden="true">
                      {buildInitials(playerLabel)}
                    </div>
                  )}
                  <div className="casino-live-sidebar__copy">
                    <strong>{playerLabel}</strong>
                    <span>
                      {item.game_type} · {item.event_type}
                    </span>
                  </div>
                  <div className="casino-live-sidebar__meta">
                    <strong className={positive ? 'casino-live-sidebar__amount casino-live-sidebar__amount_win' : 'casino-live-sidebar__amount casino-live-sidebar__amount_loss'}>
                      {positive ? `+${formatCompactNumber(item.net_result)}` : formatCompactNumber(item.net_result)}
                    </strong>
                    <small>{formatTime(item.created_at)}</small>
                  </div>
                </article>
              )
            })
          )}
        </div>
      </section>

      <section className="casino-live-sidebar__card">
        <div className="casino-live-sidebar__head">
          <div>
            <span className="casino-live-sidebar__eyebrow">Top Wins</span>
            <h3>Крупные заносы</h3>
          </div>
        </div>

        <div className="casino-live-sidebar__stack">
          {topWins.length === 0 ? (
            <p className="casino-live-sidebar__empty">Здесь появятся лучшие выигрыши суток.</p>
          ) : (
            topWins.map((item) => (
              <article key={`top-${item.id}`} className="casino-live-sidebar__mini">
                <div>
                  <strong>{item.player.display_name || item.player.username}</strong>
                  <span>{item.game_type}</span>
                </div>
                <div className="casino-live-sidebar__meta casino-live-sidebar__meta_compact">
                  <strong className="casino-live-sidebar__amount casino-live-sidebar__amount_win">
                    +{formatCompactNumber(item.net_result)}
                  </strong>
                  <small>{formatTime(item.created_at)}</small>
                </div>
              </article>
            ))
          )}
        </div>
      </section>

      <section className="casino-live-sidebar__card">
        <div className="casino-live-sidebar__head">
          <div>
            <span className="casino-live-sidebar__eyebrow">Reactions</span>
            <h3>Эмодзи</h3>
          </div>
          <EmojiPickerTrigger onSelect={onEmojiSelect} />
        </div>

        <div className="casino-live-sidebar__reaction-actions" aria-label="Быстрые реакции">
          {['🔥', '😂', '😱', '💎', '👏', '❤️'].map((emoji) => (
            <button key={emoji} type="button" className="casino-live-sidebar__reaction-chip" onClick={() => void onEmojiSelect(emoji)}>
              {emoji}
            </button>
          ))}
        </div>

        <div className="casino-live-sidebar__stack">
          {reactions.length === 0 ? (
            <p className="casino-live-sidebar__empty">Реакции полетят в активную ленту после первых событий.</p>
          ) : (
            reactions.map((item) => (
              <article key={`${item.activity_id}-${item.emoji}`} className="casino-live-sidebar__mini casino-live-sidebar__mini_reaction">
                <div>
                  <strong>
                    {item.emoji} {item.count}
                  </strong>
                  <span>{item.player.display_name || item.player.username} · {item.game_type}</span>
                </div>
                <small>{formatTime(item.latest_at)}</small>
              </article>
            ))
          )}
        </div>
      </section>

      <section className="casino-live-sidebar__card">
        <div className="casino-live-sidebar__head">
          <div>
            <span className="casino-live-sidebar__eyebrow">Fairness</span>
            <h3>Provably fair</h3>
          </div>
          {bonusAvailable > 0 ? <span className="casino-live-sidebar__live-pill">BONUS +{bonusAvailable}</span> : null}
        </div>

        <div className="casino-live-sidebar__fair-strip">{fairStrip}</div>
        <div className="casino-live-sidebar__fair-grid">
          <article className="casino-live-sidebar__fair-card">
            <span>Game</span>
            <strong>{activeGame}</strong>
          </article>
          <article className="casino-live-sidebar__fair-card">
            <span>Mode</span>
            <strong>{isAuthenticated ? 'Linked' : 'Guest'}</strong>
          </article>
          <article className="casino-live-sidebar__fair-card">
            <span>UI</span>
            <strong>Live shell</strong>
          </article>
        </div>
      </section>
    </aside>
  )
}
