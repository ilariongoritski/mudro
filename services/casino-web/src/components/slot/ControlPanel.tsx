"use client";

import { useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Minus, Plus, RotateCw, Zap } from "lucide-react";
import { useSlot } from "@/lib/slot/store";
import { spin as realSpin } from "@/lib/casino-api";
import { BET_PRESETS } from "@/lib/slot/config";
import { cn } from "@/lib/utils";

function fmt(value: number) {
  return value.toLocaleString("en-US", { maximumFractionDigits: 2 });
}

export function ControlPanel() {
  const bet = useSlot((s) => s.bet);
  const balance = useSlot((s) => s.balance);
  const phase = useSlot((s) => s.phase);
  const isLoggedIn = useSlot((s) => s.isLoggedIn);
  const turbo = useSlot((s) => s.turbo);
  const setBet = useSlot((s) => s.setBet);
  const incBet = useSlot((s) => s.incBet);
  const decBet = useSlot((s) => s.decBet);
  const toggleTurbo = useSlot((s) => s.toggleTurbo);
  const beginServerSpin = useSlot((s) => s.beginServerSpin);
  const applyServerSpin = useSlot((s) => s.applyServerSpin);
  const failServerSpin = useSlot((s) => s.failServerSpin);

  const [requestInFlight, setRequestInFlight] = useState(false);
  const [autoRemaining, setAutoRemaining] = useState(0);
  const [autoInfinite, setAutoInfinite] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const autoRef = useRef({ remaining: 0, infinite: false, stopped: false });

  const busy = phase !== "idle" && phase !== "ended";
  const canSpin = isLoggedIn && balance >= bet && !busy && !requestInFlight;

  const runSpin = async (): Promise<boolean> => {
    const startedAt = performance.now();
    if (!beginServerSpin()) return false;
    setRequestInFlight(true);
    setError(null);
    try {
      const result = await realSpin(bet);
      const minimumSpinMs = turbo ? 320 : 700;
      const elapsed = performance.now() - startedAt;
      if (elapsed < minimumSpinMs) {
        await new Promise((resolve) => window.setTimeout(resolve, minimumSpinMs - elapsed));
      }
      applyServerSpin({
        balanceAfter: result.balance,
        win: result.win,
        symbols: result.symbols,
        serverSeedHash: result.serverSeedHash,
        nonce: result.nonce,
      });
      return true;
    } catch (cause) {
      failServerSpin();
      setError(cause instanceof Error ? cause.message : "Spin was not confirmed");
      return false;
    } finally {
      setRequestInFlight(false);
    }
  };

  const handleSpin = () => void runSpin();

  const stopAuto = () => {
    autoRef.current = { remaining: 0, infinite: false, stopped: true };
    setAutoRemaining(0);
    setAutoInfinite(false);
  };

  const startAuto = async (count: number | "infinite") => {
    if (!canSpin) return;
    autoRef.current = { remaining: count === "infinite" ? 0 : count, infinite: count === "infinite", stopped: false };
    setAutoRemaining(count === "infinite" ? 0 : count);
    setAutoInfinite(count === "infinite");
    while (!autoRef.current.stopped && (autoRef.current.infinite || autoRef.current.remaining > 0)) {
      const ok = await runSpin();
      if (!ok || autoRef.current.stopped) break;
      if (!autoRef.current.infinite) {
        autoRef.current.remaining -= 1;
        setAutoRemaining(autoRef.current.remaining);
      }
      await new Promise((resolve) => window.setTimeout(resolve, turbo ? 300 : 850));
    }
    stopAuto();
  };

  useEffect(() => () => stopAuto(), []);

  return (
    <div className="mt-4 flex flex-col items-center gap-3">
      <div className="flex items-center gap-2">
        <Button variant="outline" size="icon" onClick={decBet} disabled={busy || requestInFlight || bet === BET_PRESETS[0]} className="h-10 w-10 rounded-xl border-white/20"><Minus className="h-4 w-4" /></Button>
        <div className="min-w-[132px] rounded-2xl border border-white/10 bg-white/5 px-4 py-2 text-center font-mono text-lg font-bold">BET {fmt(bet)}</div>
        <Button variant="outline" size="icon" onClick={incBet} disabled={busy || requestInFlight || bet === BET_PRESETS[BET_PRESETS.length - 1]} className="h-10 w-10 rounded-xl border-white/20"><Plus className="h-4 w-4" /></Button>
      </div>

      <div className="flex max-w-full flex-wrap justify-center gap-1.5 px-2">
        {BET_PRESETS.map((value) => <button key={value} type="button" disabled={busy || requestInFlight} onClick={() => setBet(value)} className={cn("rounded-lg border px-2 py-1 text-xs font-bold", bet === value ? "border-emerald-400 bg-emerald-500/20 text-emerald-200" : "border-white/10 bg-white/5 text-slate-300")}>{fmt(value)}</button>)}
      </div>

      <div className="flex items-center gap-2">
        <button type="button" onClick={toggleTurbo} disabled={requestInFlight} className={cn("flex h-11 items-center gap-1 rounded-xl border px-3 text-xs font-black", turbo ? "border-amber-300 bg-amber-400/20 text-amber-200" : "border-white/15 bg-white/5 text-slate-300")}><Zap className="h-4 w-4" /> TURBO</button>
        <Button onClick={handleSpin} disabled={!canSpin} className={cn("h-16 min-w-24 rounded-full text-xl font-black", canSpin ? "bg-emerald-500 hover:bg-emerald-600" : "bg-zinc-800 text-zinc-500")}>
          {requestInFlight ? <RotateCw className="h-7 w-7 animate-spin" /> : "SPIN"}
        </Button>
      </div>

      <div className="flex flex-wrap justify-center gap-1.5">
        {autoRemaining > 0 || autoInfinite ? <button type="button" onClick={stopAuto} className="rounded-lg bg-rose-500/80 px-3 py-1.5 text-xs font-black">STOP AUTO</button> : <>
          <button type="button" disabled={!canSpin} onClick={() => void startAuto(10)} className="rounded-lg border border-white/15 bg-white/5 px-3 py-1.5 text-xs font-bold">AUTO ×10</button>
          <button type="button" disabled={!canSpin} onClick={() => void startAuto(25)} className="rounded-lg border border-white/15 bg-white/5 px-3 py-1.5 text-xs font-bold">AUTO ×25</button>
          <button type="button" disabled={!canSpin} onClick={() => void startAuto("infinite")} className="rounded-lg border border-white/15 bg-white/5 px-3 py-1.5 text-xs font-bold">AUTO ∞</button>
        </>}
      </div>
      <div className="text-[11px] tracking-widest text-slate-500">{autoInfinite ? "AUTO ∞ ACTIVE" : autoRemaining > 0 ? `AUTO ${autoRemaining} LEFT` : turbo ? "TURBO ON" : "SERVER-VERIFIED SPINS"}</div>
      {error && <p role="alert" className="max-w-sm text-center text-xs text-rose-300">{error}</p>}
    </div>
  );
}
