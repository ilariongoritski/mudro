"use client";

import { create } from "zustand";
import {
  BET_PRESETS,
  BONUS_BUY_MULT,
  DEFAULT_BET,
  FREE_SPINS_AWARD,
  FREE_SPINS_RETRIGGER,
  FREE_SPINS_RETRIGGER_AWARD,
  FREE_SPINS_TRIGGER,
  STARTING_BALANCE,
  cascadeMultiplier,
  computeTier,
  type SymbolId,
  type WinTier,
} from "./config";
import {
  collectBombs,
  countScatters,
  evaluateBoard,
  generateBoard,
  round2,
  tumbleBoard,
  type Board,
  type Cell,
} from "./engine";
import { sound } from "./sound";

const BALANCE_KEY = "slot.balance.v2";

function loadBalance(): number {
  if (typeof window === "undefined") return STARTING_BALANCE;
  try {
    const raw = window.localStorage.getItem(BALANCE_KEY);
    if (raw == null) return STARTING_BALANCE;
    const n = Number(raw);
    if (!Number.isFinite(n) || n <= 0) return STARTING_BALANCE;
    return round2(n);
  } catch {
    return STARTING_BALANCE;
  }
}

/** Hydrate balance from localStorage on the client only (after mount). */
function hydrateBalance(): number | null {
  if (typeof window === "undefined") return null;
  const n = loadBalance();
  return n === STARTING_BALANCE ? null : n;
}

function saveBalance(n: number) {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(BALANCE_KEY, String(round2(n)));
  } catch {
    /* ignore */
  }
}

function emptyBoard(): Board {
  const b: Board = [];
  for (let r = 0; r < 5; r++) b.push([] as Cell[]);
  return b;
}

type Phase = "idle" | "dropping" | "celebrating" | "tumbling" | "ended";
type Timer = ReturnType<typeof setTimeout> | null;

interface User {
  id: number;
  username: string;
  telegram_id?: number;
}

export interface FairnessProof {
  serverSeedHash: string;
  nonce: number;
}

export interface ServerSpinResult {
  balanceAfter: number;
  win: number;
  symbols: string[];
  serverSeedHash: string;
  nonce: number;
}

export interface SlotState {
  balance: number;
  bet: number;
  board: Board;
  phase: Phase;
  spinKey: number; // bumps each new spin to retrigger enter animations
  tumbleKey: number; // bumps each tumble step

  winningPositions: Set<string>;
  cascade: number; // 1-based current cascade
  cascadeMult: number; // current cascade multiplier
  lastCascadeWin: number; // win of the most recent cascade step
  spinWin: number; // accumulated this spin (pre bomb-mult)
  displayWin: number; // shown total (post bomb-mult at end)
  winTier: WinTier;
  lastWins: { symbol: SymbolId; count: number; amount: number }[];

  scatterCount: number; // accumulated this spin
  bombsTotal: number; // accumulated bomb mult this spin (free spins)
  activeBombs: number; // bombs currently on board (for display)

  freeSpins: number;
  freeSpinsTotal: number;
  freeSpinsWin: number;
  inFreeSpins: boolean;
  showFreeSpinsBanner: boolean;
  showFreeSpinsEnd: boolean;
  freeSpinsEndWin: number;

  autoSpins: number;
  autoInfinite: boolean;
  turbo: boolean;
  soundOn: boolean;

  /** Bumps each time balance increases, to drive the scale-up animation. */
  balancePulse: number;
  token: string | null;
  user: User | null;
  isLoggedIn: boolean;
  fairness: FairnessProof | null;

  _timer: Timer;

  // actions
  spin: () => void;
  commitDrop: () => void;
  doTumble: () => void;
  endSequence: () => void;
  setBet: (b: number) => void;
  incBet: () => void;
  decBet: () => void;
  setAuto: (n: number) => void;
  setAutoInfinite: () => void;
  stopAuto: () => void;
  toggleTurbo: () => void;
  toggleSound: () => void;
  resetBalance: () => void;
  clearBanner: () => void;
  seedBoard: () => void;
  buyBonus: () => void;
  hydrate: () => void;
  setAuth: (token: string, user: User) => void;
  clearAuth: () => void;
  applyServerSpin: (result: ServerSpinResult) => void;
  beginServerSpin: () => boolean;
  failServerSpin: () => void;
}

