export interface ChatUser {
  id: number
  username: string
  role: string
}

export interface ChatMessage {
  id: number
  room: string
  body: string
  created_at: string
  user: ChatUser
}

export interface ChatMessagesResponse {
  items: ChatMessage[]
}

export interface GetChatMessagesArgs {
  room?: string
  limit?: number
  before_id?: number
}

export interface SendChatMessageRequest {
  room?: string
  body: string
}

export interface ChatSocketEvent {
  type: 'ready' | 'message'
  message?: ChatMessage
}
