"use client";

import { AnimatePresence, motion } from "framer-motion";
import { TumbleTile } from "./TumbleTile";
import { useSlot } from "@/lib/slot/store";
import { REELS, ROWS } from "@/lib/slot/config";
import { anticipationReel, type Cell } from "@/lib/slot/engine";
import { useMemo } from "react";

export function TumbleGrid() {
  const board = useSlot((s) => s.board);
  const spinKey = useSlot((s) => s.spinKey);
  const tumbleKey = useSlot((s) => s.tumbleKey);
  const winningPositions = useSlot((s) => s.winningPositions);
  const phase = useSlot((s) => s.phase);
  const scatterCount = useSlot((s) => s.scatterCount);
  const inFreeSpins = useSlot((s) => s.inFreeSpins);
  const turbo = useSlot((s) => s.turbo);

  const showWins = phase === "celebrating" && winningPositions.size > 0;
  const anticipateReelIdx =
    phase === "dropping" && scatterCount >= 3 && scatterCount < 4
      ? anticipationReel(scatterCount)
      : -1;

  // Drop animation config
  const dropDuration = turbo ? 0.32 : 0.5;
  const tumbleDuration = turbo ? 0.28 : 0.42;

  // Whether this render is an initial spin drop (vs a tumble step)
  const isInitialDrop = phase === "dropping" && tumbleKey === spinKey + 1;

  const columns = useMemo(() => board, [board]);

  return (
    <div
      className="relative slot-grid rounded-2xl p-2 sm:p-2.5"
      style={
        { "--cell": "clamp(48px, 13.5vw, 72px)" } as React.CSSProperties
      }
    >
      <div
        className="relative grid gap-1.5 sm:gap-2"
        style={{ gridTemplateColumns: `repeat(${REELS}, minmax(0, 1fr))` }}
      >
        {Array.from({ length: REELS }).map((_, reelIdx) => {
          const col: Cell[] = columns[reelIdx] ?? [];
          const anticipate = reelIdx === anticipateReelIdx;
          return (
            <div
              key={reelIdx}
              className="relative slot-reel-col rounded-xl overflow-hidden"
              style={{ height: "calc(var(--cell, 64px) * 5)" }}
            >
              {/* edge fades */}
              <div className="absolute inset-x-0 top-0 h-6 z-20 pointer-events-none slot-fade-top" />
              <div className="absolute inset-x-0 bottom-0 h-6 z-20 pointer-events-none slot-fade-bottom" />

              <div className="relative flex flex-col h-full">
                <AnimatePresence mode="popLayout">
                  {col.map((cell, rowIdx) => {
                    const key = cell.id;
                    const winning = showWins && winningPositions.has(`${reelIdx}-${rowIdx}`);
                    const dim = showWins && !winning;
                    const isFill = rowIdx < ROWS - col.length + (ROWS - col.length); // unused
                    return (
                      <motion.div
                        key={key}
                        layout
                        className="w-full flex items-center justify-center"
                        style={{ height: "var(--cell, 64px)", flexShrink: 0 }}
                        initial={{ y: -160, opacity: 0, scale: 0.8 }}
                        animate={{ y: 0, opacity: 1, scale: 1 }}
                        exit={{
                          scale: 0,
                          opacity: 0,
                          rotate: 90,
                          transition: { duration: 0.28, ease: "easeIn" },
                        }}
                        transition={{
                          layout: {
                            duration: tumbleDuration,
                            ease: [0.25, 0.8, 0.3, 1],
                          },
                          default: {
                            duration: isInitialDrop ? dropDuration : tumbleDuration,
                            delay: isInitialDrop ? reelIdx * 0.05 : 0,
                            ease: [0.2, 0.7, 0.3, 1],
                          },
                        }}
                      >
                        <TumbleTile
                          symbol={cell.symbol}
                          mult={cell.mult}
                          winning={!!winning}
                          dim={!!dim}
                        />
                      </motion.div>
                    );
                  })}
                </AnimatePresence>
              </div>

              {anticipate && (
                <div className="absolute inset-0 z-10 pointer-events-none slot-anticipate-glow" />
              )}
            </div>
          );
        })}
      </div>

      {/* Free-spins ambient glow on the grid border */}
      {inFreeSpins && (
        <div className="absolute inset-0 rounded-2xl pointer-events-none slot-fs-grid-glow" />
      )}
    </div>
  );
}
