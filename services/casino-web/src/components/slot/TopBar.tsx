"use client";

import { useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Volume2, VolumeX, Info, Wallet, RefreshCw } from "lucide-react";
import { useSlot } from "@/lib/slot/store";
import { sound } from "@/lib/slot/sound";
import { Paytable } from "./Paytable";
import { cn } from "@/lib/utils";

function fmt(n: number) {
  return n.toLocaleString("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

export function TopBar() {
  const balance = useSlot((s) => s.balance);
  const balancePulse = useSlot((s) => s.balancePulse);
  const soundOn = useSlot((s) => s.soundOn);
  const toggleSound = useSlot((s) => s.toggleSound);
  const resetBalance = useSlot((s) => s.resetBalance);
  const phase = useSlot((s) => s.phase);
  const inFreeSpins = useSlot((s) => s.inFreeSpins);

  const [payOpen, setPayOpen] = useState(false);
  const busy = phase !== "idle" && phase !== "ended";

  // Animated counting balance
  const balRef = useRef<HTMLDivElement>(null);
  const rafRef = useRef<number | null>(null);
  const prevBalanceRef = useRef(balance);

  useEffect(() => {
    const el = balRef.current;
    if (!el) return;
    const from = prevBalanceRef.current;
    const to = balance;
    prevBalanceRef.current = to;
    if (from === to) return;
    const duration = 700;
    const start = performance.now();
    const tick = (now: number) => {
      const t = Math.min(1, (now - start) / duration);
      const eased = 1 - Math.pow(1 - t, 3);
      const v = from + (to - from) * eased;
      el.textContent = fmt(v);
      if (t < 1) rafRef.current = requestAnimationFrame(tick);
    };
    rafRef.current = requestAnimationFrame(tick);
    return () => {
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
    };
  }, [balance]);

  const balanceGrowing = balancePulse > 0;

  return (
    <header
      className={cn(
        "w-full flex items-center justify-between gap-3 px-3 sm:px-5 py-3 border-b backdrop-blur sticky top-0 z-40 transition-colors",
        inFreeSpins
          ? "bg-[#1a0a2e]/90 border-pink-500/30"
          : "bg-[#0a0f1e]/90 border-white/10"
      )}
    >
      <div className="flex items-center gap-2.5 min-w-0">
        <div
          className={cn(
            "relative size-9 rounded-xl flex items-center justify-center shrink-0 transition-colors",
            inFreeSpins
              ? "bg-gradient-to-br from-pink-400 to-fuchsia-600 shadow-[0_4px_14px_rgba(236,72,153,.5)]"
              : "bg-gradient-to-br from-pink-400 to-fuchsia-600 shadow-[0_4px_14px_rgba(236,72,153,.4)]"
          )}
        >
          <span className="text-xl">🍭</span>
        </div>
        <div className="leading-tight min-w-0">
          <div className="font-black text-white text-base sm:text-lg tracking-tight truncate">
            SWEET<span className="text-pink-400">BONANZA</span>
          </div>
          <div className="text-[10px] text-slate-400 -mt-0.5 tracking-widest hidden sm:block">
            5×5 • PAY ANYWHERE • TUMBLE
          </div>
        </div>
      </div>

      <div className="flex items-center gap-2 sm:gap-3">
        <div
          className={cn(
            "flex items-center gap-2 rounded-xl border px-3 py-1.5 transition-colors",
            inFreeSpins
              ? "bg-[#2a1042] border-pink-500/40"
              : "bg-[#0c1322] border-pink-500/30"
          )}
        >
          <Wallet
            className={cn(
              "size-4 text-pink-400 transition-transform",
              balanceGrowing && "slot-wallet-bounce"
            )}
          />
          <div className="leading-none">
            <div className="text-[9px] text-slate-400 tracking-widest font-bold">
              BALANCE
            </div>
            <div
              key={balancePulse}
              className="font-black text-white text-sm sm:text-lg tabular-nums slot-balance-pop origin-left"
            >
              <span ref={balRef}>{fmt(balance)}</span>
            </div>
          </div>
          <button
            onClick={() => {
              if (busy) return;
              if (soundOn) sound.click();
              resetBalance();
            }}
            disabled={busy}
            className="ml-1 size-6 rounded-md bg-white/5 hover:bg-amber-500/20 border border-white/10 flex items-center justify-center text-amber-300 transition disabled:opacity-40"
            title="Reset balance to 1000"
          >
            <RefreshCw className="size-3" />
          </button>
        </div>

        <Button
          variant="ghost"
          size="icon"
          onClick={() => {
            if (soundOn) sound.click();
            setPayOpen(true);
          }}
          className="size-9 rounded-lg bg-white/5 hover:bg-white/10 text-slate-200 border border-white/10"
          title="Paytable"
        >
          <Info className="size-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => {
            if (soundOn) sound.click();
            toggleSound();
          }}
          className={cn(
            "size-9 rounded-lg border border-white/10",
            soundOn
              ? "bg-white/5 hover:bg-white/10 text-pink-400"
              : "bg-white/5 text-slate-500"
          )}
          title={soundOn ? "Mute" : "Unmute"}
        >
          {soundOn ? (
            <Volume2 className="size-4" />
          ) : (
            <VolumeX className="size-4" />
          )}
        </Button>
      </div>

      <Paytable open={payOpen} onOpenChange={setPayOpen} />
    </header>
  );
}
