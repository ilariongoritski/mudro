"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Volume2, VolumeX, Info, Wallet, LogOut, User } from "lucide-react";
import { useSlot } from "@/lib/slot/store";
import { sound } from "@/lib/slot/sound";
import { Paytable } from "./Paytable";
import { TelegramLoginButton } from "../TelegramLogin";
import { logout } from "@/lib/auth";
import { cn } from "@/lib/utils";

function fmt(n: number) {
  return n.toLocaleString("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

export function TopBar() {
  const balance = useSlot((s) => s.balance);
  const balancePulse = useSlot((s) => s.balancePulse);
  const soundOn = useSlot((s) => s.soundOn);
  const toggleSound = useSlot((s) => s.toggleSound);
  const phase = useSlot((s) => s.phase);
  const inFreeSpins = useSlot((s) => s.inFreeSpins);
  const isLoggedIn = useSlot((s) => s.isLoggedIn);
  const user = useSlot((s) => s.user);

  const [payOpen, setPayOpen] = useState(false);
  const busy = phase !== "idle" && phase !== "ended";

  const [showLogin, setShowLogin] = useState(false);

  const handleLogout = () => {
    logout();
    window.location.reload();
  };

  return (
    <header className="sticky top-0 z-50 border-b border-white/10 bg-black/60 backdrop-blur-xl">
      <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-3">
        {/* Left: Logo + User */}
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 text-xl font-black tracking-[-1px]">
            <span className="text-2xl">🍭</span>
            <span>Sweet Bonanza</span>
          </div>

          {isLoggedIn && user && (
            <div className="flex items-center gap-2 rounded-full bg-white/5 px-3 py-1 text-sm">
              <User className="h-4 w-4" />
              <span className="font-medium">{user.username}</span>
            </div>
          )}
        </div>

        {/* Center: Balance */}
        <div className="flex items-center gap-2">
          <div
            className={cn(
              "flex items-center gap-2 rounded-2xl border border-white/10 bg-white/5 px-4 py-2 font-mono text-lg font-bold",
              balancePulse && "animate-pulse"
            )}
          >
            <Wallet className="h-4 w-4 text-emerald-400" />
            <span>{fmt(balance)}</span>
          </div>
        </div>

        {/* Right: Controls */}
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            onClick={toggleSound}
            className="h-9 w-9 text-white/60 hover:text-white"
          >
            {soundOn ? <Volume2 className="h-4 w-4" /> : <VolumeX className="h-4 w-4" />}
          </Button>

          <Button
            variant="ghost"
            size="icon"
            onClick={() => setPayOpen(true)}
            className="h-9 w-9 text-white/60 hover:text-white"
          >
            <Info className="h-4 w-4" />
          </Button>

          {isLoggedIn ? (
            <Button
              variant="outline"
              size="sm"
              onClick={handleLogout}
              className="gap-2 border-white/20 hover:bg-white/5"
            >
              <LogOut className="h-4 w-4" />
              Logout
            </Button>
          ) : (
            <Button
              onClick={() => setShowLogin(true)}
              className="gap-2 bg-[#54a9eb] hover:bg-[#4a9ad6]"
            >
              <span>📱</span>
              Login with Telegram
            </Button>
          )}
        </div>
      </div>

      <Paytable open={payOpen} onOpenChange={setPayOpen} />

      {/* Login Modal */}
      {showLogin && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/70">
          <div className="w-full max-w-sm rounded-3xl border border-white/10 bg-zinc-950 p-8">
            <div className="mb-6 text-center">
              <div className="text-4xl mb-2">🍭</div>
              <h2 className="text-2xl font-bold">Welcome to Sweet Bonanza</h2>
              <p className="text-sm text-slate-400 mt-1">Login to play with real balance</p>
            </div>

            <TelegramLoginButton />

            <div className="mt-6 text-center">
              <Button
                variant="ghost"
                onClick={() => setShowLogin(false)}
                className="text-sm text-slate-400"
              >
                Cancel
              </Button>
            </div>
          </div>
        </div>
      )}
    </header>
  );
}
