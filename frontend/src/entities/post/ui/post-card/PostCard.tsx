import type { Post, PostComment } from "@/entities/post/model/types";
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
  const [liked, setLiked] = useState(false)
  const [localLikes, setLocalLikes] = useState(post.likes_count)
  const [toggleLike] = useToggleLikeMutation()

  const handleLike = async (e: React.MouseEvent) => {
    e.stopPropagation()
    setLiked(!liked)
    setLocalLikes(liked ? localLikes - 1 : localLikes + 1)
    try {
      const result = await toggleLike(post.id).unwrap()
      setLiked(result.liked)
      setLocalLikes(result.likes_count)
    } catch {
      setLiked(liked)
      setLocalLikes(localLikes)
    }
  }

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
    <article
      className={`post-card mudro-fade-up ${onOpen ? "post-card_interactive" : ""}`}
      onClick={() => onOpen?.(post)}
    >
      <header className={`post-card__head post-card__source_${post.source}`}>
        <div className="post-card__source-avatar">
          {post.source[0].toUpperCase()}
        </div>
        <div className="post-card__head-info">
          <div className="post-card__source-name">
            {post.source === 'tg' ? 'Telegram' : 'ВКонтакте'}
          </div>
          <div className="post-card__meta">{formatDateTime(post.published_at)}</div>
        </div>
      </header>

      <div className="post-card__body">
        <p className="post-card__text">{bodyText}</p>
      </div>

      <div className="post-card__actions">
        <div className="post-card__action">
          <svg viewBox="0 0 24 24"><path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg>
          {metricLabel(post.likes_count)}
        </div>
        <div className="post-card__action">
          <svg viewBox="0 0 24 24"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
          {metricLabel(totalComments)}
        </div>
        <div className="post-card__action">
          <svg viewBox="0 0 24 24"><path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"/></svg>
          {viewsMetric.value}
        </div>
      </div>

      {reactions.length > 0 && (
        <div className="post-card__reactions">
          {reactions.map(([reaction, count]) => (
            <span key={reaction} className="post-reaction" title={reaction}>
              {reactionLabel(reaction)} {count}
            </span>
          ))}
        </div>
      )}

      {mediaItems.length > 0 && (
        <div className="post-card__media-grid">
          {visibleMedia.map((item, index) => {
            const kind = resolveMediaKind(item);
            const title = resolveMediaTitle(item);
            const mediaUrl = resolveMediaUrl(item.url);
            const displayUrl = resolveMediaDisplayUrl(item);
            const showOverlay = hiddenMediaCount > 0 && index === visibleMedia.length - 1;

            return (
              <div
                key={`${item.url ?? item.title ?? item.kind}-${index}`}
                className="post-media-card"
              >
                {(kind === "image" || kind === "video") && displayUrl ? (
                  <img src={displayUrl} loading="lazy" alt={title} />
                ) : null}
                {showOverlay ? (
                  <span className="post-media-card__more">+{hiddenMediaCount}</span>
                ) : null}

                <div className="post-media-card__info">
                  <strong>{mediaKindLabel(kind)}</strong>
                  <span>{title}</span>
                  {mediaUrl ? (
                    <a
                      href={mediaUrl}
                      target="_blank"
                      rel="noreferrer"
                      onClick={(event) => event.stopPropagation()}
                    >
                      Открыть оригинал
                    </a>
                  ) : null}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {previewComments.length > 0 && (
        <section className="post-card__thread-preview" aria-label="Превью комментариев">
          <div className="post-card__thread-head">
            <span>Обсуждение</span>
            <strong>{metricLabel(totalComments)} в треде</strong>
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

    </article>
  );
};


