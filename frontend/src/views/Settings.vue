<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { useSettingsStore } from '../stores/useSettingsStore'
import { useSystemStore } from '../stores/useSystemStore'
import IToggle from '../components/ios/IToggle.vue'
import {
  createDefaultSettings,
  formToSettings,
  quotaPolicyOptions,
  settingsToForm,
  type SettingsForm,
} from '../utils/settingsModel'
import SwitchPlanFilterControl from '../components/settings/SwitchPlanFilterControl.vue'
import { Network, RefreshCcw, CheckCircle2, Loader2, FolderOpen } from 'lucide-vue-next'

const settingsStore = useSettingsStore()
const systemStore = useSystemStore()
const isSaving = ref(false)
const showSaved = ref(false)

const local = reactive<SettingsForm>(settingsToForm(createDefaultSettings()))

onMounted(() => {
  void settingsStore.fetchSettings()
})

watch(
  () => settingsStore.settings,
  (s) => {
    if (s && !isSaving.value) {
      Object.assign(local, settingsToForm(s))
      if (!local.windsurf_path.trim() && systemStore.windsurfPath) {
        local.windsurf_path = systemStore.windsurfPath
      }
    }
  },
  { immediate: true },
)

const handleDetectPath = async () => {
  const p = await systemStore.detectWindsurfPath()
  if (p) {
    local.windsurf_path = p
  }
}

const handleSave = async () => {
  isSaving.value = true
  try {
    const payload = formToSettings(local, systemStore.patchStatus)
    await settingsStore.updateSettings(payload)
    showSaved.value = true
    setTimeout(() => {
      showSaved.value = false
    }, 2000)
  } catch (e) {
    console.error(e)
    alert(`保存失败: ${String(e)}`)
  } finally {
    isSaving.value = false
  }
}
</script>

