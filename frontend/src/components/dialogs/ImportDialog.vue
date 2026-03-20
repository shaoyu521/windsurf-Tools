<script setup lang="ts">
import { computed } from 'vue'
import { ElDialog } from 'element-plus/es/components/dialog/index'
import type { ImportMode } from '../../types/windsurf'

const props = defineProps<{
  modelValue: boolean
  loading: boolean
  mode: ImportMode
  text: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:mode': [value: ImportMode]
  'update:text': [value: string]
  submit: []
}>()

const modeOptions: Array<{ value: ImportMode; label: string; hint: string }> = [
  { value: 'api_key', label: 'API Key', hint: '适合已拿到 sk-ws-* 凭证' },
  { value: 'refresh_token', label: 'Refresh Token', hint: '适合保留长期刷新的旧号' },
  { value: 'jwt', label: 'JWT', hint: '适合快速补录现成票据' },
  { value: 'password', label: '邮箱密码', hint: '让后端重新走登录流程' },
]

const helperTitle = computed(() => {
  switch (props.mode) {
    case 'api_key':
      return '每行一个 API Key，可在空格后追加备注。'
    case 'refresh_token':
      return '每行一个 Refresh Token，可在空格后追加备注。'
    case 'jwt':
      return '每行一个 JWT，可在空格后追加备注。'
    default:
      return '格式为：邮箱 密码 备注（备注可选）。'
  }
})

const placeholder = computed(() => {
  switch (props.mode) {
    case 'api_key':
      return 'sk-ws-01-demo-xxxx 创作主力\nsk-ws-01-demo-yyyy 备用池'
    case 'refresh_token':
      return 'AMf-vBx-preview-token 协作号\nAMf-vBy-preview-token 兜底号'
    case 'jwt':
      return 'eyJhbGciOi... 训练营\neyJhbGciOi... 夜间池'
    default:
      return 'name@mail.com password 备注\nsecond@mail.com password2 备用'
  }
})

const lineCount = computed(
  () =>
    props.text
      .trim()
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean).length,
)
</script>

<template>
  <ElDialog
    class="wt-dialog"
    :model-value="props.modelValue"
    width="720px"
    :close-on-click-modal="false"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <template #header>
      <div class="dialog-title">
        <p class="dialog-title__eyebrow">Bulk Intake</p>
        <h3>批量导入账号</h3>
      </div>
    </template>

    <section class="dialog-shell">
      <div class="dialog-section">
        <p class="section-label">导入类型</p>
        <div class="segment-grid">
          <button
            v-for="option in modeOptions"
            :key="option.value"
            type="button"
            class="segment-card"
            :class="{ 'is-active': props.mode === option.value }"
            @click="emit('update:mode', option.value)"
          >
            <strong>{{ option.label }}</strong>
            <span>{{ option.hint }}</span>
          </button>
        </div>
      </div>

      <div class="dialog-section dialog-section--fill">
        <div class="section-row">
          <p class="section-label">内容</p>
          <span class="section-meta">共 {{ lineCount }} 条待导入</span>
        </div>
        <p class="dialog-helper">{{ helperTitle }}</p>
        <textarea
          class="surface-textarea"
          :value="props.text"
          :placeholder="placeholder"
          @input="emit('update:text', ($event.target as HTMLTextAreaElement).value)"
        />
      </div>
    </section>

    <template #footer>
      <div class="dialog-footer">
        <button class="control-button control-button--ghost" type="button" @click="emit('update:modelValue', false)">
          取消
        </button>
        <button class="control-button control-button--primary" type="button" :disabled="props.loading" @click="emit('submit')">
          {{ props.loading ? '导入中...' : `导入 ${lineCount || 0} 条` }}
        </button>
      </div>
    </template>
  </ElDialog>
</template>
