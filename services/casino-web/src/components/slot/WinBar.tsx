"use client";

import { useEffect, useRef } from "react";
import { useSlot } from "@/lib/slot/store";
import { cn } from "@/lib/utils";

function fmt(n: number) {
  return n.toLocaleString("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

function easeOutCubic(t: number) {
  return 1 - Math.pow(1 - t, 3);
}

interface StatTileProps {
  label: string;
  value: number;
  active: boolean;
  accent: string;
  icon: string;
  pulseKey: number;
}

function StatTile({ label, value, active, accent, icon, pulseKey }: StatTileProps) {
  const numRef = useRef<HTMLSpanElement>(null);
  const rafRef = useRef<number | null>(null);
  const prevRef = useRef(value);

  useEffect(() => {
    const el = numRef.current;
    if (!el) return;
    const from = prevRef.current;
    const to = value;
    prevRef.current = to;
    if (from === to) return;
    const duration = 600;
    const start = performance.now();
    const tick = (now: number) => {
      const t = Math.min(1, (now - start) / duration);
      const v = from + (to - from) * easeOutCubic(t);
      el.textContent = fmt(v);
      if (t < 1) rafRef.current = requestAnimationFrame(tick);
    };
    rafRef.current = requestAnimationFrame(tick);
    return () => {
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
    };
  }, [value]);

  return (
    <div
      className={cn(
        "flex-1 rounded-xl px-3 py-2 border transition-all duration-300 flex items-center gap-2.5",
        active
          ? "bg-white/8 border-white/15"
          : "bg-white/3 border-white/8 opacity-60"
      )}
      style={
        active
          ? { boxShadow: `0 0 14px ${accent}33, inset 0 0 12px ${accent}11` }
          : undefined
      }
    >
      <span
        className="text-lg leading-none"
        style={{ filter: active ? `drop-shadow(0 0 6px ${accent}aa)` : "none" }}
      >
        {icon}
      </span>
      <div className="leading-tight min-w-0">
        <div
          className="text-[9px] font-bold tracking-widest"
          style={{ color: active ? accent : "#94a3b8" }}
        >
          {label}
        </div>
        <div
          key={pulseKey}
          className="font-black tabular-nums text-white text-sm sm:text-base origin-left"
          style={active ? { animation: "slot-balance-pop 0.6s cubic-bezier(0.18,1.4,0.4,1) both" } : undefined}
        >
          <span ref={numRef}>{fmt(value)}</span>
        </div>
      </div>
    </div>
  );
}

export function WinBar() {
  const displayWin = useSlot((s) => s.displayWin);
  const lastSpinWin = useSlot((s) => s.lastSpinWin);
  const freeSpinsWin = useSlot((s) => s.freeSpinsWin);
  const inFreeSpins = useSlot((s) => s.inFreeSpins);
  const freeSpins = useSlot((s) => s.freeSpins);
  const freeSpinsTotal = useSlot((s) => s.freeSpinsTotal);
  const spinKey = useSlot((s) => s.spinKey);
  const phase = useSlot((s) => s.phase);

  // Show animated displayWin during spin, else lastSpinWin
  const showWin = phase === "celebrating" || phase === "tumbling" || phase === "ended" ? displayWin : lastSpinWin;

  return (
    <div className="mt-2.5 flex items-stretch gap-2">
      <StatTile
        label="SPIN WIN"
        value={showWin}
        active={showWin > 0 || lastSpinWin > 0}
        accent="#f9a8d4"
        icon="🎯"
        pulseKey={spinKey}
      />
      <StatTile
        label={inFreeSpins ? `BONUS WIN · ${freeSpins}/${freeSpinsTotal}` : "BONUS WIN"}
        value={freeSpinsWin}
        active={inFreeSpins}
        accent="#fde047"
        icon="🍭"
        pulseKey={freeSpinsTotal}
      />
    </div>
  );
}
