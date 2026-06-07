<script setup lang="ts">
import { useLogsStore } from '@/stores/logs'
import { ref, nextTick, watch } from 'vue'

const logsStore = useLogsStore()
const scrollEl = ref<HTMLElement>()

watch(() => logsStore.recent.length, async () => {
  await nextTick()
  if (scrollEl.value) {
    scrollEl.value.scrollTop = scrollEl.value.scrollHeight
  }
})

function levelClass(level: string) {
  if (level === 'ERROR') return 'log-error'
  if (level === 'WARN') return 'log-warn'
  return 'log-info'
}
</script>

<template>
  <div class="log-panel">
    <div class="log-header">
      <span>日志</span>
      <button @click="logsStore.clear()">清空</button>
    </div>
    <div ref="scrollEl" class="log-scroll">
      <div v-for="(entry, i) in logsStore.recent" :key="i" class="log-line" :class="levelClass(entry.level)">
        <span class="log-level">{{ entry.level }}</span>
        <span class="log-msg">{{ entry.msg }}</span>
      </div>
      <div v-if="!logsStore.recent.length" class="empty">暂无日志</div>
    </div>
  </div>
</template>

<style scoped>
.log-panel {
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  overflow: hidden;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border);
  font-size: 13px;
  font-weight: 500;
}

.log-header button {
  padding: 4px 10px;
  border-radius: var(--radius);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: 12px;
}
.log-header button:hover {
  background: var(--bg-hover);
}

.log-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 8px 14px;
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.6;
}

.log-line {
  display: flex;
  gap: 8px;
  padding: 1px 4px;
  border-radius: 4px;
  transition: background 0.1s;
}
.log-line:hover { background: var(--bg-hover); }

.log-level {
  flex-shrink: 0;
  width: 44px;
  font-weight: 600;
}

.log-info .log-level { color: var(--text-muted); }
.log-warn .log-level { color: var(--warning); }
.log-error .log-level { color: var(--error); }

.log-msg {
  color: var(--text-secondary);
  word-break: break-all;
}

.empty {
  padding: 40px;
  text-align: center;
  color: var(--text-muted);
}
</style>
