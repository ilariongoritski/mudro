"use client";

import { useEffect, useState } from "react";
import { useSlot } from "@/lib/slot/store";
import { getHistory } from "@/lib/casino-api";

interface Stats {
  totalSpins: number;
  totalBet: number;
  totalWin: number;
  winRate: number;
  biggestWin: number;
}

export function StatisticsPanel() {
  const [stats, setStats] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(false);
  const isLoggedIn = useSlot((s) => s.isLoggedIn);

  const loadStats = async () => {
    if (!isLoggedIn) return;
    setLoading(true);
    try {
      const history = await getHistory(100);
      
      const totalSpins = history.length;
      const totalBet = history.reduce((sum, s) => sum + s.bet, 0);
      const totalWin = history.reduce((sum, s) => sum + s.win, 0);
      const winRate = totalSpins > 0 ? (history.filter(s => s.win > 0).length / totalSpins) * 100 : 0;
      const biggestWin = history.length > 0 ? Math.max(...history.map(s => s.win)) : 0;

      setStats({ totalSpins, totalBet, totalWin, winRate, biggestWin });
    } catch (e) {
      console.error("Failed to load stats");
    }
    setLoading(false);
  };

  useEffect(() => {
    if (isLoggedIn) loadStats();
  }, [isLoggedIn]);

  if (!isLoggedIn || !stats) return null;

  return (
    <div className="w-full max-w-2xl mx-auto mt-8 px-4">
      <h3 className="text-lg font-bold tracking-widest mb-4">Your Statistics</h3>
      
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        <div className="rounded-2xl border border-white/10 bg-zinc-950/50 p-4">
          <div className="text-xs text-slate-400">Total Spins</div>
          <div className="text-2xl font-mono font-bold">{stats.totalSpins}</div>
        </div>
        <div className="rounded-2xl border border-white/10 bg-zinc-950/50 p-4">
          <div className="text-xs text-slate-400">Total Bet</div>
          <div className="text-2xl font-mono font-bold text-red-400">{stats.totalBet}</div>
        </div>
        <div className="rounded-2xl border border-white/10 bg-zinc-950/50 p-4">
          <div className="text-xs text-slate-400">Total Win</div>
          <div className="text-2xl font-mono font-bold text-emerald-400">{stats.totalWin}</div>
        </div>
        <div className="rounded-2xl border border-white/10 bg-zinc-950/50 p-4">
          <div className="text-xs text-slate-400">Win Rate</div>
          <div className="text-2xl font-mono font-bold">{stats.winRate.toFixed(1)}%</div>
        </div>
        <div className="rounded-2xl border border-white/10 bg-zinc-950/50 p-4">
          <div className="text-xs text-slate-400">Biggest Win</div>
          <div className="text-2xl font-mono font-bold text-amber-400">{stats.biggestWin}</div>
        </div>
      </div>
    </div>
  );
}
