"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  Minus,
  Plus,
  Zap,
  RotateCw,
  Square,
  Infinity as InfinityIcon,
  ChevronDown,
  Sparkles,
} from "lucide-react";
import {
  BET_PRESETS,
  BONUS_BUY_MULT,
} from "@/lib/slot/config";
import { useSlot } from "@/lib/slot/store";
import { sound } from "@/lib/slot/sound";
import { cn } from "@/lib/utils";

function fmt(n: number) {
  return n.toLocaleString("en-US", {
    minimumFractionDigits: n < 10 ? 2 : 0,
    maximumFractionDigits: 2,
  });
}

export function ControlPanel() {
  const bet = useSlot((s) => s.bet);
  const balance = useSlot((s) => s.balance);
  const phase = useSlot((s) => s.phase);
  const turbo = useSlot((s) => s.turbo);
  const inFreeSpins = useSlot((s) => s.inFreeSpins);
  const freeSpins = useSlot((s) => s.freeSpins);
  const autoSpins = useSlot((s) => s.autoSpins);
  const autoInfinite = useSlot((s) => s.autoInfinite);
  const soundOn = useSlot((s) => s.soundOn);

  const spin = useSlot((s) => s.spin);
  const stopAuto = useSlot((s) => s.stopAuto);
  const setBet = useSlot((s) => s.setBet);
  const incBet = useSlot((s) => s.incBet);
  const decBet = useSlot((s) => s.decBet);
  const toggleTurbo = useSlot((s) => s.toggleTurbo);
  const setAuto = useSlot((s) => s.setAuto);
  const setAutoInfinite = useSlot((s) => s.setAutoInfinite);
  const buyBonus = useSlot((s) => s.buyBonus);

  const [autoOpen, setAutoOpen] = useState(false);
  const [betOpen, setBetOpen] = useState(false);
  const [buyOpen, setBuyOpen] = useState(false);

  const autoActive = autoSpins > 0 || autoInfinite;
  const busy = phase !== "idle" && phase !== "ended";
  const canChangeBet = !busy && !autoActive && !inFreeSpins;

  const buyPrice = BONUS_BUY_MULT * bet;
  const canBuyBonus = !busy && !inFreeSpins && !autoActive && balance >= buyPrice;

  const handleSpin = () => {
    if (inFreeSpins) return;
    if (autoActive) {
      if (soundOn) sound.click();
      stopAuto();
      return;
    }
    if (busy) return;
    if (balance < bet) return;
    if (soundOn) sound.click();
    spin();
  };

  const handleBuyBonus = () => {
    if (soundOn) sound.click();
    setBuyOpen(false);
    buyBonus();
  };

  const autoLabel = autoActive
    ? autoInfinite
      ? "∞"
      : String(autoSpins)
    : "AUTO";

  return (
    <div className="mt-3 sm:mt-4 space-y-2">
      <div className="flex items-center justify-between gap-2 sm:gap-3 rounded-xl bg-[#0c1322]/80 border border-white/10 px-2 py-2 sm:px-4 sm:py-3 backdrop-blur">
        {/* BET */}
        <div className="flex items-center gap-1.5 sm:gap-2 min-w-0">
          <span className="hidden sm:block text-[10px] font-bold tracking-widest text-pink-400/80">
            BET
          </span>
          <Button
            variant="ghost"
            size="icon"
            className="size-8 sm:size-9 rounded-lg bg-white/5 hover:bg-white/10 text-white border border-white/10 disabled:opacity-40"
            disabled={!canChangeBet}
            onClick={() => decBet()}
          >
            <Minus className="size-4" />
          </Button>
          <Popover open={betOpen} onOpenChange={setBetOpen}>
            <PopoverTrigger asChild>
              <button
                disabled={!canChangeBet}
                className={cn(
                  "h-9 sm:h-10 min-w-[78px] sm:min-w-[96px] px-2 rounded-lg bg-[#0a0f1e] border border-pink-500/30 text-white font-bold text-sm sm:text-base flex items-center justify-center gap-1 transition",
                  canChangeBet
                    ? "hover:border-pink-400/60 cursor-pointer"
                    : "opacity-70 cursor-not-allowed"
                )}
              >
                <span className="text-pink-400 text-xs sm:text-sm">
                  {fmt(bet)}
                </span>
                <ChevronDown className="size-3 text-pink-400/60" />
              </button>
            </PopoverTrigger>
            <PopoverContent className="w-44 p-2 bg-[#0c1322] border-white/10">
              <div className="grid grid-cols-3 gap-1.5">
                {BET_PRESETS.map((b) => (
                  <button
                    key={b}
                    onClick={() => {
                      setBet(b);
                      setBetOpen(false);
                    }}
                    className={cn(
                      "py-1.5 rounded-md text-xs font-bold border transition",
                      b === bet
                        ? "bg-pink-500 text-[#1a0a2e] border-pink-400"
                        : "bg-white/5 text-white border-white/10 hover:bg-white/10"
                    )}
                  >
                    {fmt(b)}
                  </button>
                ))}
              </div>
            </PopoverContent>
          </Popover>
          <Button
            variant="ghost"
            size="icon"
            className="size-8 sm:size-9 rounded-lg bg-white/5 hover:bg-white/10 text-white border border-white/10 disabled:opacity-40"
            disabled={!canChangeBet}
            onClick={() => incBet()}
          >
            <Plus className="size-4" />
          </Button>
        </div>

        {/* TURBO (with tooltip explaining the difference) */}
        <TooltipProvider delayDuration={250}>
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                onClick={() => {
                  if (soundOn) sound.click();
                  toggleTurbo();
                }}
                className={cn(
                  "flex items-center gap-1.5 h-9 px-2.5 sm:px-3 rounded-lg border text-xs font-bold tracking-wide transition",
                  turbo
                    ? "bg-amber-500/20 border-amber-400/60 text-amber-300"
                    : "bg-white/5 border-white/10 text-slate-300 hover:bg-white/10"
                )}
              >
                <Zap
                  className={cn(
                    "size-4",
                    turbo && "fill-amber-400 text-amber-400"
                  )}
                />
                <span className="hidden sm:inline">TURBO</span>
              </button>
            </TooltipTrigger>
            <TooltipContent className="max-w-[220px] bg-[#1a0a2e] border-pink-500/30 text-slate-200 text-xs">
              <b className="text-amber-300">Turbo ON:</b> ~2× faster — shorter
              drops, quicker tumbles & auto-spins.
              <br />
              <b className="text-slate-300">Turbo OFF:</b> full cinematic timing.
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>

        {/* SPIN */}
        <button
          onClick={handleSpin}
          disabled={inFreeSpins || (busy && !autoActive)}
          className={cn(
            "relative shrink-0 size-16 sm:size-20 rounded-full flex flex-col items-center justify-center font-black text-sm sm:text-base transition-all duration-150 select-none",
            "bg-gradient-to-b from-pink-400 to-fuchsia-600 text-white shadow-[0_8px_24px_rgba(236,72,153,.5)]",
            "hover:from-pink-300 hover:to-fuchsia-500 hover:shadow-[0_10px_30px_rgba(236,72,153,.65)] active:scale-95",
            "border-2 border-pink-300/50",
            (inFreeSpins || (busy && !autoActive)) &&
              "opacity-60 cursor-not-allowed",
            autoActive &&
              !inFreeSpins &&
              "from-rose-400 to-rose-600 text-white border-rose-300/50 shadow-[0_8px_24px_rgba(244,63,94,.45)]"
          )}
        >
          {inFreeSpins ? (
            <>
              <span className="text-[9px] sm:text-[10px] leading-none">
                FREE
              </span>
              <span className="text-base sm:text-lg leading-none">
                {freeSpins}
              </span>
            </>
          ) : autoActive ? (
            <>
              <Square className="size-5 sm:size-6 fill-current" />
              <span className="text-[9px] sm:text-[10px] mt-0.5">STOP</span>
            </>
          ) : busy ? (
            <RotateCw className="size-6 sm:size-7 animate-spin" />
          ) : (
            <>
              <span className="text-base sm:text-lg leading-none">SPIN</span>
              <span className="text-[8px] sm:text-[9px] font-bold opacity-70 mt-0.5">
                {fmt(bet)}
              </span>
            </>
          )}
        </button>

        {/* AUTO */}
        <Popover open={autoOpen} onOpenChange={setAutoOpen}>
          <PopoverTrigger asChild>
            <button
              disabled={inFreeSpins || autoActive}
              className={cn(
                "flex items-center gap-1.5 h-9 px-2.5 sm:px-3 rounded-lg border text-xs font-bold tracking-wide transition min-w-[64px] justify-center",
                autoActive
                  ? "bg-rose-500/20 border-rose-400/60 text-rose-300"
                  : "bg-white/5 border-white/10 text-slate-300 hover:bg-white/10",
                (inFreeSpins || autoActive) && "opacity-40 cursor-not-allowed"
              )}
            >
              {autoActive ? (
                <span className="flex items-center gap-1">
                  <RotateCw className="size-3.5 animate-spin" />
                  {autoLabel}
                </span>
              ) : (
                <>
                  <InfinityIcon className="size-4" />
                  <span className="hidden sm:inline">AUTO</span>
                </>
              )}
            </button>
          </PopoverTrigger>
          <PopoverContent
            className="w-44 p-2 bg-[#0c1322] border-white/10"
            align="end"
          >
            <div className="text-[10px] font-bold tracking-widest text-slate-400 px-1 pb-1.5">
              AUTO SPINS
            </div>
            <div className="grid grid-cols-3 gap-1.5">
              {[10, 25, 50, 100].map((n) => (
                <button
                  key={n}
                  onClick={() => {
                    setAuto(n);
                    setAutoOpen(false);
                  }}
                  className="py-1.5 rounded-md text-xs font-bold bg-white/5 text-white border border-white/10 hover:bg-pink-500/20 hover:border-pink-400/40 transition"
                >
                  {n}
                </button>
              ))}
              <button
                onClick={() => {
                  setAutoInfinite();
                  setAutoOpen(false);
                }}
                className="py-1.5 rounded-md text-xs font-bold bg-white/5 text-white border border-white/10 hover:bg-pink-500/20 hover:border-pink-400/40 transition flex items-center justify-center"
              >
                <InfinityIcon className="size-4" />
              </button>
            </div>
            <div className="text-[10px] text-slate-500 px-1 pt-2 leading-relaxed">
              Auto stops on insufficient balance.
            </div>
          </PopoverContent>
        </Popover>
      </div>

      {/* BUY BONUS row */}
      <Popover open={buyOpen} onOpenChange={setBuyOpen}>
        <PopoverTrigger asChild>
          <button
            disabled={!canBuyBonus}
            className={cn(
              "w-full flex items-center justify-center gap-2 h-10 rounded-xl border-2 font-black text-sm tracking-wide transition",
              "bg-gradient-to-r from-amber-500/25 via-pink-500/25 to-fuchsia-500/25",
              "border-amber-400/50 text-amber-200",
              "hover:from-amber-500/35 hover:via-pink-500/35 hover:to-fuchsia-500/35 hover:border-amber-300/70",
              !canBuyBonus && "opacity-40 cursor-not-allowed grayscale-[0.4]"
            )}
          >
            <Sparkles className="size-4 text-amber-300" />
            <span>BUY BONUS</span>
            <span className="text-amber-300/90 tabular-nums">
              {fmt(buyPrice)}
            </span>
            <span className="hidden sm:inline text-[10px] text-slate-400 font-bold">
              · 10 FREE SPINS + 💣 BOMBS
            </span>
          </button>
        </PopoverTrigger>
        <PopoverContent className="w-72 p-3 bg-[#1a0a2e] border-amber-400/40" align="center">
          <div className="text-center">
            <div className="text-amber-300 font-black text-sm flex items-center justify-center gap-1.5 mb-1">
              <Sparkles className="size-4" /> BUY FREE SPINS
            </div>
            <p className="text-slate-300 text-xs leading-relaxed mb-3">
              Pay{" "}
              <b className="text-amber-200">{fmt(buyPrice)}</b> ({BONUS_BUY_MULT}×
              bet) to instantly trigger the bonus round: 10 free spins with{" "}
              💣 multiplier bombs and retriggers.
            </p>
            <div className="flex gap-2">
              <Button
                variant="ghost"
                size="sm"
                className="flex-1 bg-white/5 text-slate-300 border border-white/10 hover:bg-white/10"
                onClick={() => setBuyOpen(false)}
              >
                Cancel
              </Button>
              <Button
                size="sm"
                className="flex-1 bg-gradient-to-b from-amber-400 to-amber-600 text-[#1a0a2e] font-black border-amber-300 hover:from-amber-300 hover:to-amber-500"
                onClick={handleBuyBonus}
              >
                Buy {fmt(buyPrice)}
              </Button>
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}
