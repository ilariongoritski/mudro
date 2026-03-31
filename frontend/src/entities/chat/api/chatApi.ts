import { mudroApi } from '@/shared/api/mudroApi'
import type {
  ChatMessage,
  ChatMessagesResponse,
  GetChatMessagesArgs,
  SendChatMessageRequest,
} from '@/entities/chat/model/types'

export const chatApi = mudroApi.injectEndpoints({
  endpoints: (build) => ({
    getChatMessages: build.query<ChatMessagesResponse, GetChatMessagesArgs | void>({
      query: (params) => ({
        url: '/chat/messages',
        params: {
          room: params?.room,
          limit: params?.limit,
          before_id: params?.before_id,
        },
      }),
      providesTags: ['Chat'],
    }),
    sendChatMessage: build.mutation<ChatMessage, SendChatMessageRequest>({
      query: (body) => ({
        url: '/chat/messages',
        method: 'POST',
        body,
      }),
    }),
  }),
})

export const { useGetChatMessagesQuery, useSendChatMessageMutation } = chatApi
