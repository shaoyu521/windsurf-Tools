<template>
  <header
    class="drag-region h-[52px] w-full flex items-center justify-between px-5 bg-white/80 dark:bg-[#1C1C1E]/85 backdrop-blur-2xl border-b border-ios-divider dark:border-ios-dividerDark select-none z-50 shrink-0"
  >
    <div class="w-16" aria-hidden="true" />
    <div class="flex flex-col items-center justify-center flex-1 min-w-0">
      <span class="text-[15px] font-semibold tracking-tight text-ios-text dark:text-ios-textDark truncate">
        Windsurf Tools
      </span>
      <span class="text-[10px] font-medium text-ios-textSecondary dark:text-ios-textSecondaryDark tracking-wide uppercase">
        Control Deck
      </span>
    </div>
    <div class="w-16 no-drag-region flex justify-end">
      <!-- 预留给将来：关于 / 检查更新等 -->
    </div>
  </header>
</template>
