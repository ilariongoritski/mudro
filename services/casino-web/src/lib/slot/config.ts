"use client";
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
    weight: 3,
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
    pay: { 6: 4, 7: 6, 8: 12, 9: 20, 10: 35, 11: 55, 12: 90 },
  },
  grape: {
    id: "grape",
    label: "GRAPE",
    emoji: "🍇",
    tier: "high",
    color: "#a855f7",
    glow: "#c084fc",
    weight: 6,
    pay: { 6: 2, 7: 5, 8: 8, 9: 15, 10: 25, 11: 40, 12: 70 },
  },
  watermelon: {
    id: "watermelon",
    label: "MELON",
    emoji: "🍉",
    tier: "high",
    color: "#22c55e",
    glow: "#86efac",
    weight: 7,
    pay: { 7: 1, 8: 3, 9: 6, 10: 12, 11: 20, 12: 35 },
  },
  apple: {
    id: "apple",
    label: "APPLE",
    emoji: "🍎",
    tier: "mid",
    color: "#dc2626",
    glow: "#f87171",
    weight: 8,
    pay: { 7: 1, 8: 2, 9: 5, 10: 8, 11: 15, 12: 25 },
  },
  blueberry: {
    id: "blueberry",
    label: "BERRY",
    emoji: "🫐",
    tier: "mid",
    color: "#3b82f6",
    glow: "#93c5fd",
    weight: 9,
    pay: { 7: 1, 8: 1, 9: 3, 10: 6, 11: 12, 12: 20 },
  },
  orange: {
    id: "orange",
    label: "ORANGE",
    emoji: "🍊",
    tier: "low",
    color: "#f97316",
    glow: "#fdba74",
    weight: 10,
    pay: { 7: 1, 8: 1, 9: 2, 10: 5, 11: 8, 12: 16 },
  },
  pear: {
    id: "pear",
    label: "PEAR",
    emoji: "🍐",
    tier: "low",
    color: "#84cc16",
    glow: "#bef264",
    weight: 11,
    pay: { 7: 1, 8: 1, 9: 1, 10: 4, 11: 6, 12: 12 },
  },
  strawberry: {
    id: "strawberry",
    label: "BERRY",
    emoji: "🍓",
    tier: "low",
    color: "#ec4899",
    glow: "#f9a8d4",
    weight: 12,
    pay: { 7: 1, 8: 1, 9: 1, 10: 3, 11: 5, 12: 10 },
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
export const MIN_PAY_COUNT = 6;

/** Bomb multiplier values that can spawn during free spins. Trimmed for ~96% RTP. */
export const BOMB_VALUES: number[] = [2, 3, 4, 5, 6, 8, 10, 12, 15, 20];
/** Weighted toward small values. */
export const BOMB_WEIGHTS: number[] = [
  30, 25, 20, 15, 10, 8, 6, 4, 3, 2,
];
/** Probability that a newly-filled cell becomes a bomb during free spins. */
export const BOMB_SPAWN_CHANCE = 0.02;

export const BET_PRESETS = [0.2, 0.5, 1, 2, 5, 10, 25, 50, 100];
export const DEFAULT_BET = 1;
export const STARTING_BALANCE = 1000;

export const FREE_SPINS_TRIGGER = 4; // scatters needed
export const FREE_SPINS_AWARD = 10;
export const FREE_SPINS_RETRIGGER = 3; // scatters needed during FS
export const FREE_SPINS_RETRIGGER_AWARD = 5;

/** Bonus Buy: pay this multiple of the bet to instantly trigger free spins. */
export const BONUS_BUY_MULT = 100;

/** Cascade multiplier by 1-based cascade index. */
const CASCADE_TABLE = [1, 1, 1, 2, 3];
export function cascadeMultiplier(cascade1Based: number): number {
  const i = Math.max(0, Math.min(cascade1Based - 1, CASCADE_TABLE.length - 1));
  return CASCADE_TABLE[i] ?? 3;
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
