"use client";

import { Sparkles } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import {
  SYMBOLS,
  PAYING_SYMBOLS,
  MIN_PAY_COUNT,
  FREE_SPINS_AWARD,
  FREE_SPINS_TRIGGER,
  FREE_SPINS_RETRIGGER,
  FREE_SPINS_RETRIGGER_AWARD,
  BOMB_VALUES,
  BONUS_BUY_MULT,
} from "@/lib/slot/config";

interface Props {
  open: boolean;
  onOpenChange: (v: boolean) => void;
}

function fmtPay(n: number) {
  return n % 1 === 0 ? String(n) : n.toFixed(2);
}

export function Paytable({ open, onOpenChange }: Props) {
  const ordered = [
    "scatter",
    "bomb",
    ...PAYING_SYMBOLS,
  ] as const;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl bg-[#1a0a2e] border-pink-500/20 text-slate-200 max-h-[88vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle className="text-white text-xl font-black tracking-tight">
            🍭 Paytable &amp; Rules
          </DialogTitle>
          <DialogDescription className="text-slate-400">
            Sweet Bonanza — 5×5 grid. Wins pay <b>anywhere</b> (no paylines):
            land {MIN_PAY_COUNT}+ matching symbols anywhere on the grid. Winning
            symbols tumble away and new ones drop in — chain cascades for bigger
            multipliers!
          </DialogDescription>
        </DialogHeader>

        <div className="overflow-y-auto pr-1 -mr-1 slot-scroll">
          {/* Symbol payouts */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            {ordered.map((id) => {
              const s = SYMBOLS[id];
              const isBomb = s.isBomb;
              return (
                <div
                  key={id}
                  className="flex items-center gap-3 rounded-lg bg-white/5 border border-white/10 p-2.5"
                >
                  <div
                    className="size-12 rounded-xl flex items-center justify-center shrink-0 relative overflow-hidden"
                    style={{
                      background: `radial-gradient(120% 120% at 50% 14%, ${s.glow}55 0%, ${s.color} 45%, #1a0f2e 130%)`,
                      boxShadow: `inset 0 -6px 10px rgba(0,0,0,.4), inset 0 4px 8px rgba(255,255,255,.12)`,
                    }}
                  >
                    <span
                      className="text-2xl"
                      style={{ filter: `drop-shadow(0 0 6px ${s.glow}aa)` }}
                    >
                      {s.emoji}
                    </span>
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="font-bold text-white text-sm">
                        {s.label}
                      </span>
                      {s.isScatter && (
                        <span className="text-[9px] px-1.5 py-0.5 rounded-full bg-pink-500/20 text-pink-300 font-bold border border-pink-400/40">
                          BONUS
                        </span>
                      )}
                      {isBomb && (
                        <span className="text-[9px] px-1.5 py-0.5 rounded-full bg-amber-500/20 text-amber-300 font-bold border border-amber-400/40">
                          MULTIPLIER
                        </span>
                      )}
                    </div>
                    {isBomb ? (
                      <div className="text-[11px] text-amber-300/90 mt-0.5 leading-tight">
                        Spawns in free spins. Values: {BOMB_VALUES.join(", ")}×.
                        Sum applies to spin win.
                      </div>
                    ) : s.isScatter ? (
                      <div className="flex gap-2 mt-0.5 text-[11px] tabular-nums flex-wrap">
                        <span className="text-slate-400">
                          4× <b className="text-pink-300">{fmtPay(s.pay[4])}</b>
                        </span>
                        <span className="text-slate-400">
                          5× <b className="text-pink-300">{fmtPay(s.pay[5])}</b>
                        </span>
                        <span className="text-slate-400">
                          6× <b className="text-pink-300">{fmtPay(s.pay[6])}</b>
                        </span>
                      </div>
                    ) : (
                      <div className="flex gap-2 mt-0.5 text-[11px] tabular-nums flex-wrap">
                        <span className="text-slate-400">
                          {MIN_PAY_COUNT}×{" "}
                          <b className="text-emerald-300">
                            {fmtPay(s.pay[MIN_PAY_COUNT])}
                          </b>
                        </span>
                        <span className="text-slate-400">
                          7× <b className="text-emerald-300">{fmtPay(s.pay[7])}</b>
                        </span>
                        <span className="text-slate-400">
                          9× <b className="text-emerald-300">{fmtPay(s.pay[9])}</b>
                        </span>
                        <span className="text-slate-400">
                          12+{" "}
                          <b className="text-emerald-300">
                            {fmtPay(s.pay[12])}
                          </b>
                        </span>
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>

          {/* Rules */}
          <div className="mt-4 space-y-3 text-sm">
            <div className="rounded-lg bg-pink-500/10 border border-pink-400/30 p-3">
              <div className="font-bold text-pink-300 mb-1 flex items-center gap-2">
                <span>🍭</span> Free Spins (BONUS)
              </div>
              <p className="text-slate-300 text-[13px] leading-relaxed">
                Land <b className="text-white">{FREE_SPINS_TRIGGER}+ lollipops</b>{" "}
                anywhere to trigger{" "}
                <b className="text-pink-200">{FREE_SPINS_AWARD} Free Spins</b>.
                Land {FREE_SPINS_RETRIGGER}+ during the bonus to retrigger +{FREE_SPINS_RETRIGGER_AWARD} spins.
              </p>
            </div>

            <div className="rounded-lg bg-amber-500/10 border border-amber-400/30 p-3">
              <div className="font-bold text-amber-300 mb-1 flex items-center gap-2">
                <Sparkles className="size-4" />
                Buy Bonus
              </div>
              <p className="text-slate-300 text-[13px] leading-relaxed">
                Can&apos;t wait? Pay{" "}
                <b className="text-amber-200">{BONUS_BUY_MULT}× your bet</b> to
                instantly trigger {FREE_SPINS_AWARD} free spins with multiplier
                bombs — no lollipops required. Tap the gold{" "}
                <b className="text-amber-200">BUY BONUS</b> bar under the reels.
              </p>
            </div>

            <div className="rounded-lg bg-amber-500/10 border border-amber-400/30 p-3">
              <div className="font-bold text-amber-300 mb-1 flex items-center gap-2">
                <span>💣</span> Multiplier Bombs
              </div>
              <p className="text-slate-300 text-[13px] leading-relaxed">
                During free spins, multiplier bombs (×2 up to ×100) drop onto the
                grid. At the end of each spin's tumble sequence, all bomb values
                are summed and multiply your total spin win!
              </p>
            </div>

            <div className="rounded-lg bg-emerald-500/10 border border-emerald-400/30 p-3">
              <div className="font-bold text-emerald-300 mb-1 flex items-center gap-2">
                <span>⚡</span> Tumble &amp; Cascade Multiplier
              </div>
              <p className="text-slate-300 text-[13px] leading-relaxed">
                Winning symbols explode and disappear; remaining symbols fall
                down and new ones drop in — this can chain into more wins. Each
                cascade step increases your win multiplier (×1 → ×2 → ×3 → ×5 →
                ×8...) for explosive payouts!
              </p>
            </div>

            <div className="rounded-lg bg-white/5 border border-white/10 p-3">
              <div className="font-bold text-slate-200 mb-1">🎯 How to Win</div>
              <p className="text-slate-400 text-[13px] leading-relaxed">
                No paylines! Count matching symbols anywhere on the 5×5 grid.{" "}
                {MIN_PAY_COUNT}+ of the same symbol = win. More symbols = bigger
                payout. Multiple symbol types can win in the same cascade.
              </p>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
