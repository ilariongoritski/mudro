import { create } from "zustand";
import { persist } from "zustand/middleware";

interface User {
  id: number;
  username: string;
  telegram_id?: number;
}

interface SlotState {
  // Game state
  phase: "idle" | "spinning" | "ended";
  balance: number;
  bet: number;
  lastWin: number;
  lastSymbols: string[];
  history: any[];
  inFreeSpins: boolean;
  freeSpins: number;
  freeSpinsTotal: number;
  freeSpinsWin: number;
  winTier: string;
  cascade: number;
  activeBombs: number;
  spinKey: number;
  balancePulse: boolean;
  soundOn: boolean;

  // Auth
  token: string | null;
  user: User | null;
  isLoggedIn: boolean;

  // Actions
  setBalance: (b: number) => void;
  setBet: (b: number) => void;
  setLastWin: (w: number) => void;
  setLastSymbols: (s: string[]) => void;
  setHistory: (h: any[]) => void;
  setAuth: (token: string, user: User) => void;
  clearAuth: () => void;
  spin: () => void;
  toggleSound: () => void;
  resetBalance: () => void;
  hydrate: () => void;
}

export const useSlot = create<SlotState>()(
  persist(
    (set, get) => ({
      phase: "idle",
      balance: 1000,
      bet: 10,
      lastWin: 0,
      lastSymbols: [],
      history: [],
      inFreeSpins: false,
      freeSpins: 0,
      freeSpinsTotal: 0,
      freeSpinsWin: 0,
      winTier: "",
      cascade: 0,
      activeBombs: 0,
      spinKey: 0,
      balancePulse: false,
      soundOn: true,

      token: null,
      user: null,
      isLoggedIn: false,

      setBalance: (b) => set({ balance: b }),
      setBet: (b) => set({ bet: b }),
      setLastWin: (w) => set({ lastWin: w }),
      setLastSymbols: (s) => set({ lastSymbols: s }),
      setHistory: (h) => set({ history: h }),

      setAuth: (token, user) => {
        localStorage.setItem("mudro_token", token);
        localStorage.setItem("mudro_user", JSON.stringify(user));
        set({ token, user, isLoggedIn: true });
      },

      clearAuth: () => {
        localStorage.removeItem("mudro_token");
        localStorage.removeItem("mudro_user");
        set({ token: null, user: null, isLoggedIn: false });
      },

      spin: () => {
        const state = get();
        if (!state.isLoggedIn) return;

        // TODO: call real API when integrated
        const newBalance = Math.max(0, state.balance - state.bet);
        set({
          phase: "spinning",
          balance: newBalance,
          spinKey: state.spinKey + 1,
          lastWin: 0,
        });

        setTimeout(() => {
          set({ phase: "ended" });
        }, 1200);
      },

      toggleSound: () => set((s) => ({ soundOn: !s.soundOn })),
      resetBalance: () => set({ balance: 1000 }),

      hydrate: () => {
        const token = localStorage.getItem("mudro_token");
        const userStr = localStorage.getItem("mudro_user");
        if (token && userStr) {
          try {
            const user = JSON.parse(userStr);
            set({ token, user, isLoggedIn: true });
          } catch {}
        }
      },
    }),
    {
      name: "mudro-slot-storage",
      partialize: (state) => ({
        balance: state.balance,
        bet: state.bet,
        soundOn: state.soundOn,
      }),
    }
  )
);
