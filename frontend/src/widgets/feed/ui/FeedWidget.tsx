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
  query?: string
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
      setLoadError('Не удалось подгрузить следующую страницу. Проверь API и повтори запрос.')
    }
  }

  return (
    <section className="feed-widget">
      <FeedControls
        totalPosts={frontData?.meta.total_posts ?? 0}
        vkPosts={sourceTotals.vk}
        tgPosts={sourceTotals.tg}
      />

      {isInitialLoading ? (
        <section className="feed-widget__state-surface feed-widget__state-surface_loading">
          <div className="feed-widget__state-copy">
            <span className="feed-widget__state-eyebrow">Загрузка</span>
            <h3>Лента поднимается из API и собирает живую выборку</h3>
            <p>Карточки уже готовятся из реальных данных. Через пару секунд появятся посты, media и обсуждения.</p>
          </div>
          <FeedLoadingSkeleton />
        </section>
      ) : null}

      {isFrontError ? (
        <div className="feed-widget__error">
          <div className="feed-widget__state-copy">
            <span className="feed-widget__state-eyebrow">Недоступно</span>
            <h3>Лента временно недоступна</h3>
            <p>Сервер не отвечает. Убедитесь, что backend запущен, и повторите попытку.</p>
          </div>
          <button type="button" onClick={() => refetch()}>
            Повторить
          </button>
        </div>
      ) : null}

      {showEmpty ? (
        <section className="feed-widget__empty">
          <div className="feed-widget__state-copy">
            <span className="feed-widget__state-eyebrow">Пусто</span>
            <h3>Постов пока нет</h3>
            <p>Попробуй сбросить фильтры или выбрать другой источник. Если лента пуста — значит данные ещё не загружены в архив.</p>
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
          {isLoadingMore
            ? Array.from({ length: 3 }).map((_, i) => (
                <article key={`skeleton-append-${i}`} className="feed-widget__skeleton-card">
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
              ))
            : null}
        </div>
      ) : null}

      {loadError ? <p className="feed-widget__error-text">{loadError}</p> : null}

      {hasMore && !isLoadingMore ? (
        <button type="button" className="feed-widget__load-more" onClick={handleLoadMore}>
          Показать ещё
        </button>
      ) : null}

      <PostDetailDrawer post={selectedPost} onClose={() => setSelectedPost(null)} />
    </section>
  )
}

export const FeedWidget = () => {
  const { source, sort, limit, query } = useAppSelector((state) => state.feedFilters)
  const feedKey = `${source}-${sort}-${limit}-${query ?? ''}`

  return <FeedWidgetInner key={feedKey} source={source} sort={sort} limit={limit} query={query} />
}
