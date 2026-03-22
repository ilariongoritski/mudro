import { mudroApi } from '@/shared/api/mudroApi'
import type { FeedQueryArgs, FeedResponse, FeedSource, FrontResponse, PostsQueryArgs } from '@/entities/post/model/types'

const toApiSource = (source: FeedSource) => (source === 'all' ? undefined : source)

export const postsApi = mudroApi.injectEndpoints({
  endpoints: (build) => ({
    getFront: build.query<FrontResponse, FeedQueryArgs>({
      query: ({ limit, source, sort, q }) => ({
        url: '/front',
        cache: 'no-store',
        params: {
          limit,
          source: toApiSource(source),
          sort,
          q,
        },
      }),
      providesTags: ['Feed'],
    }),
    toggleLike: build.mutation<{ liked: boolean; likes_count: number }, number>({
      query: (postId) => ({ url: `/api/posts/${postId}/like`, method: 'POST' }),
      invalidatesTags: ['Feed'],
    }),
    createComment: build.mutation<
      { id: number; post_id: number; author_name: string; text: string; published_at: string },
      { postId: number; text: string; parent_comment_id?: number }
    >({
      query: ({ postId, ...body }) => ({ url: `/api/posts/${postId}/comments`, method: 'POST', body }),
      invalidatesTags: ['Feed'],
    }),
    getPosts: build.query<FeedResponse, PostsQueryArgs>({
      query: ({ limit, page, source, sort, before_ts, before_id, q }) => ({
        url: '/posts',
        cache: 'no-store',
        params: {
          limit,
          ...(page ? { page } : {}),
          ...(before_ts ? { before_ts } : {}),
          ...(typeof before_id === 'number' ? { before_id } : {}),
          ...(toApiSource(source) ? { source: toApiSource(source) } : {}),
          sort,
          q,
        },
      }),
      providesTags: ['Feed'],
    }),
  }),
})

export const { useGetFrontQuery, useLazyGetPostsQuery, useToggleLikeMutation, useCreateCommentMutation } = postsApi
