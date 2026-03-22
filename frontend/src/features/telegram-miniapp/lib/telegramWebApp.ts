import type { TelegramWebApp } from '@/features/telegram-miniapp/types/telegram-webapp'

export const getTelegramWebApp = (): TelegramWebApp | null => {
  if (typeof window === 'undefined') {
    return null
  }
  return window.Telegram?.WebApp ?? null
}

export const getTelegramInitData = (): string => {
  const app = getTelegramWebApp()
  return app?.initData?.trim() ?? ''
}

export const isTelegramMiniApp = (): boolean => getTelegramInitData() !== ''

