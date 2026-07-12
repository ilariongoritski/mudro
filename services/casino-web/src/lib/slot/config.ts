// Sweet Bonanza style config: 5×5 grid, pay-anywhere, tumble, bombs, free spins.

export type SymbolId =
  | "scatter" // 🍭 lollipop — triggers free spins
  | "bomb" // 💣 multiplier (free spins only)
  | "heart" // ❤️ premium
  | "grape" // 🍇
  | "watermelon" // 🍉
  | "apple" // 🍎
  | "blueberry" // 🫐
  | "orange" // 🍊
  | "pear" // 🍐
  | "strawberry"; // 🍓

export type SymbolTier = "special" | "premium" | "high" | "mid" | "low";

export interface SymbolDef {
  id: SymbolId;
  label: string;
  emoji: string;
  tier: SymbolTier;
  color: string;
  glow: string;
  /** Reel weight (higher = more common). */
  weight: number;
  /** Pay-anywhere: multiplier of TOTAL bet for count ≥ key. 12 means "12 or more". */
  pay: Record<number, number>;
  isScatter?: boolean;
  isBomb?: boolean;
}

export const SYMBOLS: Record<SymbolId, SymbolDef> = {
  scatter: {
    id: "scatter",
    label: "LOLLIPOP",
    emoji: "🍭",
    tier: "special",
    color: "#f472b6",
    glow: "#f9a8d4",
    weight: 2.4,
    pay: { 4: 2, 5: 3, 6: 5, 7: 10, 8: 20 },
    isScatter: true,
  },
  bomb: {
    id: "bomb",
    label: "MULTIPLIER",
    emoji: "💣",
    tier: "special",
    color: "#fbbf24",
    glow: "#fde047",
    weight: 0,
    pay: {},
    isBomb: true,
  },
  heart: {
    id: "heart",
    label: "HEART",
    emoji: "❤️",
    tier: "premium",
    color: "#ef4444",
    glow: "#fca5a5",
    weight: 5,
    pay: { 5: 0.3, 6: 0.6, 7: 1.2, 8: 2.5, 9: 5, 10: 8, 11: 12, 12: 20 },
  },
  grape: {
    id: "grape",
    label: "GRAPE",
    emoji: "🍇",
    tier: "high",
    color: "#a855f7",
    glow: "#c084fc",
    weight: 6,
    pay: { 5: 0.25, 6: 0.5, 7: 1, 8: 2, 9: 4, 10: 6, 11: 9, 12: 15 },
  },
  watermelon: {
    id: "watermelon",
    label: "MELON",
    emoji: "🍉",
    tier: "high",
    color: "#22c55e",
    glow: "#86efac",
    weight: 7,
    pay: { 5: 0.2, 6: 0.4, 7: 0.8, 8: 1.5, 9: 3, 10: 5, 11: 7, 12: 12 },
  },
  apple: {
    id: "apple",
    label: "APPLE",
    emoji: "🍎",
    tier: "mid",
    color: "#dc2626",
    glow: "#f87171",
    weight: 8,
    pay: { 5: 0.15, 6: 0.3, 7: 0.6, 8: 1.2, 9: 2.5, 10: 4, 11: 6, 12: 10 },
  },
  blueberry: {
    id: "blueberry",
    label: "BERRY",
    emoji: "🫐",
    tier: "mid",
    color: "#3b82f6",
    glow: "#93c5fd",
    weight: 9,
    pay: { 5: 0.12, 6: 0.25, 7: 0.5, 8: 1, 9: 2, 10: 3, 11: 5, 12: 8 },
  },
  orange: {
    id: "orange",
    label: "ORANGE",
    emoji: "🍊",
    tier: "low",
    color: "#f97316",
    glow: "#fdba74",
    weight: 10,
    pay: { 5: 0.1, 6: 0.2, 7: 0.4, 8: 0.8, 9: 1.5, 10: 2.5, 11: 4, 12: 6 },
  },
  pear: {
    id: "pear",
    label: "PEAR",
    emoji: "🍐",
    tier: "low",
    color: "#84cc16",
    glow: "#bef264",
    weight: 11,
    pay: { 5: 0.08, 6: 0.15, 7: 0.3, 8: 0.6, 9: 1.2, 10: 2, 11: 3, 12: 5 },
  },
  strawberry: {
    id: "strawberry",
    label: "BERRY",
    emoji: "🍓",
    tier: "low",
    color: "#ec4899",
    glow: "#f9a8d4",
    weight: 12,
    pay: { 5: 0.06, 6: 0.12, 7: 0.25, 8: 0.5, 9: 1, 10: 1.5, 11: 2.5, 12: 4 },
  },
};

export const SYMBOL_IDS = Object.keys(SYMBOLS) as SymbolId[];

/** Paying symbols only (excludes scatter & bomb). */
export const PAYING_SYMBOLS: SymbolId[] = SYMBOL_IDS.filter(
  (s) => !SYMBOLS[s].isScatter && !SYMBOLS[s].isBomb
);

export const REELS = 5;
export const ROWS = 5;
export const CELLS = REELS * ROWS;

/** Minimum count of matching symbols anywhere to pay. */
export const MIN_PAY_COUNT = 5;

/** Bomb multiplier values that can spawn during free spins. Trimmed for lower RTP. */
export const BOMB_VALUES: number[] = [2, 3, 4, 5, 6, 8, 10, 15, 20, 25, 50];
/** Weighted toward small values. */
export const BOMB_WEIGHTS: number[] = [
  26, 22, 18, 14, 10, 8, 6, 4, 3, 2, 1,
];
/** Probability that a newly-filled cell becomes a bomb during free spins. */
export const BOMB_SPAWN_CHANCE = 0.045;

export const BET_PRESETS = [1, 2, 5, 10, 25, 50, 100];
export const DEFAULT_BET = 1;
export const STARTING_BALANCE = 1000;

export const FREE_SPINS_TRIGGER = 4; // scatters needed
export const FREE_SPINS_AWARD = 10;
export const FREE_SPINS_RETRIGGER = 3; // scatters needed during FS
export const FREE_SPINS_RETRIGGER_AWARD = 5;

/** Bonus Buy: pay this multiple of the bet to instantly trigger free spins. */
export const BONUS_BUY_MULT = 100;

/** Cascade multiplier by 1-based cascade index. Flattened to lower RTP. */
const CASCADE_TABLE = [0, 1, 1, 1, 2, 2, 3, 3, 4, 5];
export function cascadeMultiplier(cascade1Based: number): number {
  const i = Math.max(0, Math.min(cascade1Based, CASCADE_TABLE.length - 1));
  return CASCADE_TABLE[i] ?? 5;
}

export type WinTier = "none" | "normal" | "big" | "mega" | "epic";

export function computeTier(win: number, bet: number): WinTier {
  if (win <= 0) return "none";
  const r = win / bet;
  if (r >= 40) return "epic";
  if (r >= 12) return "mega";
  if (r >= 4) return "big";
  return "normal";
}

export const TIER_LABEL: Record<WinTier, string> = {
  none: "",
  normal: "WIN",
  big: "BIG WIN",
  mega: "MEGA WIN",
  epic: "EPIC WIN",
};
