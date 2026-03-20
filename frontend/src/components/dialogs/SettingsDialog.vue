<script setup lang="ts">
import { computed } from 'vue'
import { ElDialog } from 'element-plus/es/components/dialog/index'

const props = defineProps<{
  modelValue: boolean
  loading: boolean
  previewMode: boolean
  patchStatus: boolean | null
  windsurfPath: string
  proxyEnabled: boolean
  proxyUrl: string
  autoRefresh: boolean
  autoRefreshQuotas: boolean
  quotaRefreshPolicy: string
  quotaCustomIntervalMinutes: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:windsurfPath': [value: string]
  'update:proxyEnabled': [value: boolean]
  'update:proxyUrl': [value: string]
  'update:autoRefresh': [value: boolean]
  'update:autoRefreshQuotas': [value: boolean]
  'update:quotaRefreshPolicy': [value: string]
  'update:quotaCustomIntervalMinutes': [value: number]
  detectPath: []
  applyPatch: []
  restorePatch: []
  save: []
}>()

const patchLabel = computed(() => {
  if (props.patchStatus === null) {
    return '未检测'
  }

  return props.patchStatus ? '已启用' : '未启用'
})
</script>

<template>
  <ElDialog class="wt-dialog" :model-value="props.modelValue" width="680px" @update:model-value="emit('update:modelValue', $event)">
    <template #header>
      <div class="dialog-title">
        <p class="dialog-title__eyebrow">Control Room</p>
        <h3>运行设置</h3>
      </div>
    </template>

    <section class="dialog-shell">
      <div v-if="props.previewMode" class="inline-banner">
        当前是浏览器预览模式，保存设置只会更新界面状态，不会写入本机配置。
      </div>

      <div class="settings-grid">
        <article class="settings-card">
          <div class="settings-card__header">
            <div>
              <p class="section-label">无感切号补丁</p>
              <p class="settings-card__desc">检测或修补 Windsurf 本地扩展文件，让切号流程更平滑。</p>
            </div>
            <span class="badge badge--health" :class="props.patchStatus ? 'badge--healthy' : 'badge--unknown'">
              {{ patchLabel }}
            </span>
          </div>

          <div class="settings-card__row">
            <input
              class="surface-input"
              :value="props.windsurfPath"
              placeholder="自动检测或手动指定 Windsurf 路径"
              @input="emit('update:windsurfPath', ($event.target as HTMLInputElement).value)"
            />
            <button class="control-button control-button--ghost" type="button" @click="emit('detectPath')">检测</button>
          </div>

          <div class="settings-card__actions">
            <button
              v-if="!props.patchStatus"
              class="control-button control-button--primary"
              type="button"
              :disabled="!props.windsurfPath || props.loading"
              @click="emit('applyPatch')"
            >
              应用补丁
            </button>
            <button
              v-else
              class="control-button control-button--danger"
              type="button"
              :disabled="props.loading"
              @click="emit('restorePatch')"
            >
              还原补丁
            </button>
          </div>
        </article>

        <article class="settings-card">
          <div class="settings-card__header">
            <div>
              <p class="section-label">网络与保活</p>
              <p class="settings-card__desc">代理会影响登录与刷新链路，自动刷新用于维持账号活性。</p>
            </div>
          </div>

          <label class="toggle-row">
            <span>
              <strong>代理模式</strong>
              <small>国内网络下建议开启</small>
            </span>
            <input
              class="toggle-switch"
              type="checkbox"
              :checked="props.proxyEnabled"
              @change="emit('update:proxyEnabled', ($event.target as HTMLInputElement).checked)"
            />
          </label>

          <input
            class="surface-input"
            :value="props.proxyUrl"
            :disabled="!props.proxyEnabled"
            placeholder="http://127.0.0.1:7890"
            @input="emit('update:proxyUrl', ($event.target as HTMLInputElement).value)"
          />

          <label class="toggle-row">
            <span>
              <strong>自动刷新</strong>
              <small>每 10 分钟自动刷新全部 Token / JWT</small>
            </span>
            <input
              class="toggle-switch"
              type="checkbox"
              :checked="props.autoRefresh"
              @change="emit('update:autoRefresh', ($event.target as HTMLInputElement).checked)"
            />
          </label>

          <label class="toggle-row">
            <span>
              <strong>定期同步额度</strong>
              <small>约每 5 分钟检查一次是否到期；按下方策略拉取日/周额度</small>
            </span>
            <input
              class="toggle-switch"
              type="checkbox"
              :checked="props.autoRefreshQuotas"
              @change="emit('update:autoRefreshQuotas', ($event.target as HTMLInputElement).checked)"
            />
          </label>

          <div class="dialog-section">
            <p class="section-label">额度同步策略</p>
            <select
              class="surface-input"
              style="cursor: pointer"
              :disabled="!props.autoRefreshQuotas"
              :value="props.quotaRefreshPolicy || 'hybrid'"
              @change="emit('update:quotaRefreshPolicy', ($event.target as HTMLSelectElement).value)"
            >
              <option value="hybrid">美东换日 或 距离上次已满 24 小时（推荐）</option>
              <option value="interval_24h">仅固定每 24 小时</option>
              <option value="us_calendar">仅美东日历跨日（0 点后）</option>
              <option value="local_calendar">仅本机 Windows 时区日历跨日</option>
              <option value="interval_1h">固定每 1 小时</option>
              <option value="interval_6h">固定每 6 小时</option>
              <option value="interval_12h">固定每 12 小时</option>
              <option value="custom">自定义间隔（分钟）</option>
            </select>
          </div>

          <div v-if="(props.quotaRefreshPolicy || 'hybrid') === 'custom'" class="dialog-section">
            <p class="section-label">自定义间隔（分钟）</p>
            <input
              class="surface-input"
              type="number"
              min="5"
              max="10080"
              :disabled="!props.autoRefreshQuotas"
              :value="props.quotaCustomIntervalMinutes"
              @input="
                emit(
                  'update:quotaCustomIntervalMinutes',
                  Math.min(10080, Math.max(5, Number(($event.target as HTMLInputElement).value) || 360)),
                )
              "
            />
            <p class="settings-card__desc" style="margin-top: 0.35rem">范围 5～10080（7 天），保存时与后端一致钳制。</p>
          </div>
        </article>
      </div>
    </section>

    <template #footer>
      <div class="dialog-footer">
        <button class="control-button control-button--ghost" type="button" @click="emit('update:modelValue', false)">
          取消
        </button>
        <button class="control-button control-button--primary" type="button" :disabled="props.loading" @click="emit('save')">
          {{ props.loading ? '保存中...' : '保存设置' }}
        </button>
      </div>
    </template>
  </ElDialog>
</template>
