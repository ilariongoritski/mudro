import { useEffect, useMemo, useState } from 'react'

import { useGetChatMessagesQuery, useSendChatMessageMutation } from '@/entities/chat/api/chatApi'
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

  const {
    data,
    isLoading,
    isFetching,
    error,
    refetch,
  } = useGetChatMessagesQuery(
    { room, limit },
    { skip: !token },
  )
  const [sendChatMessage, { isLoading: isSending }] = useSendChatMessageMutation()

  useEffect(() => {
    if (!data?.items) {
      return
    }
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setMessages((current) => mergeMessages(current, data.items))
  }, [data])

  useEffect(() => {
    if (!token) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setConnectionState('idle')
       
      setMessages([])
      return
    }

    const socket = new WebSocket(buildChatWsUrl(room, token))
    setConnectionState('connecting')

    socket.onopen = () => {
      setConnectionState('open')
    }

    socket.onclose = () => {
      setConnectionState('closed')
    }

    socket.onerror = () => {
      setConnectionState('error')
    }

    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data) as ChatSocketEvent
        if (payload.type !== 'message' || !payload.message) {
          return
        }
        setMessages((current) => mergeMessages(current, [payload.message!]))
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

    const message = await sendChatMessage({ room, body: trimmed }).unwrap()
    setMessages((current) => mergeMessages(current, [message]))
  }

  return {
    room,
    messages,
    isLoading,
    isFetching,
    error,
    isSending,
    connectionState,
    connectionLabel,
    refetch,
    sendMessage,
  }
}
