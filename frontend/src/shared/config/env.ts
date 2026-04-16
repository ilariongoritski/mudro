export const env = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? '/api',
  skaroDashboardUrl: import.meta.env.VITE_SKARO_DASHBOARD_URL ?? 'http://127.0.0.1:4700/dashboard',
  supabaseUrl: import.meta.env.VITE_SUPABASE_URL ?? '',
  supabaseAnonKey: import.meta.env.VITE_SUPABASE_ANON_KEY ?? '',
}
