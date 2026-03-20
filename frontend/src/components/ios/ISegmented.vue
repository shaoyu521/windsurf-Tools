<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  options: { label: string, value: string }[]
  modelValue: string
}>()
const emit = defineEmits<{ (e: 'update:modelValue', val: string): void }>()

const activeIndex = computed(() => props.options.findIndex(o => o.value === props.modelValue))
</script>

<template>
  <div class="relative flex items-center bg-black/5 dark:bg-white/10 bg-gray-200/80 p-0.5 rounded-[9px] w-full">
    <!-- Sliding Selection Highlight -->
    <div 
      class="absolute bg-white dark:bg-[#636366] shadow-sm rounded-[7px] h-[calc(100%-4px)] top-0.5 transition-transform duration-300 ease-[cubic-bezier(0.25,1,0.5,1)]"
      :style="{
        width: `${100 / props.options.length}%`,
        transform: `translateX(${activeIndex * 100}%)`
      }"
    ></div>
    
    <button
      v-for="opt in props.options"
      :key="opt.value"
      type="button"
      class="no-drag-region relative flex-1 py-1 text-[13px] font-semibold text-center z-10 transition-colors"
      :class="props.modelValue === opt.value ? 'text-ios-text dark:text-white' : 'text-ios-textSecondary dark:text-ios-textSecondaryDark'"
      @click="emit('update:modelValue', opt.value)"
    >
      {{ opt.label }}
    </button>
  </div>
</template>
