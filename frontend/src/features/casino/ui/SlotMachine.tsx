import { useEffect, useRef, useState } from 'react'

// ─── Symbol definitions ───────────────────────────────────────────────────────
const SYMBOLS: Record<string, { emoji: string; label: string; color: string }> = {
  cherry:  { emoji: '🍒', label: 'Cherry',  color: '#ff4466' },
  lemon:   { emoji: '🍋', label: 'Lemon',   color: '#ffe066' },
  orange:  { emoji: '🍊', label: 'Orange',  color: '#ff8c00' },
  grape:   { emoji: '🍇', label: 'Grape',   color: '#b94cf5' },
  bell:    { emoji: '🔔', label: 'Bell',    color: '#f5c842' },
  star:    { emoji: '⭐', label: 'Star',    color: '#f5c842' },
  diamond: { emoji: '💎', label: 'Diamond', color: '#00e5ff' },
  seven:   { emoji: '7️⃣',  label: 'Seven',  color: '#ff4466' },
}

const SYMBOL_KEYS = Object.keys(SYMBOLS)

// Resolve a value that may be a symbol key or a raw emoji
function resolveSymbol(value: string) {
  if (SYMBOLS[value]) return SYMBOLS[value]
  // Fallback: treat value as raw emoji
  return { emoji: value, label: '', color: '#f5c842' }
}

function randomSym(): string {
  return SYMBOL_KEYS[Math.floor(Math.random() * SYMBOL_KEYS.length)]
}

const REEL_SIZE = 5   // 1 ghost + 3 visible + 1 ghost
const CELL_PX   = 80

// ─── Single reel column ───────────────────────────────────────────────────────
interface ReelColumnProps {
  middleSym: string  // key or emoji for middle row
  isSpinning: boolean
  stopDelay: number
}

