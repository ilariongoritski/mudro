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

  try {
    return (await response.json()) as T
  } catch {
    throw new Error(`Failed to parse JSON from ${input}`)
  }
}
