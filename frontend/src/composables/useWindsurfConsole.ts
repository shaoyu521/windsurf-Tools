import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus/es/components/message/index'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import {
  AddSingleAccount,
  ApplySeamlessPatch,
  AutoSwitchToNext,
  CheckPatchStatus,
  DeleteAccount,
  DeleteExpiredAccounts,
  DeleteFreePlanAccounts,
  FindWindsurfPath,
  GetAllAccounts,
  GetSettings,
  ImportByAPIKey,
  ImportByEmailPassword,
  ImportByJWT,
  ImportByRefreshToken,
  RefreshAllTokens,
  RestoreSeamlessPatch,
  SwitchAccount,
  UpdateSettings,
} from '../../wailsjs/go/main/App'
import type { Account, AddMode, ImportMode, ImportResult, SettingsForm } from '../types/windsurf'
import { getAccountHealth, getPlanTone } from '../utils/account'
import { importBatched } from '../utils/importBatch'

function hasNativeBridge(): boolean {
  return typeof window !== 'undefined' && !!(window as { go?: unknown }).go
}

function createPreviewAccount(overrides: Partial<Account>): Account {
  const baseNow = new Date()
  return {
    id: overrides.id ?? `preview-${Math.random().toString(36).slice(2, 10)}`,
    email: overrides.email ?? 'demo@windsurf.local',
    nickname: overrides.nickname ?? 'Demo Account',
    plan_name: overrides.plan_name ?? 'pro',
    used_quota: overrides.used_quota ?? 1480,
    total_quota: overrides.total_quota ?? 5000,
    daily_remaining: overrides.daily_remaining ?? '72.4%',
    weekly_remaining: overrides.weekly_remaining ?? '58.1%',
    daily_reset_at: overrides.daily_reset_at ?? new Date(baseNow.getTime() + 5 * 60 * 60 * 1000).toISOString(),
    weekly_reset_at: overrides.weekly_reset_at ?? new Date(baseNow.getTime() + 2 * 24 * 60 * 60 * 1000).toISOString(),
    subscription_expires_at:
      overrides.subscription_expires_at ?? new Date(baseNow.getTime() + 20 * 24 * 60 * 60 * 1000).toISOString(),
    token_expires_at: overrides.token_expires_at ?? new Date(baseNow.getTime() + 30 * 60 * 1000).toISOString(),
    status: overrides.status ?? 'active',
    remark: overrides.remark ?? '',
    token: overrides.token ?? 'ey-preview.demo.token.value',
    refresh_token: overrides.refresh_token ?? 'rt-preview-demo-token',
    windsurf_api_key: overrides.windsurf_api_key,
    last_login_at: overrides.last_login_at ?? new Date(baseNow.getTime() - 2 * 60 * 60 * 1000).toISOString(),
    last_quota_update:
      overrides.last_quota_update ?? new Date(baseNow.getTime() - 30 * 60 * 1000).toISOString(),
    created_at: overrides.created_at ?? new Date(baseNow.getTime() - 12 * 24 * 60 * 60 * 1000).toISOString(),
    tags: overrides.tags ?? '',
  }
}

