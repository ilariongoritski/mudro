"use client";

import { useEffect, useRef } from "react";
import { TopBar } from "./TopBar";
import { TumbleGrid } from "./TumbleGrid";
import { ControlPanel } from "./ControlPanel";
import { WinDisplay } from "./WinDisplay";
import { Banners } from "./Banners";
import { Particles } from "./Particles";
import { useSlot } from "@/lib/slot/store";

export function SlotMachine() {
  const phase = useSlot((s) => s.phase);
  const spinKey = useSlot((s) => s.spinKey);
  const winTier = useSlot((s) => s.winTier);
  const inFreeSpins = useSlot((s) => s.inFreeSpins);
  const freeSpins = useSlot((s) => s.freeSpins);
  const freeSpinsTotal = useSlot((s) => s.freeSpinsTotal);
  const freeSpinsWin = useSlot((s) => s.freeSpinsWin);
  const balance = useSlot((s) => s.balance);
  const bet = useSlot((s) => s.bet);
  const spin = useSlot((s) => s.spin);
  const cascade = useSlot((s) => s.cascade);
  const activeBombs = useSlot((s) => s.activeBombs);
  const seedBoard = useSlot((s) => s.seedBoard);
  const hydrate = useSlot((s) => s.hydrate);

  const panelRef = useRef<HTMLDivElement>(null);
  const busy = phase !== "idle" && phase !== "ended";

  // Hydrate balance from localStorage + seed an initial display board on mount
  // (client-only, avoids hydration mismatch)
  useEffect(() => {
    hydrate();
    seedBoard();
  }, [hydrate, seedBoard]);

  // screen shake on mega/epic (DOM-only)
  useEffect(() => {
    if (winTier !== "mega" && winTier !== "epic") return;
    const el = panelRef.current;
    if (!el) return;
    el.classList.remove("slot-shake");
    void el.offsetWidth;
    el.classList.add("slot-shake");
    const t = setTimeout(() => el.classList.remove("slot-shake"), 650);
    return () => clearTimeout(t);
  }, [winTier, spinKey]);

  // spacebar to spin
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.code === "Space") {
        const target = e.target as HTMLElement;
        if (
          target &&
          (target.tagName === "INPUT" || target.tagName === "TEXTAREA")
        )
          return;
        e.preventDefault();
        if (!busy && !inFreeSpins && balance >= bet) spin();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [busy, inFreeSpins, balance, bet, spin]);

  const used = freeSpinsTotal - freeSpins;

  return (
    <div className="min-h-screen flex flex-col text-white">
      {/* ambient background */}
      <div
        className={`pointer-events-none fixed inset-0 z-0 ${
          inFreeSpins ? "slot-ambient-fs" : "slot-ambient"
        }`}
        aria-hidden
      />

      <div className="relative z-10 flex flex-col min-h-screen">
        <TopBar />

        <main className="flex-1 flex flex-col items-center justify-center px-2 sm:px-3 py-3 sm:py-5">
          <div className="w-full max-w-2xl">
            {/* free spins indicator */}
            {inFreeSpins && (
              <div className="mb-2 flex justify-center">
                <div className="flex items-center gap-3 px-4 py-1.5 rounded-full bg-pink-500/15 border border-pink-400/40">
                  <span className="text-lg">🍭</span>
                  <span className="text-pink-200 font-bold text-xs sm:text-sm tracking-widest">
                    BONUS SPIN {Math.max(0, used)} / {freeSpinsTotal}
                  </span>
                  {activeBombs > 0 && (
                    <span className="flex items-center gap-1 text-amber-300 font-black text-xs">
                      <span>💣</span>×{activeBombs}
                    </span>
                  )}
                  <span className="text-yellow-300 font-black text-xs">
                    WIN {freeSpinsWin.toFixed(2)}
                  </span>
                </div>
              </div>
            )}

            {/* cascade indicator (when active, not free spins) */}
            {cascade >= 2 && !inFreeSpins && busy && (
              <div className="mb-1.5 flex justify-center">
                <div className="px-3 py-0.5 rounded-full bg-amber-500/15 border border-amber-400/40 text-amber-300 font-bold text-xs tracking-widest slot-cascade-pulse">
                  CASCADE ×{[0, 1, 1, 2, 3, 5, 8][Math.min(cascade, 6)]}
                </div>
              </div>
            )}

            {/* game panel */}
            <div ref={panelRef} className="relative rounded-3xl p-2.5 sm:p-4 slot-panel">
              <div className="relative">
                <TumbleGrid />
                <WinDisplay />
                <Banners />
              </div>
              <ControlPanel />
            </div>

            {/* status row */}
            <div className="mt-2 flex items-center justify-between text-[11px] text-slate-500 px-1">
              <span className="font-bold tracking-widest">
                5×5 • PAY ANYWHERE • BET {bet.toFixed(2)}
              </span>
              <span className="tracking-wide">SPACE = SPIN</span>
            </div>
          </div>
        </main>

        <footer className="mt-auto border-t border-white/5 px-4 py-3 text-center">
          <p className="text-[11px] text-slate-500 tracking-wide">
            🍭 Sweet Bonanza — a demo slot for entertainment only. No real money.
            Play responsibly.
          </p>
        </footer>
      </div>

      <Particles />
    </div>
  );
}
