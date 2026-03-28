export function buildSearchParams(
  values: Record<string, string | number | Array<string> | undefined>,
): string {
  const params = new URLSearchParams()

  Object.entries(values).forEach(([key, value]) => {
    if (value == null || value === '') {
      return
    }

    if (Array.isArray(value)) {
      value.forEach((item) => {
        if (item.trim()) {
          params.append(key, item)
        }
      })
      return
    }

    params.set(key, String(value))
  })

  const rendered = params.toString()
  return rendered ? `?${rendered}` : ''
}
