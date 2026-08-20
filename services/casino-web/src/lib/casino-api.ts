import { useSlot } from "./slot/store";

const CASINO_API = "/api/v1/casino";

type APIError = { error?: string };

export interface SweetBonanzaCell {
  id: number;
  symbol: string;
  mult?: number;
}
export interface SweetBonanzaStep {
  board: SweetBonanzaCell[][];
  winning_positions: { reel: number; row: number }[];
  cascade: number;
  multiplier: number;
  win: number;
}
export interface SweetBonanzaResult {
  initial_board: SweetBonanzaCell[][];
  steps: SweetBonanzaStep[];
  final_board: SweetBonanzaCell[][];
  scatter_count: number;
  bomb_multiplier?: number;
  free_spins_awarded?: number;
  total_win: number;
}

export interface SpinResult {
  balance: number;
  win: number;
  symbols: string[];
  sweet_bonanza?: SweetBonanzaResult;
  free_spins_balance?: number;
  free_spin_used?: boolean;
  serverSeedHash?: string;
  nonce?: number;
}

export interface SpinHistoryItem {
  id: number;
  bet: number;
  win: number;
  symbols: string[];
  createdAt: string;
}

type SpinHistoryWireItem = Omit<SpinHistoryItem, "createdAt"> & { created_at?: string; createdAt?: string };

export type PlinkoRisk = "low" | "medium" | "high";
export interface PlinkoResult {
  balance: number;
  payout: number;
  multiplier: number;
  path: number[];
  risk: PlinkoRisk;
}
export interface RouletteResult {
  winning_number: number;
  winning_color: string;
  payout_amount: number;
  balance: number;
}
export interface BlackjackCard { suit: string; rank: string; value: number }
export interface BlackjackState {
  id: number;
  bet: number;
  player_hand: { cards: BlackjackCard[]; score: number; is_bust: boolean };
  dealer_hand: { cards: BlackjackCard[]; score: number; is_bust: boolean };
  status: "player_turn" | "dealer_turn" | "resolved";
  winner?: "player" | "dealer" | "push";
  payout: number;
}

function token(): string {
  const value = useSlot.getState().token;
  if (!value) throw new Error("Not authenticated");
  return value;
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Authorization", `Bearer ${token()}`);
  if (init.body) headers.set("Content-Type", "application/json");
  const res = await fetch(`${CASINO_API}${path}`, { ...init, headers });
  if (!res.ok) {
    const body = await res.json().catch(() => ({} as APIError));
    if (res.status === 401 || res.status === 403) {
      useSlot.getState().clearAuth();
      throw new Error("Session expired. Please reopen the Mini App from Telegram.");
    }
    throw new Error((body as APIError).error || `Casino request failed (${res.status})`);
  }
  return res.json() as Promise<T>;
}

export async function getBalance(): Promise<number> {
  const data = await request<{ balance: number }>("/balance");
  return data.balance;
}

export async function spin(bet: number): Promise<SpinResult> {
  const result = await request<Partial<SpinResult>>("/spin", { method: "POST", body: JSON.stringify({ bet }) });
  return {
    balance: Number(result.balance ?? 0),
    win: Number(result.win ?? 0),
    symbols: Array.isArray(result.symbols) ? result.symbols : [],
    sweet_bonanza: result.sweet_bonanza,
    free_spins_balance: result.free_spins_balance,
    free_spin_used: result.free_spin_used,
    serverSeedHash: result.serverSeedHash,
    nonce: result.nonce,
  };
}

export interface BonusBuyResult {
  balance: number;
  free_spins_balance?: number;
}

export async function buyBonus(bet: number): Promise<BonusBuyResult> {
  const result = await request<Partial<BonusBuyResult>>("/bonus/buy", { method: "POST", body: JSON.stringify({ bet }) });
  return {
    balance: Number(result.balance ?? 0),
    free_spins_balance: result.free_spins_balance,
  };
}

export async function getHistory(limit = 20): Promise<SpinHistoryItem[]> {
  const data = await request<{ items: SpinHistoryWireItem[] }>(`/history?limit=${limit}`);
  return data.items.map((item) => ({ ...item, createdAt: item.createdAt ?? item.created_at ?? new Date(0).toISOString() }));
}

export function dropPlinko(bet: number, risk: PlinkoRisk): Promise<PlinkoResult> {
  return request<PlinkoResult>("/plinko/drop", { method: "POST", body: JSON.stringify({ bet, risk }) });
}

export function spinRoulette(bet: number, betType: "red" | "black" | "odd" | "even" = "red"): Promise<RouletteResult> {
  return request<RouletteResult>("/roulette/instant-spin", {
    method: "POST",
    body: JSON.stringify({ bets: [{ bet_type: betType, stake: bet }] }),
  });
}

export function startBlackjack(bet: number): Promise<BlackjackState> {
  return request<BlackjackState>("/blackjack/start", { method: "POST", body: JSON.stringify({ bet }) });
}

export function blackjackAction(action: "hit" | "stand"): Promise<BlackjackState> {
  return request<BlackjackState>("/blackjack/action", { method: "POST", body: JSON.stringify({ action }) });
}