function createPreviewAccounts(): Account[] {
  const now = new Date()

  return [
    createPreviewAccount({
      id: 'preview-pro',
      email: 'alpha@studio.dev',
      nickname: 'Alpha Pro',
      plan_name: 'pro',
      remark: '主力创作账号',
      daily_remaining: '72.4%',
      weekly_remaining: '58.1%',
      used_quota: 1480,
      total_quota: 5000,
      windsurf_api_key: 'sk-ws-01-preview-alpha-control',
      last_login_at: new Date(now.getTime() - 35 * 60 * 1000).toISOString(),
    }),
    createPreviewAccount({
      id: 'preview-team',
      email: 'ops@team.dev',
      nickname: 'Ops Team',
      plan_name: 'teams',
      remark: '协作池',
      daily_remaining: '38.8%',
      weekly_remaining: '64.3%',
      used_quota: 2120,
      total_quota: 8000,
      refresh_token: 'rt-preview-team-ops',
      last_quota_update: new Date(now.getTime() - 10 * 60 * 1000).toISOString(),
    }),
    createPreviewAccount({
      id: 'preview-trial',
      email: 'trial@camp.dev',
      nickname: 'Trial Camp',
      plan_name: 'trial',
      remark: '需要尽快补位',
      daily_remaining: '12.0%',
      weekly_remaining: '18.6%',
      used_quota: 860,
      total_quota: 1000,
      token: 'ey-preview-low-balance-token',
      token_expires_at: new Date(now.getTime() + 10 * 60 * 1000).toISOString(),
      subscription_expires_at: new Date(now.getTime() + 4 * 24 * 60 * 60 * 1000).toISOString(),
    }),
    createPreviewAccount({
      id: 'preview-expired',
      email: 'legacy@archive.dev',
      nickname: 'Legacy Seat',
      plan_name: 'free',
      remark: '示例过期状态',
      daily_remaining: '',
      weekly_remaining: '',
      used_quota: 0,
      total_quota: 0,
      status: 'expired',
      token: '',
      refresh_token: '',
      subscription_expires_at: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000).toISOString(),
      last_quota_update: new Date(now.getTime() - 8 * 24 * 60 * 60 * 1000).toISOString(),
    }),
  ]
}

function createPreviewSettings(): SettingsForm {
  return {
    proxy_enabled: true,
    proxy_url: 'http://127.0.0.1:7890',
    windsurf_path: 'C:\\Users\\Demo\\AppData\\Roaming\\Windsurf',
    concurrent_limit: 5,
    seamless_switch: false,
    auto_refresh_tokens: true,
    auto_refresh_quotas: true,
    quota_refresh_policy: 'hybrid',
    quota_custom_interval_minutes: 360,
  }
}

