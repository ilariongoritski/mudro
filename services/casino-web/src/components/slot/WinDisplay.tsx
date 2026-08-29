"use client";

import { AnimatePresence, motion } from "framer-motion";
import { useSlot } from "@/lib/slot/store";
import { FairnessProof } from "../FairnessProof";

function winLabel(win: number, bet: number) {
  const multiplier = bet > 0 ? win / bet : 0;
  if (multiplier >= 40) return "MEGA WIN";
  if (multiplier >= 12) return "SUPER WIN";
  if (multiplier >= 4) return "BIG WIN";
  return "WIN";
}

export function WinDisplay() {
  const displayWin = useSlot((s) => s.displayWin);
  const bet = useSlot((s) => s.bet);
  const phase = useSlot((s) => s.phase);
  const fairness = useSlot((s) => s.fairness);
  const label = winLabel(displayWin, bet);
  const isTierWin = label !== "WIN";

  const visible = displayWin > 0 && phase !== "dropping";

  return (
    <div className="absolute inset-x-0 top-4 z-30 flex justify-center pointer-events-none">
      <AnimatePresence>
        {visible && (
          <motion.div
            key="win-banner"
            initial={{ opacity: 0, y: -18, scale: 0.85 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -14, scale: 0.9 }}
            transition={{ type: "spring", stiffness: 380, damping: 26 }}
            className={`rounded-2xl px-8 py-3 text-center shadow-2xl ${isTierWin ? "bg-gradient-to-br from-fuchsia-500 via-violet-600 to-indigo-700 ring-2 ring-yellow-200/80" : "bg-emerald-500/90"}`}
          >
            <div className={`text-xs font-black tracking-[3px] ${isTierWin ? "text-yellow-100" : "text-emerald-950"}`}>{label}</div>
            <div className="font-mono text-4xl font-black text-white tabular-nums drop-shadow-md">
              {displayWin.toFixed(2)}
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      <FairnessProof
        serverSeedHash={fairness?.serverSeedHash}
        nonce={fairness?.nonce}
      />
    </div>
  );
}
