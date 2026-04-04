import { useEffect, useRef, useState } from 'react'

import { env } from '@/shared/config/env'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'
import { normalizeRouletteStateResponse, type RouletteStateResponse } from '@/features/casino/api/casinoApi'

export type RouletteStreamStatus = 'idle' | 'connecting' | 'connected' | 'disconnected' | 'error'

export interface RouletteStreamMessage {
  type?: string
  event?: string
  state?: Partial<RouletteStateResponse>
  round?: Partial<RouletteStateResponse>
  history?: unknown
  [key: string]: unknown
}

interface UseRouletteStreamOptions {
  enabled: boolean
  onMessage?: (message: RouletteStreamMessage) => void
}

const decodeEventBlock = (block: string) => {
  const message: Record<string, string> = {}
  const lines = block.split(/\r?\n/)

  for (const line of lines) {
    if (!line || line.startsWith(':')) {
      continue
    }

    const colonIndex = line.indexOf(':')
    if (colonIndex === -1) {
      continue
    }

    const field = line.slice(0, colonIndex).trim()
    const value = line.slice(colonIndex + 1).trimStart()
    message[field] = message[field] ? `${message[field]}\n${value}` : value
  }

  const payload = message.data ?? ''
  if (!payload) {
    return null
  }

  try {
    return JSON.parse(payload) as RouletteStreamMessage
  } catch {
    return { type: message.event ?? 'message', raw: payload }
  }
}

const normalizeStreamMessage = (message: RouletteStreamMessage): RouletteStreamMessage => {
  const normalizedState = normalizeRouletteStateResponse(
    (message.state ?? message.round ?? message) as Parameters<typeof normalizeRouletteStateResponse>[0],
  )

  if (!normalizedState) {
    return message
  }

  return {
    ...message,
    state: normalizedState,
  }
}

export const useRouletteStream = ({ enabled, onMessage }: UseRouletteStreamOptions) => {
  const token = useAppSelector((state) => state.session.token)
  const [status, setStatus] = useState<RouletteStreamStatus>('idle')
  const [error, setError] = useState<string | null>(null)
  const lastEventAt = useRef<string | null>(null)
  const onMessageRef = useRef(onMessage)

  useEffect(() => {
    onMessageRef.current = onMessage
  }, [onMessage])

  useEffect(() => {
    if (!enabled || !token) {
      setStatus('idle')
      setError(null)
      return
    }

    const abortController = new AbortController()
    let reconnectTimer: ReturnType<typeof setTimeout> | undefined
    let reconnectDelay = 1000

    const connect = async () => {
      setStatus((current) => (current === 'connected' ? current : 'connecting'))
      setError(null)

      try {
        const response = await fetch(`${env.apiBaseUrl}/casino/roulette/stream`, {
          method: 'GET',
          cache: 'no-store',
          headers: {
            Accept: 'text/event-stream',
            Authorization: `Bearer ${token}`,
          },
          signal: abortController.signal,
        })

        if (!response.ok) {
          throw new Error(`roulette stream failed: ${response.status}`)
        }

        const contentType = response.headers.get('content-type') ?? ''
        if (!contentType.includes('text/event-stream') || !response.body) {
          const payload = normalizeStreamMessage((await response.json()) as RouletteStreamMessage)
          onMessageRef.current?.(payload)
          lastEventAt.current = new Date().toISOString()
          setStatus('connected')
          return
        }

        const reader = response.body.getReader()
        const decoder = new TextDecoder('utf-8')
        let buffer = ''

        while (true) {
          const { done, value } = await reader.read()
          if (done) {
            break
          }

          buffer += decoder.decode(value, { stream: true })
          const blocks = buffer.split(/\r?\n\r?\n/)
          buffer = blocks.pop() ?? ''

          for (const block of blocks) {
            const payload = decodeEventBlock(block)
            if (!payload) {
              continue
            }

            const normalizedPayload = normalizeStreamMessage(payload)
            lastEventAt.current = new Date().toISOString()
            onMessageRef.current?.(normalizedPayload)
            setStatus('connected')
          }
        }

        setStatus('disconnected')
      } catch (streamError) {
        if (abortController.signal.aborted) {
          return
        }

        setError(streamError instanceof Error ? streamError.message : 'roulette stream error')
        setStatus('error')
        reconnectTimer = setTimeout(() => {
          reconnectDelay = Math.min(reconnectDelay * 1.8, 8000)
          void connect()
        }, reconnectDelay)
      }
    }

    void connect()

    return () => {
      abortController.abort()
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
      }
    }
  }, [enabled, token])

  return {
    status,
    error,
    lastEventAt: lastEventAt.current,
  }
}
