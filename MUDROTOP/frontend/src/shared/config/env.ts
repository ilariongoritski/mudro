const apiBaseUrl =
  import.meta.env.VITE_MOVIE_CATALOG_API_BASE_URL?.trim() || 'http://127.0.0.1:8091/api'

try {
  new URL(apiBaseUrl)
} catch {
  throw new Error(`Invalid VITE_MOVIE_CATALOG_API_BASE_URL: ${apiBaseUrl}`)
}

export const env = {
  apiBaseUrl,
}