<template>
  <div class="p-8 h-full flex flex-col max-w-3xl mx-auto w-full relative">
    <div class="flex items-center justify-between mb-8 shrink-0">
      <div>
        <h1 class="text-[32px] font-bold tracking-tight">高级设置</h1>
        <p class="text-[13px] text-ios-textSecondary dark:text-ios-textSecondaryDark mt-1">
          选项与后端 JSON 字段一致，保存后立即作用于 Wails 服务
        </p>
      </div>
      <button
        type="button"
        @click="handleSave"
        :disabled="settingsStore.isLoading || isSaving"
        class="no-drag-region flex items-center px-6 py-2.5 bg-gradient-to-b from-[#3b82f6] to-ios-blue text-white rounded-full font-semibold text-[15px] ios-btn shadow-md shadow-ios-blue/20 ring-1 ring-black/5 ring-inset active:ring-black/10 relative overflow-hidden transition-all disabled:opacity-50 disabled:active:scale-100"
      >
        <span
          :class="{ 'opacity-0': showSaved || isSaving }"
          class="transition-opacity flex items-center tracking-wide"
          >保存配置</span
        >

        <div v-if="isSaving && !showSaved" class="absolute inset-0 flex items-center justify-center">
          <Loader2 class="w-5 h-5 animate-spin" />
        </div>

        <div
          v-if="showSaved"
          class="absolute inset-0 flex items-center justify-center bg-ios-green dark:bg-ios-greenDark transition-all duration-300"
        >
          <CheckCircle2 class="w-[18px] h-[18px] mr-1.5" stroke-width="2.5" />
          <span class="font-bold tracking-wide">已生效</span>
        </div>
      </button>
    </div>

    <Transition name="fade" mode="out-in">
      <div v-if="settingsStore.isLoading" class="space-y-7 overflow-y-auto pb-10 w-full">
        <div class="ios-glass rounded-[24px] p-6 animate-pulse flex flex-col space-y-6 opacity-60">
          <div class="h-5 w-40 bg-black/10 dark:bg-white/10 rounded-md" />
          <div class="h-4 w-full bg-black/5 dark:bg-white/5 rounded-md" />
        </div>
      </div>

      <div v-else class="space-y-7 overflow-y-auto pb-10">
        <!-- 路径 -->
        <div class="ios-glass rounded-[24px] overflow-hidden group">
          <div
            class="px-6 py-4 flex items-center border-b border-black/[0.04] dark:border-white/[0.04] bg-gradient-to-r from-violet-500/[0.08] to-transparent dark:from-violet-400/[0.12]"
          >
            <div
              class="w-8 h-8 rounded-xl bg-violet-500/20 flex items-center justify-center mr-3 text-violet-600 dark:text-violet-300"
            >
              <FolderOpen class="w-[18px] h-[18px]" stroke-width="2.5" />
            </div>
            <h2 class="text-[17px] font-semibold tracking-tight">Windsurf 安装路径</h2>
          </div>
          <div class="p-6 space-y-4">
            <p class="text-[13px] text-ios-textSecondary dark:text-ios-textSecondaryDark leading-relaxed">
              用于检测与写入无感切号补丁。可自动探测或粘贴本机路径。
            </p>
            <div class="flex gap-2">
              <input
                v-model="local.windsurf_path"
                type="text"
                class="no-drag-region flex-1 bg-black/[0.03] dark:bg-white/[0.06] border border-black/5 dark:border-white/5 px-4 py-3 rounded-[14px] font-mono text-[13px] focus:ring-[3px] focus:ring-ios-blue/30 outline-none"
                placeholder="%AppData%\...\Windsurf"
              />
              <button
                type="button"
                class="no-drag-region shrink-0 px-4 py-3 rounded-[14px] bg-ios-blue/10 text-ios-blue font-semibold text-[13px] ios-btn hover:bg-ios-blue/20"
                :disabled="systemStore.isGlobalLoading"
                @click="handleDetectPath"
              >
                自动检测
              </button>
            </div>
          </div>
        </div>

        <!-- 网络 -->
        <div class="ios-glass rounded-[24px] overflow-hidden group">
          <div
            class="px-6 py-4 flex items-center border-b border-black/[0.04] dark:border-white/[0.04] bg-gradient-to-r from-ios-blue/[0.08] to-transparent dark:from-ios-blue/[0.15]"
          >
            <div
              class="w-8 h-8 rounded-xl bg-ios-blue/20 flex items-center justify-center mr-3 text-ios-blue shadow-inner shadow-white/50 dark:shadow-black/50"
            >
              <Network class="w-[18px] h-[18px]" stroke-width="2.5" />
            </div>
            <h2 class="text-[17px] font-semibold tracking-tight">网络代理</h2>
          </div>
          <div class="p-6 space-y-6">
            <div class="flex items-center justify-between gap-4">
              <div class="pr-2">
                <div class="text-[16px] font-semibold mb-1">启用 HTTP 代理</div>
                <div class="text-[13px] text-ios-textSecondary dark:text-ios-textSecondaryDark leading-relaxed">
                  登录、刷新 Token、拉取额度等请求走该代理（对应 proxy_enabled / proxy_url）。
                </div>
              </div>
              <IToggle v-model="local.proxy_enabled" />
            </div>

            <Transition name="fade">
              <div v-if="local.proxy_enabled" class="pt-1">
                <div class="relative flex items-center">
                  <div class="absolute left-4 opacity-50 pointer-events-none">
                    <Network class="w-4 h-4" />
                  </div>
                  <input
                    v-model="local.proxy_url"
                    type="text"
                    class="no-drag-region w-full bg-black/[0.03] dark:bg-white/[0.06] border border-black/5 dark:border-white/5 pl-11 pr-4 py-[14px] rounded-[14px] font-mono text-[14px] focus:ring-[3px] focus:ring-ios-blue/30 outline-none transition-all shadow-inner"
                    placeholder="http://127.0.0.1:7890"
                  />
                </div>
              </div>
            </Transition>
          </div>
        </div>

        <!-- 自动化 -->
        <div class="ios-glass rounded-[24px] overflow-hidden group">
          <div
            class="px-6 py-4 flex items-center border-b border-black/[0.04] dark:border-white/[0.04] bg-gradient-to-r from-ios-green/[0.08] to-transparent dark:from-ios-green/[0.15]"
          >
            <div
              class="w-8 h-8 rounded-xl bg-ios-green/20 flex items-center justify-center mr-3 text-ios-greenDark shadow-inner shadow-white/50 dark:shadow-black/50"
            >
              <RefreshCcw class="w-[18px] h-[18px]" stroke-width="2.5" />
            </div>
            <h2 class="text-[17px] font-semibold tracking-tight">保活与额度同步</h2>
          </div>
          <div class="p-6 space-y-6">
            <div class="flex items-center justify-between gap-4">
              <div class="pr-2">
                <div class="text-[16px] font-semibold mb-1">自动刷新 Token</div>
                <div class="text-[13px] text-ios-textSecondary dark:text-ios-textSecondaryDark leading-relaxed">
                  定时刷新账号池 JWT（auto_refresh_tokens）。
                </div>
              </div>
              <IToggle v-model="local.auto_refresh_tokens" />
            </div>

            <div class="h-px w-full bg-black/[0.06] dark:bg-white/[0.08]" />

            <div class="flex items-center justify-between gap-4">
              <div class="pr-2">
                <div class="text-[16px] font-semibold mb-1">定期同步额度</div>
                <div class="text-[13px] text-ios-textSecondary dark:text-ios-textSecondaryDark leading-relaxed">
                  按策略拉取日/周额度展示（auto_refresh_quotas）。
                </div>
              </div>
              <IToggle v-model="local.auto_refresh_quotas" />
            </div>

            <div class="rounded-[16px] bg-black/[0.03] dark:bg-white/[0.05] p-4 space-y-3">
              <label class="block text-[12px] font-semibold uppercase tracking-wide text-ios-textSecondary dark:text-ios-textSecondaryDark"
                >额度同步策略（quota_refresh_policy）</label
              >
              <select
                v-model="local.quota_refresh_policy"
                :disabled="!local.auto_refresh_quotas"
                class="no-drag-region w-full bg-white/80 dark:bg-black/30 border border-black/10 dark:border-white/10 rounded-[12px] px-3 py-2.5 text-[14px] outline-none focus:ring-2 focus:ring-ios-blue/30"
              >
                <option v-for="opt in quotaPolicyOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>

              <div v-if="local.quota_refresh_policy === 'custom'" class="pt-1 space-y-1">
                <label class="text-[12px] text-ios-textSecondary dark:text-ios-textSecondaryDark"
                  >自定义间隔（分钟，5～10080）</label
                >
                <input
                  v-model.number="local.quota_custom_interval_minutes"
                  type="number"
                  min="5"
                  max="10080"
                  :disabled="!local.auto_refresh_quotas"
                  class="no-drag-region w-full bg-white/80 dark:bg-black/30 border border-black/10 dark:border-white/10 rounded-[12px] px-3 py-2.5 text-[14px] outline-none"
                />
              </div>
            </div>

            <div class="flex items-center justify-between gap-4 opacity-80">
              <div class="pr-2">
                <div class="text-[15px] font-semibold mb-1">并发上限</div>
                <div class="text-[12px] text-ios-textSecondary dark:text-ios-textSecondaryDark">
                  concurrent_limit（预留，当前后端默认 5）
                </div>
              </div>
              <input
                v-model.number="local.concurrent_limit"
                type="number"
                min="1"
                max="50"
                class="no-drag-region w-24 text-center bg-black/[0.03] dark:bg-white/[0.06] border border-black/5 rounded-[12px] py-2 text-[14px] font-mono"
              />
            </div>

            <div class="h-px w-full bg-black/[0.06] dark:bg-white/[0.08]" />

            <div class="rounded-[16px] bg-black/[0.03] dark:bg-white/[0.05] p-4 space-y-2">
              <label
                class="block text-[12px] font-semibold uppercase tracking-wide text-ios-textSecondary dark:text-ios-textSecondaryDark"
                >无感「下一席位」计划范围（auto_switch_plan_filter）</label
              >
              <p class="text-[12px] text-ios-textSecondary dark:text-ios-textSecondaryDark leading-relaxed">
                可勾选多个计划（例如同时 Trial + Pro），仅在所选池内轮换下一席位；与账号池分组一致。保存配置后生效。
              </p>
              <div
                class="no-drag-region rounded-[12px] border border-black/10 dark:border-white/10 bg-white/80 dark:bg-black/30 px-3 py-3"
              >
                <SwitchPlanFilterControl v-model="local.auto_switch_plan_filter" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
