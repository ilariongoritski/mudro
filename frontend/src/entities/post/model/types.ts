export type FeedSource = "all" | "vk" | "tg";
export type FeedSort = "desc" | "asc";

export interface MediaItem {
  kind: string;
  url?: string;
  preview_url?: string;
  title?: string;
  width?: number;
  height?: number;
  position?: number;
  is_image?: boolean;
  is_video?: boolean;
  is_audio?: boolean;
  is_document?: boolean;
  is_link?: boolean;
}

export interface PostComment {
  source_comment_id: string;
  parent_comment_id?: string;
  author_name: string;
  published_at: string;
  text: string;
  reactions?:
    | Record<string, number>
    | Array<{ label: string; count: number; raw: string }>;
  media?: MediaItem[];
}

export interface Post {
  id: number;
  source: "vk" | "tg";
  source_post_id: string;
  published_at: string;
  text?: string | null;
  media?: MediaItem[];
  likes_count: number;
  views_count?: number | null;
  comments_count?: number | null;
  comments?: PostComment[];
  reactions?: Record<string, number>;
  created_at: string;
  updated_at: string;
}

export interface SourceStat {
  source: "vk" | "tg";
  posts: number;
}

export interface FeedResponse {
  page?: number;
  limit: number;
  items: Post[];
  next_cursor?: FeedCursor;
}

export interface FeedCursor {
  before_ts: string;
  before_id: number;
}

export interface FrontResponse {
  meta: {
    total_posts: number;
    last_sync_at?: string;
    sources: SourceStat[];
  };
  feed: FeedResponse;
}

export interface FeedQueryArgs {
  limit: number;
  source: FeedSource;
  sort: FeedSort;
  q?: string;
}

export interface PostsQueryArgs extends FeedQueryArgs {
  page?: number;
  before_ts?: string;
  before_id?: number;
}
