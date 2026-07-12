"use client";

import { useState } from "react";
import { loginWithTelegram } from "@/lib/auth";
import { Button } from "@/components/ui/button";

export function TelegramLoginButton({ className }: { className?: string }) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleLogin = async () => {
    setLoading(true);
    setError(null);
    try {
      await loginWithTelegram();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Telegram login failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={className}>
      <Button
        type="button"
        size="lg"
        className="w-full text-lg py-6 bg-[#54a9eb] hover:bg-[#4a9ad6]"
        disabled={loading}
        onClick={handleLogin}
      >
        {loading ? "Signing in…" : "📱 Login with Telegram"}
      </Button>
      {error && <p role="alert" className="mt-3 text-sm text-red-300">{error}</p>}
    </div>
  );
}
