import { useCallback, useEffect, useMemo, useState } from 'react'

import {
  useClaimCasinoBonusSubscriptionMutation,
  useGetCasinoBonusStateQuery,
} from '@/features/casino/api/casinoApi'

import './BonusPanel.css'

interface BonusMainAction {
  label: string
  busy: boolean
  disabled: boolean
  onTrigger: () => void
}

interface BonusPanelProps {
  isAuthenticated: boolean
  isActive: boolean
  userName?: string | null
  telegramInitData?: string | null
  onMainActionChange?: (action: BonusMainAction | null) => void
}

const demoRewards = ['+10', 'x2', 'Boost', 'Lucky', 'Free', 'VIP', 'Spin', 'Bonus', 'Chest']

export const BonusPanel = ({ isAuthenticated, isActive, userName, telegramInitData, onMainActionChange }: BonusPanelProps) => {
  const [localStatus, setLocalStatus] = useState<string | null>(null)
  const { data: bonusState, isFetching } = useGetCasinoBonusStateQuery(undefined, {
    skip: !isAuthenticated,
    pollingInterval: isAuthenticated ? 15000 : 0,
    refetchOnFocus: true,
  })
  const [claimSubscription, { isLoading: isClaiming }] = useClaimCasinoBonusSubscriptionMutation()

  const subscribed = bonusState?.subscribed ?? false
  const claimReady = bonusState?.claim_ready ?? false
  const freeSpinsAvailable = bonusState?.free_spins_available ?? 0
  const freeSpinsTotal = bonusState?.free_spins_total ?? 10
  const fairStrip = bonusState?.fair_strip ?? 'Provably fair: bonus-wheel-public · nonce 10'
  const bonusHistory = bonusState?.history ?? []

  const claimLabel = useMemo(() => {
    if (isClaiming) return 'Проверяем подписку...'
    if (freeSpinsAvailable > 0) return `Доступно ${freeSpinsAvailable} spins`
    if (claimReady) return 'Забрать +10 free spins'
    if (subscribed) return 'Подписка активна'
    return 'Подписаться и забрать +10'
  }, [claimReady, freeSpinsAvailable, isClaiming, subscribed])

  const onClaim = useCallback(async () => {
    if (!isAuthenticated) {
      setLocalStatus('Нужно открыть bonus внутри Telegram mini app.')
      return
    }
    if (!telegramInitData) {
      setLocalStatus('Для claim нужна живая Telegram WebApp сессия.')
      return
    }
    try {
      await claimSubscription({ initData: telegramInitData }).unwrap()
      setLocalStatus('Bonus обновлён. Если подписка подтверждена, free spins уже зачислены.')
    } catch {
      setLocalStatus('Не удалось подтвердить подписку или получить bonus.')
    }
  }, [claimSubscription, isAuthenticated, telegramInitData])

  useEffect(() => {
    if (!isActive) {
      onMainActionChange?.(null)
      return
    }
    onMainActionChange?.({
      label: claimReady ? 'Claim +10' : 'Bonus',
      busy: isClaiming,
      disabled: !isAuthenticated || isClaiming,
      onTrigger: () => {
        void onClaim()
      },
    })
  }, [claimReady, isActive, isAuthenticated, isClaiming, onClaim, onMainActionChange])

  return (
    <section className="bonus-panel">
      <header className="bonus-panel__header">
        <div>
          <span className="bonus-panel__eyebrow">Bonus</span>
          <h2>Подписка = +10 free spins</h2>
          <p>
            {userName ? `@${userName}` : 'guest'} · {subscribed ? 'канал подтверждён' : 'нужна подписка на канал'}
          </p>
        </div>

        <div className="bonus-panel__counter">
          <span>Free spins</span>
          <strong>{isFetching ? '…' : freeSpinsAvailable}</strong>
          <small>из {freeSpinsTotal}</small>
        </div>
      </header>

      <div className="bonus-panel__layout">
        <section className="bonus-panel__hero">
          <div className="bonus-panel__wheel">
            <div className="bonus-panel__wheel-core">
              <span>+10</span>
              <small>bonus</small>
            </div>
          </div>

          <div className="bonus-panel__claim">
            <div className="bonus-panel__status-chip">{claimReady ? 'Можно забрать' : subscribed ? 'Подписка активна' : 'Нужна подписка'}</div>
            <button type="button" className="bonus-panel__primary" onClick={() => void onClaim()} disabled={!isAuthenticated || isClaiming || !telegramInitData}>
              {claimLabel}
            </button>
            <p>{localStatus ?? bonusState?.verification_message ?? 'Бонусный сценарий ждёт подтверждение подписки и начисляет бесплатные spins через casino API.'}</p>
          </div>
        </section>

        <section className="bonus-panel__grid">
          {demoRewards.map((reward, index) => (
            <article key={`${reward}-${index}`} className="bonus-panel__grid-item">
              <strong>{reward}</strong>
              <span>награда</span>
            </article>
          ))}
        </section>
      </div>

      <div className="bonus-panel__fair-strip">{fairStrip}</div>

      <section className="bonus-panel__history">
        <div className="bonus-panel__history-head">
          <strong>История бонусов</strong>
          <small>{bonusHistory.length}</small>
        </div>
        <div className="bonus-panel__history-list">
          {bonusHistory.length === 0 ? (
            <p className="bonus-panel__empty">История бонусов появится после первых claim/bonus событий.</p>
          ) : (
            bonusHistory.map((item) => (
              <article key={`${item.id}`} className="bonus-panel__history-item">
                <div>
                  <strong>{item.title}</strong>
                  <span>{item.status ?? 'BONUS'}</span>
                </div>
                <div>
                  <strong>{item.amount ?? '+10'}</strong>
                  <small>{item.created_at ?? 'только что'}</small>
                </div>
              </article>
            ))
          )}
        </div>
      </section>
    </section>
  )
}
