<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import Header from './components/layout/Header.vue'
import Sidebar from './components/layout/Sidebar.vue'
import Dashboard from './views/Dashboard.vue'
import Accounts from './views/Accounts.vue'
import Settings from './views/Settings.vue'
import { useAccountStore } from './stores/useAccountStore'
import { useSettingsStore } from './stores/useSettingsStore'
import { useSystemStore } from './stores/useSystemStore'

const activeTab = ref('Dashboard')

onMounted(() => {
  const accounts = useAccountStore()
  const settings = useSettingsStore()
  const system = useSystemStore()
  void Promise.all([accounts.fetchAccounts(), settings.fetchSettings(), system.initSystemEnvironment()])
})

const currentView = computed(() => {
  switch (activeTab.value) {
    case 'Accounts': return Accounts
    case 'Settings': return Settings
    case 'Dashboard':
    default: return Dashboard
  }
})
</script>

<template>
  <div
    class="flex flex-col h-full text-ios-text dark:text-ios-textDark overflow-hidden antialiased app-root"
  >
    <Header />
    <div class="flex flex-1 overflow-hidden relative">
      <Sidebar :activeTab="activeTab" @update:activeTab="activeTab = $event" />
      <main class="flex-1 overflow-y-auto overflow-x-hidden relative scroll-smooth">
        <Transition name="fade" mode="out-in">
          <component :is="currentView" />
        </Transition>
      </main>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s ease-out, transform 0.2s ease-out;
}
.fade-enter-from {
  opacity: 0;
  transform: translateY(6px);
}
.fade-leave-to {
  opacity: 0;
  transform: translateY(-2px);
}
</style>
