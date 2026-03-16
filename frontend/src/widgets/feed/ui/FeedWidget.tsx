import { useMemo, useState } from 'react'

import type { FeedCursor, Post } from '@/entities/post/model/types'
import { useGetFrontQuery, useLazyGetPostsQuery } from '@/entities/post/model/postsApi'
import { PostCard } from '@/entities/post/ui/post-card/PostCard'
import { PostDetailDrawer } from '@/entities/post/ui/post-detail-drawer/PostDetailDrawer'
import { FeedControls } from '@/features/feed-controls/ui/FeedControls'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'
import './FeedWidget.css'

interface FeedWidgetInnerProps {
  source: 'all' | 'vk' | 'tg'
  sort: 'desc' | 'asc'
  limit: number
}

const initialSkeletonIds = ['a', 'b', 'c', 'd', 'e', 'f']

const FeedLoadingSkeleton = () => {
  return (
    <div className="feed-widget__grid feed-widget__grid_loading" aria-hidden="true">
      {initialSkeletonIds.map((id) => (
        <article key={id} className="feed-widget__skeleton-card">
          <div className="feed-widget__skeleton-chip" />
          <div className="feed-widget__skeleton-lines">
            <span />
            <span />
            <span />
          </div>
          <div className="feed-widget__skeleton-media" />
          <div className="feed-widget__skeleton-stats">
            <span />
            <span />
            <span />
          </div>
        </article>
      ))}
    </div>
  )
}

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
    <section className="feed-widget">
      <FeedControls
        totalPosts={frontData?.meta.total_posts ?? 0}
        vkPosts={sourceTotals.vk}
        tgPosts={sourceTotals.tg}
        lastSyncAt={frontData?.meta.last_sync_at}
      />

      {isInitialLoading ? (
        <section className="feed-widget__state-surface feed-widget__state-surface_loading">
          <div className="feed-widget__state-copy">
            <span className="feed-widget__state-eyebrow">Loading surface</span>
            <h3>Лента поднимается из API и собирает mixed stream</h3>
            <p>
              Пока toolbar уже на месте, карточки готовятся из реальных данных. Это не статический
              фейк-экран, а ожидание живой выборки.
            </p>
          </div>
          <FeedLoadingSkeleton />
        </section>
      ) : null}

      {isFrontError ? (
        <div className="feed-widget__error">
          <div className="feed-widget__state-copy">
            <span className="feed-widget__state-eyebrow">Feed error</span>
            <h3>Не удалось загрузить `/api/front`</h3>
            <p>
              Это уже не визуальный баг страницы, а проблема слоя данных. Повтори запрос или
              проверь backend-контур.
            </p>
          </div>
          <button type="button" onClick={() => refetch()}>
            Повторить
          </button>
        </div>
      ) : null}

      {showEmpty ? (
        <section className="feed-widget__empty">
          <div className="feed-widget__empty-copy">
            <span className="feed-widget__state-eyebrow">Empty feed</span>
            <h3>Постов под текущие фильтры пока нет</h3>
            <p>
              Это нормальный сценарий для пустой базы, жесткого source-фильтра или еще не
              синхронизированного архива.
            </p>
          </div>
          <div className="feed-widget__empty-actions">
            <button type="button" onClick={() => refetch()}>
              Обновить ленту
            </button>
          </div>
        </section>
      ) : null}

      {!isInitialLoading && !showEmpty ? (
        <div className="feed-widget__grid">
          {items.map((post) => (
            <PostCard key={`${post.source}-${post.id}-${post.source_post_id}`} post={post} onOpen={setSelectedPost} />
          ))}
        </div>
      ) : null}

      {loadError ? <p className="feed-widget__error-text">{loadError}</p> : null}

      {hasMore ? (
        <button type="button" className="feed-widget__load-more" disabled={isLoadingMore} onClick={handleLoadMore}>
          {isLoadingMore ? 'Загружаю...' : 'Показать еще'}
        </button>
      ) : null}

      <PostDetailDrawer post={selectedPost} onClose={() => setSelectedPost(null)} />
    </section>
  )
}

export const FeedWidget = () => {
  const { source, sort, limit } = useAppSelector((state) => state.feedFilters)
  const feedKey = `${source}-${sort}-${limit}`

  return <FeedWidgetInner key={feedKey} source={source} sort={sort} limit={limit} />
}
