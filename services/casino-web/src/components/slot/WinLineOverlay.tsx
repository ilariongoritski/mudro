"use client";

import { memo } from "react";
import type { LineWin } from "@/lib/slot/engine";

interface Props {
  show: boolean;
  wins: LineWin[];
}

/**
 * Draws all winning paylines as glowing gold polylines over the 5x3 grid.
 * Coordinates use a 0..100 viewBox mapped to the grid (5 cols x 3 rows),
 * with non-scaling strokes so line width stays crisp at any size.
 */
export const WinLineOverlay = memo(function WinLineOverlay({
  show,
  wins,
}: Props) {
  if (!show || wins.length === 0) return null;

  const points = (w: LineWin) =>
    w.positions
      .map(([reel, row]) => `${(reel + 0.5) * 20},${(row + 0.5) * (100 / 3)}`)
      .join(" ");

  return (
    <svg
      className="absolute inset-0 w-full h-full pointer-events-none z-30"
      viewBox="0 0 100 100"
      preserveAspectRatio="none"
    >
      <defs>
        <filter id="winGlow" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="1.4" result="b" />
          <feMerge>
            <feMergeNode in="b" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>
      {wins.map((w, i) => (
        <polyline
          key={`${w.lineIndex}-${i}`}
          points={points(w)}
          fill="none"
          stroke="#fde047"
          strokeWidth={2.4}
          strokeLinejoin="round"
          strokeLinecap="round"
          vectorEffect="non-scaling-stroke"
          filter="url(#winGlow)"
          opacity={0.92}
          className="slot-winline"
        />
      ))}
    </svg>
  );
});
