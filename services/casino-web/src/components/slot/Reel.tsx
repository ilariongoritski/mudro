"use client";

import { memo, useLayoutEffect, useMemo, useRef, useState } from "react";
import { motion } from "framer-motion";
import { SymbolTile } from "./SymbolTile";
import { randomSymbol } from "@/lib/slot/engine";
import type { SymbolId } from "@/lib/slot/config";

interface ReelProps {
  reelIndex: number;
  prev: SymbolId[];
  result: SymbolId[];
  spinKey: number;
  spinDelay: number;
  spinDuration: number;
  turbo: boolean;
  anticipate: boolean;
  winningSet: Set<string>;
  showWins: boolean;
  onReelStop: () => void;
}

const FILLERS_BASE = 24;

function buildStrip(prev: SymbolId[], result: SymbolId[], reelIndex: number) {
  const m = FILLERS_BASE + reelIndex * 6;
  const fillers: SymbolId[] = [];
  for (let i = 0; i < m; i++) fillers.push(randomSymbol());
  return [...prev, ...fillers, ...result];
}

interface ReelSpinProps extends ReelProps {
  step: number;
}

function ReelSpin({
  reelIndex,
  prev,
  result,
  spinKey,
  spinDelay,
  spinDuration,
  anticipate,
  winningSet,
  showWins,
  onReelStop,
  step,
}: ReelSpinProps) {
  const strip = useMemo(
    () => (spinKey === 0 ? prev : buildStrip(prev, result, reelIndex)),
    [spinKey, prev, result, reelIndex]
  );

  const target = (strip.length - 3) * step;
  const resultStart = strip.length - 3;
  const animating = spinKey > 0;

  const handleComplete = () => {
    if (spinKey === 0) return;
    onReelStop();
  };

  return (
    <motion.div
      className="flex flex-col will-change-transform"
      initial={{ y: 0, filter: "blur(0px)" }}
      animate={
        animating
          ? {
              y: -target,
              filter: ["blur(0px)", "blur(1.4px)", "blur(1.4px)", "blur(0px)"],
            }
          : { y: 0, filter: "blur(0px)" }
      }
      transition={
        animating
          ? {
              duration: spinDuration / 1000,
              delay: spinDelay / 1000,
              ease: [0.08, 0.62, 0.18, 1],
              times: [0, 0.08, 0.85, 1],
            }
          : { duration: 0 }
      }
      onAnimationComplete={handleComplete}
    >
      {strip.map((s, i) => {
        const isResult = i >= resultStart;
        const row = isResult ? i - resultStart : -1;
        const winning =
          isResult && showWins && winningSet.has(`${reelIndex}-${row}`);
        const dim = isResult && showWins && !winning;
        return (
          <div
            key={i}
            className="w-full flex items-center justify-center"
            style={{ height: "var(--cell, 92px)" }}
          >
            <SymbolTile
              symbol={s}
              winning={winning}
              dim={dim}
              anticipate={anticipate && isResult}
            />
          </div>
        );
      })}
    </motion.div>
  );
}

export const Reel = memo(function Reel(props: ReelProps) {
  const windowRef = useRef<HTMLDivElement>(null);
  const [step, setStep] = useState(92);

  useLayoutEffect(() => {
    const el = windowRef.current;
    if (!el) return;
    const measure = () => {
      const h = el.getBoundingClientRect().height;
      if (h > 0) setStep(h / 3);
    };
    measure();
    const ro = new ResizeObserver(measure);
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

  return (
    <div
      ref={windowRef}
      className="relative overflow-hidden slot-reel-window"
      style={{ height: "calc(var(--cell, 92px) * 3)" }}
    >
      <div className="absolute inset-x-0 top-0 h-8 z-20 pointer-events-none slot-fade-top" />
      <div className="absolute inset-x-0 bottom-0 h-8 z-20 pointer-events-none slot-fade-bottom" />
      <ReelSpin key={props.spinKey} {...props} step={step} />
      {props.anticipate && (
        <div className="absolute inset-0 z-10 pointer-events-none slot-anticipate-glow" />
      )}
    </div>
  );
});
