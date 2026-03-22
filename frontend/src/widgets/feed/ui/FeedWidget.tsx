import { useMemo, useState } from 'react'

import type { FeedCursor, Post } from '@/entities/post/model/types'
import { useGetFrontQuery, useLazyGetPostsQuery } from '@/entities/post/model/postsApi'
import { PostCard } from '@/entities/post/ui/post-card/PostCard'
import { PostDetailDrawer } from '@/entities/post/ui/post-detail-drawer/PostDetailDrawer'
import { FeedControls } from '@/features/feed-controls/ui/FeedControls'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'
import { Button } from '@/shared/ui/button'
import { Card, CardContent } from '@/shared/ui/card'
import { Skeleton } from '@/shared/ui/skeleton'

interface FeedWidgetInnerProps {
  source: 'all' | 'vk' | 'tg'
  sort: 'desc' | 'asc'
  limit: number
}

const skeletonIds = ['a', 'b', 'c', 'd', 'e', 'f']

const FeedLoadingSkeleton = () => (
  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
    {skeletonIds.map((id) => (
      <Card key={id}>
        <CardContent className="space-y-3 p-5">
          <div className="flex items-center gap-2">
            <Skeleton className="h-5 w-12" />
            <Skeleton className="h-3 w-16" />
          </div>
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
          <Skeleton className="h-4 w-1/2" />
          <Skeleton className="h-24 w-full" />
          <div className="flex gap-4">
            <Skeleton className="h-3 w-12" />
            <Skeleton className="h-3 w-12" />
            <Skeleton className="h-3 w-12" />
          </div>
        </CardContent>
      </Card>
    ))}
  </div>
)

const FeedWidgetInner = ({ source, sort, limit }: FeedWidgetInnerProps) => {
  const [loadedItems, setLoadedItems] = useState<Post[]>([])
  const [nextCursor, setNextCursor] = useState<FeedCursor | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [selectedPost, setSelectedPost] = useState<Post | null>(null)

  const {
    data: frontData,
    isFetching: isFrontFetching,
    isError: isFrontError,
    refetch,
  } = useGetFrontQuery({ limit, source, sort })

  const [loadMorePosts, { isFetching: isLoadingMore }] = useLazyGetPostsQuery()

  const items = useMemo(() => {
    return [...(frontData?.feed.items ?? []), ...loadedItems]
  }, [frontData?.feed.items, loadedItems])

  const sourceTotals = useMemo(() => {
    const totals = { vk: 0, tg: 0 }
    for (const item of frontData?.meta.sources ?? []) {
      if (item.source === 'vk') totals.vk = item.posts
      if (item.source === 'tg') totals.tg = item.posts
    }
    return totals
  }, [frontData])

  const effectiveCursor = nextCursor ?? frontData?.feed.next_cursor ?? null
  const hasMore = Boolean(effectiveCursor)
  const isInitialLoading = isFrontFetching && items.length === 0
  const showEmpty = !isInitialLoading && !isFrontError && items.length === 0

  const handleLoadMore = async () => {
    if (!effectiveCursor) return

    try {
      setLoadError(null)
      const response = await loadMorePosts({
        limit,
        source,
        sort,
        before_ts: effectiveCursor.before_ts,
        before_id: effectiveCursor.before_id,
      }).unwrap()
      setLoadedItems((current) => [...current, ...response.items])
      setNextCursor(response.next_cursor ?? null)
    } catch {
      setLoadError('Не удалось подгрузить следующую страницу. Проверь API и повтори запрос.')
    }
  }

  return (
    <div className="space-y-4">
      <FeedControls
        totalPosts={frontData?.meta.total_posts ?? 0}
        vkPosts={sourceTotals.vk}
        tgPosts={sourceTotals.tg}
        lastSyncAt={frontData?.meta.last_sync_at}
      />

      {isInitialLoading && (
        <div className="space-y-4">
          <Card>
            <CardContent className="py-6 text-center">
              <p className="text-xs font-medium text-slate-400 uppercase tracking-wide mb-1">Загрузка</p>
              <h3 className="text-base font-semibold text-slate-700">Лента поднимается из API</h3>
              <p className="text-sm text-slate-500 mt-1">Карточки уже готовятся из реальных данных.</p>
            </CardContent>
          </Card>
          <FeedLoadingSkeleton />
        </div>
      )}

      {isFrontError && (
        <Card className="border-red-200">
          <CardContent className="py-6 text-center space-y-3">
            <p className="text-xs font-medium text-red-400 uppercase tracking-wide">Ошибка</p>
            <h3 className="text-base font-semibold text-slate-700">Не удалось загрузить ленту</h3>
            <p className="text-sm text-slate-500">Повтори запрос или проверь backend-контур.</p>
            <Button variant="outline" size="sm" onClick={() => refetch()}>Повторить</Button>
          </CardContent>
        </Card>
      )}

      {showEmpty && (
        <Card>
          <CardContent className="py-6 text-center space-y-3">
            <p className="text-xs font-medium text-slate-400 uppercase tracking-wide">Пусто</p>
            <h3 className="text-base font-semibold text-slate-700">Под текущими фильтрами постов нет</h3>
            <p className="text-sm text-slate-500">Попробуй другой source-фильтр или обнови ленту.</p>
            <Button variant="outline" size="sm" onClick={() => refetch()}>Обновить ленту</Button>
          </CardContent>
        </Card>
      )}

      {!isInitialLoading && !showEmpty && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {items.map((post) => (
            <PostCard key={`${post.source}-${post.id}-${post.source_post_id}`} post={post} onOpen={setSelectedPost} />
          ))}
        </div>
      )}

      {loadError && <p className="text-sm text-red-500 text-center">{loadError}</p>}

      {hasMore && (
        <div className="flex justify-center pt-2">
          <Button variant="outline" disabled={isLoadingMore} onClick={handleLoadMore}>
            {isLoadingMore ? 'Загружаю...' : 'Показать еще'}
          </Button>
        </div>
      )}

      <PostDetailDrawer post={selectedPost} onClose={() => setSelectedPost(null)} />
    </div>
  )
}

export const FeedWidget = () => {
  const { source, sort, limit } = useAppSelector((state) => state.feedFilters)
  const feedKey = `${source}-${sort}-${limit}`

  return <FeedWidgetInner key={feedKey} source={source} sort={sort} limit={limit} />
}
