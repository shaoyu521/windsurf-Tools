<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useAccountStore } from '../stores/useAccountStore'
import { useSettingsStore } from '../stores/useSettingsStore'
import ImportModal from '../components/accounts/ImportModal.vue'
import { Plus, Trash2, Power, RefreshCcw, Users, ChevronRight, KeyRound, BarChart3, UserX } from 'lucide-vue-next'
import { APIInfo } from '../api/wails'
import { getPlanTone } from '../utils/account'
import { models } from '../../wailsjs/go/models'
import SwitchPlanFilterControl from '../components/settings/SwitchPlanFilterControl.vue'
import { formatSwitchPlanFilterSummary, normalizeSwitchPlanFilter } from '../utils/settingsModel'

const accountStore = useAccountStore()
const settingsStore = useSettingsStore()
const showImportModal = ref(false)
const quotaRefreshingId = ref<string | null>(null)

const switchPlanFilter = ref('all')

watch(
  () => settingsStore.settings,
  (s) => {
    switchPlanFilter.value = normalizeSwitchPlanFilter(s?.auto_switch_plan_filter)
  },
  { immediate: true },
)

const planSectionOrder = ['pro', 'max', 'team', 'enterprise', 'trial', 'free', 'unknown'] as const

const planSectionLabels: Record<string, string> = {
  pro: 'Pro',
  max: 'Max / Ultimate',
  team: 'Teams',
  enterprise: 'Enterprise',
  trial: 'Trial',
  free: 'Free',
  unknown: '未识别计划',
}

const groupedAccounts = computed(() => {
  const groups = new Map<string, models.Account[]>()
  for (const k of planSectionOrder) {
    groups.set(k, [])
  }
  for (const acc of accountStore.accounts) {
    const tone = getPlanTone(acc.plan_name)
    if (!groups.has(tone)) {
      groups.set(tone, [])
    }
    groups.get(tone)!.push(acc)
  }
  return planSectionOrder
    .map((key) => ({
      key,
      title: planSectionLabels[key] || key,
      items: groups.get(key) || [],
    }))
    .filter((g) => g.items.length > 0)
})

const switchPoolLabel = computed(() => formatSwitchPlanFilterSummary(switchPlanFilter.value))

onMounted(() => {
  void accountStore.fetchAccounts()
})

const persistSwitchPool = async () => {
  try {
    await settingsStore.saveAutoSwitchPlanFilter(switchPlanFilter.value)
  } catch (e: unknown) {
    alert(`保存切号范围失败: ${String(e)}`)
  }
}

const onSwitchPlanFilterUpdate = (v: string) => {
  switchPlanFilter.value = v
  void persistSwitchPool()
}

const handleSwitch = async (id: string) => {
  try {
    await APIInfo.switchAccount(id)
    alert('账号切换成功。请重载或重启 Windsurf 后生效。')
  } catch (e: unknown) {
    alert(`切换失败: ${String(e)}`)
  }
}

const handleAutoNext = async (id: string) => {
  try {
    const email = await accountStore.autoSwitchToNext(id, switchPlanFilter.value)
    alert(`已切换到：${email}\n范围：${switchPoolLabel.value}`)
  } catch (e: unknown) {
    alert(`自动切号失败: ${String(e)}`)
  }
}

const handleDelete = async (id: string) => {
  if (confirm('是否确认移除该账号？')) {
    await accountStore.deleteAccount(id)
  }
}

const handleCleanExpired = async () => {
  try {
    const n = await accountStore.cleanExpiredAccounts()
    alert(`已清理 ${n} 个过期账号`)
  } catch (e: unknown) {
    alert(`清理失败: ${String(e)}`)
  }
}

const handleDeleteFreePlans = async () => {
  const n = freePlanAccountCount.value
  if (n === 0) {
    alert('当前没有识别为免费（Free / Basic）计划的账号')
    return
  }
  if (!confirm(`将永久删除 ${n} 个免费计划账号（计划名含 free 或 basic），不可恢复。确定？`)) {
    return
  }
  try {
    const deleted = await accountStore.deleteFreePlanAccounts()
    alert(`已删除 ${deleted} 个免费账号`)
  } catch (e: unknown) {
    alert(`删除失败: ${String(e)}`)
  }
}

