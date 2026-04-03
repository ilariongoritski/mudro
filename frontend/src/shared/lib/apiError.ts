import type { SerializedError } from '@reduxjs/toolkit'
import type { FetchBaseQueryError } from '@reduxjs/toolkit/query'

/**
 * Extracts a human-readable message from RTK Query error objects.
 * Works with both FetchBaseQueryError and SerializedError.
 */
export function getErrorMessage(
  error: FetchBaseQueryError | SerializedError | undefined,
  fallback = 'Произошла ошибка. Попробуйте ещё раз.',
): string {
  if (!error) return fallback

  if ('status' in error) {
    // FetchBaseQueryError
    const data = error.data
    if (typeof data === 'string') return data
    if (data && typeof data === 'object' && 'error' in data) {
      return String((data as { error: string }).error)
    }
    if (data && typeof data === 'object' && 'message' in data) {
      return String((data as { message: string }).message)
    }
    if (error.status === 'FETCH_ERROR') return 'Сервер недоступен. Проверьте соединение.'
    if (error.status === 'TIMEOUT_ERROR') return 'Время ожидания истекло.'
    return fallback
  }

  // SerializedError
  return error.message ?? fallback
}
