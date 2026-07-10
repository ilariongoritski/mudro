"use client";

import { useEffect } from "react";
import { loginWithTelegram } from "@/lib/auth";
import { Button } from "@/components/ui/button";

declare global {
  interface Window {
    onTelegramAuth?: (user: any) => void;
  }
}

export function TelegramLoginButton() {
  useEffect(() => {
    // Load Telegram widget script
    const script = document.createElement("script");
    script.src = "https://telegram.org/js/telegram-widget.js?22";
    script.async = true;
    document.body.appendChild(script);

    // Global callback
    window.onTelegramAuth = async (telegramUser: any) => {
      try {
        // In real implementation we would send initData
        // For now we simulate with user data
        console.log("Telegram auth:", telegramUser);
        
        // TODO: Replace with real initData from Telegram WebApp
        const fakeInitData = JSON.stringify(telegramUser);
        await loginWithTelegram(fakeInitData);
        window.location.reload();
      } catch (e) {
        alert("Telegram login failed: " + (e as Error).message);
      }
    };

    return () => {
      document.body.removeChild(script);
      delete window.onTelegramAuth;
    };
  }, []);

  return (
    <div className="flex flex-col items-center gap-3">
      <div 
        id="telegram-login-container"
        className="telegram-login"
      >
        {/* Telegram will inject the button here */}
        <a 
          href="#"
          className="inline-flex items-center gap-2 px-6 py-3 bg-[#54a9eb] hover:bg-[#4a9ad6] text-white rounded-xl font-medium transition-colors"
          onClick={(e) => {
            e.preventDefault();
            // Fallback for demo
            alert("Telegram Login Widget would open here. In production it uses real Telegram auth.");
          }}
        >
          <span>📱</span>
          Login with Telegram
        </a>
      </div>
      <p className="text-xs text-slate-400 text-center max-w-[220px]">
        Fast & secure login. We only receive your Telegram ID and name.
      </p>
    </div>
  );
}
