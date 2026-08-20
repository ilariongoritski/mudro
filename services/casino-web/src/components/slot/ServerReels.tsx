"use client";

import { motion, useReducedMotion } from "framer-motion";
import { useSlot } from "@/lib/slot/store";

const REEL_STRIP = [
  ["🍓", "#ec4899", "#f9a8d4"], ["🍊", "#f97316", "#fdba74"], ["🫐", "#3b82f6", "#93c5fd"],
  ["🍐", "#84cc16", "#bef264"], ["🍉", "#10b981", "#6ee7b7"], ["🍎", "#ef4444", "#fca5a5"],
  ["🍇", "#8b5cf6", "#c4b5fd"], ["❤️", "#e11d48", "#fda4af"], ["🍭", "#f472b6", "#f9a8d4"],
] as const;

const SYMBOL_TILE: Record<string, (typeof REEL_STRIP)[number]> = {
  strawberry: REEL_STRIP[0], orange: REEL_STRIP[1], blueberry: REEL_STRIP[2],
  pear: REEL_STRIP[3], watermelon: REEL_STRIP[4], apple: REEL_STRIP[5],
  grape: REEL_STRIP[6], heart: REEL_STRIP[7], scatter: REEL_STRIP[8],
  bomb: ["💣", "#f59e0b", "#fde68a"],
};

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
        const result = (board[reel] ?? []).map((cell) => SYMBOL_TILE[cell.symbol] ?? REEL_STRIP[0]);
        // Keep the strip length stable while the request is pending. Once the
        // server board arrives, replace only its final five cells and land on
        // their boundary; the overlay and TumbleGrid therefore show the same
        // server-authoritative initial_board at handoff.
        const strip = hasServerResult ? [...filler, ...filler, ...result] : [...filler, ...filler, ...filler];
        const landingOffset = FILLER_ROWS * 2;

        return (
          <div key={reel} className="relative overflow-hidden rounded-xl border border-white/10 bg-black/25">
            <motion.div
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
              {strip.map(([symbol, color, glow], index) => (
                <div key={`${reel}-${index}`} className="flex h-[var(--cell,64px)] items-center justify-center">
                  <span
                    className="flex items-center justify-center rounded-[26%] border border-white/25 leading-none shadow-[inset_0_-5px_10px_rgba(0,0,0,.3),inset_0_4px_8px_rgba(255,255,255,.22)]"
                    style={{
                      width: "calc(var(--cell, 64px) * .88)", height: "calc(var(--cell, 64px) * .88)",
                      fontSize: "calc(var(--cell, 64px) * .49)",
                      background: `radial-gradient(120% 120% at 50% 14%, ${glow}88 0%, ${color} 48%, #1a0f2e 140%)`,
                      filter: `drop-shadow(0 0 7px ${glow}aa)`,
                    }}
                  >{symbol}</span>
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
