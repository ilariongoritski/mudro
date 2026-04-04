export type PlinkoRisk = 'low' | 'medium' | 'high'

export const plinkoRiskOptions: Array<{ value: PlinkoRisk; label: string; description: string }> = [
  { value: 'low', label: 'Low risk', description: 'Safer center-heavy drops' },
  { value: 'medium', label: 'Medium risk', description: 'Balanced payout curve' },
  { value: 'high', label: 'High risk', description: 'Edge-weighted volatility' },
]

export const plinkoRowsOptions = [8, 10, 12]
export const plinkoQuickStakeValues = [10, 25, 50, 100, 250]

export interface PlinkoRound {
  id: string
  risk: PlinkoRisk
  rows: number
  stake: number
  path: number[]
  bucket: number
  multiplier: number
  win: number
  createdAt: string
}

const buildWeights = (bucketCount: number, risk: PlinkoRisk) => {
  const center = (bucketCount - 1) / 2

  return Array.from({ length: bucketCount }, (_, index) => {
    const distance = Math.abs(index - center) / Math.max(center, 1)

    if (risk === 'low') {
      return 1.2 / (1 + distance * 4)
    }

    if (risk === 'high') {
      return 0.6 + Math.pow(distance, 1.7) * 5.2
    }

    return 0.8 + Math.pow(distance, 1.2) * 1.6
  })
}

export const buildPlinkoMultiplierTable = (rows: number, risk: PlinkoRisk) => {
  const bucketCount = rows + 1
  const center = (bucketCount - 1) / 2

  return Array.from({ length: bucketCount }, (_, index) => {
    const distance = Math.abs(index - center) / Math.max(center, 1)

    if (risk === 'low') {
      return Number((0.8 + distance * 1.6).toFixed(2))
    }

    if (risk === 'high') {
      return Number((0.2 + Math.pow(distance, 1.45) * 14.5).toFixed(2))
    }

    return Number((0.5 + Math.pow(distance, 1.25) * 5.6).toFixed(2))
  })
}

const chooseBucket = (bucketCount: number, risk: PlinkoRisk) => {
  const weights = buildWeights(bucketCount, risk)
  const total = weights.reduce((sum, value) => sum + value, 0)
  let cursor = Math.random() * total

  for (let index = 0; index < weights.length; index += 1) {
    cursor -= weights[index]
    if (cursor <= 0) {
      return index
    }
  }

  return Math.floor(bucketCount / 2)
}

const shuffleMoves = (targetBucket: number, rows: number) => {
  const moves = Array.from({ length: rows }, (_, index) => (index < targetBucket ? 1 : 0))

  for (let index = moves.length - 1; index > 0; index -= 1) {
    const swapIndex = Math.floor(Math.random() * (index + 1))
    const temp = moves[index]
    moves[index] = moves[swapIndex]
    moves[swapIndex] = temp
  }

  return moves
}

export const generatePlinkoRound = ({
  rows,
  risk,
  stake,
}: {
  rows: number
  risk: PlinkoRisk
  stake: number
}): PlinkoRound => {
  const bucketCount = rows + 1
  const bucket = chooseBucket(bucketCount, risk)
  const moves = shuffleMoves(bucket, rows)
  const path = [0]

  let lane = 0
  for (const move of moves) {
    lane += move
    path.push(lane)
  }

  const multiplierTable = buildPlinkoMultiplierTable(rows, risk)
  const multiplier = multiplierTable[bucket] ?? 1
  const win = Number((stake * multiplier).toFixed(2))

  return {
    id: `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`,
    risk,
    rows,
    stake,
    path,
    bucket,
    multiplier,
    win,
    createdAt: new Date().toISOString(),
  }
}

