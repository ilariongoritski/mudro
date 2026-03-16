import type { Post } from "@/entities/post/model/types";
import {
  resolveMediaDisplayUrl,
  mediaKindLabel,
  metricLabel,
  normalizeReactions,
  resolveMediaKind,
  resolveMediaTitle,
  resolveMediaUrl,
} from "@/entities/post/lib/postPresentation";
import { formatDateTime } from "@/shared/lib/format/date";
import "./PostCard.css";

interface PostCardProps {
  post: Post;
  onOpen?: (post: Post) => void;
}

export const PostCard = ({ post, onOpen }: PostCardProps) => {
  const reactions = normalizeReactions(post.reactions);
  const mediaItems = post.media ?? [];
  const visibleMedia = mediaItems.slice(0, 3);
  const hiddenMediaCount = Math.max(mediaItems.length - visibleMedia.length, 0);
  const previewComments = (post.comments ?? []).slice(0, 2);
  const bodyText =
    post.text?.trim() || "Описание для этого поста пока не подтянулось.";

  return (
    <article
      className={`post-card mudro-fade-up ${onOpen ? "post-card_interactive" : ""}`}
    >
      <header className="post-card__head">
        <div className="post-card__head-main">
          <div className={`post-card__source post-card__source_${post.source}`}>
            {post.source.toUpperCase()}
          </div>
          <div className="post-card__eyebrow">Пост #{post.id}</div>
        </div>
        <div className="post-card__meta">
          {formatDateTime(post.published_at)}
        </div>
      </header>

      <div className="post-card__body">
        <p className="post-card__text">{bodyText}</p>
      </div>

      <div className="post-card__stats">
        <span className="post-card__stat">
          <small>Лайки</small>
          <strong>{metricLabel(post.likes_count)}</strong>
        </span>
        <span className="post-card__stat">
          <small>Просмотры</small>
          <strong>{metricLabel(post.views_count)}</strong>
        </span>
        <span className="post-card__stat">
          <small>Комментарии</small>
          <strong>{metricLabel(post.comments_count)}</strong>
        </span>
      </div>

      {reactions.length > 0 && (
        <div className="post-card__reactions">
          {reactions.map(([reaction, count]) => (
            <span key={reaction} className="post-reaction" title={reaction}>
              {reaction.replace("emoji:", "")} {count}
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
            const showOverlay =
              hiddenMediaCount > 0 && index === visibleMedia.length - 1;

            return (
              <div
                key={`${item.url ?? item.title ?? item.kind}-${index}`}
                className="post-media-card"
              >
                {(kind === "image" || kind === "video") && displayUrl ? (
                  <img src={displayUrl} loading="lazy" alt={title} />
                ) : null}
                {showOverlay ? (
                  <span className="post-media-card__more">
                    +{hiddenMediaCount}
                  </span>
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
                      Open
                    </a>
                  ) : null}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {previewComments.length > 0 && (
        <section
          className="post-card__thread-preview"
          aria-label="Превью комментариев"
        >
          <div className="post-card__thread-head">
            <span>Тред</span>
            <strong>{metricLabel(post.comments_count)} комментария</strong>
          </div>
          <div className="post-card__thread-list">
            {previewComments.map((comment) => (
              <article
                key={`${comment.source_comment_id}-${comment.parent_comment_id ?? "root"}`}
                className="post-card__thread-item"
              >
                <div className="post-card__thread-meta">
                  <strong>{comment.author_name || "Без имени"}</strong>
                  {comment.parent_comment_id ? (
                    <span>ответ на #{comment.parent_comment_id}</span>
                  ) : null}
                </div>
                <p>{comment.text?.trim() || "Без текста"}</p>
                {(comment.media ?? []).length > 0 ? (
                  <div className="post-card__thread-media">
                    {(comment.media ?? []).slice(0, 2).map((item, index) => {
                      const kind = resolveMediaKind(item);
                      const title = resolveMediaTitle(item);
                      const displayUrl = resolveMediaDisplayUrl(item);

                      if (
                        (kind === "image" || kind === "video") &&
                        displayUrl
                      ) {
                        return (
                          <img
                            key={`${comment.source_comment_id}-${item.url ?? item.title ?? item.kind}-${index}`}
                            className="post-card__thread-media-thumb"
                            src={displayUrl}
                            loading="lazy"
                            alt={title}
                          />
                        );
                      }

                      return (
                        <span
                          key={`${comment.source_comment_id}-${item.url ?? item.title ?? item.kind}-${index}`}
                          className="post-card__thread-media-badge"
                          title={title}
                        >
                          {mediaKindLabel(kind)}
                        </span>
                      );
                    })}
                  </div>
                ) : null}
              </article>
            ))}
          </div>
        </section>
      )}

      {onOpen ? (
        <footer className="post-card__footer">
          <button
            type="button"
            className="post-card__open"
            onClick={() => onOpen(post)}
          >
            Открыть пост
          </button>
        </footer>
      ) : null}
    </article>
  );
};
