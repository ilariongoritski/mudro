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

export async function loginWithTelegram(initData: string): Promise<AuthResponse> {
  const res = await fetch(`${AUTH_API}/api/auth/telegram`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ initData }),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || "Telegram login failed");
  }

  const data: AuthResponse = await res.json();

  // Save to store
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
