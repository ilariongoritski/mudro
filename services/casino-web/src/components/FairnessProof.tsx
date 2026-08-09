"use client";

import { useSlot } from "@/lib/slot/store";
import { Button } from "@/components/ui/button";

export function FairnessProof({ 
  serverSeedHash, 
  nonce 
}: { 
  serverSeedHash?: string; 
  nonce?: number;
}) {
  const lastWins = useSlot((s) => s.lastWins);

  if (!serverSeedHash) return null;

  return (
    <div className="mt-4 rounded-2xl border border-white/10 bg-zinc-950/50 p-4 text-xs">
      <div className="font-bold mb-2 flex items-center gap-2">
        <span>🔐</span> Fair Play Proof
      </div>
      <div className="space-y-1 font-mono text-slate-400">
        <div>Server Seed Hash: <span className="text-emerald-400 break-all">{serverSeedHash}</span></div>
        <div>Nonce: <span className="text-emerald-400">{nonce}</span></div>
        <div>Winning Symbols: <span className="text-emerald-400">{lastWins.map(w => w.symbol).join(", ") || "none"}</span></div>
      </div>
      <p className="mt-2 text-[10px] text-slate-500">
        This proves the outcome was generated fairly before you placed your bet.
      </p>
    </div>
  );
}
