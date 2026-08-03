"use client";

import { motion, useReducedMotion } from "framer-motion";
import { useSlot } from "@/lib/slot/store";

const REEL_STRIP = [
  ["🍓", "#ec4899", "#f9a8d4"], ["🍊", "#f97316", "#fdba74"], ["🫐", "#3b82f6", "#93c5fd"],
  ["🍐", "#84cc16", "#bef264"], ["🍉", "#10b981", "#6ee7b7"], ["🍎", "#ef4444", "#fca5a5"],
  ["🍇", "#8b5cf6", "#c4b5fd"], ["❤️", "#e11d48", "#fda4af"], ["🍭", "#f472b6", "#f9a8d4"],
] as const;

export function ServerReels() {
  const phase = useSlot((s) => s.phase);
  const reduceMotion = useReducedMotion();
  if (phase !== "dropping") return null;

  return (
    <div className="absolute inset-0 z-20 grid grid-cols-5 gap-1.5 rounded-2xl bg-[#180d30]/94 p-2 sm:gap-2 sm:p-2.5" aria-label="Slot reels are spinning" aria-live="polite">
      {Array.from({ length: 5 }).map((_, reel) => (
        <div key={reel} className="relative overflow-hidden rounded-xl border border-white/10 bg-black/25">
          <motion.div
            className="absolute inset-x-0 flex flex-col items-center gap-1.5 py-1.5"
            initial={{ y: 0 }}
            animate={reduceMotion ? { opacity: 0.72 } : { y: [0, -392] }}
            transition={reduceMotion ? { duration: 0.15 } : { duration: 0.52 + reel * 0.045, repeat: Infinity, ease: "linear" }}
          >
            {[...REEL_STRIP, ...REEL_STRIP, ...REEL_STRIP].map(([symbol, color, glow], index) => (
              <span
                key={`${reel}-${index}`}
                className="flex items-center justify-center rounded-[26%] border border-white/25 leading-none shadow-[inset_0_-5px_10px_rgba(0,0,0,.3),inset_0_4px_8px_rgba(255,255,255,.22)]"
                style={{
                  width: "calc(var(--cell, 64px) * .88)", height: "calc(var(--cell, 64px) * .88)",
                  fontSize: "calc(var(--cell, 64px) * .49)",
                  background: `radial-gradient(120% 120% at 50% 14%, ${glow}88 0%, ${color} 48%, #1a0f2e 140%)`,
                  filter: `drop-shadow(0 0 7px ${glow}aa)`,
                }}
              >{symbol}</span>
            ))}
          </motion.div>
        </div>
      ))}
      <div className="pointer-events-none absolute inset-x-4 top-1/2 h-px bg-white/55 shadow-[0_0_18px_4px_rgba(255,255,255,.22)]" />
      <p className="pointer-events-none absolute bottom-2 left-0 right-0 text-center text-[9px] font-bold tracking-[0.18em] text-white/65">SERVER SPIN IN PROGRESS</p>
    </div>
  );
}
