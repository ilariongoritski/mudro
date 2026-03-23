const apiBaseUrl =
  import.meta.env.VITE_MOVIE_CATALOG_API_BASE_URL?.trim() || 'http://127.0.0.1:8091/api'

export const env = {
  apiBaseUrl,
}