function clearTimer(s: SlotState) {
  if (s._timer) {
    clearTimeout(s._timer);
    s._timer = null;
  }
}

function delayMs(base: number, turbo: boolean) {
  return turbo ? Math.max(180, base * 0.5) : base;
}

export const useSlot = create<SlotState>((set, get) => ({
  balance: STARTING_BALANCE,
  bet: DEFAULT_BET,
  board: emptyBoard(),
  phase: "idle",
  spinKey: 0,
  tumbleKey: 0,

  winningPositions: new Set(),
  cascade: 0,
  cascadeMult: 1,
  lastCascadeWin: 0,
  spinWin: 0,
  displayWin: 0,
  winTier: "none",
  lastWins: [],

  scatterCount: 0,
  bombsTotal: 0,
  activeBombs: 0,

  freeSpins: 0,
  freeSpinsTotal: 0,
  freeSpinsWin: 0,
  inFreeSpins: false,
  showFreeSpinsBanner: false,
  showFreeSpinsEnd: false,
  freeSpinsEndWin: 0,

  autoSpins: 0,
  autoInfinite: false,
  turbo: false,
  soundOn: true,

  balancePulse: 0,
  token: null,
  user: null,
  isLoggedIn: false,
  fairness: null,

  _timer: null,

  spin: () => {
    const s = get();
    if (s.phase !== "idle" && s.phase !== "ended") return;
    clearTimer(s);

    const isFree = s.freeSpins > 0;
    if (!isFree) {
      if (s.balance < s.bet) return;
      const nb = round2(s.balance - s.bet);
      set({ balance: nb });
      saveBalance(nb);
    } else {
      set({ freeSpins: s.freeSpins - 1 });
    }

    const board = generateBoard(isFree);
    const scatters = countScatters(board);
    const bombs = collectBombs(board);

    if (s.soundOn) sound.spinStart();

    set({
      board,
      phase: "dropping",
      spinKey: s.spinKey + 1,
      tumbleKey: s.tumbleKey + 1,
      winningPositions: new Set(),
      cascade: 0,
      cascadeMult: 1,
      lastCascadeWin: 0,
      spinWin: 0,
      displayWin: 0,
      winTier: "none",
      lastWins: [],
      scatterCount: scatters,
      bombsTotal: bombs.total,
      activeBombs: bombs.total,
      showFreeSpinsBanner: false,
      showFreeSpinsEnd: false,
    });

    const t = setTimeout(
      () => get().commitDrop(),
      delayMs(820, s.turbo)
    );
    set({ _timer: t });
  },

  commitDrop: () => {
    const s = get();
    const evalRes = evaluateBoard(s.board, s.bet);

    if (evalRes.totalWin <= 0) {
      get().endSequence();
      return;
    }

    const cascade = s.cascade + 1;
    const mult = cascadeMultiplier(cascade);
    const stepWin = round2(evalRes.totalWin * mult);
    const spinWin = round2(s.spinWin + stepWin);
    const tier = computeTier(spinWin, s.bet);

    if (s.soundOn) {
      if (cascade >= 3) sound.winBig();
      else sound.winSmall();
      if (tier === "mega" || tier === "epic") sound.coinDrop();
    }

    set({
      phase: "celebrating",
      winningPositions: evalRes.winningPositions,
      cascade,
      cascadeMult: mult,
      lastCascadeWin: stepWin,
      spinWin,
      displayWin: spinWin,
      winTier: tier,
      lastWins: evalRes.wins.map((w) => ({
        symbol: w.symbol,
        count: w.count,
        amount: round2(w.amount * mult),
      })),
    });

    const t = setTimeout(
      () => get().doTumble(),
      delayMs(720, s.turbo)
    );
    set({ _timer: t });
  },

  doTumble: () => {
    const s = get();
    if (s.soundOn) sound.tumblePop();

    // Scatters from tumbles are intentionally ignored for trigger/retrigger
    // (only the initial board's scatters count) — prevents free-spin loops.
    const { board: newBoard } = tumbleBoard(
      s.board,
      s.winningPositions,
      s.inFreeSpins
    );
    const bombs = collectBombs(newBoard);

    set({
      board: newBoard,
      phase: "tumbling",
      tumbleKey: s.tumbleKey + 1,
      winningPositions: new Set(),
      // NOTE: scatters from tumbles do NOT accumulate toward the trigger/
      // retrigger — only the initial board's scatters count. This prevents
      // free spins from looping forever when scatters appear in cascades.
      scatterCount: s.scatterCount,
      activeBombs: bombs.total,
    });

    const t = setTimeout(
      () => get().commitDrop(),
      delayMs(620, s.turbo)
    );
    set({ _timer: t });
  },

  endSequence: () => {
    const s = get();
    // Apply bomb multiplier (free spins only): bombs on the final board sum
    // and multiply the accumulated spin win.
    let finalWin = s.spinWin;
    let bombMult = 0;
    if (s.inFreeSpins) {
      const bombs = collectBombs(s.board);
      bombMult = bombs.total;
      if (bombMult > 0) {
        finalWin = round2(s.spinWin * bombMult);
        if (s.soundOn) sound.bomb();
      }
    }

    const tier = computeTier(finalWin, s.bet);
    const newBalance = round2(s.balance + finalWin);
    if (finalWin > 0) saveBalance(newBalance);

    set({
      phase: "ended",
      displayWin: finalWin,
      winTier: tier,
      winningPositions: new Set(),
      cascadeMult: bombMult > 0 ? bombMult : s.cascadeMult,
      balance: newBalance,
      // pulse the balance whenever it grows
      balancePulse: finalWin > 0 ? s.balancePulse + 1 : s.balancePulse,
    });

    if (finalWin > 0 && s.soundOn) {
      if (tier === "epic" || tier === "mega") sound.jackpot();
      else if (tier === "big") sound.winBig();
    }

    // Free-spins win accumulation
    if (s.inFreeSpins && finalWin > 0) {
      set({ freeSpinsWin: round2(s.freeSpinsWin + finalWin) });
    }

    // Free spins trigger / retrigger
    const trigger = s.inFreeSpins
      ? FREE_SPINS_RETRIGGER
      : FREE_SPINS_TRIGGER;
    const award = s.inFreeSpins
      ? FREE_SPINS_RETRIGGER_AWARD
      : FREE_SPINS_AWARD;
    if (s.scatterCount >= trigger) {
      if (!s.inFreeSpins) {
        set({
          freeSpins: award,
          freeSpinsTotal: award,
          freeSpinsWin: 0,
          inFreeSpins: true,
          showFreeSpinsBanner: true,
        });
      } else {
        set({
          freeSpins: s.freeSpins + award,
          freeSpinsTotal: s.freeSpinsTotal + award,
        });
      }
      if (s.soundOn) sound.freeSpinsTrigger();
    }

    // Schedule next
    const cur = get();
    let delay = 650;
    if (cur.showFreeSpinsBanner) delay = 2400;
    else if (tier === "epic") delay = 2800;
    else if (tier === "mega") delay = 2400;
    else if (tier === "big") delay = 1900;
    else if (tier === "normal") delay = 1100;
    if (cur.turbo) delay = Math.max(380, delay * 0.55);

    // Free spins continue automatically
    if (cur.freeSpins > 0) {
      const t = setTimeout(() => {
        set({ phase: "ended" });
        get().spin();
      }, delay);
      set({ _timer: t });
      return;
    }

    // Free spins just ended
    if (cur.inFreeSpins) {
      set({
        inFreeSpins: false,
        showFreeSpinsEnd: true,
        freeSpinsEndWin: cur.freeSpinsWin,
        freeSpinsTotal: 0,
        freeSpinsWin: 0,
      });
      const t = setTimeout(() => {
        set({ showFreeSpinsEnd: false, phase: "ended" });
        const c = get();
        if (
          (c.autoSpins > 0 || c.autoInfinite) &&
          c.balance >= c.bet
        ) {
          if (!c.autoInfinite)
            set({ autoSpins: Math.max(0, c.autoSpins - 1) });
          get().spin();
        }
      }, 2800);
      set({ _timer: t });
      return;
    }

    // Base-game auto-spin
    if (cur.autoSpins > 0 || cur.autoInfinite) {
      if (cur.balance < cur.bet) {
        set({ autoSpins: 0, autoInfinite: false, phase: "idle" });
        return;
      }
      let remaining = cur.autoSpins;
      if (!cur.autoInfinite) {
        remaining = Math.max(0, cur.autoSpins - 1);
        set({ autoSpins: remaining });
      }
      if (cur.autoInfinite || remaining > 0) {
        const t = setTimeout(() => {
          set({ phase: "idle" });
          get().spin();
        }, delay);
        set({ _timer: t });
      } else {
        set({ phase: "idle" });
      }
    } else {
      set({ phase: "idle" });
    }
  },

  setBet: (b) => {
    const s = get();
    if (s.phase !== "idle" && s.phase !== "ended") return;
    if (s.inFreeSpins || s.autoSpins > 0 || s.autoInfinite) return;
    if (!BET_PRESETS.includes(b)) return;
    if (s.soundOn) sound.tick();
    set({ bet: b });
  },
  incBet: () => {
    const s = get();
    if (s.phase !== "idle" && s.phase !== "ended") return;
    if (s.inFreeSpins || s.autoSpins > 0 || s.autoInfinite) return;
    const i = BET_PRESETS.indexOf(s.bet);
    const next = BET_PRESETS[Math.min(BET_PRESETS.length - 1, i + 1)];
    if (next !== s.bet) {
      if (s.soundOn) sound.tick();
      set({ bet: next });
    }
  },
  decBet: () => {
    const s = get();
    if (s.phase !== "idle" && s.phase !== "ended") return;
    if (s.inFreeSpins || s.autoSpins > 0 || s.autoInfinite) return;
    const i = BET_PRESETS.indexOf(s.bet);
    const next = BET_PRESETS[Math.max(0, i - 1)];
    if (next !== s.bet) {
      if (s.soundOn) sound.tick();
      set({ bet: next });
    }
  },

  setAuto: (n) => {
    const s = get();
    if (s.inFreeSpins) return;
    if (s.balance < s.bet) return;
    if (s.phase !== "idle" && s.phase !== "ended") {
      // queue: just set the count, current spin continues
      set({ autoSpins: n, autoInfinite: false });
      return;
    }
    set({ autoSpins: n, autoInfinite: false });
    get().spin();
  },
  setAutoInfinite: () => {
    const s = get();
    if (s.inFreeSpins) return;
    if (s.balance < s.bet) return;
    if (s.phase !== "idle" && s.phase !== "ended") {
      set({ autoInfinite: true, autoSpins: 0 });
      return;
    }
    set({ autoInfinite: true, autoSpins: 0 });
    get().spin();
  },
  stopAuto: () => {
    const s = get();
    clearTimer(s);
    set({ autoSpins: 0, autoInfinite: false });
  },

  toggleTurbo: () => set({ turbo: !get().turbo }),
  toggleSound: () => {
    const on = !get().soundOn;
    sound.setMuted(!on);
    set({ soundOn: on });
  },

  resetBalance: () => {
    const s = get();
    if (s.phase !== "idle" && s.phase !== "ended") return;
    clearTimer(s);
    set({
      balance: STARTING_BALANCE,
      autoSpins: 0,
      autoInfinite: false,
      inFreeSpins: false,
      freeSpins: 0,
      freeSpinsTotal: 0,
      freeSpinsWin: 0,
      showFreeSpinsBanner: false,
      showFreeSpinsEnd: false,
      displayWin: 0,
      winTier: "none",
      spinWin: 0,
      cascade: 0,
      phase: "idle",
      board: emptyBoard(),
    });
    saveBalance(STARTING_BALANCE);
  },

  clearBanner: () => set({ showFreeSpinsBanner: false }),

  seedBoard: () => {
    const s = get();
    if (s.phase !== "idle" && s.phase !== "ended") return;
    if (s.board.some((col) => col.length > 0)) return;
    set({ board: generateBoard(false), spinKey: 0, tumbleKey: 0 });
  },

  buyBonus: () => {
    const s = get();
    if (s.phase !== "idle" && s.phase !== "ended") return;
    if (s.inFreeSpins) return;
    const price = round2(BONUS_BUY_MULT * s.bet);
    if (s.balance < price) return;
    clearTimer(s);

    const nb = round2(s.balance - price);
    set({ balance: nb });
    saveBalance(nb);

    const board = generateBoard(true); // free-spins board (bombs can spawn)
    const scatters = countScatters(board);
    const bombs = collectBombs(board);

    if (s.soundOn) sound.freeSpinsTrigger();

    set({
      board,
      phase: "dropping",
      spinKey: s.spinKey + 1,
      tumbleKey: s.tumbleKey + 1,
      winningPositions: new Set(),
      cascade: 0,
      cascadeMult: 1,
      lastCascadeWin: 0,
      spinWin: 0,
      displayWin: 0,
      winTier: "none",
      lastWins: [],
      scatterCount: scatters,
      bombsTotal: bombs.total,
      activeBombs: bombs.total,
      showFreeSpinsBanner: true,
      showFreeSpinsEnd: false,
      freeSpins: FREE_SPINS_AWARD,
      freeSpinsTotal: FREE_SPINS_AWARD,
      freeSpinsWin: 0,
      inFreeSpins: true,
    });

    const t = setTimeout(() => get().commitDrop(), delayMs(1400, s.turbo));
    set({ _timer: t });
  },

  hydrate: () => {
    const n = hydrateBalance();
    if (n != null && n !== get().balance) set({ balance: n });
    if (typeof window === "undefined") return;
    try {
      const token = window.localStorage.getItem("mudro_token");
      const rawUser = window.localStorage.getItem("mudro_user");
      if (token && rawUser) set({ token, user: JSON.parse(rawUser), isLoggedIn: true });
    } catch {
      /* Invalid persisted auth is ignored. */
    }
  },

  setAuth: (token, user) => {
    if (typeof window !== "undefined") {
      window.localStorage.setItem("mudro_token", token);
      window.localStorage.setItem("mudro_user", JSON.stringify(user));
    }
    set({ token, user, isLoggedIn: true });
  },

  clearAuth: () => {
    if (typeof window !== "undefined") {
      window.localStorage.removeItem("mudro_token");
      window.localStorage.removeItem("mudro_user");
    }
    set({ token: null, user: null, isLoggedIn: false, fairness: null });
  },

  beginServerSpin: () => {
    const s = get();
    if (!s.isLoggedIn || (s.phase !== "idle" && s.phase !== "ended") || s.balance < s.bet) return false;
    set({ phase: "dropping" });
    return true;
  },

  failServerSpin: () => {
    set({ phase: "ended" });
  },

  applyServerSpin: (result) => {
    const s = get();
    clearTimer(s);
    const winTier = computeTier(result.win, s.bet);
    set({
      balance: round2(result.balanceAfter),
      phase: "ended",
      spinKey: s.spinKey + 1,
      tumbleKey: s.tumbleKey + 1,
      displayWin: round2(result.win),
      spinWin: round2(result.win),
      lastCascadeWin: round2(result.win),
      winTier,
      fairness: { serverSeedHash: result.serverSeedHash, nonce: result.nonce },
      winningPositions: new Set(),
      cascade: 0,
      activeBombs: 0,
      balancePulse: result.win > 0 ? s.balancePulse + 1 : s.balancePulse,
    });
  },
}));