const handleRefreshTokens = async () => {
  try {
    const map = await accountStore.refreshAllTokens()
    const entries = Object.entries(map || {})
    const ok = entries.filter(([, v]) => String(v).includes('成功')).length
    alert(`刷新完成：${ok} / ${entries.length}`)
  } catch (e: unknown) {
    alert(`刷新失败: ${String(e)}`)
  }
}

const handleRefreshAllQuotas = async () => {
  try {
    const map = await accountStore.refreshAllQuotas()
    const entries = Object.entries(map || {})
    const synced = entries.filter(([, v]) => String(v).includes('已同步')).length
    const skipped = entries.filter(([, v]) => String(v).includes('跳过')).length
    alert(`额度同步完成：已更新 ${synced} 个，跳过 ${skipped} 个，共 ${entries.length} 条`)
  } catch (e: unknown) {
    alert(`同步额度失败: ${String(e)}`)
  }
}

const handleRefreshOneQuota = async (id: string, email: string) => {
  quotaRefreshingId.value = id
  try {
    await accountStore.refreshAccountQuota(id)
    alert(`${email || '账号'} 额度已更新`)
  } catch (e: unknown) {
    alert(`刷新额度失败: ${String(e)}`)
  } finally {
    quotaRefreshingId.value = null
  }
}

const parseQuotaWidth = (str: string) => {
  if (!str) {
    return '0%'
  }
  const n = parseFloat(String(str).replace('%', '').trim())
  if (!Number.isFinite(n)) {
    return '0%'
  }
  return `${Math.max(0, Math.min(100, n))}%`
}

const getQuotaColor = (str: string) => {
  const n = parseFloat(String(str).replace('%', '').trim())
  if (!Number.isFinite(n)) {
    return 'bg-gray-400'
  }
  if (n > 50) {
    return 'bg-ios-green'
  }
  if (n > 20) {
    return 'bg-yellow-500'
  }
  return 'bg-ios-red'
}
</script>

