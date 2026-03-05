import { formatDateTime } from '@/shared/lib/format/date'
import type { MediaItem, Post } from '@/entities/post/model/types'
import './PostCard.css'

interface PostCardProps {
  post: Post
}

const mediaTypeLabel = (item: MediaItem): string => {
  if (item.is_image) return 'Изображение'
  if (item.is_video) return 'Видео'
  if (item.is_audio) return 'Аудио'
  if (item.is_document) return 'Документ'
  if (item.is_link) return 'Ссылка'
  return item.kind || 'Вложение'
}

const normalizeReactions = (reactions?: Record<string, number>) => {
  if (!reactions) return []

  return Object.entries(reactions)
    .filter(([, count]) => count > 0)
    .sort((a, b) => b[1] - a[1])
}

export const PostCard = ({ post }: PostCardProps) => {
  const reactions = normalizeReactions(post.reactions)

  return (
    <article className="post-card mudro-fade-up">
      <header className="post-card__head">
        <div className={`post-card__source post-card__source_${post.source}`}>{post.source.toUpperCase()}</div>
        <div className="post-card__meta">#{post.id} · {formatDateTime(post.published_at)}</div>
      </header>

      <p className="post-card__text">{post.text?.trim() || 'Без текста'}</p>

      <div className="post-card__stats">
        <span>❤️ {post.likes_count}</span>
        <span>👁 {post.views_count ?? '—'}</span>
        <span>💬 {post.comments_count ?? '—'}</span>
      </div>

      {reactions.length > 0 && (
        <div className="post-card__reactions">
          {reactions.map(([reaction, count]) => (
            <span key={reaction} className="post-reaction" title={reaction}>
              {reaction.replace('emoji:', '')} {count}
            </span>
          ))}
        </div>
      )}

      {post.media && post.media.length > 0 && (
        <div className="post-card__media-grid">
          {post.media.map((item, index) => (
            <div key={`${item.url ?? item.title ?? item.kind}-${index}`} className="post-media-card">
              {item.is_image && item.url ? <img src={item.url} loading="lazy" alt={item.title || 'Media'} /> : null}
              <div className="post-media-card__info">
                <strong>{mediaTypeLabel(item)}</strong>
                <span>{item.title || 'Без названия'}</span>
                {item.url ? (
                  <a href={item.url} target="_blank" rel="noreferrer">
                    Открыть
                  </a>
                ) : null}
              </div>
            </div>
          ))}
        </div>
      )}
    </article>
  )
}
