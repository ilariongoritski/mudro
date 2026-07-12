"use client";

import { useEffect, useState } from "react";
import { useSlot } from "@/lib/slot/store";
import { getHistory, SpinHistoryItem } from "@/lib/casino-api";
import { Button } from "@/components/ui/button";

export function HistoryPanel() {
  const [history, setHistory] = useState<SpinHistoryItem[]>([]);
  const [loading, setLoading] = useState(false);
  const isLoggedIn = useSlot((s) => s.isLoggedIn);

  const loadHistory = async () => {
    if (!isLoggedIn) return;
    setLoading(true);
    try {
      const data = await getHistory(20);
      setHistory(data);
    } catch (e) {
      console.error("Failed to load history");
    }
    setLoading(false);
  };

  useEffect(() => {
    if (isLoggedIn) {
      loadHistory();
    }
  }, [isLoggedIn]);

  if (!isLoggedIn) return null;

  return (
    <div className="w-full max-w-2xl mx-auto mt-8 px-4">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-bold tracking-widest">Recent Spins</h3>
        <Button
          variant="ghost"
          size="sm"
          onClick={loadHistory}
          disabled={loading}
        >
          {loading ? "Loading..." : "Refresh"}
        </Button>
      </div>

      <div className="rounded-2xl border border-white/10 bg-zinc-950/50 overflow-hidden">
        {history.length === 0 ? (
          <div className="p-8 text-center text-slate-400 text-sm">
            No spins yet. Make your first spin!
          </div>
        ) : (
          <div className="divide-y divide-white/10 text-sm">
            {history.slice(0, 10).map((spin, index) => (
              <div key={index} className="flex items-center justify-between px-4 py-3 hover:bg-white/5">
                <div className="flex items-center gap-4">
                  <div className="font-mono text-emerald-400">#{spin.id}</div>
                  <div className="text-slate-400">
                    {new Date(spin.createdAt).toLocaleDateString()} {new Date(spin.createdAt).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
                  </div>
                </div>
                <div className="flex items-center gap-4 font-mono">
                  <div className="text-red-400">-{spin.bet}</div>
                  <div className={spin.win > 0 ? "text-emerald-400" : "text-slate-500"}>
                    {spin.win > 0 ? `+${spin.win}` : "0"}
                  </div>
                  <div className="text-xs text-slate-500 w-16 text-right">
                    {spin.symbols?.slice(0, 3).join(" ")}...
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