<template>
  <div class="p-8 h-full flex flex-col max-w-6xl mx-auto w-full">
    <div class="flex flex-wrap items-center justify-between gap-4 mb-6 shrink-0">
      <div>
        <h1 class="text-[32px] font-bold tracking-tight">账号池</h1>
        <p class="text-[13px] text-ios-textSecondary dark:text-ios-textSecondaryDark mt-1">
          按 Pro / Teams / Trial 等分组；「下一席位」仅在下方勾选的计划池内切换（可多选）。
        </p>
      </div>
      <div class="flex flex-wrap gap-2 justify-end items-center">
        <div
          class="flex flex-col gap-1.5 mr-1 px-3 py-2 rounded-2xl bg-violet-500/10 border border-violet-500/15 max-w-full sm:max-w-[420px]"
        >
          <span class="text-[11px] font-semibold text-violet-800 dark:text-violet-300">下一席位范围</span>
          <SwitchPlanFilterControl
            variant="compact"
            :model-value="switchPlanFilter"
            @update:model-value="onSwitchPlanFilterUpdate"
          />
        </div>
        <button
          type="button"
          class="no-drag-region flex items-center px-4 py-2 bg-black/5 dark:bg-white/10 text-ios-text dark:text-ios-textDark rounded-full font-semibold text-[14px] ios-btn hover:bg-black/10 dark:hover:bg-white/15 transition-colors disabled:opacity-50"
          :disabled="accountStore.actionLoading"
          @click="handleRefreshTokens"
        >
          <KeyRound class="w-[18px] h-[18px] mr-1.5" stroke-width="2.5" />
          刷新凭证
        </button>
        <button
          type="button"
          class="no-drag-region flex items-center px-4 py-2 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 rounded-full font-semibold text-[14px] ios-btn hover:bg-emerald-500/15 transition-colors disabled:opacity-50"
          :disabled="accountStore.actionLoading"
          title="忽略定时策略，立即拉取所有账号日/周额度"
          @click="handleRefreshAllQuotas"
        >
          <BarChart3 class="w-[18px] h-[18px] mr-1.5" stroke-width="2.5" />
          同步全部额度
        </button>
        <button
          type="button"
          class="no-drag-region flex items-center px-4 py-2 bg-ios-red/10 text-ios-red dark:text-ios-redDark rounded-full font-semibold text-[14px] ios-btn hover:bg-ios-red/20 transition-colors"
          @click="handleCleanExpired"
        >
          <Trash2 class="w-[18px] h-[18px] mr-1.5" stroke-width="2.5" />
          清理过期
        </button>
        <button
          type="button"
          class="no-drag-region flex items-center px-4 py-2 bg-amber-500/12 text-amber-900 dark:text-amber-300 rounded-full font-semibold text-[14px] ios-btn hover:bg-amber-500/18 transition-colors disabled:opacity-50"
          :disabled="accountStore.actionLoading || freePlanAccountCount === 0"
          title="删除计划归类为 Free / Basic 的账号（与号池「Free」分组一致）"
          @click="handleDeleteFreePlans"
        >
          <UserX class="w-[18px] h-[18px] mr-1.5" stroke-width="2.5" />
          删除免费
        </button>
        <button
          type="button"
          class="no-drag-region flex items-center px-5 py-2.5 bg-gradient-to-b from-[#3b82f6] to-ios-blue text-white rounded-full font-semibold text-[14px] ios-btn shadow-md shadow-ios-blue/20 ring-1 ring-black/5 ring-inset active:ring-black/10 transition-all"
          @click="showImportModal = true"
        >
          <Plus class="w-[18px] h-[18px] mr-1" stroke-width="2.5" />
          批量导入
        </button>
      </div>
    </div>

    <div v-if="accountStore.isLoading" class="flex justify-center items-center p-20 flex-1">
      <RefreshCcw class="w-10 h-10 animate-spin text-ios-textSecondary opacity-30" />
    </div>

    <div
      v-else-if="accountStore.accounts.length === 0"
      class="flex flex-col items-center justify-center flex-1 text-ios-textSecondary"
    >
      <div class="w-24 h-24 mb-6 rounded-3xl bg-black/5 dark:bg-white/5 flex items-center justify-center">
        <Users class="w-12 h-12 opacity-50" />
      </div>
      <p class="text-[17px] font-medium">账号簿为空</p>
      <p class="text-[15px] opacity-70 mt-1">点击「批量导入」添加凭证</p>
    </div>

    <div v-else class="overflow-y-auto pb-10 space-y-10">
      <section v-for="section in groupedAccounts" :key="section.key" class="space-y-4">
        <div class="flex items-center gap-3 sticky top-0 z-10 py-2 -mx-1 px-1 bg-ios-bg/90 dark:bg-ios-bgDark/90 backdrop-blur-md border-b border-black/[0.06] dark:border-white/[0.08]">
          <h2 class="text-[18px] font-bold tracking-tight">{{ section.title }}</h2>
          <span
            class="text-[12px] font-semibold px-2.5 py-0.5 rounded-full bg-black/5 dark:bg-white/10 text-ios-textSecondary dark:text-ios-textSecondaryDark"
          >
            {{ section.items.length }} 个
          </span>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5 auto-rows-max">
          <div
            v-for="acc in section.items"
            :key="acc.id"
            class="ios-glass p-5 rounded-[22px] flex flex-col relative overflow-hidden group hover:shadow-xl hover:-translate-y-0.5 transition-all duration-300 ease-out border border-black/[0.04] dark:border-white/[0.06]"
          >
            <div
              class="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-white/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"
            />

            <div class="flex justify-between items-start mb-4 z-10">
              <div class="overflow-hidden pr-2 min-w-0">
                <h3 class="font-semibold text-[17px] mb-1 truncate" :title="acc.email">
                  {{ acc.nickname || acc.email || '未命名' }}
                </h3>
                <p
                  class="text-[12px] text-ios-textSecondary dark:text-ios-textSecondaryDark truncate"
                  :title="acc.email"
                >
                  {{ acc.email }}
                </p>
                <div class="flex items-center flex-wrap gap-2 mt-2">
                  <span
                    class="px-2 py-0.5 bg-gradient-to-br from-ios-blue/10 to-[#60A5FA]/20 dark:from-ios-blue/20 text-ios-blue dark:text-[#60A5FA] rounded-md font-bold text-[10px] uppercase tracking-wider ring-1 ring-ios-blue/10"
                  >
                    {{ acc.plan_name || 'unknown' }}
                  </span>
                  <span
                    v-if="acc.remark"
                    class="text-[12px] bg-black/5 dark:bg-white/10 px-2 py-0.5 rounded-md text-ios-textSecondary font-medium truncate max-w-[140px]"
                  >
                    {{ acc.remark }}
                  </span>
                </div>
              </div>

              <div class="flex gap-1.5 shrink-0">
                <button
                  type="button"
                  class="no-drag-region w-9 h-9 flex items-center justify-center rounded-full bg-ios-blue/10 text-ios-blue hover:bg-ios-blue/20 transition ios-btn"
                  title="切换到此账号"
                  @click="handleSwitch(acc.id)"
                >
                  <Power class="w-[18px] h-[18px]" stroke-width="2.5" />
                </button>
                <button
                  type="button"
                  class="no-drag-region w-9 h-9 flex items-center justify-center rounded-full bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 hover:bg-emerald-500/18 transition ios-btn disabled:opacity-40"
                  title="手动刷新此账号额度"
                  :disabled="quotaRefreshingId === acc.id || accountStore.actionLoading"
                  @click="handleRefreshOneQuota(acc.id, acc.email)"
                >
                  <RefreshCcw
                    class="w-[16px] h-[16px]"
                    :class="{ 'animate-spin': quotaRefreshingId === acc.id }"
                    stroke-width="2.5"
                  />
                </button>
                <button
                  type="button"
                  class="no-drag-region w-9 h-9 flex items-center justify-center rounded-full bg-black/5 dark:bg-white/10 text-ios-textSecondary hover:bg-black/10 dark:hover:bg-white/15 transition ios-btn"
                  :title="`下一席位（${switchPoolLabel}）`"
                  @click="handleAutoNext(acc.id)"
                >
                  <ChevronRight class="w-[18px] h-[18px]" stroke-width="2.5" />
                </button>
                <button
                  type="button"
                  class="no-drag-region w-9 h-9 flex items-center justify-center rounded-full bg-ios-red/10 text-ios-red hover:bg-ios-red/20 transition ios-btn"
                  title="删除"
                  @click="handleDelete(acc.id)"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
            </div>

            <p
              v-if="acc.last_quota_update"
              class="text-[10px] text-ios-textSecondary dark:text-ios-textSecondaryDark opacity-80 mb-2 truncate"
              :title="acc.last_quota_update"
            >
              额度同步：{{ acc.last_quota_update.slice(0, 16).replace('T', ' ') }}
            </p>

            <div class="mt-auto space-y-3">
              <div>
                <div
                  class="flex items-center justify-between text-[11px] font-semibold text-ios-textSecondary dark:text-ios-textSecondaryDark mb-1.5 uppercase tracking-wide"
                >
                  <span>每日额度</span>
                  <span>{{ acc.daily_remaining || '—' }}</span>
                </div>
                <div class="h-1.5 w-full bg-black/5 dark:bg-white/10 rounded-full overflow-hidden p-px">
                  <div
                    class="h-full rounded-full transition-all duration-500 ease-out"
                    :class="getQuotaColor(acc.daily_remaining || '')"
                    :style="{ width: parseQuotaWidth(acc.daily_remaining || '') }"
                  />
                </div>
              </div>

              <div v-if="acc.weekly_remaining">
                <div
                  class="flex items-center justify-between text-[11px] font-semibold text-ios-textSecondary dark:text-ios-textSecondaryDark mb-1.5 uppercase tracking-wide"
                >
                  <span>周期额度</span>
                  <span>{{ acc.weekly_remaining }}</span>
                </div>
                <div class="h-1.5 w-full bg-black/5 dark:bg-white/10 rounded-full overflow-hidden p-px">
                  <div
                    class="h-full rounded-full transition-all duration-500 ease-out"
                    :class="getQuotaColor(acc.weekly_remaining || '')"
                    :style="{ width: parseQuotaWidth(acc.weekly_remaining || '') }"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <ImportModal :isOpen="showImportModal" @close="showImportModal = false" />
  </div>
</template>
