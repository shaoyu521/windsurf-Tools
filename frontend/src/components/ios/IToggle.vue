<script setup lang="ts">
const props = withDefaults(defineProps<{ modelValue: boolean; disabled?: boolean }>(), {
  disabled: false,
})
const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void }>()

const handleClick = () => {
  if (props.disabled) {
    return
  }
  emit('update:modelValue', !props.modelValue)
}
</script>
<template>
  <button
    type="button"
    class="no-drag-region w-[50px] h-8 rounded-full p-0.5 flex items-center transition-colors duration-200 shrink-0"
    :class="[
      props.modelValue ? 'bg-ios-green dark:bg-ios-greenDark' : 'bg-gray-300 dark:bg-white/15',
      props.disabled ? 'cursor-not-allowed opacity-45' : 'cursor-pointer',
    ]"
    :disabled="props.disabled"
    @click="handleClick"
  >
    <div
      class="w-7 h-7 bg-white rounded-full shadow-[0_2px_4px_rgba(0,0,0,0.2)] dark:shadow-sm transition-transform duration-200"
      :class="props.modelValue ? 'translate-x-[18px]' : 'translate-x-[2px]'"
    />
  </button>
</template>
