export interface TelegramThemeParams {
  bg_color?: string
  text_color?: string
  hint_color?: string
  link_color?: string
  button_color?: string
  button_text_color?: string
  secondary_bg_color?: string
}

export interface TelegramMainButton {
  setText(text: string): void
  show(): void
  hide(): void
  onClick(callback: () => void): void
  offClick(callback: () => void): void
}

export interface TelegramBackButton {
  show(): void
  hide(): void
  onClick(callback: () => void): void
  offClick(callback: () => void): void
}

export interface TelegramHapticFeedback {
  impactOccurred(style: 'light' | 'medium' | 'heavy' | 'rigid' | 'soft'): void
  notificationOccurred(type: 'error' | 'success' | 'warning'): void
  selectionChanged(): void
}

export interface TelegramWebApp {
  initData: string
  initDataUnsafe?: Record<string, unknown>
  colorScheme?: 'light' | 'dark'
  themeParams?: TelegramThemeParams
  MainButton: TelegramMainButton
  BackButton: TelegramBackButton
  HapticFeedback?: TelegramHapticFeedback
  ready(): void
  expand(): void
}

declare global {
  interface Window {
    Telegram?: {
      WebApp?: TelegramWebApp
    }
  }
}

export {}

