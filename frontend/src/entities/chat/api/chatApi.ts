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
    uploadKeys: build.mutation<void, any>({
      query: (body) => ({
        url: '/chat/keys',
        method: 'POST',
        body,
      }),
    }),
    getKeysBundle: build.query<any, number>({
      query: (userId) => ({
        url: '/chat/keys',
        params: { user_id: userId },
      }),
    }),
  }),
})

export const { 
  useGetChatMessagesQuery, 
  useLazyGetChatMessagesQuery, 
  useSendChatMessageMutation,
  useUploadKeysMutation,
  useGetKeysBundleQuery,
  useLazyGetKeysBundleQuery
} = chatApi
