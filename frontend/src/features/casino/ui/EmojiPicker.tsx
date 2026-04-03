import { useEffect, useRef, useState } from 'react'

import './EmojiPicker.css'

// ─── Emoji Data ────────────────────────────────────────
const CATEGORIES: Record<string, { label: string; emojis: string[] }> = {
  slots: {
    label: '🎰',
    emojis: ['🎰', '🍒', '🍋', '🍊', '🍇', '🔔', '⭐', '💎', '7️⃣', '🃏', '🀄', '🎲', '🎯', '🎳', '🏆', '💰'],
  },
  animals: {
    label: '🐾',
    emojis: ['🐯', '🦁', '🐻', '🐼', '🦊', '🐺', '🦝', '🐶', '🐱', '🐸', '🐙', '🦋', '🐉', '🦄', '🐢', '🦈'],
  },
  symbols: {
    label: '✨',
    emojis: ['🔥', '💥', '⚡', '❄️', '🌊', '💫', '✨', '🌟', '💯', '❤️', '💜', '💛', '🤑', '🎉', '🥳', '🚀'],
  },
  faces: {
    label: '😊',
    emojis: ['😎', '🤩', '😏', '🤑', '😤', '🥶', '😈', '👹', '🤯', '😂', '🥹', '😤', '😑', '😬', '🤫', '😤'],
  },
}

// ─── Types ─────────────────────────────────────────────
interface EmojiPickerProps {
  onSelect: (emoji: string) => void
}

// ─── Component ────────────────────────────────────────
export const EmojiPickerTrigger = ({ onSelect }: EmojiPickerProps) => {
  const [isOpen, setIsOpen] = useState(false)
  const [activeCategory, setActiveCategory] = useState<keyof typeof CATEGORIES>('slots')
  const containerRef = useRef<HTMLDivElement>(null)

  // Close on outside click
  useEffect(() => {
    if (!isOpen) return
    const handleClick = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [isOpen])

  const handleSelect = (emoji: string) => {
    onSelect(emoji)
    setIsOpen(false)
  }

  return (
    <div className="emoji-picker-trigger" ref={containerRef}>
      <button
        type="button"
        className={`emoji-picker-trigger__btn${isOpen ? ' emoji-picker-trigger__btn--open' : ''}`}
        onClick={() => setIsOpen((v) => !v)}
        aria-label="Открыть выбор эмодзи"
        title="Эмодзи и стикеры"
      >
        😊
      </button>

      {isOpen ? (
        <div className="emoji-picker" role="dialog" aria-label="Выбор эмодзи">
          {/* Tabs */}
          <div className="emoji-picker__tabs">
            {Object.entries(CATEGORIES).map(([key, { label }]) => (
              <button
                key={key}
                type="button"
                className={`emoji-picker__tab${activeCategory === key ? ' emoji-picker__tab--active' : ''}`}
                onClick={() => setActiveCategory(key)}
                aria-label={key}
                title={key}
              >
                {label}
              </button>
            ))}
          </div>

          {/* Grid */}
          <div className="emoji-picker__grid">
            {CATEGORIES[activeCategory].emojis.map((emoji) => (
              <button
                key={emoji}
                type="button"
                className="emoji-picker__item"
                onClick={() => handleSelect(emoji)}
                aria-label={emoji}
              >
                {emoji}
              </button>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  )
}
