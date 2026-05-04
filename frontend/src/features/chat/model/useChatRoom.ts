import { useCallback, useEffect, useEffectEvent, useMemo, useState } from 'react'

import { useLazyGetChatMessagesQuery, useSendChatMessageMutation, useGetChatMessagesQuery } from '@/entities/chat/api/chatApi'
import type { ChatMessage, ChatSocketEvent } from '@/entities/chat/model/types'
import { env } from '@/shared/config/env'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'

interface UseChatRoomOptions {
  room?: string
  limit?: number
}

const DEFAULT_LIMIT = 50

const mergeMessages = (current: ChatMessage[], incoming: ChatMessage[]) => {
  const indexed = new Map<number, ChatMessage>()

  for (const item of current) {
    indexed.set(item.id, item)
  }
  for (const item of incoming) {
    indexed.set(item.id, item)
  }

  return Array.from(indexed.values()).sort((left, right) => left.id - right.id)
}

const buildChatWsUrl = (room: string, token: string) => {
  const base = env.apiBaseUrl.startsWith('http')
    ? new URL(env.apiBaseUrl)
    : new URL(env.apiBaseUrl, window.location.origin)

  base.protocol = base.protocol === 'https:' ? 'wss:' : 'ws:'
  base.pathname = `${base.pathname.replace(/\/$/, '')}/chat/ws`
  base.search = ''
  base.searchParams.set('room', room)
  base.searchParams.set('token', token)

  return base.toString()
}

export const useChatRoom = ({ room = 'main', limit = DEFAULT_LIMIT }: UseChatRoomOptions = {}) => {
  const token = useAppSelector((state) => state.session.token)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [connectionState, setConnectionState] = useState<'idle' | 'connecting' | 'open' | 'closed' | 'error'>('idle')
  const [hasMore, setHasMore] = useState(true)

  // Initial load
  const {
    data: initialData,
    isLoading: isInitialLoading,
    error: initialError,
    refetch,
  } = useGetChatMessagesQuery(
    { room, limit },
    { skip: !token },
  )

  // Lazy query for loading history (older messages)
  const [triggerLoadMore, { isFetching: isMoreLoading }] = useLazyGetChatMessagesQuery()

  const [sendChatMessage, { isLoading: isSending }] = useSendChatMessageMutation()

  const resetRoomState = useEffectEvent(() => {
    setMessages([])
    setHasMore(true)
    setConnectionState('idle')
  })

  const syncInitialMessages = useEffectEvent((items: ChatMessage[]) => {
    setMessages((current) => mergeMessages(current, items))
    setHasMore(items.length >= limit)
  })

  const markSocketConnecting = useEffectEvent(() => {
    setConnectionState('connecting')
  })

  useEffect(() => {
    resetRoomState()
  }, [room, token])

  useEffect(() => {
    if (initialData?.items) {
      syncInitialMessages(initialData.items)
    }
  }, [initialData?.items])

  const loadMore = useCallback(async () => {
    if (isMoreLoading || !hasMore || messages.length === 0) {
      return
    }

    const oldestId = messages[0].id
    try {
      const result = await triggerLoadMore({ room, limit, before_id: oldestId }).unwrap()
      if (result.items.length === 0) {
        setHasMore(false)
        return
      }
      if (result.items.length < limit) {
        setHasMore(false)
      }
      setMessages((current) => mergeMessages(current, result.items))
    } catch (err) {
      console.error('Failed to load more chat messages', err)
    }
  }, [isMoreLoading, hasMore, messages, room, limit, triggerLoadMore])

  useEffect(() => {
    if (!token) {
      return
    }

    const socket = new WebSocket(buildChatWsUrl(room, token))
    markSocketConnecting()

    socket.onopen = () => {
      setConnectionState('open')
    }

    socket.onclose = () => {
      setConnectionState('closed')
    }

    socket.onerror = () => {
      setConnectionState('error')
    }

    socket.onmessage = async (event) => {
      try {
        const payload = JSON.parse(event.data) as ChatSocketEvent
        if (payload.type !== 'message' || !payload.message) {
          return
        }

        const message = payload.message
        setMessages((current) => mergeMessages(current, [message]))
      } catch (err) {
        console.error('Chat socket message parse failed', err)
      }
    }

    return () => {
      socket.close(1000, 'chat-room-dispose')
    }
  }, [room, token])

  const connectionLabel = useMemo(() => {
    switch (connectionState) {
      case 'connecting':
        return 'Подключаем realtime'
      case 'open':
        return 'Realtime подключён'
      case 'closed':
        return 'Realtime отключён'
      case 'error':
        return 'Ошибка realtime'
      default:
        return 'Ожидаем авторизацию'
    }
  }, [connectionState])

  const sendMessage = async (body: string) => {
    const trimmed = body.trim()
    if (!trimmed) {
      return
    }

    const message = await sendChatMessage({
      room,
      body: trimmed,
    }).unwrap()
    setMessages((current) => mergeMessages(current, [message]))
  }

  return {
    room,
    messages,
    isLoading: isInitialLoading,
    isFetching: isMoreLoading,
    error: initialError,
    isSending,
    hasMore,
    connectionState,
    connectionLabel,
    refetch,
    loadMore,
    sendMessage,
  }
}
