"use client";

import { SYMBOLS, type SymbolId } from "@/lib/slot/config";
import { cn } from "@/lib/utils";

interface Props {
  symbol: SymbolId;
  winning?: boolean;
  dim?: boolean;
  anticipate?: boolean;
}

export function SymbolTile({ symbol, winning, dim, anticipate }: Props) {
  const def = SYMBOLS[symbol];
  const c = def.color;
  const g = def.glow;

  return (
    <div
      className={cn(
        "relative w-full h-full flex items-center justify-center select-none",
        winning && "z-10"
      )}
      style={{
        opacity: dim ? 0.32 : 1,
        filter: dim ? "saturate(0.6)" : "none",
        transition: "opacity .3s ease, filter .3s ease, transform .3s ease",
        transform: winning ? "scale(1.06)" : "scale(1)",
      }}
    >
      <div
        className={cn(
          "relative rounded-[22%] overflow-hidden flex items-center justify-center"
        )}
        style={{
          width: "86%",
          height: "86%",
          background: `radial-gradient(120% 120% at 50% 16%, ${g}40 0%, ${c} 42%, ${c}cc 68%, #0a0f1e 135%)`,
          boxShadow: winning
            ? `0 0 0 2px ${g}, 0 0 26px 5px ${g}cc, inset 0 0 20px ${g}77`
            : `inset 0 -10px 18px rgba(0,0,0,.45), inset 0 7px 14px rgba(255,255,255,.14), 0 2px 6px rgba(0,0,0,.35)`,
          transition: "box-shadow .3s ease",
        }}
      >
        {/* top sheen */}
        <div
          className="absolute inset-x-0 top-0 h-1/2 opacity-70 pointer-events-none"
          style={{
            background:
              "linear-gradient(to bottom, rgba(255,255,255,.28), rgba(255,255,255,0))",
          }}
        />
        {/* shimmer for special */}
        {def.tier === "special" && (
          <div className="absolute inset-0 slot-shimmer pointer-events-none" />
        )}
        {/* emoji */}
        <span
          className="relative leading-none"
          style={{
            fontSize: "calc(var(--cell, 92px) * 0.46)",
            filter: `drop-shadow(0 2px 3px rgba(0,0,0,.55)) drop-shadow(0 0 10px ${g}aa)`,
          }}
        >
          {def.emoji}
        </span>
        {/* label badge for special */}
        {def.tier === "special" && (
          <span
            className="absolute bottom-[5%] left-1/2 -translate-x-1/2 px-1 rounded-full font-black tracking-wider"
            style={{
              background: g,
              color: "#0a0f1e",
              fontSize: "calc(var(--cell, 92px) * 0.085)",
              lineHeight: 1.3,
            }}
          >
            {def.label}
          </span>
        )}
        {/* anticipate ring */}
        {anticipate && (
          <div className="absolute inset-0 rounded-[22%] ring-2 ring-amber-300/80 slot-anticipate-ring pointer-events-none" />
        )}
      </div>
      {winning && (
        <div className="absolute inset-0 pointer-events-none slot-win-pulse rounded-[22%]" />
      )}
    </div>
  );
}
