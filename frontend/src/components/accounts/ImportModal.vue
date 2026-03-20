<script setup lang="ts">
import { ref } from 'vue'
import ISegmented from '../ios/ISegmented.vue'
import { APIInfo } from '../../api/wails'
import { useAccountStore } from '../../stores/useAccountStore'
import { X, Loader2, CheckCircle2, AlertCircle } from 'lucide-vue-next'
import { toAPIKeyItems, toEmailPasswordItems, toJWTItems, toTokenItems } from '../../utils/importParse'
import { importBatched } from '../../utils/importBatch'
import { main } from '../../../wailsjs/go/models'

const props = defineProps<{ isOpen: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()
const accountStore = useAccountStore()

const modes = [
  { label: '邮箱/密码', value: 'password' },
  { label: 'Refresh Token', value: 'refresh_token' },
  { label: 'API Key', value: 'api_key' },
  { label: 'JWT', value: 'jwt' },
]

const currentMode = ref('password')
const inputText = ref('')
const isLoading = ref(false)
const results = ref<main.ImportResult[]>([])

const handleImport = async () => {
  const lines = inputText.value.split('\n').map((l) => l.trim()).filter(Boolean)
  if (!lines.length) {
    return
  }
  isLoading.value = true
  results.value = []
  try {
    let batch: main.ImportResult[] = []
    switch (currentMode.value) {
      case 'api_key':
        batch = await importBatched(
          toAPIKeyItems(lines),
          (slice) => APIInfo.importByAPIKey(slice),
          (acc) => {
            results.value = acc
          },
        )
        break
      case 'jwt':
        batch = await importBatched(
          toJWTItems(lines),
          (slice) => APIInfo.importByJWT(slice),
          (acc) => {
            results.value = acc
          },
        )
        break
      case 'refresh_token':
        batch = await importBatched(
          toTokenItems(lines),
          (slice) => APIInfo.importByRefreshToken(slice),
          (acc) => {
            results.value = acc
          },
        )
        break
      case 'password': {
        const items = toEmailPasswordItems(lines)
        if (!items.length) {
          alert(
            '未解析到有效行。支持：JSON；账号:/邮箱:/卡号: + 密码:；---- 分隔；tab/逗号分隔；「邮箱 密码」；含「密码不对就这个」会尝试第二个密码；续行「密码:」可自动合并。',
          )
          isLoading.value = false
          return
        }
        batch = await importBatched(
          items,
          (slice) => APIInfo.importByEmailPassword(slice),
          (acc) => {
            results.value = acc
          },
        )
        break
      }
      default:
        break
    }
    results.value = batch || []
    await accountStore.fetchAccounts()
    inputText.value = ''
  } catch (e) {
    console.error(e)
    alert(`导入失败: ${String(e)}`)
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-[100] flex animate-in fade-in duration-300 items-end sm:items-center justify-center bg-black/40 dark:bg-black/60 backdrop-blur-md"
  >
    <div
      class="bg-ios-bg dark:bg-ios-bgDark w-full sm:w-[540px] h-[90vh] sm:h-auto sm:max-h-[85vh] rounded-t-3xl sm:rounded-[28px] shadow-[0_20px_60px_-10px_rgba(0,0,0,0.3)] dark:shadow-[0_20px_60px_-10px_rgba(0,0,0,0.8)] ring-1 ring-white/50 dark:ring-white/10 flex flex-col transform transition-transform animate-in slide-in-from-bottom-12 duration-[400ms] ease-[cubic-bezier(0.16,1,0.3,1)] overflow-hidden"
    >
      <div
        class="px-5 py-4 border-b border-black/[0.06] dark:border-white/[0.06] bg-white/70 dark:bg-[#1C1C1E]/70 backdrop-blur-xl flex justify-between items-center shrink-0"
      >
        <h3 class="font-bold text-[17px] tracking-tight">批量导入</h3>
        <button
          type="button"
          class="no-drag-region p-1.5 rounded-full bg-black/5 dark:bg-white/10 hover:bg-black/10 dark:hover:bg-white/20 transition-all ios-btn"
          @click="emit('close')"
        >
          <X class="w-5 h-5 text-ios-textSecondary dark:text-ios-textSecondaryDark" stroke-width="2.5" />
        </button>
      </div>

      <div class="p-5 flex-1 overflow-y-auto">
        <ISegmented v-model="currentMode" :options="modes" class="mb-5 h-8 flex-shrink-0" />

        <div class="mb-4 text-xs text-ios-textSecondary dark:text-ios-textSecondaryDark leading-relaxed space-y-1">
          <p>每行一条；凭证与备注用空格分隔（首列为凭证）。</p>
          <p v-if="currentMode === 'password'">
            支持多种粘贴格式（账号:/邮箱:、----、tab、引号逗号等）；同一邮箱多行只保留最后一条；主密码失败会自动试
            <code class="px-1 rounded bg-black/5 dark:bg-white/10">alt_password</code>
            或行内「密码不对就这个」后的第二个密码。大批量易卡顿可先改用下方 JWT 模式。
          </p>
          <p v-if="currentMode === 'jwt'">
            推荐与
            <code class="px-1 rounded bg-black/5 dark:bg-white/10">_quick_key.py</code>
            同链路：
            <code class="px-1 rounded bg-black/5 dark:bg-white/10">python tools/batch_quick_jwt.py</code>
            批量得到 Windsurf
            <code class="px-1 rounded bg-black/5 dark:bg-white/10">GetUserJwt</code>
            票据。若仅需 Firebase
            <code class="px-1 rounded bg-black/5 dark:bg-white/10">idToken</code>
            ，可用
            <code class="px-1 rounded bg-black/5 dark:bg-white/10">tools/email-password-to-firebase-jwt.mjs</code>
            。
          </p>
        </div>

        <textarea
          v-model="inputText"
          class="no-drag-region w-full h-[180px] bg-white/80 dark:bg-black/40 border border-black/10 dark:border-white/10 p-4 rounded-[18px] focus:outline-none focus:ring-2 focus:ring-ios-blue/50 dark:focus:ring-ios-blue/30 resize-none font-mono text-[13px] shadow-sm transition-all"
          placeholder="粘贴多行内容..."
        />

        <div v-if="results.length" class="mt-5 space-y-2 max-h-40 overflow-y-auto">
          <h4 class="text-xs font-semibold uppercase tracking-wider text-ios-textSecondary dark:text-ios-textSecondaryDark mb-2">
            导入结果
          </h4>
          <div
            v-for="(r, i) in results"
            :key="i"
            class="text-xs p-2.5 rounded-xl flex items-center justify-between shadow-sm border"
            :class="
              r.success
                ? 'bg-ios-green/10 border-ios-green/20 text-ios-greenDark'
                : 'bg-ios-red/10 border-ios-red/20 text-ios-redDark'
            "
          >
            <span class="font-semibold truncate max-w-[280px] mr-2" :title="r.email">{{ r.email }}</span>
            <div class="flex items-center shrink-0 font-medium">
              <CheckCircle2 v-if="r.success" class="w-4 h-4 mr-1" />
              <AlertCircle v-else class="w-4 h-4 mr-1" />
              {{ r.success ? '成功' : r.error || '失败' }}
            </div>
          </div>
        </div>
      </div>

      <div
        class="p-5 border-t border-black/[0.06] dark:border-white/[0.06] bg-white/70 dark:bg-[#1C1C1E]/70 backdrop-blur-xl shrink-0"
      >
        <button
          type="button"
          class="no-drag-region w-full h-[48px] bg-gradient-to-b from-[#3b82f6] to-ios-blue text-white rounded-[14px] font-semibold text-[17px] ios-btn flex items-center justify-center disabled:opacity-50 shadow-md shadow-ios-blue/20 ring-1 ring-black/5 ring-inset active:ring-black/10"
          :disabled="isLoading || !inputText.trim()"
          @click="handleImport"
        >
          <Loader2 v-if="isLoading" class="w-5 h-5 animate-spin mr-2" />
          {{ isLoading ? '导入中…' : '开始导入' }}
        </button>
      </div>
    </div>
  </div>
</template>