export function useWindsurfConsole() {
  const accounts = ref<Account[]>([])
  const previewMode = ref(false)
  const bootstrapping = ref(true)
  const listLoading = ref(false)
  const actionLoading = ref(false)
  const patchStatus = ref<boolean | null>(null)
  const windsurfPath = ref('')
  const proxyEnabled = ref(false)
  const proxyURL = ref('')
  const autoRefresh = ref(false)
  const autoRefreshQuotas = ref(false)
  const quotaRefreshPolicy = ref('hybrid')
  const quotaCustomIntervalMinutes = ref(360)

  const busy = computed(() => listLoading.value || actionLoading.value)

  function clampQuotaCustomMinutes(m: number): number {
    if (!Number.isFinite(m) || m <= 0) {
      return 360
    }
    return Math.min(10080, Math.max(5, Math.round(m)))
  }

  const quotaRefreshLabel = computed(() => {
    if (!autoRefreshQuotas.value) {
      return '额度同步：关'
    }
    const p = quotaRefreshPolicy.value || 'hybrid'
    const cm = clampQuotaCustomMinutes(quotaCustomIntervalMinutes.value)
    const labels: Record<string, string> = {
      hybrid: '开（美东换日或满24h）',
      interval_24h: '开（每24小时）',
      us_calendar: '开（仅美东换日）',
      local_calendar: '开（本机时区跨日）',
      interval_1h: '开（每1小时）',
      interval_6h: '开（每6小时）',
      interval_12h: '开（每12小时）',
      custom: `开（自定义每 ${cm} 分钟）`,
    }
    return labels[p] || `开（${p}）`
  })
  const accountCount = computed(() => accounts.value.length)
  const proCount = computed(() => accounts.value.filter((account) => getPlanTone(account.plan_name) === 'pro').length)
  const maxCount = computed(() => accounts.value.filter((account) => getPlanTone(account.plan_name) === 'max').length)
  const teamCount = computed(() => accounts.value.filter((account) => getPlanTone(account.plan_name) === 'team').length)
  const enterpriseCount = computed(() => accounts.value.filter((account) => getPlanTone(account.plan_name) === 'enterprise').length)
  const criticalCount = computed(() => accounts.value.filter((account) => getAccountHealth(account) === 'critical').length)
  const healthyCount = computed(() => accounts.value.filter((account) => getAccountHealth(account) === 'healthy').length)
  const paidCount = computed(() => proCount.value + maxCount.value + teamCount.value + enterpriseCount.value)

  function syncSettingsLocally(settings: SettingsForm) {
    proxyEnabled.value = settings.proxy_enabled
    proxyURL.value = settings.proxy_url || ''
    windsurfPath.value = settings.windsurf_path || ''
    autoRefresh.value = settings.auto_refresh_tokens
    autoRefreshQuotas.value = settings.auto_refresh_quotas ?? false
    quotaRefreshPolicy.value = settings.quota_refresh_policy || 'hybrid'
    quotaCustomIntervalMinutes.value = clampQuotaCustomMinutes(
      settings.quota_custom_interval_minutes ?? 360,
    )
  }

  function loadPreviewState() {
    const previewSettings = createPreviewSettings()
    previewMode.value = true
    accounts.value = createPreviewAccounts()
    syncSettingsLocally(previewSettings)
    patchStatus.value = previewSettings.seamless_switch
    bootstrapping.value = false
  }

  async function loadAccounts() {
    if (previewMode.value) {
      accounts.value = createPreviewAccounts()
      return
    }

    listLoading.value = true
    try {
      accounts.value = (await GetAllAccounts()) || []
    } catch (error) {
      ElMessage.error(`加载账号失败: ${String(error)}`)
    } finally {
      listLoading.value = false
    }
  }

  async function detectWindsurfPath() {
    if (previewMode.value) {
      if (!windsurfPath.value) {
        windsurfPath.value = createPreviewSettings().windsurf_path
      }
      return
    }

    try {
      windsurfPath.value = await FindWindsurfPath()
      patchStatus.value = windsurfPath.value ? await CheckPatchStatus(windsurfPath.value) : null
    } catch {
      windsurfPath.value = ''
      patchStatus.value = null
    }
  }

  async function loadSettings() {
    if (previewMode.value) {
      syncSettingsLocally(createPreviewSettings())
      return
    }

    try {
      const settings = await GetSettings()
      syncSettingsLocally({
        proxy_enabled: settings.proxy_enabled || false,
        proxy_url: settings.proxy_url || '',
        windsurf_path: settings.windsurf_path || '',
        concurrent_limit: settings.concurrent_limit || 5,
        seamless_switch: settings.seamless_switch || false,
        auto_refresh_tokens: settings.auto_refresh_tokens || false,
        auto_refresh_quotas: settings.auto_refresh_quotas || false,
        quota_refresh_policy: settings.quota_refresh_policy || 'hybrid',
        quota_custom_interval_minutes: settings.quota_custom_interval_minutes ?? 360,
      })
    } catch {
      // Settings are optional during bootstrap. Silent fallback keeps the shell usable.
    }
  }

  async function bootstrap() {
    bootstrapping.value = true

    if (!hasNativeBridge()) {
      loadPreviewState()
      return
    }

    previewMode.value = false

    try {
      await Promise.all([loadAccounts(), detectWindsurfPath(), loadSettings()])
    } finally {
      bootstrapping.value = false
    }
  }

  function buildPreviewImportResult(email: string, success = true): ImportResult {
    return { email, success }
  }

  async function handleImport(importMode: ImportMode, rawText: string): Promise<boolean> {
    const lines = rawText
      .trim()
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)

    if (!lines.length) {
      ElMessage.warning('请输入导入数据')
      return false
    }

    actionLoading.value = true

    try {
      let results: ImportResult[] = []

      if (previewMode.value) {
        const imported = lines.map((line, index) => {
          const pieces = line.split(/\s+/)
          const hint = pieces[0]
          const remark = pieces.slice(1).join(' ')
          const email =
            importMode === 'password'
              ? pieces[0] || `preview-${Date.now()}-${index}@demo.local`
              : hint.includes('@')
                ? hint
                : `preview-${hint.slice(0, 6) || index}@demo.local`

          return createPreviewAccount({
            id: `preview-import-${Date.now()}-${index}`,
            email,
            nickname: remark || email.split('@')[0],
            remark,
            plan_name: index % 2 === 0 ? 'pro' : 'trial',
            daily_remaining: index % 2 === 0 ? '66.0%' : '27.4%',
            weekly_remaining: index % 2 === 0 ? '81.2%' : '41.0%',
            windsurf_api_key: importMode === 'api_key' ? hint : undefined,
            refresh_token: importMode === 'refresh_token' ? hint : undefined,
            token: importMode === 'jwt' ? hint : 'ey-preview-imported-token',
          })
        })

        accounts.value = [...imported, ...accounts.value]
        results = imported.map((account) => buildPreviewImportResult(account.email))
      } else if (importMode === 'api_key') {
        const items = lines.map((line) => {
          const parts = line.split(/\s+/)
          return {
            api_key: parts[0],
            remark: parts.slice(1).join(' '),
          }
        })
        results = await importBatched(items, (slice) => ImportByAPIKey(slice))
      } else if (importMode === 'refresh_token') {
        const items = lines.map((line) => {
          const parts = line.split(/\s+/)
          return {
            token: parts[0],
            remark: parts.slice(1).join(' '),
          }
        })
        results = await importBatched(items, (slice) => ImportByRefreshToken(slice))
      } else if (importMode === 'jwt') {
        const items = lines.map((line) => {
          const parts = line.split(/\s+/)
          return {
            jwt: parts[0],
            remark: parts.slice(1).join(' '),
          }
        })
        results = await importBatched(items, (slice) => ImportByJWT(slice))
      } else {
        const items = lines.map((line) => {
          const parts = line.split(/\s+/)
          return {
            email: parts[0] || '',
            password: parts[1] || '',
            alt_password: '',
            remark: parts.slice(2).join(' '),
          }
        })
        results = await importBatched(items, (slice) => ImportByEmailPassword(slice))
      }

      const succeeded = results.filter((item) => item.success).length
      const failed = results.length - succeeded

      if (succeeded > 0) {
        ElMessage.success(`导入完成，成功 ${succeeded} 条${failed ? `，失败 ${failed} 条` : ''}`)
      } else {
        ElMessage.error(`导入失败：${results.find((item) => item.error)?.error || '未知错误'}`)
        return false
      }

      if (!previewMode.value) {
        await loadAccounts()
      }

      return true
    } catch (error) {
      ElMessage.error(`导入失败: ${String(error)}`)
      return false
    } finally {
      actionLoading.value = false
    }
  }

  async function handleAddSingle(addMode: AddMode, value: string, remark: string): Promise<boolean> {
    if (addMode === 'password') {
      try {
        const cred = JSON.parse(value) as { email?: string; password?: string }
        if (!cred.email?.trim() || !cred.password) {
          ElMessage.warning('请填写邮箱和密码')
          return false
        }
      } catch {
        ElMessage.warning('邮箱密码格式错误')
        return false
      }
    } else if (!value.trim()) {
      ElMessage.warning('请输入账号凭证')
      return false
    }

    actionLoading.value = true

    try {
      if (previewMode.value) {
        let syntheticEmail = `added-${Date.now().toString().slice(-4)}@demo.local`
        if (addMode === 'jwt') {
          syntheticEmail = `jwt-${Date.now().toString().slice(-4)}@demo.local`
        }
        if (addMode === 'password') {
          try {
            const cred = JSON.parse(value) as { email?: string }
            if (cred.email?.trim()) {
              syntheticEmail = cred.email.trim()
            }
          } catch {
            /* keep default */
          }
        }

        accounts.value = [
          createPreviewAccount({
            id: `preview-single-${Date.now()}`,
            email: syntheticEmail,
            nickname: remark || syntheticEmail.split('@')[0],
            remark,
            plan_name: 'pro',
            daily_remaining: '64.0%',
            weekly_remaining: '73.0%',
            windsurf_api_key: addMode === 'api_key' ? value : undefined,
            refresh_token: addMode === 'refresh_token' ? value : undefined,
            token: addMode === 'jwt' ? value : 'ey-preview-added-token',
            password: addMode === 'password' ? '••••' : undefined,
          }),
          ...accounts.value,
        ]

        ElMessage.success(`添加成功: ${syntheticEmail}`)
        return true
      }

      const result = await AddSingleAccount(addMode, value.trim(), remark.trim())
      if (result.success) {
        ElMessage.success(`添加成功: ${result.email}`)
        await loadAccounts()
        return true
      } else {
        ElMessage.error(`添加失败: ${result.error || '未知错误'}`)
        return false
      }
    } catch (error) {
      ElMessage.error(`添加失败: ${String(error)}`)
      return false
    } finally {
      actionLoading.value = false
    }
  }

  async function handleSwitch(id: string, email: string) {
    actionLoading.value = true

    try {
      if (previewMode.value) {
        ElMessage.success(`预览模式：已模拟切换到 ${email}`)
        return
      }

      await SwitchAccount(id)
      ElMessage.success(`已切换到: ${email}`)
    } catch (error) {
      ElMessage.error(`切换失败: ${String(error)}`)
    } finally {
      actionLoading.value = false
    }
  }

  async function handleAutoSwitch(currentId: string) {
    actionLoading.value = true

    try {
      if (previewMode.value) {
        const candidate = accounts.value.find(
          (account) => account.id !== currentId && getAccountHealth(account) !== 'expired',
        )

        if (!candidate) {
          ElMessage.warning('预览模式下没有可切换的示例账号')
          return
        }

        ElMessage.success(`预览模式：已模拟切换到 ${candidate.email}`)
        return
      }

      const email = await AutoSwitchToNext(currentId, 'all')
      ElMessage.success(`额度不足，已切换到: ${email}`)
      await loadAccounts()
    } catch (error) {
      ElMessage.error(`自动切换失败: ${String(error)}`)
    } finally {
      actionLoading.value = false
    }
  }

  async function handleDelete(id: string, email: string) {
    try {
      await ElMessageBox.confirm(`确定删除账号 ${email} 吗？`, '删除确认', {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消',
      })

      actionLoading.value = true

      if (previewMode.value) {
        accounts.value = accounts.value.filter((account) => account.id !== id)
        ElMessage.success('预览数据已删除')
        return
      }

      await DeleteAccount(id)
      ElMessage.success('账号已删除')
      await loadAccounts()
    } catch (error) {
      if (error !== 'cancel') {
        ElMessage.error(`删除失败: ${String(error)}`)
      }
    } finally {
      actionLoading.value = false
    }
  }

  async function handleDeleteExpired() {
    try {
      await ElMessageBox.confirm('确定清理所有已过期账号吗？', '批量清理', {
        type: 'warning',
        confirmButtonText: '清理',
        cancelButtonText: '取消',
      })

      actionLoading.value = true

      if (previewMode.value) {
        const before = accounts.value.length
        accounts.value = accounts.value.filter((account) => getAccountHealth(account) !== 'expired')
        ElMessage.success(`预览数据已清理 ${before - accounts.value.length} 个账号`)
        return
      }

      const deleted = await DeleteExpiredAccounts()
      ElMessage.success(`已删除 ${deleted} 个过期账号`)
      await loadAccounts()
    } catch (error) {
      if (error !== 'cancel') {
        ElMessage.error(`清理失败: ${String(error)}`)
      }
    } finally {
      actionLoading.value = false
    }
  }

  async function handleDeleteFreePlans() {
    const freeCount = accounts.value.filter((a) => getPlanTone(a.plan_name) === 'free').length
    if (freeCount === 0) {
      ElMessage.info('当前没有免费（Free / Basic）计划账号')
      return
    }
    try {
      await ElMessageBox.confirm(
        `将永久删除 ${freeCount} 个免费计划账号（计划名含 free 或 basic），不可恢复。`,
        '删除免费账号',
        { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
      )

      actionLoading.value = true

      if (previewMode.value) {
        const before = accounts.value.length
        accounts.value = accounts.value.filter((account) => getPlanTone(account.plan_name) !== 'free')
        ElMessage.success(`预览数据已删除 ${before - accounts.value.length} 个免费账号`)
        return
      }

      const deleted = await DeleteFreePlanAccounts()
      ElMessage.success(`已删除 ${deleted} 个免费账号`)
      await loadAccounts()
    } catch (error) {
      if (error !== 'cancel') {
        ElMessage.error(`删除失败: ${String(error)}`)
      }
    } finally {
      actionLoading.value = false
    }
  }

  async function handleApplyPatch() {
    if (!windsurfPath.value) {
      ElMessage.warning('请先检测 Windsurf 路径')
      return
    }

    actionLoading.value = true

    try {
      if (previewMode.value) {
        patchStatus.value = true
        ElMessage.success('预览模式：已模拟应用补丁')
        return
      }

      const result = await ApplySeamlessPatch(windsurfPath.value)
      patchStatus.value = true
      ElMessage.success(result.already_patched ? '补丁已存在' : '补丁已应用')
    } catch (error) {
      ElMessage.error(`补丁失败: ${String(error)}`)
    } finally {
      actionLoading.value = false
    }
  }

  async function handleRestorePatch() {
    try {
      await ElMessageBox.confirm('确定还原当前补丁吗？', '还原确认', {
        type: 'warning',
        confirmButtonText: '还原',
        cancelButtonText: '取消',
      })

      actionLoading.value = true

      if (previewMode.value) {
        patchStatus.value = false
        ElMessage.success('预览模式：已模拟还原补丁')
        return
      }

      await RestoreSeamlessPatch(windsurfPath.value)
      patchStatus.value = false
      ElMessage.success('补丁已还原')
    } catch (error) {
      if (error !== 'cancel') {
        ElMessage.error(`还原失败: ${String(error)}`)
      }
    } finally {
      actionLoading.value = false
    }
  }

  async function handleRefreshAll() {
    actionLoading.value = true

    try {
      if (previewMode.value) {
        accounts.value = accounts.value.map((account, index) => ({
          ...account,
          last_quota_update: new Date().toISOString(),
          daily_remaining: index === 2 ? '18.0%' : account.daily_remaining,
        }))

        ElMessage.success(`预览模式：已刷新 ${accounts.value.length} 条示例数据`)
        return
      }

      const result = await RefreshAllTokens()
      const entries = Object.entries(result || {})
      const successCount = entries.filter(([, value]) => String(value).includes('成功')).length
      ElMessage.success(`刷新完成: ${successCount}/${entries.length}`)
      await loadAccounts()
    } catch (error) {
      ElMessage.error(`刷新失败: ${String(error)}`)
    } finally {
      actionLoading.value = false
    }
  }

  async function saveSettings(): Promise<boolean> {
    actionLoading.value = true

    const payload: SettingsForm = {
      proxy_enabled: proxyEnabled.value,
      proxy_url: proxyURL.value.trim(),
      windsurf_path: windsurfPath.value.trim(),
      concurrent_limit: 5,
      seamless_switch: !!patchStatus.value,
      auto_refresh_tokens: autoRefresh.value,
      auto_refresh_quotas: autoRefreshQuotas.value,
      quota_refresh_policy: quotaRefreshPolicy.value || 'hybrid',
      quota_custom_interval_minutes: clampQuotaCustomMinutes(quotaCustomIntervalMinutes.value),
    }

    try {
      if (previewMode.value) {
        syncSettingsLocally(payload)
        ElMessage.success('预览模式：设置已保存到本地界面状态')
        return true
      }

      await UpdateSettings(payload)
      ElMessage.success('设置已保存')
      return true
    } catch (error) {
      ElMessage.error(`保存失败: ${String(error)}`)
      return false
    } finally {
      actionLoading.value = false
    }
  }

  onMounted(() => {
    void bootstrap()
  })

  return {
    accounts,
    previewMode,
    bootstrapping,
    listLoading,
    actionLoading,
    busy,
    patchStatus,
    windsurfPath,
    proxyEnabled,
    proxyURL,
    autoRefresh,
    autoRefreshQuotas,
    quotaRefreshPolicy,
    quotaCustomIntervalMinutes,
    quotaRefreshLabel,
    accountCount,
    proCount,
    maxCount,
    teamCount,
    enterpriseCount,
    paidCount,
    criticalCount,
    healthyCount,
    bootstrap,
    loadAccounts,
    detectWindsurfPath,
    loadSettings,
    handleImport,
    handleAddSingle,
    handleSwitch,
    handleAutoSwitch,
    handleDelete,
    handleDeleteExpired,
    handleDeleteFreePlans,
    handleApplyPatch,
    handleRestorePatch,
    handleRefreshAll,
    saveSettings,
  }
}
