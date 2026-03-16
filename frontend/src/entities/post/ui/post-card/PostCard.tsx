import type { Post, PostComment } from "@/entities/post/model/types";
import {
  mediaKindLabel,
  metricDisplay,
  metricLabel,
  normalizeReactions,
  reactionLabel,
  resolveMediaDisplayUrl,
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

const normalizeCommentReactions = (reactions?: PostComment["reactions"]) => {
  if (!reactions) return [] as Array<[string, number]>;

  if (Array.isArray(reactions)) {
    return reactions
      .filter((reaction) => reaction.count > 0)
      .map((reaction) => [reaction.raw || reaction.label, reaction.count] as [string, number]);
  }

  return normalizeReactions(reactions);
};

export const PostCard = ({ post, onOpen }: PostCardProps) => {
  const reactions = normalizeReactions(post.reactions);
  const mediaItems = post.media ?? [];
  const visibleMedia = mediaItems.slice(0, 3);
  const hiddenMediaCount = Math.max(mediaItems.length - visibleMedia.length, 0);
  const previewComments = (post.comments ?? []).slice(0, 2);
  const totalComments = post.comments_count ?? post.comments?.length ?? 0;
  const bodyText = post.text?.trim() || "Описание для этого поста пока не подтянулось.";
  const viewsMetric = metricDisplay(post.views_count);

  return (
    <article
      className={`post-card mudro-fade-up ${onOpen ? "post-card_interactive" : ""}`}
    >
      <header className="post-card__head">
        <div className="post-card__head-main">
          <div className={`post-card__source post-card__source_${post.source}`}>
            {post.source.toUpperCase()}
          </div>
          <div className="post-card__eyebrow">
            {post.source.toUpperCase()} #{post.source_post_id}
          </div>
        </div>
        <div className="post-card__meta">{formatDateTime(post.published_at)}</div>
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
          <strong className={viewsMetric.missing ? "post-card__metric-missing" : undefined}>
            {viewsMetric.value}
          </strong>
        </span>
        <span className="post-card__stat">
          <small>Комментарии</small>
          <strong>{metricLabel(totalComments)}</strong>
        </span>
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
          <div className="post-card__thread-list">
            {previewComments.map((comment) => {
              const commentReactions = normalizeCommentReactions(comment.reactions);

              return (
                <article
                  key={`${comment.source_comment_id}-${comment.parent_comment_id ?? "root"}`}
                  className={`post-card__thread-item ${comment.parent_comment_id ? "post-card__thread-item_reply" : ""}`}
                >
                  <div className="post-card__thread-meta">
                    <strong>{comment.author_name || "Без имени"}</strong>
                    {comment.parent_comment_id ? (
                      <span>ответ на #{comment.parent_comment_id}</span>
                    ) : null}
                  </div>
                  <p>{comment.text?.trim() || "Без текста"}</p>

                  {commentReactions.length > 0 ? (
                    <div className="post-card__thread-reactions">
                      {commentReactions.map(([reaction, count]) => (
                        <span
                          key={`${comment.source_comment_id}-${reaction}`}
                          className="post-card__thread-reaction"
                          title={reaction}
                        >
                          {reactionLabel(reaction)} {count}
                        </span>
                      ))}
                    </div>
                  ) : null}

                  {(comment.media ?? []).length > 0 ? (
                    <div className="post-card__thread-media">
                      {(comment.media ?? []).slice(0, 2).map((item, index) => {
                        const kind = resolveMediaKind(item);
                        const title = resolveMediaTitle(item);
                        const displayUrl = resolveMediaDisplayUrl(item);

                        if ((kind === "image" || kind === "video") && displayUrl) {
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
              );
            })}
          </div>
        </section>
      )}

      {onOpen ? (
        <footer className="post-card__footer">
          <button type="button" className="post-card__open" onClick={() => onOpen(post)}>
            Открыть пост
          </button>
        </footer>
      ) : null}
    </article>
  );
};


