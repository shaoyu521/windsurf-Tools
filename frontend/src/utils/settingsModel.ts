import { models } from '../../wailsjs/go/models'

/** 与 backend/utils/plan_tone.go PlanTone 顺序一致，用于排序与全选判定 */
export const SWITCH_PLAN_FILTER_TONES = [
  'pro',
  'max',
  'team',
  'enterprise',
  'trial',
  'free',
  'unknown',
] as const

export type SwitchPlanTone = (typeof SWITCH_PLAN_FILTER_TONES)[number]

/** 多选勾选列表（不含「全部」） */
export const switchPlanFilterToneOptions: Array<{ value: SwitchPlanTone; label: string }> = [
  { value: 'pro', label: 'Pro' },
  { value: 'max', label: 'Max / Ultimate' },
  { value: 'team', label: 'Teams' },
  { value: 'enterprise', label: 'Enterprise' },
  { value: 'trial', label: 'Trial' },
  { value: 'free', label: 'Free' },
  { value: 'unknown', label: '未识别' },
]

const TONE_ORDER = new Map(SWITCH_PLAN_FILTER_TONES.map((t, i) => [t, i]))

/** 下拉/兼容：含「全部」与单选旧值 */
export const switchPlanFilterOptions: Array<{ value: string; label: string }> = [
  { value: 'all', label: '全部计划（不限制）' },
  ...switchPlanFilterToneOptions.map((o) => ({ value: o.value, label: `仅 ${o.label}` })),
]

/** 与 backend/models/settings.go + wailsjs models.Settings 对齐 */
export function createDefaultSettings(): models.Settings {
  return new models.Settings({
    proxy_enabled: false,
    proxy_url: '',
    windsurf_path: '',
    concurrent_limit: 5,
    seamless_switch: false,
    auto_refresh_tokens: false,
    auto_refresh_quotas: false,
    quota_refresh_policy: 'hybrid',
    quota_custom_interval_minutes: 360,
    auto_switch_plan_filter: 'all',
  })
}

export function normalizeSettings(raw: unknown): models.Settings {
  const base = createDefaultSettings()
  if (!raw || typeof raw !== 'object') {
    return base
  }
  const s = raw as Record<string, unknown>
  return new models.Settings({
    proxy_enabled: Boolean(s.proxy_enabled),
    proxy_url: String(s.proxy_url ?? ''),
    windsurf_path: String(s.windsurf_path ?? ''),
    concurrent_limit: Math.max(1, Number(s.concurrent_limit) || 5),
    seamless_switch: Boolean(s.seamless_switch),
    auto_refresh_tokens: Boolean(s.auto_refresh_tokens),
    auto_refresh_quotas: Boolean(s.auto_refresh_quotas),
    quota_refresh_policy: String(s.quota_refresh_policy || 'hybrid'),
    quota_custom_interval_minutes: clampQuotaMinutes(Number(s.quota_custom_interval_minutes)),
    auto_switch_plan_filter: normalizeSwitchPlanFilter(String(s.auto_switch_plan_filter ?? 'all')),
  })
}

/** 规范化存储：all；或逗号分隔的合法 tone（去重、按固定顺序排序）。支持旧版单值 pro / trial 等。 */
export function normalizeSwitchPlanFilter(v: string | undefined | null): string {
  if (v == null || v === '' || v === 'undefined') {
    return 'all'
  }
  let s = String(v).trim().toLowerCase().replace(/，/g, ',')
  if (s === 'all') {
    return 'all'
  }
  const allowed = new Set<string>(SWITCH_PLAN_FILTER_TONES as unknown as string[])
  const parts = [
    ...new Set(
      s
        .split(',')
        .map((x) => x.trim())
        .filter(Boolean)
        .filter((x) => allowed.has(x)),
    ),
  ]
  if (parts.length === 0) {
    return 'all'
  }
  if (parts.length >= SWITCH_PLAN_FILTER_TONES.length) {
    return 'all'
  }
  parts.sort((a, b) => (TONE_ORDER.get(a as SwitchPlanTone) ?? 0) - (TONE_ORDER.get(b as SwitchPlanTone) ?? 0))
  return parts.join(',')
}

/** 用于界面展示当前范围文案 */
export function formatSwitchPlanFilterSummary(filter: string | undefined | null): string {
  const n = normalizeSwitchPlanFilter(filter ?? 'all')
  if (n === 'all') {
    return '全部计划（不限制）'
  }
  const labelByValue = Object.fromEntries(switchPlanFilterToneOptions.map((o) => [o.value, o.label]))
  return n
    .split(',')
    .map((t) => labelByValue[t] || t)
    .join('、')
}

export function clampQuotaMinutes(m: number): number {
  if (!Number.isFinite(m) || m <= 0) {
    return 360
  }
  return Math.min(10080, Math.max(5, Math.round(m)))
}

/** 与后端 JSON 字段一致，便于 reactive + v-model */
export type SettingsForm = {
  proxy_enabled: boolean
  proxy_url: string
  windsurf_path: string
  concurrent_limit: number
  seamless_switch: boolean
  auto_refresh_tokens: boolean
  auto_refresh_quotas: boolean
  quota_refresh_policy: string
  quota_custom_interval_minutes: number
  /** 无感下一席位：all 或逗号分隔多选，如 trial,pro */
  auto_switch_plan_filter: string
}

export function settingsToForm(s: models.Settings): SettingsForm {
  return {
    proxy_enabled: s.proxy_enabled,
    proxy_url: s.proxy_url || '',
    windsurf_path: s.windsurf_path || '',
    concurrent_limit: s.concurrent_limit || 5,
    seamless_switch: s.seamless_switch,
    auto_refresh_tokens: s.auto_refresh_tokens,
    auto_refresh_quotas: s.auto_refresh_quotas,
    quota_refresh_policy: s.quota_refresh_policy || 'hybrid',
    quota_custom_interval_minutes: clampQuotaMinutes(s.quota_custom_interval_minutes),
    auto_switch_plan_filter: normalizeSwitchPlanFilter(s.auto_switch_plan_filter),
  }
}

export function formToSettings(form: SettingsForm, patchApplied: boolean): models.Settings {
  return new models.Settings({
    proxy_enabled: form.proxy_enabled,
    proxy_url: form.proxy_url.trim(),
    windsurf_path: form.windsurf_path.trim(),
    concurrent_limit: Math.max(1, Math.round(form.concurrent_limit) || 5),
    seamless_switch: patchApplied,
    auto_refresh_tokens: form.auto_refresh_tokens,
    auto_refresh_quotas: form.auto_refresh_quotas,
    quota_refresh_policy: form.quota_refresh_policy || 'hybrid',
    quota_custom_interval_minutes: clampQuotaMinutes(form.quota_custom_interval_minutes),
    auto_switch_plan_filter: normalizeSwitchPlanFilter(form.auto_switch_plan_filter),
  })
}

export const quotaPolicyOptions: Array<{ value: string; label: string }> = [
  { value: 'hybrid', label: '美东换日或满 24h（推荐）' },
  { value: 'interval_24h', label: '固定每 24 小时' },
  { value: 'us_calendar', label: '仅美东日历跨日' },
  { value: 'local_calendar', label: '本机时区跨日' },
  { value: 'interval_1h', label: '每 1 小时' },
  { value: 'interval_6h', label: '每 6 小时' },
  { value: 'interval_12h', label: '每 12 小时' },
  { value: 'custom', label: '自定义间隔（分钟）' },
]
