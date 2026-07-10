"use client";

import { useSlot } from "@/lib/slot/store";

function fmt(n: number) {
  return n.toLocaleString("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

export function Banners() {
  const showBanner = useSlot((s) => s.showFreeSpinsBanner);
  const showEnd = useSlot((s) => s.showFreeSpinsEnd);
  const endWin = useSlot((s) => s.freeSpinsEndWin);
  const freeSpinsTotal = useSlot((s) => s.freeSpinsTotal);

  return (
    <>
      {showBanner && (
        <div className="absolute inset-0 z-40 pointer-events-none flex items-center justify-center">
          <div className="slot-banner-enter flex flex-col items-center text-center px-6">
            {/* spinning lollipops */}
            <div className="flex gap-3 mb-3">
              <span className="text-4xl sm:text-5xl slot-spin-slow">🍭</span>
              <span className="text-5xl sm:text-7xl slot-float">🍭</span>
              <span className="text-4xl sm:text-5xl slot-spin-slow" style={{ animationDirection: "reverse" }}>🍭</span>
            </div>
            <div
              className="font-black tracking-tight leading-none"
              style={{
                fontSize: "clamp(30px, 7.5vw, 64px)",
                color: "#f9a8d4",
                textShadow:
                  "0 0 30px #f472b6cc, 0 0 60px #ec489977, 0 4px 10px rgba(0,0,0,.5)",
              }}
            >
              BONUS!
            </div>
            <div
              className="font-black mt-1"
              style={{
                fontSize: "clamp(20px, 5vw, 38px)",
                color: "#fde047",
                textShadow: "0 0 18px #fde047aa",
              }}
            >
              {freeSpinsTotal} FREE SPINS
            </div>
            <div className="mt-3 px-4 py-1.5 rounded-full bg-pink-500/20 border border-pink-400/50 text-pink-100 font-bold text-xs sm:text-sm tracking-widest flex items-center gap-2">
              <span>💣</span>
              MULTIPLIER BOMBS ACTIVE
            </div>
          </div>
        </div>
      )}

      {showEnd && (
        <div className="absolute inset-0 z-40 pointer-events-none flex items-center justify-center">
          <div className="slot-banner-enter flex flex-col items-center text-center px-6">
            <div className="flex gap-2 mb-2 text-4xl sm:text-5xl">
              <span className="slot-float">🎉</span>
              <span className="slot-float" style={{ animationDelay: "0.15s" }}>🍭</span>
              <span className="slot-float" style={{ animationDelay: "0.3s" }}>💰</span>
            </div>
            <div
              className="font-black tracking-tight"
              style={{
                fontSize: "clamp(24px, 5.5vw, 48px)",
                color: "#f9a8d4",
                textShadow: "0 0 26px #f472b6aa",
              }}
            >
              BONUS COMPLETE
            </div>
            <div className="text-pink-200 font-bold mt-2 text-xs sm:text-sm tracking-widest">
              TOTAL BONUS WIN
            </div>
            <div
              className="font-black tabular-nums"
              style={{
                fontSize: "clamp(28px, 6vw, 50px)",
                color: "#fde047",
                textShadow: "0 0 20px #fde047aa",
              }}
            >
              {fmt(endWin)}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
