import { mudroApi } from '@/shared/api/mudroApi'
import type { FeedQueryArgs, FeedResponse, FeedSource, FrontResponse, PostsQueryArgs } from '@/entities/post/model/types'

const toApiSource = (source: FeedSource) => (source === 'all' ? undefined : source)

export const postsApi = mudroApi.injectEndpoints({
  endpoints: (build) => ({
    getFront: build.query<FrontResponse, FeedQueryArgs>({
      query: ({ limit, source, sort }) => ({
        url: '/api/front',
        params: {
          limit,
          source: toApiSource(source),
          sort,
        },
      }),
      providesTags: ['Feed'],
    }),
    getPosts: build.query<FeedResponse, PostsQueryArgs>({
      query: ({ limit, page, source, sort }) => ({
        url: '/api/posts',
        params: {
          limit,
          page,
          source: toApiSource(source),
          sort,
        },
      }),
      providesTags: ['Feed'],
    }),
  }),
})

export const { useGetFrontQuery, useLazyGetPostsQuery } = postsApi
