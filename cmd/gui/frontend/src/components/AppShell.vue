<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { useQueueStore } from '@/stores/queue'
import { useLogsStore } from '@/stores/logs'
import { useWailsEvent } from '@/composables/useWails'
import DropZoneCard from './DropZoneCard.vue'
import FileQueueTable from './FileQueueTable.vue'
import SettingsPanel from './SettingsPanel.vue'
import ProgressPanel from './ProgressPanel.vue'
import LogPanel from './LogPanel.vue'

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
      </nav>
    </header>

    <main class="content">
      <template v-if="activeTab === 'queue'">
        <div class="queue-split">
          <DropZoneCard class="split-left" />
          <FileQueueTable class="split-right" />
        </div>
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

.content {
  flex: 1;
  overflow: hidden;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.queue-split {
  flex: 1;
  display: flex;
  gap: 12px;
  min-height: 0;
}

.split-left {
  flex: 1;
  min-width: 0;
}

.split-right {
  flex: 1;
  min-width: 0;
}
</style>
