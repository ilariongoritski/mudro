"use client";

import { useEffect, useRef } from "react";
import { useSlot } from "@/lib/slot/store";
import { TIER_LABEL } from "@/lib/slot/config";

function fmt(n: number) {
  return n.toLocaleString("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

function easeOutCubic(t: number) {
  return 1 - Math.pow(1 - t, 3);
}

export function WinDisplay() {
  const displayWin = useSlot((s) => s.displayWin);
  const winTier = useSlot((s) => s.winTier);
  const cascade = useSlot((s) => s.cascade);
  const cascadeMult = useSlot((s) => s.cascadeMult);
  const phase = useSlot((s) => s.phase);
  const inFreeSpins = useSlot((s) => s.inFreeSpins);

  const numRef = useRef<HTMLSpanElement>(null);
  const rafRef = useRef<number | null>(null);
  const lastTargetRef = useRef(0);

  useEffect(() => {
    if (displayWin <= 0) {
      lastTargetRef.current = 0;
      return;
    }
    const el = numRef.current;
    if (!el) return;
    const from = lastTargetRef.current;
    const to = displayWin;
    lastTargetRef.current = to;
    const duration =
      winTier === "epic" ? 2000 : winTier === "mega" ? 1600 : winTier === "big" ? 1200 : 700;
    const start = performance.now();
    const tick = (now: number) => {
      const t = Math.min(1, (now - start) / duration);
      const v = from + (to - from) * easeOutCubic(t);
      el.textContent = fmt(v);
      if (t < 1) rafRef.current = requestAnimationFrame(tick);
    };
    el.textContent = fmt(from);
    rafRef.current = requestAnimationFrame(tick);
    return () => {
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
    };
  }, [displayWin, winTier]);

  // Cascade multiplier badge (during celebrating / tumbling)
  const showCascade =
    cascade >= 1 && (phase === "celebrating" || phase === "tumbling");

  if (displayWin <= 0 && !showCascade) return null;

  const isBig = winTier === "big" || winTier === "mega" || winTier === "epic";
  const tierColor =
    winTier === "epic"
      ? "#fde047"
      : winTier === "mega"
        ? "#fb923c"
        : winTier === "big"
          ? "#34d399"
          : "#a7f3d0";

  // Cascade badge color
  const cascadeColor = cascade >= 4 ? "#fde047" : cascade >= 2 ? "#fb923c" : "#34d399";

  return (
    <div className="absolute inset-0 z-30 pointer-events-none flex flex-col items-center justify-center">
      {/* Cascade multiplier (top) */}
      {showCascade && (
        <div
          className="absolute top-1.5 left-1/2 -translate-x-1/2 slot-cascade-badge"
          style={{ marginTop: "0px" }}
        >
          <div
            className="px-3 py-1 rounded-full font-black flex items-center gap-1.5 border-2"
            style={{
              background: "rgba(10,5,25,.85)",
              borderColor: cascadeColor,
              color: cascadeColor,
              boxShadow: `0 0 16px ${cascadeColor}88`,
            }}
          >
            {inFreeSpins && <span className="text-base">💣</span>}
            <span style={{ fontSize: "clamp(14px, 3vw, 20px)" }}>
              ×{cascadeMult}
            </span>
            {cascade >= 2 && (
              <span className="text-[10px] font-bold opacity-80">
                CASCADE {cascade}
              </span>
            )}
          </div>
        </div>
      )}

      {displayWin > 0 && (
        <>
          {isBig ? (
            <div className="flex flex-col items-center slot-bigwin-enter">
              <div
                className="font-black tracking-tight text-center leading-none"
                style={{
                  fontSize: "clamp(26px, 6.5vw, 58px)",
                  color: tierColor,
                  textShadow: `0 0 24px ${tierColor}cc, 0 0 48px ${tierColor}77, 0 4px 8px rgba(0,0,0,.6)`,
                }}
              >
                {TIER_LABEL[winTier]}
              </div>
              <span
                ref={numRef}
                className="font-black tabular-nums mt-1"
                style={{
                  fontSize: "clamp(22px, 5.5vw, 48px)",
                  color: "#fff",
                  textShadow: `0 0 18px ${tierColor}aa, 0 3px 6px rgba(0,0,0,.7)`,
                }}
              >
                0.00
              </span>
            </div>
          ) : (
            <div className="slot-smallwin-enter bg-[#1a0a2e]/85 border border-emerald-400/40 rounded-full px-4 py-1.5 flex items-center gap-2 backdrop-blur">
              <span className="text-emerald-300 text-xs font-bold tracking-widest">
                WIN
              </span>
              <span
                ref={numRef}
                className="font-black tabular-nums text-white"
                style={{ fontSize: "clamp(15px, 3.2vw, 22px)" }}
              >
                0.00
              </span>
            </div>
          )}
        </>
      )}
    </div>
  );
}
