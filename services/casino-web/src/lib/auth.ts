import { useSlot } from "./slot/store";

const authAPI = process.env.NEXT_PUBLIC_AUTH_API_URL ?? "";
export const telegramAuthEndpoint = "/api/v1/auth/telegram";

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

export async function loginWithTelegram(initData?: string): Promise<AuthResponse> {
  let dataToSend = initData;

  if (typeof window !== "undefined" && window.Telegram?.WebApp) {
    const telegram = window.Telegram.WebApp;
    telegram.ready();
    dataToSend = telegram.initData;
  }

  if (!dataToSend) {
    throw new Error("Open the game from Telegram to sign in");
  }

  const res = await fetch(`${authAPI}${telegramAuthEndpoint}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ initData: dataToSend }),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || "Telegram login failed");
  }

  const data: AuthResponse = await res.json();
  useSlot.getState().setAuth(data.token, data.user);
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
      useSlot.getState().setAuth(token, JSON.parse(userStr));
    } catch {
      // Ignore corrupted persisted auth state.
    }
  }
}

export function isTelegramWebApp(): boolean {
  return typeof window !== "undefined" && !!window.Telegram?.WebApp;
}
