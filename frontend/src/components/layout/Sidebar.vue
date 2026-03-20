<script setup lang="ts">
import { LayoutDashboard, Users, Settings } from 'lucide-vue-next'

const props = defineProps<{ activeTab: string }>()
const emit = defineEmits<{ (e: 'update:activeTab', tab: string): void }>()

const menuItems = [
  { id: 'Dashboard', icon: LayoutDashboard, label: '总览' },
  { id: 'Accounts', icon: Users, label: '账号池' },
  { id: 'Settings', icon: Settings, label: '设置' },
]
</script>

<template>
  <nav class="w-60 h-full ios-glass border-r flex flex-col pt-6 pb-6 z-40 shrink-0">
    <div class="px-5 pb-2 mb-2 text-xs font-semibold uppercase text-ios-textSecondary dark:text-ios-textSecondaryDark tracking-wider">
      导航
    </div>
    <ul class="flex-1 space-y-1.5 px-3">
      <li v-for="item in menuItems" :key="item.id">
        <button
          type="button"
          class="no-drag-region"
          @click="emit('update:activeTab', item.id)"
          :class="[
            'w-full flex items-center px-4 py-2.5 rounded-[14px] text-[14px] transition-all duration-[250ms] font-medium ios-btn',
            activeTab === item.id 
              ? 'bg-gradient-to-b from-[#3b82f6] to-ios-blue text-white shadow-md shadow-ios-blue/25 ring-1 ring-black/5 dark:ring-white/10 ring-inset' 
              : 'text-ios-text dark:text-ios-textDark hover:bg-black/5 dark:hover:bg-white/10'
          ]"
        >
          <component :is="item.icon" class="w-5 h-5 mr-3 transition-opacity duration-300" :class="activeTab === item.id ? 'opacity-100' : 'opacity-70'" stroke-width="2.2" />
          {{ item.label }}
        </button>
      </li>
    </ul>
  </nav>
</template>
