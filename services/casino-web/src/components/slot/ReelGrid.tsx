"use client";

import { useCallback, useEffect, useMemo, useRef } from "react";
import { Reel } from "./Reel";
import { WinLineOverlay } from "./WinLineOverlay";
import { useSlot } from "@/lib/slot/store";
import { anticipationReel, type Grid, countScatters } from "@/lib/slot/engine";
import { REELS } from "@/lib/slot/config";
import { sound } from "@/lib/slot/sound";

const BASE_DURATION = [680, 800, 920, 1040, 1180];
const BASE_DELAY = [0, 80, 160, 240, 320];

export function ReelGrid() {
  const grid = useSlot((s) => s.grid);
  const pendingResult = useSlot((s) => s.pendingResult);
  const spinning = useSlot((s) => s.spinning);
  const spinKey = useSlot((s) => s.spinKey);
  const turbo = useSlot((s) => s.turbo);
  const lastResult = useSlot((s) => s.lastResult);
  const commitSpin = useSlot((s) => s.commitSpin);

  const stoppedRef = useRef(0);

  useEffect(() => {
    stoppedRef.current = 0;
  }, [spinKey]);

  const handleReelStop = useCallback(() => {
    if (useSlot.getState().soundOn) sound.reelStop();
    stoppedRef.current += 1;
    if (stoppedRef.current >= REELS) {
      stoppedRef.current = 0;
      commitSpin();
    }
  }, [commitSpin]);

  const activeGrid: Grid = pendingResult?.grid ?? grid;
  const anticipateFrom =
    spinning && pendingResult ? anticipationReel(countScatters(pendingResult.grid)) : -1;

  // Extract symbols from grid cells for Reel component
  const prevGrid = useMemo(() => grid.map(col => col.map(c => c.symbol)), [grid]);
  const resultGrid = useMemo(() => activeGrid.map(col => col.map(c => c.symbol)), [activeGrid]);

  // winning positions (stabilized)
  const winningSet = useMemo(() => {
    const set = new Set<string>();
    if (lastResult && lastResult.totalWin > 0) {
      for (const w of lastResult.wins)
        for (const [r, row] of w.positions) set.add(`${r}-${row}`);
      for (const [r, row] of lastResult.scatterPositions)
        set.add(`${r}-${row}`);
    }
    return set;
  }, [lastResult]);
  const showWins = !spinning && !!lastResult && lastResult.totalWin > 0;

  return (
    <div
      className="relative slot-grid rounded-2xl p-2 sm:p-3"
      style={{
        "--cell": "clamp(58px, 15.5vw, 96px)",
      } as React.CSSProperties}
    >
      <div className="relative">
        <div className="flex gap-1.5 sm:gap-2">
          {Array.from({ length: REELS }).map((_, i) => {
            const prev = prevGrid[i] ?? ["seven", "diamond", "crown"];
            const result = resultGrid[i] ?? prev;
            const anticipate = anticipateFrom >= 0 && i >= anticipateFrom;
            const dur = BASE_DURATION[i] * (turbo ? 0.5 : 1) + (anticipate ? 480 : 0);
            const delay = BASE_DELAY[i] * (turbo ? 0.5 : 1) + (anticipate ? 150 : 0);
            return (
              <div key={i} className="relative flex-1 slot-reel-col">
                <Reel
                  reelIndex={i}
                  prev={prev}
                  result={result}
                  spinKey={spinKey}
                  spinDelay={delay}
                  spinDuration={dur}
                  turbo={turbo}
                  anticipate={anticipate}
                  winningSet={winningSet}
                  showWins={showWins}
                  onReelStop={handleReelStop}
                />
              </div>
            );
          })}
        </div>

        <WinLineOverlay show={showWins} wins={lastResult?.wins ?? []} />
      </div>
    </div>
  );
}
