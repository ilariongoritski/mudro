import type { MediaItem, Post } from '@/entities/post/model/types'
import { formatDateTime } from '@/shared/lib/format/date'
import './PostCard.css'

interface PostCardProps {
  post: Post
}

type MediaKind = 'image' | 'video' | 'audio' | 'document' | 'link' | 'unknown'

const imageExt = /\.(jpg|jpeg|png|gif|webp|bmp|svg)$/i
const videoExt = /\.(mp4|mov|avi|mkv|webm)$/i
const audioExt = /\.(mp3|ogg|wav|m4a|aac|flac)$/i
const docExt = /\.(pdf|doc|docx|txt|zip|rar|7z)$/i

const resolveMediaKind = (item: MediaItem): MediaKind => {
  if (item.is_image) return 'image'
  if (item.is_video) return 'video'
  if (item.is_audio) return 'audio'
  if (item.is_document) return 'document'
  if (item.is_link) return 'link'

  const kind = (item.kind || '').toLowerCase()
  if (['photo', 'image', 'gif'].includes(kind)) return 'image'
  if (kind === 'video') return 'video'
  if (kind === 'audio' || kind === 'voice') return 'audio'
  if (['doc', 'file', 'document'].includes(kind)) return 'document'
  if (kind === 'link') return 'link'

  const probe = `${item.url ?? ''} ${item.title ?? ''}`.toLowerCase()
  if (imageExt.test(probe)) return 'image'
  if (videoExt.test(probe)) return 'video'
  if (audioExt.test(probe)) return 'audio'
  if (docExt.test(probe)) return 'document'
  if (probe.includes('http://') || probe.includes('https://')) return 'link'

  return 'unknown'
}

const mediaKindLabel = (kind: MediaKind): string => {
  switch (kind) {
    case 'image':
      return 'Image'
    case 'video':
      return 'Video'
    case 'audio':
      return 'Audio'
    case 'document':
      return 'Document'
    case 'link':
      return 'Link'
    default:
      return 'Attachment'
  }
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

      <p className="post-card__text">{post.text?.trim() || 'No text'}</p>

      <div className="post-card__stats">
        <span>Likes: {post.likes_count}</span>
        <span>Views: {post.views_count ?? '—'}</span>
        <span>Comments: {post.comments_count ?? '—'}</span>
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
          {post.media.map((item, index) => {
            const kind = resolveMediaKind(item)
            const title = item.title || 'Untitled'

            return (
              <div key={`${item.url ?? item.title ?? item.kind}-${index}`} className="post-media-card">
                {kind === 'image' && item.url ? <img src={item.url} loading="lazy" alt={title} /> : null}
                {kind === 'video' && item.preview_url ? <img src={item.preview_url} loading="lazy" alt={title} /> : null}

                <div className="post-media-card__info">
                  <strong>{mediaKindLabel(kind)}</strong>
                  <span>{title}</span>
                  {item.url ? (
                    <a href={item.url} target="_blank" rel="noreferrer">
                      Open
                    </a>
                  ) : null}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </article>
  )
}
