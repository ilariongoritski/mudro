"use client";

import { useSlot } from "@/lib/slot/store";
import { FairnessProof } from "../FairnessProof";

export function WinDisplay() {
  const displayWin = useSlot((s) => s.displayWin);
  const fairness = useSlot((s) => s.fairness);

  if (displayWin <= 0) return null;

  return (
    <div className="absolute inset-x-0 top-4 z-30 flex justify-center">
      <div className="rounded-2xl bg-emerald-500/90 px-8 py-3 text-center shadow-2xl">
        <div className="text-xs font-bold tracking-[3px] text-emerald-950">WIN</div>
        <div className="font-mono text-4xl font-black text-white tabular-nums">
          {displayWin.toFixed(2)}
        </div>
      </div>

      <FairnessProof 
        serverSeedHash={fairness?.serverSeedHash} 
        nonce={fairness?.nonce} 
      />
    </div>
  );
}
