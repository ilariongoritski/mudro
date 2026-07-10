"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Minus, Plus, RotateCw } from "lucide-react";
import { useSlot } from "@/lib/slot/store";
import { spin as realSpin } from "@/lib/casino-api";
import { cn } from "@/lib/utils";

function fmt(n: number) {
  return n.toLocaleString("en-US", {
    minimumFractionDigits: n < 10 ? 2 : 0,
    maximumFractionDigits: 2,
  });
}

export function ControlPanel() {
  const bet = useSlot((s) => s.bet);
  const balance = useSlot((s) => s.balance);
  const phase = useSlot((s) => s.phase);
  const isLoggedIn = useSlot((s) => s.isLoggedIn);
  const setBet = useSlot((s) => s.setBet);
  const setBalance = useSlot((s) => s.setBalance);
  const setLastWin = useSlot((s) => s.setLastWin);
  const setLastSymbols = useSlot((s) => s.setLastSymbols);

  const [spinning, setSpinning] = useState(false);

  const busy = phase !== "idle" && phase !== "ended";
  const canSpin = isLoggedIn && balance >= bet && !busy && !spinning;

  const handleSpin = async () => {
    if (!canSpin) return;

    setSpinning(true);

    try {
      // Real backend spin
      const result = await realSpin(bet);

      setBalance(result.balanceAfter);
      setLastWin(result.win);
      setLastSymbols(result.symbols);

      // Trigger visual spin
      useSlot.getState().spin();

      // Reset after animation
      setTimeout(() => {
        setSpinning(false);
      }, 1400);
    } catch (e) {
      console.error("Spin failed:", e);
      // Fallback to local
      useSlot.getState().spin();
      setSpinning(false);
    }
  };

  const changeBet = (delta: number) => {
    const newBet = Math.max(1, Math.min(1000, bet + delta));
    setBet(newBet);
  };

  return (
    <div className="mt-4 flex flex-col items-center gap-4">
      {/* Bet controls */}
      <div className="flex items-center gap-3">
        <Button
          variant="outline"
          size="icon"
          onClick={() => changeBet(-5)}
          disabled={busy || bet <= 1}
          className="h-10 w-10 rounded-xl border-white/20"
        >
          <Minus className="h-4 w-4" />
        </Button>

        <div className="flex items-center gap-2 rounded-2xl border border-white/10 bg-white/5 px-6 py-2 font-mono text-xl font-bold min-w-[140px] justify-center">
          BET {fmt(bet)}
        </div>

        <Button
          variant="outline"
          size="icon"
          onClick={() => changeBet(5)}
          disabled={busy || bet >= 1000}
          className="h-10 w-10 rounded-xl border-white/20"
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>

      {/* Spin button */}
      <Button
        onClick={handleSpin}
        disabled={!canSpin}
        className={cn(
          "h-16 w-16 rounded-full text-2xl font-black transition-all active:scale-95",
          canSpin
            ? "bg-emerald-500 hover:bg-emerald-600 shadow-lg shadow-emerald-500/30"
            : "bg-zinc-800 text-zinc-500 cursor-not-allowed"
        )}
      >
        {spinning ? (
          <RotateCw className="h-8 w-8 animate-spin" />
        ) : (
          "SPIN"
        )}
      </Button>

      <div className="text-[11px] text-slate-500 tracking-widest">
        {isLoggedIn ? "SPACE = SPIN" : "LOGIN TO PLAY"}
      </div>
    </div>
  );
}
