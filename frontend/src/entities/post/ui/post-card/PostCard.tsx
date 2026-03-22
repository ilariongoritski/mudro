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
import { Card, CardContent, CardFooter, CardHeader } from '@/shared/ui/card'

interface PostCardProps {
  post: Post
  onOpen?: (post: Post) => void
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

export const PostCard = ({ post, onOpen }: PostCardProps) => {
  const reactions = normalizeReactions(post.reactions)
  const mediaItems = post.media ?? []
  const visibleMedia = mediaItems.slice(0, 3)
  const hiddenMediaCount = Math.max(mediaItems.length - visibleMedia.length, 0)
  const previewComments = (post.comments ?? []).slice(0, 2)
  const totalComments = post.comments_count ?? post.comments?.length ?? 0
  const remainingComments = Math.max(totalComments - previewComments.length, 0)
  const bodyText = post.text?.trim() || 'Для этого поста текст пока не подтянулся.'
  const viewsMetric = metricDisplay(post.views_count)
  const originalPostUrl = buildOriginalPostUrl(post.source, post.source_post_id)

  return (
    <Card className={cn('transition-shadow', onOpen && 'hover:shadow-md cursor-pointer')}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Badge variant={post.source === 'vk' ? 'vk' : 'tg'}>{post.source.toUpperCase()}</Badge>
            <span className="text-xs text-slate-400">#{post.source_post_id}</span>
          </div>
          <time className="text-xs text-slate-400">{formatDateTime(post.published_at)}</time>
        </div>
      </CardHeader>

      <CardContent className="space-y-3">
        <p className="text-sm text-slate-700 leading-relaxed line-clamp-4">{bodyText}</p>

        <div className="grid grid-cols-3 gap-2 rounded-lg bg-slate-50 p-2.5">
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

        {mediaItems.length > 0 && (
          <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
            {visibleMedia.map((item, index) => {
              const kind = resolveMediaKind(item)
              const title = resolveMediaTitle(item)
              const mediaUrl = resolveMediaUrl(item.url)
              const displayUrl = resolveMediaDisplayUrl(item)
              const showOverlay = hiddenMediaCount > 0 && index === visibleMedia.length - 1
              const ctaHref = originalPostUrl ?? mediaUrl

              return (
                <div key={`${item.url ?? item.title ?? item.kind}-${index}`} className="relative rounded-lg overflow-hidden bg-slate-100 border border-slate-200">
                  {(kind === 'image' || kind === 'video') && displayUrl ? (
                    <img src={displayUrl} loading="lazy" alt={title} className="w-full h-28 object-cover" />
                  ) : null}
                  {showOverlay && (
                    <span className="absolute inset-0 flex items-center justify-center bg-black/50 text-white font-bold text-lg">
                      +{hiddenMediaCount}
                    </span>
                  )}
                  <div className="p-1.5 text-[10px] space-y-0.5">
                    <strong className="block text-slate-600">{mediaKindLabel(kind)}</strong>
                    <span className="block text-slate-400 truncate" title={title}>{title}</span>
                    {ctaHref && (
                      <a href={ctaHref} target="_blank" rel="noreferrer" onClick={(e) => e.stopPropagation()} className="text-sky-600 hover:underline">
                        {originalPostUrl ? 'Оригинал' : 'Открыть'}
                      </a>
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        )}

        {previewComments.length > 0 && (
          <div className="border-t border-slate-100 pt-3 space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-slate-500">Обсуждение</span>
              <span className="text-xs text-slate-400">{metricLabel(totalComments)} в треде</span>
            </div>
            <div className="space-y-2">
              {previewComments.map((comment) => {
                const commentReactions = normalizeCommentReactions(comment.reactions)

                return (
                  <div
                    key={`${comment.source_comment_id}-${comment.parent_comment_id ?? 'root'}`}
                    className={cn('text-xs space-y-1', comment.parent_comment_id && 'ml-4 pl-3 border-l-2 border-slate-200')}
                  >
                    <div className="flex items-center gap-2 text-slate-400">
                      <strong className="text-slate-600">{humanizeCommentAuthor(comment.author_name)}</strong>
                      {comment.parent_comment_id && <span>ответ на #{comment.parent_comment_id}</span>}
                    </div>
                    <p className="text-slate-600">{comment.text?.trim() || 'Без текста'}</p>

                    {commentReactions.length > 0 && (
                      <div className="flex flex-wrap gap-1">
                        {commentReactions.map(([reaction, count]) => (
                          <span key={`${comment.source_comment_id}-${reaction}`} className="inline-flex items-center gap-0.5 rounded-full bg-slate-50 px-1.5 py-0.5 text-[10px]" title={reaction}>
                            {reactionLabel(reaction)} {count}
                          </span>
                        ))}
                      </div>
                    )}

                    {(comment.media ?? []).length > 0 && (
                      <div className="flex gap-1.5 mt-1">
                        {(comment.media ?? []).slice(0, 2).map((item, index) => {
                          const kind = resolveMediaKind(item)
                          const title = resolveMediaTitle(item)
                          const displayUrl = resolveMediaDisplayUrl(item)

                          if ((kind === 'image' || kind === 'video') && displayUrl) {
                            return (
                              <img
                                key={`${comment.source_comment_id}-${item.url ?? item.title ?? item.kind}-${index}`}
                                className="w-12 h-12 rounded object-cover"
                                src={displayUrl}
                                loading="lazy"
                                alt={title}
                              />
                            )
                          }

                          return (
                            <span
                              key={`${comment.source_comment_id}-${item.url ?? item.title ?? item.kind}-${index}`}
                              className="inline-flex items-center px-1.5 py-0.5 rounded bg-slate-100 text-[10px] text-slate-500"
                              title={title}
                            >
                              {mediaKindLabel(kind)}
                            </span>
                          )
                        })}
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
            {remainingComments > 0 && (
              <p className="text-xs text-slate-400">Еще {remainingComments} в обсуждении</p>
            )}
          </div>
        )}
      </CardContent>

      {(onOpen || originalPostUrl) && (
        <CardFooter className="gap-2">
          {onOpen && (
            <Button variant="outline" size="sm" onClick={() => onOpen(post)}>
              Подробнее
            </Button>
          )}
          {originalPostUrl && (
            <Button variant="ghost" size="sm" asChild>
              <a href={originalPostUrl} target="_blank" rel="noreferrer">Открыть пост</a>
            </Button>
          )}
        </CardFooter>
      )}
    </Card>
  )
}
