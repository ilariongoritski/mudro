"use client";
// Sweet Bonanza engine: pay-anywhere evaluation + tumble (cascade) logic.
import {
  BOMB_SPAWN_CHANCE,
  BOMB_VALUES,
  BOMB_WEIGHTS,
  MIN_PAY_COUNT,
  PAYING_SYMBOLS,
  REELS,
  ROWS,
  SYMBOLS,
  type SymbolId,
} from "./config";

export interface Cell {
  id: number;
  symbol: SymbolId;
  /** Present when symbol === "bomb". */
  mult?: number;
}
/** board[reel][row], row 0 = top, row ROWS-1 = bottom. */
export type Board = Cell[][];
export type Grid = Board;

export type LineWin = {
  symbol: SymbolId;
  count: number;
  positions: [number, number][];
  amount: number;
  lineIndex: number;
};

let _id = 1;
export function nextId(): number {
  return _id++;
}

const weightedPool: SymbolId[] = (() => {
  const pool: SymbolId[] = [];
  for (const id of PAYING_SYMBOLS) {
    const w = Math.max(1, Math.round(SYMBOLS[id].weight * 10));
    for (let i = 0; i < w; i++) pool.push(id);
  }
  // scatter in the same pool
  const sw = Math.max(1, Math.round(SYMBOLS.scatter.weight * 10));
  for (let i = 0; i < sw; i++) pool.push("scatter");
  return pool;
})();

export function randomSymbol(): SymbolId {
  return weightedPool[(Math.random() * weightedPool.length) | 0];
}

function randomBombValue(): number {
  const total = BOMB_WEIGHTS.reduce((a, b) => a + b, 0);
  let r = Math.random() * total;
  for (let i = 0; i < BOMB_VALUES.length; i++) {
    r -= BOMB_WEIGHTS[i];
    if (r <= 0) return BOMB_VALUES[i];
  }
  return BOMB_VALUES[0];
}

/** A new cell: paying symbol or scatter (never bomb unless free spins). */
function newCell(freeSpins: boolean): Cell {
  // In free spins, a fraction of new cells become bombs.
  if (freeSpins && Math.random() < BOMB_SPAWN_CHANCE) {
    return { id: nextId(), symbol: "bomb", mult: randomBombValue() };
  }
  return { id: nextId(), symbol: randomSymbol() };
}

export function generateBoard(freeSpins = false): Board {
  const board: Board = [];
  for (let r = 0; r < REELS; r++) {
    const col: Cell[] = [];
    for (let row = 0; row < ROWS; row++) col.push(newCell(freeSpins));
    board.push(col);
  }
  return board;
}

export interface BoardWin {
  symbol: SymbolId;
  count: number;
  positions: [number, number][]; // [reel, row]
  amount: number; // already × bet
}

export interface EvalResult {
  wins: BoardWin[];
  totalWin: number; // × bet already, before cascade multiplier
  winningPositions: Set<string>; // "reel-row"
  symbolCounts: Record<string, number>;
}

export function evaluateBoard(board: Board, bet: number): EvalResult {
  const counts: Partial<Record<SymbolId, number>> = {};
  const positions: Partial<Record<SymbolId, [number, number][]>> = {};

  for (let r = 0; r < REELS; r++) {
    for (let row = 0; row < ROWS; row++) {
      const c = board[r][row];
      if (c.symbol === "bomb") continue;
      counts[c.symbol] = (counts[c.symbol] ?? 0) + 1;
      (positions[c.symbol] ??= []).push([r, row]);
    }
  }

  const wins: BoardWin[] = [];
  const winningPositions = new Set<string>();
  let totalWin = 0;

  for (const sym of PAYING_SYMBOLS) {
    const count = counts[sym] ?? 0;
    if (count < MIN_PAY_COUNT) continue;
    const def = SYMBOLS[sym];
    // pick the highest bracket reached
    let mult = 0;
    const keys = Object.keys(def.pay)
      .map(Number)
      .sort((a, b) => a - b);
    for (const k of keys) {
      if (count >= k) mult = def.pay[k];
    }
    if (mult <= 0) continue;
    const amount = mult * bet;
    wins.push({ symbol: sym, count, positions: positions[sym]!, amount });
    totalWin += amount;
    for (const [r, row] of positions[sym]!) winningPositions.add(`${r}-${row}`);
  }

  return {
    wins,
    totalWin,
    winningPositions,
    symbolCounts: counts as Record<string, number>,
  };
}

export function countScatters(board: Board): number {
  let n = 0;
  for (let r = 0; r < REELS; r++)
    for (let row = 0; row < ROWS; row++)
      if (board[r][row].symbol === "scatter") n++;
  return n;
}

export function collectBombs(board: Board): { positions: [number, number][]; total: number } {
  const positions: [number, number][] = [];
  let total = 0;
  for (let r = 0; r < REELS; r++)
    for (let row = 0; row < ROWS; row++) {
      const c = board[r][row];
      if (c.symbol === "bomb" && c.mult) {
        positions.push([r, row]);
        total += c.mult;
      }
    }
  return { positions, total };
}

/**
 * Tumble: remove winning cells, apply gravity (survivors fall to bottom of
 * each reel), fill new cells at the top. Scatter & bomb cells are NOT removed
 * by wins (they persist and fall with gravity).
 *
 * Returns the new board plus counts of newly-introduced scatters & bombs.
 */
export function tumbleBoard(
  board: Board,
  winningPositions: Set<string>,
  freeSpins: boolean
): { board: Board; newScatters: number; newBombs: number } {
  let newScatters = 0;
  let newBombs = 0;

  const next: Board = board.map((col, reel) => {
    // survivors keep order (top→bottom), winners removed
    const survivors = col.filter((_, row) => !winningPositions.has(`${reel}-${row}`));
    const fillCount = ROWS - survivors.length;
    const fills: Cell[] = [];
    for (let i = 0; i < fillCount; i++) {
      const c = newCell(freeSpins);
      if (c.symbol === "scatter") newScatters++;
      if (c.symbol === "bomb") newBombs++;
      fills.push(c);
    }
    // new cells on top, survivors below (gravity → survivors sink to bottom)
    return [...fills, ...survivors];
  });

  return { board: next, newScatters, newBombs };
}

/** Anticipation: which reel (index of the LAST reel that hasn't "settled") to
 *  slow down / glow when 3 scatters are already visible and a 4th could still
 *  complete the trigger. With drop mechanics we anticipate the rightmost reel
 *  column when scatters ≥ 3 and < 4. Returns reel index to anticipate, or -1. */
export function anticipationReel(scatterCount: number): number {
  if (scatterCount === 3) return REELS - 1; // highlight last column
  return -1;
}

export function round2(n: number): number {
  return Math.round(n * 100) / 100;
}
