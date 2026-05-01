import { useMemo, useState } from 'react'

import type { FeedCursor, Post } from '@/entities/post/model/types'
import { useGetFrontQuery, useLazyGetPostsQuery } from '@/entities/post/model/postsApi'
import { PostCard } from '@/entities/post/ui/post-card/PostCard'
import { PostDetailDrawer } from '@/entities/post/ui/post-detail-drawer/PostDetailDrawer'
import { FeedControls } from '@/features/feed-controls/ui/FeedControls'
import { Button, Skeleton } from '@/shared/ui'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'

import './FeedWidget.css'

interface FeedWidgetInnerProps {
  source: 'all' | 'vk' | 'tg'
  sort: 'desc' | 'asc'
  limit: number
  query?: string
}

const SKELETON_IDS = ['a', 'b', 'c', 'd', 'e', 'f']

const PostCardSkeleton = () => (
  <article className="feed-widget__skeleton-card">
    <div className="feed-widget__skeleton-row">
      <Skeleton type="circle" width={36} height={36} />
      <div className="feed-widget__skeleton-col">
        <Skeleton type="text" width="40%" />
        <Skeleton type="text" width="25%" />
      </div>
    </div>
    <Skeleton type="text" width="90%" />
    <Skeleton type="text" width="95%" />
    <Skeleton type="text" width="70%" />
    <Skeleton type="rect" height={180} />
    <div className="feed-widget__skeleton-row">
      <Skeleton type="text" width={60} height={20} />
      <Skeleton type="text" width={60} height={20} />
      <Skeleton type="text" width={60} height={20} />
    </div>
  </article>
)

const FeedWidgetInner = ({ source, sort, limit, query }: FeedWidgetInnerProps) => {
  const [loadedItems, setLoadedItems] = useState<Post[]>([])
  const [nextCursor, setNextCursor] = useState<FeedCursor | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [selectedPost, setSelectedPost] = useState<Post | null>(null)

  const {
    data: frontData,
    isFetching: isFrontFetching,
    isError: isFrontError,
    refetch,
  } = useGetFrontQuery({ limit, source, sort, q: query })

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
      setLoadError('Не удалось загрузить следующую страницу.')
    }
  }

  return (
    <section className="feed-widget">
      <FeedControls
        totalPosts={frontData?.meta.total_posts ?? 0}
        vkPosts={sourceTotals.vk}
        tgPosts={sourceTotals.tg}
      />

      {/* Начальная загрузка */}
      {isInitialLoading && (
        <div className="feed-widget__grid feed-widget__grid_loading" aria-busy="true" aria-label="Загрузка ленты">
          {SKELETON_IDS.map((id) => <PostCardSkeleton key={id} />)}
        </div>
      )}

      {/* Ошибка */}
      {isFrontError && (
        <div className="feed-widget__state-box feed-widget__state-box_error">
          <p className="feed-widget__state-title">Лента временно недоступна</p>
          <p className="feed-widget__state-text">Сервер не отвечает. Убедитесь что backend запущен.</p>
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            Повторить
          </Button>
        </div>
      )}

      {/* Пусто */}
      {showEmpty && (
        <div className="feed-widget__state-box">
          <p className="feed-widget__state-title">Постов пока нет</p>
          <p className="feed-widget__state-text">Попробуй сбросить фильтры или выбрать другой источник.</p>
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            Обновить ленту
          </Button>
        </div>
      )}

      {/* Список постов */}
      {!isInitialLoading && !showEmpty && (
        <div className="feed-widget__grid">
          {items.map((post) => (
            <PostCard
              key={`${post.source}-${post.id}-${post.source_post_id}`}
              post={post}
              onOpen={setSelectedPost}
            />
          ))}
          {isLoadingMore && (
            <>
              <PostCardSkeleton />
              <PostCardSkeleton />
              <PostCardSkeleton />
            </>
          )}
        </div>
      )}

      {/* Ошибка догрузки */}
      {loadError && <p className="feed-widget__error-text">{loadError}</p>}

      {/* Показать ещё */}
      {hasMore && !isLoadingMore && (
        <div className="feed-widget__load-more-wrap">
          <Button variant="outline" onClick={handleLoadMore}>
            Показать ещё
          </Button>
        </div>
      )}

      {/* Конец ленты */}
      {!hasMore && !isInitialLoading && !isLoadingMore && items.length > 0 && (
        <div className="feed-widget__end">
          <span className="feed-widget__end-line" aria-hidden="true" />
          <span className="feed-widget__end-label">Вы дочитали ленту</span>
          <span className="feed-widget__end-line" aria-hidden="true" />
        </div>
      )}

      <PostDetailDrawer post={selectedPost} onClose={() => setSelectedPost(null)} />
    </section>
  )
}

export const FeedWidget = () => {
  const { source, sort, limit, query } = useAppSelector((state) => state.feedFilters)
  const feedKey = `${source}-${sort}-${limit}-${query ?? ''}`
  return <FeedWidgetInner key={feedKey} source={source} sort={sort} limit={limit} query={query} />
}
