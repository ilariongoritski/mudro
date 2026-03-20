import type { MediaItem } from "@/entities/post/model/types";
import { env } from "@/shared/config/env";

export type MediaKind =
  | "image"
  | "video"
  | "audio"
  | "document"
  | "link"
  | "unknown";

const imageExt = /\.(jpg|jpeg|png|gif|webp|bmp|svg)(?:$|[?#])/i;
const videoExt = /\.(mp4|mov|avi|mkv|webm)(?:$|[?#])/i;
const audioExt = /\.(mp3|ogg|wav|m4a|aac|flac)(?:$|[?#])/i;
const docExt = /\.(pdf|doc|docx|txt|zip|rar|7z)(?:$|[?#])/i;
const trailingCombiningMarks = /[\u0300-\u036f]+/g;

const resolveApiOrigin = (): string | null => {
  const base = env.apiBaseUrl.trim();
  if (!base || base === "/") return null;

  try {
    return new URL(base, window.location.origin).origin;
  } catch {
    return null;
  }
};

export const resolveMediaKind = (item: MediaItem): MediaKind => {
  if (item.is_image) return "image";
  if (item.is_video) return "video";
  if (item.is_audio) return "audio";
  if (item.is_document) return "document";
  if (item.is_link) return "link";

  const kind = (item.kind || "").toLowerCase();
  if (["photo", "image", "gif"].includes(kind)) return "image";
  if (kind === "video") return "video";
  if (kind === "audio" || kind === "voice") return "audio";
  if (["doc", "file", "document"].includes(kind)) return "document";
  if (kind === "link") return "link";

  const probe = `${item.url ?? ""} ${item.title ?? ""}`.toLowerCase();
  if (imageExt.test(probe)) return "image";
  if (videoExt.test(probe)) return "video";
  if (audioExt.test(probe)) return "audio";
  if (docExt.test(probe)) return "document";
  if (probe.includes("http://") || probe.includes("https://")) return "link";

  return "unknown";
};

export const resolveMediaUrl = (raw?: string): string | undefined => {
  const value = raw?.trim();
  if (!value) return undefined;
  if (value.startsWith("missing://")) return undefined;
  if (value.startsWith("http://") || value.startsWith("https://")) return value;

  const apiOrigin = resolveApiOrigin();
  if (value.startsWith("/")) {
    return apiOrigin ? new URL(value, apiOrigin).toString() : value;
  }

  const normalized = value.replace(/^\.?[\\/]/, "").replace(/\\/g, "/");
  const mediaPath = `/media/${normalized}`;
  return apiOrigin ? new URL(mediaPath, apiOrigin).toString() : mediaPath;
};

export const resolveMediaTitle = (item: MediaItem): string => {
  const explicit = item.title?.trim();
  if (explicit) return explicit;

  const url = item.url?.trim() ?? "";
  if (!url) return mediaKindLabel(resolveMediaKind(item));

  try {
    const parsed = new URL(url.replace(/\\/g, "/"));
    const fromPath = parsed.pathname.split("/").filter(Boolean).at(-1);
    if (fromPath) return fromPath;
  } catch {
    const normalizedUrl = url.replace(/\\/g, "/");
    const fromUrl = normalizedUrl.split("/").filter(Boolean).at(-1)?.split("?")[0];
    if (fromUrl) return fromUrl;
  }

  return mediaKindLabel(resolveMediaKind(item));
};

const isImageUrl = (value?: string): boolean => {
  if (!value) return false;
  return imageExt.test(value);
};

export const resolveMediaDisplayUrl = (item: MediaItem): string | undefined => {
  const kind = resolveMediaKind(item);
  const mediaUrl = resolveMediaUrl(item.url);
  const previewUrl = resolveMediaUrl(item.preview_url);

  if (kind === "image") {
    return mediaUrl ?? previewUrl;
  }

  if (kind === "video") {
    if (previewUrl) return previewUrl;
    if (isImageUrl(mediaUrl)) return mediaUrl;
  }

  return undefined;
};

export const mediaKindLabel = (kind: MediaKind): string => {
  switch (kind) {
    case "image":
      return "Изображение";
    case "video":
      return "Видео";
    case "audio":
      return "Аудио";
    case "document":
      return "Файл";
    case "link":
      return "Ссылка";
    default:
      return "Вложение";
  }
};

export const normalizeReactions = (reactions?: Record<string, number>) => {
  if (!reactions) return [];

  return Object.entries(reactions)
    .filter(([, count]) => count > 0)
    .sort((a, b) => b[1] - a[1]);
};

export const metricLabel = (value?: number | null): string => {
  if (value == null) return "-";

  return new Intl.NumberFormat("ru-RU").format(value);
};

export const metricDisplay = (value?: number | null) => {
  if (value == null) {
    return {
      value: "нет данных",
      missing: true,
    };
  }

  return {
    value: new Intl.NumberFormat("ru-RU").format(value),
    missing: false,
  };
};

export const reactionLabel = (raw: string): string => {
  if (!raw) return "•";
  if (raw.startsWith("emoji:")) return raw.slice("emoji:".length).trim() || "🙂";
  if (raw.startsWith("custom:")) return "✨";
  if (raw.startsWith("unknown:")) return "•";
  return raw;
};

export const buildOriginalPostUrl = (source: "vk" | "tg", sourcePostID: string): string | undefined => {
  const normalized = sourcePostID.trim();
  if (!normalized) return undefined;

  if (source === "vk") {
    if (normalized.includes("_")) {
      return `https://vk.com/wall${normalized}`;
    }
    return undefined;
  }

  if (/^\d+$/.test(normalized)) {
    return `https://t.me/tgmydro/${normalized}`;
  }

  return undefined;
};

export const humanizeCommentAuthor = (author?: string | null): string => {
  const normalized = author?.trim();
  if (!normalized) return "Без имени";
  if (/^Участник #\d+$/.test(normalized)) return normalized;
  return normalized.replace(trailingCombiningMarks, "");
};