function ReelColumn({ middleSym, isSpinning, stopDelay }: ReelColumnProps) {
  const [spinning, setSpinning] = useState(false)

  const [strip, setStrip] = useState<string[]>(() => [
    randomSym(), randomSym(), middleSym, randomSym(), randomSym(),
  ])

  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const stopRef     = useRef<ReturnType<typeof setTimeout>  | null>(null)

  useEffect(() => {
    if (isSpinning) {
      setSpinning(true)
      intervalRef.current = setInterval(() => {
        setStrip([randomSym(), randomSym(), randomSym(), randomSym(), randomSym()])
      }, 75)
    } else {
      stopRef.current = setTimeout(() => {
        if (intervalRef.current) clearInterval(intervalRef.current)
        setSpinning(false)
        setStrip([randomSym(), randomSym(), middleSym, randomSym(), randomSym()])
      }, stopDelay)
    }

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
      if (stopRef.current)     clearTimeout(stopRef.current)
    }
  }, [isSpinning, middleSym, stopDelay])

  const clipHeight   = CELL_PX * 4
  const ghostOffset  = CELL_PX * 0.5

  return (
    <div
      style={{
        flex: 1,
        position: 'relative',
        overflow: 'hidden',
        height: clipHeight,
        borderRadius: 14,
        border: spinning
          ? '1.5px solid rgba(245,200,66,0.55)'
          : '1.5px solid rgba(255,255,255,0.07)',
        background: 'linear-gradient(180deg, #0d0d1a 0%, #12121f 100%)',
        boxShadow: spinning
          ? 'inset 0 0 18px rgba(245,200,66,0.18)'
          : 'inset 0 0 12px rgba(0,0,0,0.5)',
        transition: 'border-color 0.3s, box-shadow 0.3s',
      }}
    >
      {/* Fade top */}
      <div
        style={{
          position: 'absolute', top: 0, left: 0, right: 0, zIndex: 20,
          height: CELL_PX * 0.6, pointerEvents: 'none',
          background: 'linear-gradient(180deg, rgba(13,13,26,0.92) 0%, transparent 100%)',
        }}
      />
      {/* Fade bottom */}
      <div
        style={{
          position: 'absolute', bottom: 0, left: 0, right: 0, zIndex: 20,
          height: CELL_PX * 0.6, pointerEvents: 'none',
          background: 'linear-gradient(0deg, rgba(13,13,26,0.92) 0%, transparent 100%)',
        }}
      />

      {/* Middle-row highlight */}
      {!spinning && (
        <div
          style={{
            position: 'absolute', left: 0, right: 0, zIndex: 10, pointerEvents: 'none',
            top: ghostOffset + CELL_PX, height: CELL_PX,
            background: 'rgba(245,200,66,0.04)',
            borderTop: '1px solid rgba(245,200,66,0.15)',
            borderBottom: '1px solid rgba(245,200,66,0.15)',
          }}
        />
      )}

      {/* Symbol strip */}
      <div style={{ transform: `translateY(-${ghostOffset}px)` }}>
        {strip.map((sym, rowIdx) => {
          const info    = resolveSymbol(sym)
          const isMiddle = rowIdx === 2
          const isGhost  = rowIdx === 0 || rowIdx === REEL_SIZE - 1
          return (
            <div
              key={rowIdx}
              style={{
                height: CELL_PX,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                userSelect: 'none',
                opacity: isGhost ? 0.3 : isMiddle ? 1 : 0.7,
                filter: spinning
                  ? 'blur(2.5px) brightness(1.2)'
                  : isGhost ? 'blur(0.8px)' : 'none',
                transition: spinning ? 'none' : 'filter 0.2s',
              }}
            >
              <div
                style={{
                  fontSize: 38, lineHeight: 1,
                  transform: isMiddle && !spinning ? 'scale(1.08)' : 'scale(1)',
                  textShadow: isMiddle && !spinning ? `0 0 14px ${info.color}99` : 'none',
                  transition: 'transform 0.2s, text-shadow 0.2s',
                }}
              >
                {info.emoji}
              </div>
              {isMiddle && !spinning && info.label && (
                <div
                  style={{
                    fontSize: 8, fontFamily: 'monospace', textTransform: 'uppercase',
                    letterSpacing: '0.25em', opacity: 0.5, marginTop: 2, color: info.color,
                  }}
                >
                  {info.label}
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}

// ─── SlotMachine — 3 columns ──────────────────────────────────────────────────
interface SlotMachineProps {
  symbols: string[]   // [col0, col1, col2] middle-row landing symbols (keys or emojis)
  isSpinning: boolean
  lastWin: number
}

export function SlotMachine({ symbols, isSpinning, lastWin }: SlotMachineProps) {
  const midSyms = symbols.length === 3 ? symbols : ['cherry', 'cherry', 'cherry']
  const isWin   = lastWin > 0

  return (
    <div style={{ width: '100%' }}>
      <div
        style={{
          position: 'relative',
          borderRadius: 24,
          padding: 16,
          background: 'linear-gradient(145deg, #1a1a2e 0%, #0f0f23 100%)',
          border: isWin
            ? '2px solid rgba(0,255,136,0.5)'
            : '2px solid rgba(245,200,66,0.18)',
          boxShadow: isWin
            ? '0 0 40px rgba(0,255,136,0.18), 0 4px 32px rgba(0,0,0,0.7)'
            : '0 0 18px rgba(245,200,66,0.07), 0 4px 32px rgba(0,0,0,0.7)',
          transition: 'border-color 0.4s, box-shadow 0.4s',
        }}
      >
        {/* Title bar */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: 12, gap: 8 }}>
          <div style={{ height: 1, flex: 1, background: 'linear-gradient(90deg, transparent, rgba(245,200,66,0.35))' }} />
          <span style={{
            fontSize: 11, fontFamily: "'Exo 2', sans-serif", fontWeight: 700,
            letterSpacing: '0.25em', textTransform: 'uppercase', color: '#f5c842',
          }}>
            MUDRO SLOTS
          </span>
          <div style={{ height: 1, flex: 1, background: 'linear-gradient(90deg, rgba(245,200,66,0.35), transparent)' }} />
        </div>

        {/* 3 reel columns */}
        <div style={{ display: 'flex', gap: 8 }}>
          {[0, 1, 2].map((col) => (
            <ReelColumn
              key={col}
              middleSym={midSyms[col]}
              isSpinning={isSpinning}
              stopDelay={col * 160}
            />
          ))}
        </div>

        {/* Payline label */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 8, padding: '0 4px' }}>
          <div style={{ height: 1, flex: 1, opacity: 0.15, background: '#f5c842' }} />
          <span style={{ fontSize: 9, opacity: 0.25, textTransform: 'uppercase', letterSpacing: '0.3em' }}>payline</span>
          <div style={{ height: 1, flex: 1, opacity: 0.15, background: '#f5c842' }} />
        </div>

        {/* Result area */}
        <div style={{ marginTop: 12, textAlign: 'center', minHeight: 44, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          {!isSpinning && lastWin > 0 && (
            <div className="casino-bounce-in" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
              <div style={{
                fontSize: 24, fontFamily: "'Exo 2', sans-serif", fontWeight: 900, letterSpacing: '0.05em',
                color: '#00ff88',
                textShadow: '0 0 20px rgba(0,255,136,0.7), 0 0 40px rgba(0,255,136,0.4)',
              }}>
                +{lastWin.toLocaleString()} 🪙
              </div>
              <div style={{ fontSize: 10, textTransform: 'uppercase', letterSpacing: '0.3em', opacity: 0.7, color: '#00ff88' }}>
                {lastWin >= 1000 ? 'JACKPOT! 🎉' : lastWin >= 500 ? 'BIG WIN! 🔥' : 'WIN!'}
              </div>
            </div>
          )}
          {!isSpinning && lastWin === 0 && symbols.length > 0 && (
            <div style={{ fontSize: 12, opacity: 0.3, letterSpacing: '0.1em', color: '#ff4466' }}>
              Нет совпадений
            </div>
          )}
          {isSpinning && (
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              {[0, 1, 2].map((i) => (
                <div
                  key={i}
                  style={{
                    width: 8, height: 8, borderRadius: '50%', background: '#f5c842',
                    animation: `casinoBounce 0.6s ease-in-out ${i * 0.15}s infinite`,
                  }}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
