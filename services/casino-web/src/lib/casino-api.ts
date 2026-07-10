// services/casino-web/src/lib/casino-api.ts
// Minimal adapter to casino-api (shared wallet, spin, history)

const CASINO_API = process.env.NEXT_PUBLIC_CASINO_API_URL || 'http://localhost:8082';
const AUTH_API = process.env.NEXT_PUBLIC_AUTH_API_URL || 'http://localhost:8080';

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

export async function getBalance(token: string): Promise<number> {
  const res = await fetch(`${CASINO_API}/wallet/balance`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error('Failed to fetch balance');
  const data = await res.json();
  return data.balance ?? 0;
}

export async function spin(bet: number, token: string): Promise<SpinResult> {
  const res = await fetch(`${CASINO_API}/spin`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({ bet, game: 'slot' }),
  });
  if (!res.ok) {
    const err = await res.text();
    throw new Error(err || 'Spin failed');
  }
  return res.json();
}

export async function getHistory(limit = 20, token: string): Promise<SpinHistoryItem[]> {
  const res = await fetch(`${CASINO_API}/history?limit=${limit}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error('Failed to fetch history');
  return res.json();
}
