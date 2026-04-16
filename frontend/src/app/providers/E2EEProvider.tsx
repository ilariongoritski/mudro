import React from 'react'
import { useE2EEInitialization } from '@/features/chat/lib/useE2EEInitialization'

export const E2EEProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  useE2EEInitialization()
  return <>{children}</>
}
