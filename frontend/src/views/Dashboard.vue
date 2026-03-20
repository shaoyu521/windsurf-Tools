<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useAccountStore } from '../stores/useAccountStore'
import { useSystemStore } from '../stores/useSystemStore'
import { useSettingsStore } from '../stores/useSettingsStore'
import IToggle from '../components/ios/IToggle.vue'
import { Power, Users, Activity, FileLock2, Info, BarChart3, KeyRound } from 'lucide-vue-next'
import SwitchPlanFilterControl from '../components/settings/SwitchPlanFilterControl.vue'
import { formatSwitchPlanFilterSummary, normalizeSwitchPlanFilter } from '../utils/settingsModel'

const accountStore = useAccountStore()
const systemStore = useSystemStore()
const settingsStore = useSettingsStore()

const switchPlanFilter = ref('all')

watch(
  () => settingsStore.settings,
  (s) => {
    switchPlanFilter.value = normalizeSwitchPlanFilter(s?.auto_switch_plan_filter)
  },
  { immediate: true },
)

const switchPoolLabel = computed(() => formatSwitchPlanFilterSummary(switchPlanFilter.value))

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

onMounted(() => {
  accountStore.fetchAccounts()
  settingsStore.fetchSettings()
  systemStore.initSystemEnvironment()
})

const handlePatchToggle = async (newVal: boolean) => {
  if (newVal) {
    await systemStore.applySeamlessPatch()
  } else {
    await systemStore.restoreSeamlessPatch()
  }
}

const handleRefreshAllQuotas = async () => {
  try {
    const map = await accountStore.refreshAllQuotas()
    const entries = Object.entries(map || {})
    const synced = entries.filter(([, v]) => String(v).includes('已同步')).length
    alert(`额度已同步：${synced} / ${entries.length} 个账号`)
  } catch (e: unknown) {
    alert(`同步额度失败: ${String(e)}`)
  }
}

const handleRefreshAllTokens = async () => {
  try {
    const map = await accountStore.refreshAllTokens()
    const entries = Object.entries(map || {})
    const ok = entries.filter(([, v]) => String(v).includes('成功')).length
    alert(`凭证刷新：${ok} / ${entries.length}`)
  } catch (e: unknown) {
    alert(`刷新凭证失败: ${String(e)}`)
  }
}
</script>

