import { useEffect } from 'react'

import type { Post, PostComment } from '@/entities/post/model/types'
import {
  buildOriginalPostUrl,
  humanizeCommentAuthor,
  mediaKindLabel,
  metricDisplay,
  metricLabel,
  normalizeReactions,
  reactionLabel,
  resolveMediaDisplayUrl,
  resolveMediaKind,
  resolveMediaTitle,
  resolveMediaUrl,
} from '@/entities/post/lib/postPresentation'
import { formatDateTime } from '@/shared/lib/format/date'
import './PostDetailDrawer.css'

interface PostDetailDrawerProps {
  post: Post | null
  onClose: () => void
}

const normalizeCommentReactions = (reactions?: PostComment['reactions']) => {
  if (!reactions) return [] as Array<[string, number]>

  if (Array.isArray(reactions)) {
    return reactions
      .filter((reaction) => reaction.count > 0)
      .map((reaction) => [reaction.raw || reaction.label, reaction.count] as [string, number])
  }

  return normalizeReactions(reactions)
}

export const PostDetailDrawer = ({ post, onClose }: PostDetailDrawerProps) => {
  useEffect(() => {
    if (!post) return

    const previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }

    window.addEventListener('keydown', handleEscape)
    return () => {
      document.body.style.overflow = previousOverflow
      window.removeEventListener('keydown', handleEscape)
    }
  }, [post, onClose])

  if (!post) return null

  const mediaItems = post.media ?? []
  const reactions = normalizeReactions(post.reactions)
  const comments = post.comments ?? []
  const totalComments = post.comments_count ?? comments.length
  const fullText = post.text?.trim() || 'Для этого поста нет развернутого текста.'
  const viewsMetric = metricDisplay(post.views_count)
  const originalPostUrl = buildOriginalPostUrl(post.source, post.source_post_id)

  return (
    <div className="post-drawer" role="dialog" aria-modal="true" aria-labelledby="post-drawer-title">
      <button type="button" className="post-drawer__backdrop" aria-label="Закрыть карточку поста" onClick={onClose} />

      <aside className="post-drawer__panel">
        <header className="post-drawer__head">
          <div className="post-drawer__head-main">
            <div className={`post-drawer__source post-drawer__source_${post.source}`}>{post.source.toUpperCase()}</div>
            <div className="post-drawer__eyebrow">{post.source.toUpperCase()} #{post.source_post_id}</div>
            <h2 id="post-drawer-title">Развернутый просмотр поста</h2>
            <p>{formatDateTime(post.published_at)}</p>
          </div>

          <div className="post-drawer__head-actions">
            {originalPostUrl ? (
              <a className="post-drawer__origin-link" href={originalPostUrl} target="_blank" rel="noreferrer">
                Оригинальный пост
              </a>
            ) : null}
            <button type="button" className="post-drawer__close" onClick={onClose}>
              Закрыть
            </button>
          </div>
        </header>

        <section className="post-drawer__stats" aria-label="Метрики поста">
          <article>
            <span>Лайки</span>
            <strong>{metricLabel(post.likes_count)}</strong>
          </article>
          <article>
            <span>Просмотры</span>
            <strong className={viewsMetric.missing ? 'post-drawer__metric-missing' : undefined}>{viewsMetric.value}</strong>
          </article>
          <article>
            <span>Комментарии</span>
            <strong>{metricLabel(totalComments)}</strong>
          </article>
        </section>

        {reactions.length > 0 ? (
          <section className="post-drawer__reactions" aria-label="Реакции">
            {reactions.map(([reaction, count]) => (
              <span key={reaction} className="post-drawer__reaction" title={reaction}>
                {reactionLabel(reaction)} {count}
              </span>
            ))}
          </section>
        ) : null}

        <section className="post-drawer__text-block">
          <span className="post-drawer__section-label">Текст поста</span>
          <p>{fullText}</p>
        </section>

        {mediaItems.length > 0 ? (
          <section className="post-drawer__media" aria-label="Вложения">
            <span className="post-drawer__section-label">Вложения</span>
            <div className="post-drawer__media-grid">
              {mediaItems.map((item, index) => {
                const kind = resolveMediaKind(item)
                const title = resolveMediaTitle(item)
                const mediaUrl = resolveMediaUrl(item.url)
                const displayUrl = resolveMediaDisplayUrl(item)
                const ctaHref = originalPostUrl ?? mediaUrl
                const ctaLabel = originalPostUrl ? 'Оригинальный пост' : 'Открыть файл'

                return (
                  <article key={`${item.url ?? item.title ?? item.kind}-${index}`} className="post-drawer__media-card">
                    {(kind === 'image' || kind === 'video') && displayUrl ? <img src={displayUrl} alt={title} loading="lazy" /> : null}

                    <div className="post-drawer__media-info">
                      <strong>{mediaKindLabel(kind)}</strong>
                      <span title={title}>{title}</span>
                      {ctaHref ? (
                        <a href={ctaHref} target="_blank" rel="noreferrer">
                          {ctaLabel}
                        </a>
                      ) : null}
                    </div>
                  </article>
                )
              })}
            </div>
          </section>
        ) : null}

        {comments.length > 0 ? (
          <section className="post-drawer__comments" aria-label="Комментарии">
            <span className="post-drawer__section-label">Комментарии и ответы</span>
            <div className="post-drawer__comment-list">
              {comments.map((comment) => {
                const commentReactions = normalizeCommentReactions(comment.reactions)

                return (
                  <article
                    key={`${comment.source_comment_id}-${comment.parent_comment_id ?? 'root'}`}
                    className={`post-drawer__comment-card ${comment.parent_comment_id ? 'post-drawer__comment-card_reply' : ''}`}
                  >
                    <div className="post-drawer__comment-meta">
                      <strong>{humanizeCommentAuthor(comment.author_name)}</strong>
                      <span>{formatDateTime(comment.published_at)}</span>
                      {comment.parent_comment_id ? <span>ответ на #{comment.parent_comment_id}</span> : null}
                    </div>
                    <p>{comment.text?.trim() || 'Без текста'}</p>

                    {commentReactions.length > 0 ? (
                      <div className="post-drawer__comment-reactions">
                        {commentReactions.map(([reaction, count]) => (
                          <span key={`${comment.source_comment_id}-${reaction}`} className="post-drawer__comment-reaction" title={reaction}>
                            {reactionLabel(reaction)} {count}
                          </span>
                        ))}
                      </div>
                    ) : null}

                    {(comment.media ?? []).length > 0 ? (
                      <div className="post-drawer__comment-media-grid">
                        {(comment.media ?? []).map((item, index) => {
                          const kind = resolveMediaKind(item)
                          const title = resolveMediaTitle(item)
                          const mediaUrl = resolveMediaUrl(item.url)
                          const displayUrl = resolveMediaDisplayUrl(item)
                          const ctaHref = originalPostUrl ?? mediaUrl
                          const ctaLabel = originalPostUrl ? 'Оригинальный пост' : 'Открыть файл'

                          return (
                            <article
                              key={`${comment.source_comment_id}-${item.url ?? item.title ?? item.kind}-${index}`}
                              className="post-drawer__comment-media-card"
                            >
                              {(kind === 'image' || kind === 'video') && displayUrl ? <img src={displayUrl} alt={title} loading="lazy" /> : null}

                              <div className="post-drawer__comment-media-info">
                                <strong>{mediaKindLabel(kind)}</strong>
                                <span title={title}>{title}</span>
                                {ctaHref ? (
                                  <a href={ctaHref} target="_blank" rel="noreferrer">
                                    {ctaLabel}
                                  </a>
                                ) : null}
                              </div>
                            </article>
                          )
                        })}
                      </div>
                    ) : null}
                  </article>
                )
              })}
            </div>
          </section>
        ) : null}
      </aside>
    </div>
  )
}
