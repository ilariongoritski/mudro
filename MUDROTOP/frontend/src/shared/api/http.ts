export async function getJSON<T>(input: string, signal?: AbortSignal): Promise<T> {
  const response = await fetch(input, {
    signal,
    headers: {
      Accept: 'application/json',
    },
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  return response.json() as Promise<T>
}
