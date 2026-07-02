import { useEffect, useRef, useState } from 'react'

import './SlotMachine.css'

const SYMBOLS: Record<string, { emoji: string; color: string }> = {
  cherry:  { emoji: '🍒', color: '#ff4466' },
  lemon:   { emoji: '🍋', color: '#ffe066' },
  bar:     { emoji: '🍫', color: '#c08040' },
  seven:   { emoji: '7️⃣', color: '#ff4466' },
  diamond: { emoji: '💎', color: '#00e5ff' },
}

const SYMBOL_KEYS = Object.keys(SYMBOLS)

function resolveSymbol(value: string) {
  return SYMBOLS[value] || { emoji: value, color: '#f5c842' }
}

function randomSym(): string {
  return SYMBOL_KEYS[Math.floor(Math.random() * SYMBOL_KEYS.length)]
}

const REEL_COUNT = 5
const VISIBLE_ROWS = 3
const CELL_PX = 72

interface ReelColumnProps {
  resultSym: string
  isSpinning: boolean
  stopDelay: number
  }

function ReelColumn({ resultSym, isSpinning, stopDelay }: ReelColumnProps) {
  const [strip, setStrip] = useState<string[]>(() => {
    const s: string[] = []
    for (let i = 0; i < VISIBLE_ROWS + 2; i++) s.push(i === 1 ? resultSym : randomSym())
    return s
  })
  const [spinning, setSpinning] = useState(false)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const stopRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => { setSpinning(isSpinning) }, [isSpinning])

  useEffect(() => {
    if (!isSpinning) return
    intervalRef.current = setInterval(() => {
      const s: string[] = []
      for (let i = 0; i < VISIBLE_ROWS + 2; i++) s.push(randomSym())
      setStrip(s)
    }, 70)
    return () => { if (intervalRef.current) clearInterval(intervalRef.current) }
  }, [isSpinning])

  useEffect(() => {
    if (isSpinning) return
    stopRef.current = setTimeout(() => {
      if (intervalRef.current) clearInterval(intervalRef.current)
      const s: string[] = []
      for (let i = 0; i < VISIBLE_ROWS + 2; i++) s.push(i === 1 ? resultSym : randomSym())
      setStrip(s)
    }, stopDelay)
    return () => { if (stopRef.current) clearTimeout(stopRef.current) }
  }, [isSpinning, resultSym, stopDelay])

  useEffect(() => {
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
      if (stopRef.current) clearTimeout(stopRef.current)
    }
  }, [])

  const clipHeight = CELL_PX * VISIBLE_ROWS
  const ghostOffset = CELL_PX * 0.5

  return (
    <div className="slot-reel" style={{
      height: clipHeight,
      borderColor: spinning ? 'rgba(245,200,66,0.55)' : 'rgba(255,255,255,0.07)',
      boxShadow: spinning ? 'inset 0 0 18px rgba(245,200,66,0.18)' : 'inset 0 0 12px rgba(0,0,0,0.5)',
    }}>
      <div className="slot-reel__fade slot-reel__fade_top" />
      <div className="slot-reel__fade slot-reel__fade_bottom" />
      {!spinning && <div className="slot-reel__highlight" style={{ top: ghostOffset + CELL_PX, height: CELL_PX }} />}
      <div style={{ transform: `translateY(-${ghostOffset}px)` }}>
        {strip.map((sym, rowIdx) => {
          const info = resolveSymbol(sym)
          const isMiddle = rowIdx === 1
          return (
            <div key={rowIdx} className="slot-cell" style={{
              height: CELL_PX,
              fontSize: isMiddle ? 38 : 30,
              opacity: isMiddle ? 1 : 0.45,
              filter: isMiddle ? 'none' : 'grayscale(0.3)',
            }}>
              <span style={{ color: info.color }}>{info.emoji}</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}

interface SlotMachineProps {
  symbols: string[]
  isSpinning: boolean
  lastWin: number
}

export const SlotMachine = ({ symbols, isSpinning, lastWin }: SlotMachineProps) => {
  const displaySyms = symbols.length >= REEL_COUNT
    ? symbols.slice(0, REEL_COUNT)
    : [...symbols, ...Array(REEL_COUNT - symbols.length).fill('cherry')]

  return (
    <div className="slot-machine">
      <div className="slot-machine__reels">
        {displaySyms.map((sym, idx) => (
          <ReelColumn
            key={idx}
            resultSym={sym}
            isSpinning={isSpinning}
            stopDelay={idx * 200}
            
          />
        ))}
      </div>
      <div className="slot-machine__result">
        {isSpinning ? (
          <div className="slot-machine__dots">
            {[0,1,2].map(i => <span key={i} style={{ animationDelay: `${i*0.15}s` }} />)}
          </div>
        ) : lastWin > 0 ? (
          <div className="slot-machine__win">
            <span className="slot-machine__win-amount">+{lastWin.toLocaleString()} 🪙</span>
            <span className="slot-machine__win-label">{lastWin >= 1000 ? 'JACKPOT! 🎉' : lastWin >= 500 ? 'BIG WIN! 🔥' : 'WIN!'}</span>
          </div>
        ) : symbols.length > 0 ? (
          <span className="slot-machine__no-win">Нет совпадений</span>
        ) : null}
      </div>
    </div>
  )
}
