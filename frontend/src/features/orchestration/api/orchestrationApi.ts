import { mudroApi } from '@/shared/api/mudroApi'

export interface OrchestrationStat {
  label: string
  value: string
  tone?: 'neutral' | 'ok' | 'warn' | 'accent'
}

export interface OrchestrationStatusResponse {
  branch: string
  commit: string
  updated_at: string
  moscow_time: string
  dashboard_url: string
  api_endpoint: string
  state: string[]
  todo: string[]
  done: string[]
  status: OrchestrationStat[]
}

const SKARO_DASHBOARD_URL = 'http://127.0.0.1:4700/dashboard'
const ORCHESTRATION_ENDPOINT = '/orchestration/status'

const asString = (value: unknown, fallback = '') => {
  if (typeof value === 'string') return value.trim() || fallback
  if (typeof value === 'number') return String(value)
  return fallback
}

const asList = (value: unknown): string[] => {
  if (!Array.isArray(value)) return []
  return value
    .map((item) => asString(item))
    .filter(Boolean)
}

const asStatusList = (value: unknown): OrchestrationStat[] => {
  if (!Array.isArray(value)) return []
  const out: OrchestrationStat[] = []
  for (const item of value) {
    if (!item || typeof item !== 'object') continue
    const record = item as Record<string, unknown>
    const label = asString(record.label)
    const rawValue = asString(record.value)
    if (!label || !rawValue) continue
    const rawTone = asString(record.tone, 'neutral')
    const tone = rawTone === 'ok' || rawTone === 'warn' || rawTone === 'accent' ? rawTone : 'neutral'
    out.push({ label, value: rawValue, tone })
  }
  return out
}

const asSection = (value: unknown): Record<string, unknown> | null => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null
  return value as Record<string, unknown>
}

const extractNested = (raw: Record<string, unknown>, key: string): unknown => {
  if (key in raw) return raw[key]
  const sections = asSection(raw.sections)
  if (!sections) return undefined
  return sections[key]
}

export const normalizeOrchestrationStatus = (raw: unknown): OrchestrationStatusResponse => {
  const record = asSection(raw) ?? {}
  const now = new Date()

  return {
    branch: asString(record.branch, 'codex/casino-mvp'),
    commit: asString(record.commit, 'unknown'),
    updated_at: asString(record.updated_at, now.toISOString()),
    moscow_time: asString(record.moscow_time, now.toLocaleTimeString('ru-RU', {
      timeZone: 'Europe/Moscow',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })),
    dashboard_url: asString(record.dashboard_url, SKARO_DASHBOARD_URL),
    api_endpoint: asString(record.api_endpoint, ORCHESTRATION_ENDPOINT),
    state: asList(extractNested(record, 'state')),
    todo: asList(extractNested(record, 'todo')),
    done: asList(extractNested(record, 'done')),
    status: asStatusList(extractNested(record, 'status')),
  }
}

export const orchestrationApi = mudroApi.injectEndpoints({
  endpoints: (build) => ({
    getOrchestrationStatus: build.query<OrchestrationStatusResponse, void>({
      query: () => ({
        url: ORCHESTRATION_ENDPOINT,
        cache: 'no-store',
      }),
      transformResponse: normalizeOrchestrationStatus,
    }),
  }),
})

export const { useGetOrchestrationStatusQuery } = orchestrationApi
