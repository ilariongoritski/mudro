import type { RouletteBetType, RouletteColor } from '@/features/casino/api/casinoApi'

export const rouletteWheelOrder = [
  0, 32, 15, 19, 4, 21, 2, 25, 17, 34, 6, 27, 13, 36, 11, 30, 8, 23, 10,
  5, 24, 16, 33, 1, 20, 14, 31, 9, 22, 18, 29, 7, 28, 12, 35, 3, 26,
]

export const rouletteBetTypeLabels: Record<RouletteBetType, string> = {
  straight: 'Straight',
  red: 'Red',
  black: 'Black',
  green: 'Green',
  odd: 'Odd',
  even: 'Even',
  low: 'Low',
  high: 'High',
}

export const rouletteColorLabels: Record<RouletteColor, string> = {
  red: 'Red',
  black: 'Black',
  green: 'Green',
  unknown: 'Unknown',
}

const redNumbers = new Set([1, 3, 5, 7, 9, 12, 14, 16, 18, 19, 21, 23, 25, 27, 30, 32, 34, 36])

export const rouletteQuickStakeValues = [10, 25, 50, 100, 250]

export const rouletteStakeTypeOptions: Array<{ value: RouletteBetType; label: string; description: string }> = [
  { value: 'straight', label: 'Straight', description: 'Number 0-36' },
  { value: 'red', label: 'Red', description: 'Pays 2x' },
  { value: 'black', label: 'Black', description: 'Pays 2x' },
  { value: 'green', label: 'Green', description: 'Pays 14x' },
  { value: 'odd', label: 'Odd', description: 'Pays 2x' },
  { value: 'even', label: 'Even', description: 'Pays 2x' },
  { value: 'low', label: 'Low', description: '1-18' },
  { value: 'high', label: 'High', description: '19-36' },
]

export const formatRouletteClock = (milliseconds: number) => {
  if (!Number.isFinite(milliseconds) || milliseconds <= 0) {
    return '0s'
  }

  const seconds = Math.ceil(milliseconds / 1000)
  const wholeMinutes = Math.floor(seconds / 60)
  const restSeconds = seconds % 60

  if (wholeMinutes <= 0) {
    return `${restSeconds}s`
  }

  return `${wholeMinutes}:${restSeconds.toString().padStart(2, '0')}`
}

export const getRouletteColor = (value: number): RouletteColor => {
  if (value === 0) {
    return 'green'
  }

  return redNumbers.has(value) ? 'red' : 'black'
}

export const formatRouletteNumber = (value: number | null | undefined) => {
  if (typeof value !== 'number' || Number.isNaN(value)) {
    return '—'
  }

  return value.toString().padStart(2, '0')
}

