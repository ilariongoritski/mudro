"use client";

import { AnimatePresence, motion, useReducedMotion } from "framer-motion";
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
  const serverLanded = useSlot((s) => s.serverLanded);
  const reduceMotion = useReducedMotion();

  const showWins = phase === "celebrating" && winningPositions.size > 0;
  const anticipateReelIdx =
    phase === "dropping" && scatterCount >= 3 && scatterCount < 4
      ? anticipationReel(scatterCount)
      : -1;

  const dropDuration = turbo || reduceMotion ? 0.26 : 0.54;
  const tumbleDuration = turbo || reduceMotion ? 0.22 : 0.38;
  // A new spin increments both keys together; tumbles advance only tumbleKey.
  const isInitialDrop = phase === "dropping" && tumbleKey === spinKey;
  const columns = useMemo(() => board, [board]);

  return (
    <div
      className="relative slot-grid rounded-2xl p-2 sm:p-2.5"
      style={{ "--cell": "clamp(48px, 13.5vw, 72px)" } as React.CSSProperties}
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
              <div className="absolute inset-x-0 top-0 h-6 z-20 pointer-events-none slot-fade-top" />
              <div className="absolute inset-x-0 bottom-0 h-6 z-20 pointer-events-none slot-fade-bottom" />

              <div className="relative flex flex-col h-full">
                <AnimatePresence mode="popLayout">
                  {col.map((cell, rowIdx) => {
                    const key = cell.id;
                    const winning = showWins && winningPositions.has(`${reelIdx}-${rowIdx}`);
                    const dim = showWins && !winning;
                    const landingDelay = isInitialDrop && !serverLanded ? reelIdx * (turbo ? 0.025 : 0.075) + rowIdx * (turbo ? 0.008 : 0.018) : 0;
                    return (
                      <motion.div
                        key={key}
                        layout={!reduceMotion}
                        className="w-full flex items-center justify-center will-change-transform"
                        style={{ height: "var(--cell, 64px)", flexShrink: 0 }}
                        initial={serverLanded ? false : reduceMotion ? { opacity: 0 } : { y: -220 - rowIdx * 26, opacity: 0, scale: 0.92 }}
                        animate={reduceMotion ? { opacity: 1 } : { y: [null, 8, -3, 0], opacity: 1, scale: [null, 1.015, 0.995, 1] }}
                        exit={{
                          scale: reduceMotion ? 1 : 0.72,
                          opacity: 0,
                          rotate: reduceMotion ? 0 : 12,
                          transition: { duration: turbo ? 0.14 : 0.22, ease: "easeIn" },
                        }}
                        transition={{
                          layout: { duration: tumbleDuration, ease: [0.18, 0.8, 0.24, 1] },
                          default: {
                            duration: isInitialDrop ? dropDuration : tumbleDuration,
                            delay: landingDelay,
                            times: reduceMotion ? undefined : [0, 0.76, 0.9, 1],
                            ease: reduceMotion ? "easeOut" : [0.18, 0.82, 0.22, 1],
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

              {anticipate && <div className="absolute inset-0 z-10 pointer-events-none slot-anticipate-glow" />}
            </div>
          );
        })}
      </div>

      {inFreeSpins && <div className="absolute inset-0 rounded-2xl pointer-events-none slot-fs-grid-glow" />}
    </div>
  );
}
