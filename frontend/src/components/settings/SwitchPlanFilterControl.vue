<script setup lang="ts">
import { computed } from 'vue'
import {
  formatSwitchPlanFilterSummary,
  normalizeSwitchPlanFilter,
  switchPlanFilterToneOptions,
  type SwitchPlanTone,
} from '../../utils/settingsModel'

const props = withDefaults(
  defineProps<{
    modelValue: string
    /** compact：顶栏等窄区域 */
    variant?: 'default' | 'compact'
  }>(),
  { variant: 'default' },
)

const emit = defineEmits<{
  'update:modelValue': [string]
}>()

const normalized = computed(() => normalizeSwitchPlanFilter(props.modelValue))

const isUnrestricted = computed(() => normalized.value === 'all')

const selectedTones = computed(() => {
  if (normalized.value === 'all') {
    return new Set<string>()
  }
  return new Set(normalized.value.split(',').filter(Boolean))
})

const summary = computed(() => formatSwitchPlanFilterSummary(props.modelValue))

function emitValue(v: string) {
  emit('update:modelValue', normalizeSwitchPlanFilter(v))
}

function onAllChange(checked: boolean) {
  if (checked) {
    emitValue('all')
    return
  }
  // 取消「全部」时给常用默认组合，避免空选
  emitValue('pro,trial')
}

function toggleTone(tone: SwitchPlanTone) {
  if (isUnrestricted.value) {
    emitValue(tone)
    return
  }
  const next = new Set(selectedTones.value)
  if (next.has(tone)) {
    next.delete(tone)
    if (next.size === 0) {
      emitValue('all')
      return
    }
  } else {
    next.add(tone)
  }
  emitValue([...next].join(','))
}
</script>

<template>
  <div :class="variant === 'compact' ? 'space-y-2' : 'space-y-3'">
    <label
      class="flex items-start gap-2 cursor-pointer select-none"
      :class="variant === 'compact' ? 'text-[11px]' : 'text-[14px]'"
    >
      <input
        type="checkbox"
        class="no-drag-region mt-0.5 rounded border-black/20 dark:border-white/30"
        :checked="isUnrestricted"
        @change="onAllChange(($event.target as HTMLInputElement).checked)"
      />
      <span class="font-semibold">全部计划（不限制）</span>
    </label>

    <div
      v-if="!isUnrestricted"
      class="flex flex-wrap gap-x-3 gap-y-1.5"
      :class="variant === 'compact' ? 'text-[11px] pl-0.5' : 'text-[13px]'"
    >
      <label
        v-for="opt in switchPlanFilterToneOptions"
        :key="opt.value"
        class="inline-flex items-center gap-1 cursor-pointer select-none whitespace-nowrap"
      >
        <input
          type="checkbox"
          class="no-drag-region rounded border-black/20 dark:border-white/30 shrink-0"
          :checked="selectedTones.has(opt.value)"
          @change="toggleTone(opt.value)"
        />
        <span>{{ opt.label }}</span>
      </label>
    </div>

    <p
      v-if="variant === 'default'"
      class="text-[11px] text-ios-textSecondary dark:text-ios-textSecondaryDark leading-relaxed"
    >
      当前范围：<span class="font-medium text-ios-text dark:text-ios-textDark">{{ summary }}</span>
    </p>
  </div>
</template>
