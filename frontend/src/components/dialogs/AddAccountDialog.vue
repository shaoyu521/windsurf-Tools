<script setup lang="ts">
import { computed } from 'vue'
import { ElDialog } from 'element-plus/es/components/dialog/index'
import type { AddMode } from '../../types/windsurf'

const props = defineProps<{
  modelValue: boolean
  loading: boolean
  mode: AddMode
  value: string
  remark: string
  email: string
  password: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:mode': [value: AddMode]
  'update:value': [value: string]
  'update:remark': [value: string]
  'update:email': [value: string]
  'update:password': [value: string]
  submit: []
}>()

const modeOptions: Array<{ value: AddMode; label: string; hint: string }> = [
  { value: 'password', label: '邮箱密码', hint: 'Firebase 登录，密码仅存本机' },
  { value: 'api_key', label: 'API Key', hint: 'sk-ws-* 最稳定' },
  { value: 'jwt', label: 'JWT', hint: '现成登录票据' },
  { value: 'refresh_token', label: 'Refresh Token', hint: '长期刷新' },
]

const placeholder = computed(() => {
  switch (props.mode) {
    case 'api_key':
      return 'sk-ws-01-...'
    case 'jwt':
      return 'eyJhbGciOi...'
    default:
      return 'AMf-vBx...'
  }
})

const isPasswordMode = computed(() => props.mode === 'password')
</script>

<template>
  <ElDialog class="wt-dialog" :model-value="props.modelValue" width="560px" @update:model-value="emit('update:modelValue', $event)">
    <template #header>
      <div class="dialog-title">
        <p class="dialog-title__eyebrow">Single Insert</p>
        <h3>添加单个账号</h3>
      </div>
    </template>

    <section class="dialog-shell">
      <div class="dialog-section">
        <p class="section-label">导入方式</p>
        <div class="segment-grid segment-grid--four">
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

      <template v-if="isPasswordMode">
        <div class="dialog-section">
          <p class="section-label">邮箱</p>
          <input
            class="surface-input"
            type="email"
            autocomplete="username"
            :value="props.email"
            placeholder="name@example.com"
            @input="emit('update:email', ($event.target as HTMLInputElement).value)"
          />
        </div>
        <div class="dialog-section">
          <p class="section-label">密码</p>
          <input
            class="surface-input"
            type="password"
            autocomplete="current-password"
            :value="props.password"
            placeholder="登录密码"
            @input="emit('update:password', ($event.target as HTMLInputElement).value)"
          />
        </div>
      </template>

      <div v-else class="dialog-section">
        <p class="section-label">凭证内容</p>
        <textarea
          class="surface-textarea surface-textarea--compact"
          :value="props.value"
          :placeholder="placeholder"
          @input="emit('update:value', ($event.target as HTMLTextAreaElement).value)"
        />
      </div>

      <div class="dialog-section">
        <p class="section-label">备注</p>
        <input
          class="surface-input"
          :value="props.remark"
          placeholder="例如：主力创作池 / 备用号 / 团队共享"
          @input="emit('update:remark', ($event.target as HTMLInputElement).value)"
        />
      </div>
    </section>

    <template #footer>
      <div class="dialog-footer">
        <button class="control-button control-button--ghost" type="button" @click="emit('update:modelValue', false)">
          取消
        </button>
        <button class="control-button control-button--primary" type="button" :disabled="props.loading" @click="emit('submit')">
          {{ props.loading ? '添加中...' : '确认添加' }}
        </button>
      </div>
    </template>
  </ElDialog>
</template>