<template>
  <div class="p-8 h-full flex flex-col max-w-6xl mx-auto w-full">
    <div class="flex items-center justify-between mb-8 shrink-0">
      <h1 class="text-[32px] font-bold tracking-tight">Windsurf 控制台</h1>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6 auto-rows-max">
      <!-- System Status Card -->
      <div class="ios-glass p-6 rounded-[24px] flex flex-col space-y-6">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 rounded-2xl bg-ios-blue/10 flex items-center justify-center text-ios-blue">
            <FileLock2 class="w-5 h-5" stroke-width="2.5" />
          </div>
          <div>
            <h2 class="text-[17px] font-semibold">无感切换底层补丁</h2>
            <p class="text-[13px] text-ios-textSecondary dark:text-ios-textSecondaryDark font-medium">拦截扣费协议并自动切换额度</p>
          </div>
        </div>
        
        <div class="p-4 bg-black/5 dark:bg-white/5 rounded-[16px] flex items-center justify-between">
          <div class="flex flex-col">
            <span class="text-[15px] font-semibold mb-0.5">运行状态</span>
            <span class="text-[13px] text-ios-textSecondary dark:text-ios-textSecondaryDark flex items-center">
              <span class="w-2 h-2 rounded-full mr-2" :class="systemStore.patchStatus ? 'bg-ios-green' : 'bg-ios-red'"></span>
              {{ systemStore.patchStatus ? '已开启并生效中' : '未开启 (可能无法无感切换)' }}
            </span>
          </div>
          <div class="flex flex-col items-center justify-center">
             <IToggle 
               :modelValue="systemStore.patchStatus" 
               @update:modelValue="handlePatchToggle" 
               :disabled="systemStore.isGlobalLoading || !systemStore.windsurfPath" 
             />
          </div>
        </div>

        <div class="rounded-[16px] bg-violet-500/8 dark:bg-violet-500/10 border border-violet-500/15 p-3 space-y-2">
          <p class="text-[13px] font-semibold">下一席位计划范围</p>
          <p class="text-[11px] text-ios-textSecondary dark:text-ios-textSecondaryDark">
            与「账号池」共用：{{ switchPoolLabel }}
          </p>
          <SwitchPlanFilterControl
            variant="compact"
            :model-value="switchPlanFilter"
            @update:model-value="onSwitchPlanFilterUpdate"
          />
        </div>

        <div v-if="!systemStore.windsurfPath" class="text-xs text-ios-red dark:text-ios-redDark flex items-start">
          <Info class="w-4 h-4 mr-1.5 shrink-0 mt-0.5" />
          无法定位本地 Windsurf 安装路径，请在「设置」中检测或手动填写 windsurf_path。
        </div>
        <div
          v-else
          class="text-[11px] font-mono text-ios-textSecondary dark:text-ios-textSecondaryDark truncate opacity-80"
          :title="systemStore.windsurfPath"
        >
          {{ systemStore.windsurfPath }}
        </div>
      </div>

      <!-- Quick Stats Card -->
      <div class="ios-glass p-6 rounded-[24px] flex flex-col space-y-6">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 rounded-2xl bg-ios-green/10 flex items-center justify-center text-ios-greenDark">
            <Activity class="w-5 h-5" stroke-width="2.5" />
          </div>
          <div>
            <h2 class="text-[17px] font-semibold">资产概览</h2>
            <p class="text-[13px] text-ios-textSecondary dark:text-ios-textSecondaryDark font-medium">账号池监控</p>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4 flex-1">
          <div class="p-4 bg-black/5 dark:bg-white/5 rounded-[16px] flex flex-col justify-center">
             <span class="text-[32px] font-bold text-ios-text tracking-tight h-10 flex items-center">
               {{ accountStore.accounts.length }}
             </span>
             <span class="text-[13px] font-medium text-ios-textSecondary dark:text-ios-textSecondaryDark flex items-center mt-1">
               <Users class="w-4 h-4 mr-1.5" />
               活跃账号数
             </span>
          </div>
          <div class="p-4 bg-black/5 dark:bg-white/5 rounded-[16px] flex flex-col justify-center">
             <span class="text-[32px] font-bold text-ios-text tracking-tight h-10 flex items-center">
               {{ settingsStore.settings?.auto_refresh_tokens ? '开启' : '关闭' }}
             </span>
             <span class="text-[13px] font-medium text-ios-textSecondary dark:text-ios-textSecondaryDark flex items-center mt-1">
               <Power class="w-4 h-4 mr-1.5" />
               自动保活策略
             </span>
          </div>
        </div>

        <div class="flex flex-wrap gap-2">
          <button
            type="button"
            class="no-drag-region flex-1 min-w-[140px] flex items-center justify-center gap-2 px-4 py-2.5 rounded-[14px] bg-emerald-500/10 text-emerald-800 dark:text-emerald-400 font-semibold text-[13px] ios-btn hover:bg-emerald-500/15 disabled:opacity-50"
            :disabled="accountStore.actionLoading"
            @click="handleRefreshAllQuotas"
          >
            <BarChart3 class="w-4 h-4" stroke-width="2.5" />
            同步全部额度
          </button>
          <button
            type="button"
            class="no-drag-region flex-1 min-w-[140px] flex items-center justify-center gap-2 px-4 py-2.5 rounded-[14px] bg-black/5 dark:bg-white/10 font-semibold text-[13px] ios-btn hover:bg-black/10 dark:hover:bg-white/15 disabled:opacity-50"
            :disabled="accountStore.actionLoading"
            @click="handleRefreshAllTokens"
          >
            <KeyRound class="w-4 h-4" stroke-width="2.5" />
            刷新全部凭证
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
