import { useEffect, useMemo } from 'react'
import { getTelegramInitData, getTelegramWebApp } from '@/features/telegram-miniapp/lib/telegramWebApp'

export const useTelegramWebApp = () => {
  const webApp = useMemo(() => getTelegramWebApp(), [])
  const initData = useMemo(() => getTelegramInitData(), [])
  const isTelegram = Boolean(webApp && initData)

  useEffect(() => {
    if (!webApp) {
      return
    }
    try {
      webApp.ready()
      webApp.expand()
    } catch {
      // Ignore runtime Telegram bridge errors and keep web fallback.
    }
  }, [webApp])

  return {
    webApp,
    initData,
    isTelegram,
    colorScheme: webApp?.colorScheme ?? 'light',
    themeParams: webApp?.themeParams,
  }
}

