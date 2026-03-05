import { useMemo, useState } from 'react'

import type { FeedCursor, Post } from '@/entities/post/model/types'
import { useGetFrontQuery, useLazyGetPostsQuery } from '@/entities/post/model/postsApi'
import { PostCard } from '@/entities/post/ui/post-card/PostCard'
import { FeedControls } from '@/features/feed-controls/ui/FeedControls'
import { formatDateTime } from '@/shared/lib/format/date'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'
import './FeedWidget.css'

interface FeedWidgetInnerProps {
  source: 'all' | 'vk' | 'tg'
  sort: 'desc' | 'asc'
  limit: number
}

const FeedWidgetInner = ({ source, sort, limit }: FeedWidgetInnerProps) => {
  const [loadedItems, setLoadedItems] = useState<Post[]>([])
  const [nextCursor, setNextCursor] = useState<FeedCursor | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)

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
      setLoadError('Could not load next page. Retry after API check.')
    }
  }

  return (
    <section className="feed-widget">
      <div className="feed-widget__meta mudro-fade-up">
        <div>
          <div className="feed-widget__meta-label">Total posts</div>
          <div className="feed-widget__meta-value">{frontData?.meta.total_posts ?? '0'}</div>
        </div>
        <div>
          <div className="feed-widget__meta-label">VK</div>
          <div className="feed-widget__meta-value">{sourceTotals.vk}</div>
        </div>
        <div>
          <div className="feed-widget__meta-label">Telegram</div>
          <div className="feed-widget__meta-value">{sourceTotals.tg}</div>
        </div>
        <div>
          <div className="feed-widget__meta-label">Last sync</div>
          <div className="feed-widget__meta-value feed-widget__meta-value_small">
            {formatDateTime(frontData?.meta.last_sync_at)}
          </div>
        </div>
      </div>

      <FeedControls />

      {isFrontFetching && items.length === 0 ? <p className="feed-widget__status">Loading feed...</p> : null}

      {isFrontError ? (
        <div className="feed-widget__error">
          <p>Failed to load `/api/front`.</p>
          <button type="button" onClick={() => refetch()}>
            Retry
          </button>
        </div>
      ) : null}

      {!isFrontFetching && !isFrontError && items.length === 0 ? (
        <p className="feed-widget__status">No data in posts yet.</p>
      ) : null}

      <div className="feed-widget__grid">
        {items.map((post) => (
          <PostCard key={`${post.source}-${post.id}-${post.source_post_id}`} post={post} />
        ))}
      </div>

      {loadError ? <p className="feed-widget__error-text">{loadError}</p> : null}

      {hasMore ? (
        <button type="button" className="feed-widget__load-more" disabled={isLoadingMore} onClick={handleLoadMore}>
          {isLoadingMore ? 'Loading...' : 'Load more'}
        </button>
      ) : null}
    </section>
  )
}

export const FeedWidget = () => {
  const { source, sort, limit } = useAppSelector((state) => state.feedFilters)
  const feedKey = `${source}-${sort}-${limit}`

  return <FeedWidgetInner key={feedKey} source={source} sort={sort} limit={limit} />
}
