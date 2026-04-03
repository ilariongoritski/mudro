import { mudroApi } from '@/shared/api/mudroApi'
import type { FeedQueryArgs, FeedResponse, FeedSource, FrontResponse, PostsQueryArgs } from '@/entities/post/model/types'

const toApiSource = (source: FeedSource) => (source === 'all' ? undefined : source)

export const postsApi = mudroApi.injectEndpoints({
  endpoints: (build) => ({
    getFront: build.query<FrontResponse, FeedQueryArgs>({
      query: ({ limit, source, sort, q }) => ({
        url: '/front',
        // 'cache' is a fetch option, not RTK Query cache control.
        // We use 'no-store' to ensure we bypass browser/CDN caches for the main feed.
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
      query: (postId) => ({ url: `/posts/${postId}/like`, method: 'POST' }),
      invalidatesTags: ['Feed'],
    }),
    createComment: build.mutation<
      { id: number; post_id: number; author_name: string; text: string; published_at: string },
      { postId: number; text: string; parent_comment_id?: number }
    >({
      query: ({ postId, ...body }) => ({ url: `/posts/${postId}/comments`, method: 'POST', body }),
      invalidatesTags: ['Feed'],
    }),
    getPosts: build.query<FeedResponse, PostsQueryArgs>({
      query: ({ limit, page, source, sort, before_ts, before_id, q }) => ({
        url: '/posts',
        // Bypass browser cache for post history requests.
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
