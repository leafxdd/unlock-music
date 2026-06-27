<script setup lang="ts">
import { useQueueStore } from '@/stores/queue'

const queueStore = useQueueStore()

function statusLabel(s: string) {
  const map: Record<string, string> = {
    queued: '排队中',
    validating: '验证中',
    decrypting: '解密中',
    metadata: '元数据',
    writing: '写入中',
    done: '完成',
    skipped: '跳过',
    failed: '失败',
  }
  return map[s] || s
}

function statusClass(s: string) {
  if (s === 'done') return 'done'
  if (s === 'failed') return 'failed'
  if (s === 'skipped') return 'skipped'
  return 'active'
}

function basename(p: string) {
  return p.split(/[\\/]/).pop() || p
}
</script>

<template>
  <div class="queue-panel">
    <div class="panel-header">
      <span class="label">队列</span>
      <span class="count">{{ queueStore.totalCount }}</span>
      <div class="spacer" />
      <span v-if="queueStore.doneCount" class="stat stat-done">{{ queueStore.doneCount }} 完成</span>
      <span v-if="queueStore.failedCount" class="stat stat-failed">{{ queueStore.failedCount }} 失败</span>
      <button class="btn-clear" @click="queueStore.clear()">清空</button>
    </div>
    <div class="list-scroll">
      <div v-if="queueStore.list.length === 0" class="empty">
        <span>等待文件输入…</span>
      </div>
      <div
        v-for="item in queueStore.list"
        :key="item.path"
        class="item"
        :class="statusClass(item.status)"
      >
        <div class="item-top">
          <span class="filename" :title="item.path">{{ basename(item.path) }}</span>
          <span class="badge" :class="'badge-' + statusClass(item.status)">{{ statusLabel(item.status) }}</span>
        </div>
        <div v-if="item.outputPath && item.status === 'done'" class="item-output">
          → {{ basename(item.outputPath) }}
        </div>
        <div v-if="item.status !== 'done' && item.status !== 'failed' && item.status !== 'skipped' && item.status !== 'queued'" class="item-bar">
          <div class="bar-track">
            <div class="bar-fill" :style="{ width: item.progress + '%' }" />
          </div>
          <span class="pct">{{ item.progress }}%</span>
        </div>
        <div v-if="item.error" class="item-error">{{ item.error }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.queue-panel {
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 12px;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

.count {
  font-size: 11px;
  background: var(--bg-tertiary);
  color: var(--text-muted);
  padding: 1px 7px;
  border-radius: 10px;
}

.spacer { flex: 1; }

.stat { font-size: 11px; }
.stat-done { color: var(--success); }
.stat-failed { color: var(--error); }

.btn-clear {
  padding: 3px 10px;
  border-radius: var(--radius);
  font-size: 11px;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}
.btn-clear:hover { background: var(--bg-hover); }

.list-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 6px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  font-size: 13px;
}

.item {
  padding: 8px 10px;
  border-radius: 6px;
  background: var(--bg-tertiary);
  border-left: 3px solid var(--border);
  transition: background 0.15s;
}
.item:hover { background: var(--bg-hover); }
.item.done { border-left-color: var(--success); }
.item.failed { border-left-color: var(--error); }
.item.skipped { border-left-color: var(--warning); }
.item.active { border-left-color: var(--accent); }

.item-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.filename {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
  flex: 1;
}

.badge {
  padding: 1px 7px;
  border-radius: 10px;
  font-size: 10px;
  font-weight: 500;
  flex-shrink: 0;
}
.badge-done { background: color-mix(in srgb, var(--success) 15%, transparent); color: var(--success); }
.badge-failed { background: color-mix(in srgb, var(--error) 15%, transparent); color: var(--error); }
.badge-skipped { background: color-mix(in srgb, var(--warning) 15%, transparent); color: var(--warning); }
.badge-active { background: color-mix(in srgb, var(--accent) 15%, transparent); color: var(--accent); }

.item-output {
  font-size: 10px;
  color: var(--text-muted);
  margin-top: 3px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
}

.bar-track {
  flex: 1;
  height: 3px;
  background: var(--bg-hover);
  border-radius: 2px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: 2px;
  background: var(--accent);
  transition: width 0.3s ease-out;
}

.pct {
  font-size: 10px;
  color: var(--text-muted);
  min-width: 28px;
  text-align: right;
}

.item-error {
  margin-top: 3px;
  font-size: 10px;
  color: var(--error);
  word-break: break-all;
}
</style>
