import React from "react";
import type { Post, PostComment } from "@/entities/post/model/types";
import {
  mediaKindLabel,
  metricDisplay,
  metricLabel,
  normalizeReactions,
  reactionLabel,
  resolveMediaDisplayUrl,
  resolveMediaKind,
  resolveMediaPosterUrl,
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
      onClick={() => onOpen?.(post)}
      {...(onOpen && {
        tabIndex: 0,
        role: "button",
        "aria-label": `Открыть пост${post.text ? ": " + post.text.slice(0, 60) : ""}`,
        onKeyDown: (e: React.KeyboardEvent) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            onOpen(post);
          }
        },
      })}
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
        <div className="post-card__action" aria-label={`${metricLabel(post.likes_count)} лайков`}>
          <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg>
          <span aria-hidden="true">{metricLabel(post.likes_count)}</span>
        </div>
        <div className="post-card__action" aria-label={`${metricLabel(totalComments)} комментариев`}>
          <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
          <span aria-hidden="true">{metricLabel(totalComments)}</span>
        </div>
        <div className="post-card__action" aria-label={`${viewsMetric.value} просмотров`}>
          <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"/></svg>
          <span aria-hidden="true">{viewsMetric.value}</span>
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
            const posterUrl = resolveMediaPosterUrl(item);
            const showOverlay = hiddenMediaCount > 0 && index === visibleMedia.length - 1;

            return (
              <div
                key={`${item.url ?? item.title ?? item.kind}-${index}`}
                className="post-media-card"
              >
                {kind === "image" && displayUrl ? (
                  <img src={displayUrl} loading="lazy" alt={title} />
                ) : null}
                {kind === "video" && mediaUrl ? (
                  <video
                    src={mediaUrl}
                    poster={posterUrl}
                    preload="metadata"
                    muted
                    loop
                    playsInline
                  />
                ) : null}
                {showOverlay ? (
                  <span
                    className="post-media-card__more"
                    aria-label={`Ещё ${hiddenMediaCount} фото`}
                  >
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

    </article>
  );
};

