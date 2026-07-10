"use client";

import { SYMBOLS, type SymbolId } from "@/lib/slot/config";
import { cn } from "@/lib/utils";

interface Props {
  symbol: SymbolId;
  mult?: number;
  winning?: boolean;
  dim?: boolean;
}

export function TumbleTile({ symbol, mult, winning, dim }: Props) {
  const def = SYMBOLS[symbol];
  const c = def.color;
  const g = def.glow;
  const isBomb = def.isBomb;
  const isScatter = def.isScatter;

  return (
    <div
      className={cn(
        "relative w-full h-full flex items-center justify-center select-none",
        winning && "z-10",
        isBomb && "slot-bomb-float"
      )}
      style={{
        opacity: dim ? 0.35 : 1,
        filter: dim ? "saturate(0.55) brightness(0.8)" : "none",
        transition: "opacity .25s ease, filter .25s ease",
      }}
    >
      <div
        className={cn(
          "relative rounded-[26%] overflow-hidden flex items-center justify-center",
          winning && "slot-win-pop"
        )}
        style={{
          width: "90%",
          height: "90%",
          background: isBomb
            ? `radial-gradient(120% 120% at 50% 16%, #fffbeb 0%, ${c} 45%, #7c2d12 130%)`
            : `radial-gradient(120% 120% at 50% 14%, ${g}55 0%, ${c} 45%, ${c}cc 70%, #1a0f2e 135%)`,
          boxShadow: winning
            ? `0 0 0 2px #fff, 0 0 22px 4px ${g}, inset 0 0 16px ${g}88`
            : isBomb
              ? `inset 0 -6px 12px rgba(0,0,0,.35), inset 0 4px 10px rgba(255,255,255,.2), 0 0 14px ${g}66`
              : `inset 0 -8px 14px rgba(0,0,0,.4), inset 0 6px 12px rgba(255,255,255,.16), 0 2px 5px rgba(0,0,0,.35)`,
          transition: "box-shadow .25s ease",
        }}
      >
        {/* top sheen */}
        <div
          className="absolute inset-x-0 top-0 h-1/2 opacity-70 pointer-events-none"
          style={{
            background:
              "linear-gradient(to bottom, rgba(255,255,255,.3), rgba(255,255,255,0))",
          }}
        />
        {/* shimmer sweep for special / winning */}
        {(isBomb || isScatter || winning) && (
          <div className="absolute inset-0 slot-shimmer pointer-events-none" />
        )}
        {/* emoji */}
        <span
          className={cn(
            "relative leading-none",
            isBomb && "slot-bomb-pulse"
          )}
          style={{
            fontSize: "calc(var(--cell, 64px) * 0.5)",
            filter: `drop-shadow(0 2px 3px rgba(0,0,0,.55)) drop-shadow(0 0 8px ${g}aa)`,
          }}
        >
          {def.emoji}
        </span>

        {/* bomb fuse spark */}
        {isBomb && (
          <span className="absolute -top-1 left-1/2 -translate-x-1/2 text-[10px] slot-fuse-spark pointer-events-none">
            ✨
          </span>
        )}

        {/* bomb multiplier badge */}
        {isBomb && mult != null && (
          <span
            className="absolute -bottom-0.5 -right-0.5 px-1.5 rounded-full font-black leading-none border-2 border-yellow-200 slot-bomb-badge"
            style={{
              background: "#7c2d12",
              color: "#fde047",
              fontSize: "calc(var(--cell, 64px) * 0.2)",
            }}
          >
            ×{mult}
          </span>
        )}
        {/* scatter label */}
        {isScatter && (
          <span
            className="absolute bottom-[4%] left-1/2 -translate-x-1/2 px-1 rounded-full font-black tracking-wider"
            style={{
              background: g,
              color: "#3b0a2e",
              fontSize: "calc(var(--cell, 64px) * 0.082)",
              lineHeight: 1.3,
            }}
          >
            BONUS
          </span>
        )}
      </div>
      {winning && (
        <>
          <div className="absolute inset-0 pointer-events-none slot-win-pulse rounded-[26%]" />
          <div className="absolute inset-0 pointer-events-none slot-win-ring rounded-[26%]" />
        </>
      )}
    </div>
  );
}
