import type { MediaItem } from "@/entities/post/model/types";
import { env } from "@/shared/config/env";

export type MediaKind =
  | "image"
  | "video"
  | "audio"
  | "document"
  | "link"
  | "unknown";

const imageExt = /\.(jpg|jpeg|png|gif|webp|bmp|svg)$/i;
const videoExt = /\.(mp4|mov|avi|mkv|webm)$/i;
const audioExt = /\.(mp3|ogg|wav|m4a|aac|flac)$/i;
const docExt = /\.(pdf|doc|docx|txt|zip|rar|7z)$/i;

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
  const normalizedUrl = url.replace(/\\/g, "/");
  const fromUrl = normalizedUrl.split("/").filter(Boolean).at(-1);
  if (fromUrl) return fromUrl;

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
    if (isImageUrl(mediaUrl)) return mediaUrl;
    if (isImageUrl(previewUrl)) return previewUrl;
  }

  if (kind === "video") {
    return previewUrl ?? mediaUrl;
  }

  return undefined;
};

export const mediaKindLabel = (kind: MediaKind): string => {
  switch (kind) {
    case "image":
      return "Image";
    case "video":
      return "Video";
    case "audio":
      return "Audio";
    case "document":
      return "Document";
    case "link":
      return "Link";
    default:
      return "Attachment";
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
