import type {
  CasinoActivityItem as CasinoApiActivityItem,
  CasinoHistoryItem,
  CasinoProfileResponse,
} from '@/features/casino/api/casinoApi'
import type { User } from '@/entities/session/model/sessionSlice'

export type CasinoActivityStatus = 'WIN' | 'LOST' | 'CASHOUT'
export type CasinoFairnessStatus = 'verified' | 'pending' | 'unknown'

export interface CasinoActivityItem {
  id: string | number
  gameType: 'Slots' | 'Roulette' | 'Plinko' | 'Crash' | 'Coinflip' | 'Mines' | 'Bonus' | string
  title: string
  bet: number
  amount: number
  status: CasinoActivityStatus
  createdAt: string
  details?: string
  fairnessStatus?: CasinoFairnessStatus
  markers?: string[]
}

export interface CasinoProfileSummary {
  userId: number
  username: string
  displayName: string
  balance: number
  currency: string
  totalWagered: number
  totalWon: number
  gamesPlayed: number
  slotsRoundsPlayed: number
  rouletteRoundsPlayed: number
  otherGamesPlayed: number
  level: number
  currentLevelProgress: number
  nextLevelTarget: number
  rtp?: number
  avatarUrl?: string | null
  lastGameAt?: string | null
}

export interface BuildCasinoProfileInput {
  profile?: CasinoProfileResponse | null
  user?: User | null
  balance: number
  currency?: string
  rtp?: number
  history?: CasinoHistoryItem[]
  activity?: CasinoActivityItem[]
}

const LEVEL_STEP = 1000

const normalizeStatus = (bet: number, win: number): CasinoActivityStatus => {
  const net = win - bet
  if (net > 0) return 'WIN'
  if (net === 0 && win > 0) return 'CASHOUT'
  return 'LOST'
}

const normalizeStatusFromNet = (netResult: number, payoutAmount = 0): CasinoActivityStatus => {
  if (netResult > 0) return 'WIN'
  if (netResult === 0 && payoutAmount > 0) return 'CASHOUT'
  return 'LOST'
}

const resolveGameLabel = (value: string) => {
  const normalized = value.trim().toLowerCase()

  switch (normalized) {
    case 'slots':
      return 'Slots'
    case 'roulette':
      return 'Roulette'
    case 'plinko':
      return 'Plinko'
    case 'crash':
      return 'Crash'
    case 'coinflip':
      return 'Coinflip'
    case 'mines':
      return 'Mines'
    case 'bonus':
      return 'Bonus'
    default:
      return value || 'Game'
  }
}

const resolveGameBucket = (value: string) => {
  const normalized = value.trim().toLowerCase()

  if (normalized === 'slots') return 'slots'
  if (normalized === 'roulette') return 'roulette'
  return 'other'
}

export const formatCasinoDateTime = (value: string) => {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) {
    return 'Только что'
  }

  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  }).format(parsed)
}

export const buildCasinoActivityFromHistory = (history: CasinoHistoryItem[] = []): CasinoActivityItem[] =>
  history.map((item) => {
    const amount = item.win - item.bet

    return {
      id: item.id,
      gameType: 'Slots',
      title: 'Slot spin',
      bet: item.bet,
      amount,
      status: normalizeStatus(item.bet, item.win),
      createdAt: item.created_at,
      details: item.symbols.join(' · '),
    }
  })

export const buildCasinoActivityFromApi = (items: CasinoApiActivityItem[] = []): CasinoActivityItem[] =>
  items.map((item) => {
    const metadata = item.metadata ?? {}
    const fairnessStatus: CasinoFairnessStatus =
      typeof metadata.fairness === 'string'
        ? (metadata.fairness as CasinoFairnessStatus)
        : typeof metadata.proof === 'string' || typeof metadata.fair_proof === 'string'
          ? 'verified'
          : 'unknown'

    return {
      id: item.id,
      gameType: resolveGameLabel(item.game_type),
      title: item.game_ref ?? resolveGameLabel(item.game_type),
      bet: item.bet_amount,
      amount: item.net_result,
      status: normalizeStatusFromNet(item.net_result, item.payout_amount),
      createdAt: item.created_at,
      details:
        Object.entries(metadata)
          .filter(([key]) => !['fairness', 'proof', 'fair_proof'].includes(key))
          .map(([key, value]) => `${key}: ${String(value)}`)
          .join(' · ') || undefined,
      fairnessStatus,
      markers: [resolveGameLabel(item.game_type), fairnessStatus, item.net_result >= 0 ? 'positive' : 'negative'],
    }
  })

export const buildCasinoProfileSummary = ({
  profile,
  user,
  balance,
  currency = 'credits',
  rtp,
  history = [],
  activity,
}: BuildCasinoProfileInput): CasinoProfileSummary => {
  const resolvedActivity = activity ?? buildCasinoActivityFromHistory(history)
  const totalWagered = profile?.total_wagered ?? history.reduce((sum, item) => sum + item.bet, 0)
  const totalWon = profile?.total_won ?? history.reduce((sum, item) => sum + item.win, 0)
  const gamesPlayed = resolvedActivity.length
  const slotsRoundsPlayed = resolvedActivity.filter((item) => resolveGameBucket(item.gameType) === 'slots').length
  const rouletteRoundsPlayed = resolvedActivity.filter((item) => resolveGameBucket(item.gameType) === 'roulette').length
  const otherGamesPlayed = Math.max(gamesPlayed - slotsRoundsPlayed - rouletteRoundsPlayed, 0)
  const level = profile?.level ?? Math.floor(totalWagered / LEVEL_STEP) + 1
  const nextLevelTarget = profile?.next_level_xp ?? profile?.progress_target ?? level * LEVEL_STEP
  const currentLevelProgress = profile ? Math.min((profile.xp_progress ?? 0) / LEVEL_STEP, 1) : Math.min(totalWagered / LEVEL_STEP, 1)

  return {
    userId: profile?.user_id ?? user?.id ?? 0,
    username: profile?.username ?? user?.username ?? 'guest',
    displayName: profile?.display_name ?? user?.username ?? 'Игрок',
    balance: profile?.balance ?? balance,
    currency: currency ?? 'credits',
    totalWagered,
    totalWon,
    gamesPlayed: profile?.games_played ?? gamesPlayed,
    slotsRoundsPlayed,
    rouletteRoundsPlayed: profile?.roulette_rounds_played ?? rouletteRoundsPlayed,
    otherGamesPlayed,
    level,
    currentLevelProgress,
    nextLevelTarget,
    rtp,
    avatarUrl: profile?.avatar_url ?? null,
    lastGameAt: profile?.last_game_at ?? null,
  }
}
