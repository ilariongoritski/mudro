export interface UserProfile {
  id: number;
  display_name: string;
  username: string;
  email?: string | null;
  age?: number | null;
  bio?: string | null;
  social_links?: Record<string, string>;
  avatar_url?: string | null;
  profile_completion: number;
  rating: number;
  telegram_username?: string | null;
  created_at: string;
  updated_at: string;
}

export interface ProfileUpdate {
  display_name?: string;
  username?: string;
  email?: string | null;
  age?: number | null;
  bio?: string | null;
  social_links?: Record<string, string>;
}

export interface CasinoStats {
  balance: string;
  games_count: number;
  max_win: string;
}

export interface Activity {
  id: number;
  type: string;
  ref_id?: number;
  metadata?: Record<string, any>;
  created_at: string;
}
