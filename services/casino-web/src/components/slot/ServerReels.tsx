"use client";

import { motion, useReducedMotion } from "framer-motion";
import { useSlot } from "@/lib/slot/store";
import { TumbleTile } from "./TumbleTile";
import type { SymbolId } from "@/lib/slot/config";

/** Decorative filler tile shown only while the request is pending. */
function FillerTile({ symbol, color, glow }: { symbol: string; color: string; glow: string }) {
  return (
    <span
      className="flex items-center justify-center rounded-[26%] border border-white/25 leading-none shadow-[inset_0_-5px_10px_rgba(0,0,0,.3),inset_0_4px_8px_rgba(255,255,255,.22)]"
      style={{
        width: "calc(var(--cell, 64px) * .88)",
        height: "calc(var(--cell, 64px) * .88)",
        fontSize: "calc(var(--cell, 64px) * .49)",
        background: `radial-gradient(120% 120% at 50% 14%, ${glow}88 0%, ${color} 48%, #1a0f2e 140%)`,
        filter: `drop-shadow(0 0 7px ${glow}aa)`,
      }}
    >{symbol}</span>
  );
}

const REEL_STRIP: [string, string, string][] = [
  ["🍓", "#ec4899", "#f9a8d4"], ["🍊", "#f97316", "#fdba74"], ["🫐", "#3b82f6", "#93c5fd"],
  ["🍐", "#84cc16", "#bef264"], ["🍉", "#10b981", "#6ee7b7"], ["🍎", "#ef4444", "#fca5a5"],
  ["🍇", "#8b5cf6", "#c4b5fd"], ["❤️", "#e11d48", "#fda4af"], ["🍭", "#f472b6", "#f9a8d4"],
];

const FILLER_ROWS = 9;

export function ServerReels() {
  const phase = useSlot((s) => s.phase);
  const board = useSlot((s) => s.board);
  const spinKey = useSlot((s) => s.spinKey);
  const serverReelsReady = useSlot((s) => s.serverReelsReady);
  const turbo = useSlot((s) => s.turbo);
  const reduceMotion = useReducedMotion();

  if (phase !== "dropping") return null;

  const hasServerResult = serverReelsReady;
  const speed = turbo ? 0.5 : 1;

  return (
    <div className="absolute inset-0 z-20 grid grid-cols-5 gap-1.5 rounded-2xl bg-[#180d30]/94 p-2 sm:gap-2 sm:p-2.5" aria-label="Slot reels are spinning" aria-live="polite">
      {Array.from({ length: 5 }).map((_, reel) => {
        const filler = Array.from({ length: FILLER_ROWS }, (_, row) => REEL_STRIP[(reel * 2 + row) % REEL_STRIP.length]);
        // Landing cells reuse TumbleTile — the exact component the settled
        // board renders — so the stopped overlay and the revealed board are
        // visually identical at handoff.
        const result = (board[reel] ?? []).map((cell) => (
          <TumbleTile key={`r-${cell.id}`} symbol={cell.symbol as SymbolId} mult={cell.mult} />
        ));
        const strip = hasServerResult
          ? [...filler.map(([s, c, g], i) => <FillerTile key={`f-${i}`} symbol={s} color={c} glow={g} />), ...result]
          : filler.map(([s, c, g], i) => <FillerTile key={`f-${i}`} symbol={s} color={c} glow={g} />);
        const landingOffset = FILLER_ROWS;
        const stopDuration = (0.72 + reel * 0.095) * speed;

        return (
          <div key={reel} className="relative overflow-hidden rounded-xl border border-white/10 bg-black/25">
            <motion.div
              key={`${spinKey}-${reel}-${hasServerResult ? "landing" : "spinning"}`}
              className="absolute inset-x-0 flex flex-col"
              initial={{ y: 0 }}
              animate={
                reduceMotion
                  ? { opacity: hasServerResult ? 1 : 0.72 }
                  : hasServerResult
                    ? { y: `calc(var(--cell, 64px) * -${landingOffset})` }
                    : { y: [0, "calc(var(--cell, 64px) * -9)"] }
              }
              transition={
                reduceMotion
                  ? { duration: 0.1 }
                  : hasServerResult
                    ? { duration: stopDuration, ease: [0.12, 0.76, 0.2, 1] }
                    : { duration: (0.42 + reel * 0.035) * speed, repeat: Infinity, ease: "linear" }
              }
            >
              {strip.map((tile, index) => (
                <div key={`${reel}-${index}`} className="flex h-[var(--cell,64px)] w-full items-center justify-center">
                  {tile}
                </div>
              ))}
            </motion.div>
          </div>
        );
      })}
      <p className="pointer-events-none absolute bottom-2 left-0 right-0 text-center text-[9px] font-bold tracking-[0.18em] text-white/65">
        {hasServerResult ? "REELS STOPPING" : "SERVER SPIN IN PROGRESS"}
      </p>
    </div>
  );
}
