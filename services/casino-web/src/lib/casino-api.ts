import { useSlot } from "./slot/store";

const CASINO_API = "/api/casino";

type APIError = { error?: string };

export interface SpinResult {
  balance: number;
  win: number;
  symbols: string[];
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
    throw new Error((body as APIError).error || `Casino request failed (${res.status})`);
  }
  return res.json() as Promise<T>;
}

export async function getBalance(): Promise<number> {
  const data = await request<{ balance: number }>("/balance");
  return data.balance;
}

export function spin(bet: number): Promise<SpinResult> {
  return request<SpinResult>("/spin", { method: "POST", body: JSON.stringify({ bet }) });
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
