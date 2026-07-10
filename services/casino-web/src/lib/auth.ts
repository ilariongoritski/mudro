import { useSlot } from "./slot/store";

const AUTH_API = process.env.NEXT_PUBLIC_AUTH_API_URL || "http://localhost:8080";

export interface TelegramUser {
  id: number;
  first_name: string;
  last_name?: string;
  username?: string;
  photo_url?: string;
}

export interface AuthResponse {
  token: string;
  user: {
    id: number;
    username: string;
    telegram_id?: number;
  };
}

// Real Telegram WebApp login
export async function loginWithTelegram(initData?: string): Promise<AuthResponse> {
  let dataToSend = initData;

  // If running inside Telegram WebApp, get real initData
  if (typeof window !== "undefined" && (window as any).Telegram?.WebApp) {
    const tg = (window as any).Telegram.WebApp;
    dataToSend = tg.initData;
    console.log("[Telegram] Using real initData from WebApp");
  }

  if (!dataToSend) {
    throw new Error("No Telegram initData available");
  }

  const res = await fetch(`${AUTH_API}/api/auth/telegram`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ initData: dataToSend }),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || "Telegram login failed");
  }

  const data: AuthResponse = await res.json();

  const store = useSlot.getState();
  store.setAuth(data.token, data.user);

  return data;
}

export function logout() {
  const store = useSlot.getState();
  store.clearAuth();
  localStorage.removeItem("mudro_token");
  localStorage.removeItem("mudro_user");
}

export function restoreAuthFromStorage() {
  const token = localStorage.getItem("mudro_token");
  const userStr = localStorage.getItem("mudro_user");
  if (token && userStr) {
    try {
      const user = JSON.parse(userStr);
      useSlot.getState().setAuth(token, user);
    } catch {}
  }
}

// Helper to check if running inside Telegram
export function isTelegramWebApp(): boolean {
  return typeof window !== "undefined" && !!(window as any).Telegram?.WebApp;
}
