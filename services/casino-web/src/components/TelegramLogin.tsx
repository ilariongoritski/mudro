"use client";

import { useEffect, useRef, useState } from "react";
import { loginWithTelegram } from "@/lib/auth";

type LoginState = "loading" | "outside-telegram" | "error";

export function TelegramLoginButton() {
  const [state, setState] = useState<LoginState>("loading");
  const [message, setMessage] = useState("");
  const startedRef = useRef(false);

  useEffect(() => {
    if (startedRef.current) return;
    startedRef.current = true;

    const telegram = window.Telegram?.WebApp;
    const initData = telegram?.initData?.trim();
    if (!initData) {
      setState("outside-telegram");
      return;
    }

    telegram.ready();
    telegram.expand();
    void loginWithTelegram(initData)
      .catch((error: unknown) => {
        setState("error");
        setMessage(error instanceof Error ? error.message : "Telegram login failed");
      });
  }, []);

  if (state === "loading") {
    return <p className="text-sm text-slate-300">Connecting your Telegram account…</p>;
  }

  if (state === "outside-telegram") {
    return <p className="text-sm text-slate-300">Open this game from the MUDRO Telegram bot to sign in securely.</p>;
  }

  return <p className="text-sm text-rose-300">{message}</p>;
}
