// services/casino-web/src/lib/casino-api.ts
// Real casino-api adapter — calls through nginx /api/casino/ proxy

import { useSlot } from "./slot/store";

const CASINO_API = "/api/casino";

export interface SpinResult {
  id: number;
  bet: number;
  win: number;
  symbols: string[];
  serverSeedHash: string;
  nonce: number;
  balanceAfter: number;
}

export interface SpinHistoryItem {
  id: number;
  bet: number;
  win: number;
  symbols: string[];
  createdAt: string;
}

function getToken(): string | null {
  return useSlot.getState().token;
}

export async function getBalance(): Promise<number> {
  const token = getToken();
  if (!token) throw new Error("Not authenticated");

  const res = await fetch(`${CASINO_API}/wallet/balance`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error("Failed to fetch balance");
  const data = await res.json();
  return data.balance ?? 0;
}

export async function spin(bet: number): Promise<SpinResult> {
  const token = getToken();
  if (!token) throw new Error("Not authenticated");

  const res = await fetch(`${CASINO_API}/spin`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({ bet, game: "slot" }),
  });
  if (!res.ok) {
    const err = await res.text();
    throw new Error(err || "Spin failed");
  }
  return res.json();
}

export async function getHistory(limit = 20): Promise<SpinHistoryItem[]> {
  const token = getToken();
  if (!token) throw new Error("Not authenticated");

  const res = await fetch(`${CASINO_API}/history?limit=${limit}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error("Failed to fetch history");
  return res.json();
}