<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { useQueueStore } from '@/stores/queue'
import { useLogsStore } from '@/stores/logs'
import { useWailsEvent } from '@/composables/useWails'
import DropZoneCard from './DropZoneCard.vue'
import SettingsPanel from './SettingsPanel.vue'
import ProgressPanel from './ProgressPanel.vue'
import LogPanel from './LogPanel.vue'
import { useTheme } from '@/composables/useTheme'

const { pref: themePref, cycle: cycleTheme } = useTheme()
const themeLabel = computed(
  () => ({ system: '跟随系统', light: '浅色', dark: '深色' })[themePref.value],
)

const settingsStore = useSettingsStore()
const queueStore = useQueueStore()
const logsStore = useLogsStore()

const activeTab = ref<'queue' | 'settings' | 'logs'>('queue')

onMounted(async () => {
  await settingsStore.load()
  if (settingsStore.settings.inputDir) {
    await queueStore.addPaths([settingsStore.settings.inputDir])
  }
})

useWailsEvent('file:event', (e: any) => queueStore.handleFileEvent(e))
useWailsEvent('file:progress', (e: any) => queueStore.handleProgress(e))
useWailsEvent('log', (e: any) => logsStore.add(e))
useWailsEvent('processing:done', () => { queueStore.processing = false })
useWailsEvent('processing:error', (msg: string) => {
  logsStore.add({ level: 'ERROR', msg })
})
</script>

<template>
  <div class="shell">
    <header class="header" style="--wails-draggable: drag">
      <div class="title">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M9 18V5l12-2v13"/>
          <circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/>
        </svg>
        <span>Unlock Music</span>
      </div>
      <nav class="tabs">
        <button :class="{ active: activeTab === 'queue' }" @click="activeTab = 'queue'">文件队列</button>
        <button :class="{ active: activeTab === 'settings' }" @click="activeTab = 'settings'">设置</button>
        <button :class="{ active: activeTab === 'logs' }" @click="activeTab = 'logs'">日志</button>
        <button
          class="theme-toggle"
          :title="`主题：${themeLabel}（点击切换）`"
          :aria-label="`主题：${themeLabel}`"
          @click="cycleTheme"
        >
          <svg v-if="themePref === 'system'" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <rect x="2" y="3.5" width="20" height="14" rx="2"/>
            <line x1="8" y1="21" x2="16" y2="21"/>
            <line x1="12" y1="17.5" x2="12" y2="21"/>
          </svg>
          <svg v-else-if="themePref === 'light'" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="4.2"/>
            <line x1="12" y1="0.8" x2="12" y2="3.6"/>
            <line x1="12" y1="20.4" x2="12" y2="23.2"/>
            <line x1="3.8" y1="3.8" x2="5.8" y2="5.8"/>
            <line x1="18.2" y1="18.2" x2="20.2" y2="20.2"/>
            <line x1="0.8" y1="12" x2="3.6" y2="12"/>
            <line x1="20.4" y1="12" x2="23.2" y2="12"/>
            <line x1="3.8" y1="20.2" x2="5.8" y2="18.2"/>
            <line x1="18.2" y1="5.8" x2="20.2" y2="3.8"/>
          </svg>
          <svg v-else width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
          </svg>
          <span>{{ themeLabel }}</span>
        </button>
      </nav>
    </header>

    <main class="content">
      <template v-if="activeTab === 'queue'">
        <DropZoneCard class="queue-main" />
        <ProgressPanel />
      </template>
      <SettingsPanel v-else-if="activeTab === 'settings'" />
      <LogPanel v-else />
    </main>
  </div>
</template>

<style scoped>
.shell {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  height: 48px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: var(--accent);
}

.tabs {
  display: flex;
  gap: 4px;
}

.tabs button {
  padding: 6px 14px;
  border-radius: var(--radius);
  background: transparent;
  color: var(--text-secondary);
  font-size: 13px;
  transition: background 0.15s, color 0.15s;
  --wails-draggable: no-drag;
}

.tabs button:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.tabs button.active {
  background: var(--accent);
  color: #fff;
}

/* Scoped under .tabs so it outweighs the `.tabs button` rule above — otherwise
   that rule's horizontal padding squeezes the icon out of the button box. */
.tabs .theme-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 30px;
  padding: 0 10px;
  margin-left: 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  font-size: 12px;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
  --wails-draggable: no-drag;
}

.tabs .theme-toggle svg {
  flex-shrink: 0;
}

.tabs .theme-toggle:hover {
  background: var(--bg-hover);
  border-color: var(--accent);
  color: var(--accent);
}

.content {
  flex: 1;
  overflow: hidden;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.queue-main {
  flex: 1;
  min-height: 0;
}
</style>
