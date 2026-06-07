<script setup lang="ts">
import { useQueueStore } from '@/stores/queue'
import { computed } from 'vue'

const queueStore = useQueueStore()

const finishedCount = computed(() => queueStore.doneCount + queueStore.failedCount)

const overallProgress = computed(() => {
  if (queueStore.totalCount === 0) return 0
  return Math.round((finishedCount.value / queueStore.totalCount) * 100)
})
</script>

<template>
  <div class="progress-panel">
    <div class="progress-row">
      <div class="track">
        <div class="fill" :style="{ width: overallProgress + '%' }" />
      </div>
      <span class="pct">{{ queueStore.totalCount > 0 ? `${finishedCount}/${queueStore.totalCount}` : '—' }}</span>
    </div>
  </div>
</template>

<style scoped>
.progress-panel {
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  padding: 10px 14px;
  flex-shrink: 0;
}

.progress-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.track {
  flex: 1;
  height: 6px;
  background: var(--bg-tertiary);
  border-radius: 3px;
  overflow: hidden;
}

.fill {
  height: 100%;
  background: linear-gradient(90deg, var(--accent), var(--accent-hover));
  border-radius: 3px;
  transition: width 0.3s ease-out;
}

.pct {
  font-size: 12px;
  color: var(--text-secondary);
  min-width: 40px;
  text-align: right;
  font-family: var(--font-mono);
}
</style>
