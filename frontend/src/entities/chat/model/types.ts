export interface ChatUser {
  id: number
  username: string
  role: string
}

export interface ChatMessage {
  id: number
  room: string
  body: string
  encrypted_body?: string | null
  nonce?: string | null
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

export interface E2EEKeyBundle {
  identity_key: string
  signed_prekey: string
  signature: string
  one_time_prekey?: {
    id: number
    key: string
  }
}

export interface UserKeys {
  identityKey: CryptoKeyPair
  signedPrekey: CryptoKeyPair
  signature: ArrayBuffer
  oneTimePrekeys: CryptoKeyPair[]
}

export interface UploadKeysRequest extends E2EEKeyBundle {
  one_time_prekeys?: Array<{
    id: number
    key: string
  }>
}
