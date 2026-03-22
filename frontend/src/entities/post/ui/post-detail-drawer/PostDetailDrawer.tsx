import { useEffect } from 'react'
import { X } from 'lucide-react'

import { CommentForm } from '@/features/comment-form/ui/CommentForm'
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
import { cn } from '@/shared/lib/utils'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'

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
    <div className="fixed inset-0 z-50" role="dialog" aria-modal="true" aria-labelledby="post-drawer-title">
      <button
        type="button"
        className="absolute inset-0 bg-black/40 cursor-default"
        aria-label="Закрыть"
        onClick={onClose}
      />

      <aside className="absolute inset-y-0 right-0 w-full max-w-lg bg-white shadow-2xl overflow-y-auto">
        <header className="sticky top-0 z-10 flex items-center justify-between gap-3 px-5 py-4 bg-white/90 backdrop-blur border-b border-slate-200">
          <div className="flex items-center gap-2 min-w-0">
            <Badge variant={post.source === 'vk' ? 'vk' : 'tg'}>{post.source.toUpperCase()}</Badge>
            <span className="text-xs text-slate-400 truncate">#{post.source_post_id}</span>
            <h2 id="post-drawer-title" className="sr-only">Развернутый просмотр поста</h2>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            {originalPostUrl && (
              <Button variant="ghost" size="sm" asChild>
                <a href={originalPostUrl} target="_blank" rel="noreferrer">Оригинал</a>
              </Button>
            )}
            <Button variant="ghost" size="icon" onClick={onClose}>
              <X size={18} />
            </Button>
          </div>
        </header>

        <div className="p-5 space-y-5">
          <time className="text-xs text-slate-400">{formatDateTime(post.published_at)}</time>

          <div className="grid grid-cols-3 gap-2 rounded-lg bg-slate-50 p-3">
            <div className="text-center">
              <span className="block text-[10px] text-slate-400 uppercase">Лайки</span>
              <strong className="text-sm font-semibold">{metricLabel(post.likes_count)}</strong>
            </div>
            <div className="text-center">
              <span className="block text-[10px] text-slate-400 uppercase">Просмотры</span>
              <strong className={cn('text-sm font-semibold', viewsMetric.missing && 'text-slate-300')}>{viewsMetric.value}</strong>
            </div>
            <div className="text-center">
              <span className="block text-[10px] text-slate-400 uppercase">Комменты</span>
              <strong className="text-sm font-semibold">{metricLabel(totalComments)}</strong>
            </div>
          </div>

          {reactions.length > 0 && (
            <div className="flex flex-wrap gap-1.5">
              {reactions.map(([reaction, count]) => (
                <span key={reaction} className="inline-flex items-center gap-1 rounded-full bg-slate-100 px-2 py-0.5 text-xs" title={reaction}>
                  {reactionLabel(reaction)} {count}
                </span>
              ))}
            </div>
          )}

          <div>
            <span className="block text-xs font-medium text-slate-400 uppercase tracking-wide mb-1.5">Текст поста</span>
            <p className="text-sm text-slate-700 leading-relaxed whitespace-pre-wrap">{fullText}</p>
          </div>

          {mediaItems.length > 0 && (
            <div>
              <span className="block text-xs font-medium text-slate-400 uppercase tracking-wide mb-2">Вложения</span>
              <div className="grid grid-cols-2 gap-2">
                {mediaItems.map((item, index) => {
                  const kind = resolveMediaKind(item)
                  const title = resolveMediaTitle(item)
                  const mediaUrl = resolveMediaUrl(item.url)
                  const displayUrl = resolveMediaDisplayUrl(item)
                  const ctaHref = originalPostUrl ?? mediaUrl

                  return (
                    <div key={`${item.url ?? item.title ?? item.kind}-${index}`} className="rounded-lg overflow-hidden bg-slate-100 border border-slate-200">
                      {(kind === 'image' || kind === 'video') && displayUrl ? (
                        <img src={displayUrl} alt={title} loading="lazy" className="w-full h-32 object-cover" />
                      ) : null}
                      <div className="p-2 text-xs space-y-0.5">
                        <strong className="block text-slate-600">{mediaKindLabel(kind)}</strong>
                        <span className="block text-slate-400 truncate" title={title}>{title}</span>
                        {ctaHref && (
                          <a href={ctaHref} target="_blank" rel="noreferrer" className="text-sky-600 hover:underline">
                            {originalPostUrl ? 'Оригинал' : 'Открыть'}
                          </a>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )}

          {comments.length > 0 && (
            <div>
              <span className="block text-xs font-medium text-slate-400 uppercase tracking-wide mb-2">Комментарии и ответы</span>
              <div className="space-y-3">
                {comments.map((comment) => {
                  const commentReactions = normalizeCommentReactions(comment.reactions)

                  return (
                    <div
                      key={`${comment.source_comment_id}-${comment.parent_comment_id ?? 'root'}`}
                      className={cn('text-xs space-y-1.5 p-3 rounded-lg bg-slate-50', comment.parent_comment_id && 'ml-4 border-l-2 border-slate-200')}
                    >
                      <div className="flex items-center gap-2 text-slate-400">
                        <strong className="text-slate-700">{humanizeCommentAuthor(comment.author_name)}</strong>
                        <span>{formatDateTime(comment.published_at)}</span>
                        {comment.parent_comment_id && <span>ответ на #{comment.parent_comment_id}</span>}
                      </div>
                      <p className="text-slate-600">{comment.text?.trim() || 'Без текста'}</p>

                      {commentReactions.length > 0 && (
                        <div className="flex flex-wrap gap-1">
                          {commentReactions.map(([reaction, count]) => (
                            <span key={`${comment.source_comment_id}-${reaction}`} className="inline-flex items-center gap-0.5 rounded-full bg-white px-1.5 py-0.5 text-[10px]" title={reaction}>
                              {reactionLabel(reaction)} {count}
                            </span>
                          ))}
                        </div>
                      )}

                      {(comment.media ?? []).length > 0 && (
                        <div className="flex gap-2 mt-1">
                          {(comment.media ?? []).map((item, index) => {
                            const kind = resolveMediaKind(item)
                            const title = resolveMediaTitle(item)
                            const mediaUrl = resolveMediaUrl(item.url)
                            const displayUrl = resolveMediaDisplayUrl(item)
                            const ctaHref = originalPostUrl ?? mediaUrl

                            if ((kind === 'image' || kind === 'video') && displayUrl) {
                              return (
                                <img
                                  key={`${comment.source_comment_id}-${item.url ?? item.title ?? item.kind}-${index}`}
                                  className="w-16 h-16 rounded object-cover"
                                  src={displayUrl}
                                  loading="lazy"
                                  alt={title}
                                />
                              )
                            }

                            return (
                              <div
                                key={`${comment.source_comment_id}-${item.url ?? item.title ?? item.kind}-${index}`}
                                className="rounded bg-white p-1.5 text-[10px] space-y-0.5"
                              >
                                <strong className="block text-slate-600">{mediaKindLabel(kind)}</strong>
                                <span className="block text-slate-400 truncate" title={title}>{title}</span>
                                {ctaHref && (
                                  <a href={ctaHref} target="_blank" rel="noreferrer" className="text-sky-600 hover:underline">
                                    Открыть
                                  </a>
                                )}
                              </div>
                            )
                          })}
                        </div>
                      )}
                    </div>
                  )
                })}
              </div>
            </div>
          )}

          <CommentForm postId={post.id} />
        </div>
      </aside>
    </div>
  )
}
