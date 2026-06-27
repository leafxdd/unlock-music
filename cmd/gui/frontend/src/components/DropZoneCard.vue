<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useQueueStore } from '@/stores/queue'
import { useSettingsStore } from '@/stores/settings'
import { backend } from '@/composables/useWails'

const queueStore = useQueueStore()
const settingsStore = useSettingsStore()
const dragging = ref(false)

async function selectFiles() {
  const files = await backend.selectInputFiles()
  if (files?.length) {
    const dir = files[0].replace(/[\\/][^\\/]+$/, '')
    settingsStore.settings.outputDir = dir
    await settingsStore.save()
    await queueStore.addPaths(files)
  }
}

async function selectDir() {
  const dir = await backend.selectInputDir()
  if (dir) {
    settingsStore.settings.inputDir = dir
    settingsStore.settings.outputDir = dir
    await settingsStore.save()
    await queueStore.addPaths([dir])
  }
}

async function pickOutput() {
  const dir = await backend.selectOutputDir()
  if (dir) {
    settingsStore.settings.outputDir = dir
    await settingsStore.save()
  }
}

function onOutputInput(e: Event) {
  settingsStore.settings.outputDir = (e.target as HTMLInputElement).value
}

async function onOutputBlur() {
  await settingsStore.save()
}

async function startProcessing() {
  if (queueStore.processing || queueStore.pendingPaths.length === 0) return
  queueStore.processing = true
  const paths = [...queueStore.pendingPaths]
  queueStore.pendingPaths = []
  await backend.startProcessingBatch(paths)
}

onMounted(() => {
  window.runtime.OnFileDrop(async (_x: number, _y: number, paths: string[]) => {
    if (!paths?.length) return
    const t = await backend.resolveDrop(paths)
    if (t.dir) {
      settingsStore.settings.outputDir = t.dir
      // Persist the input dir on a folder drop so it is restored next launch.
      if (t.isDir) settingsStore.settings.inputDir = t.dir
      await settingsStore.save()
    }
    await queueStore.addPaths(paths)
  }, true)
})

onUnmounted(() => {
  window.runtime.OnFileDropOff()
})
</script>

<template>
  <div
    class="dropzone"
    :class="{ dragging }"
    style="--wails-drop-target: drop"
    @dragover.prevent="dragging = true"
    @dragleave="dragging = false"
    @drop.prevent="dragging = false"
  >
    <div class="dropzone-inner">
      <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.4">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
        <polyline points="17 8 12 3 7 8"/>
        <line x1="12" y1="3" x2="12" y2="15"/>
      </svg>
      <p class="hint">拖放文件或文件夹到此处</p>
      <div class="actions">
        <button class="btn btn-primary" @click="selectFiles">选择文件</button>
        <button class="btn btn-secondary" @click="selectDir">选择目录</button>
      </div>
    </div>

    <div class="bottom-section">
      <div class="output-row">
        <label>输出目录</label>
        <div class="path-row">
          <input :value="settingsStore.settings.outputDir" placeholder="默认与输入目录相同" @input="onOutputInput" @blur="onOutputBlur" />
          <button class="btn btn-secondary btn-sm" @click="pickOutput">浏览</button>
        </div>
      </div>
      <button
        v-if="queueStore.pendingPaths.length > 0 && !queueStore.processing"
        class="btn btn-start"
        @click="startProcessing"
      >
        开始转换 ({{ queueStore.totalCount }})
      </button>
      <button
        v-else-if="queueStore.processing"
        class="btn btn-stop"
        @click="backend.stopProcessing().then(() => queueStore.processing = false)"
      >
        停止
      </button>
    </div>
  </div>
</template>

<style scoped>
.dropzone {
  border: 2px dashed var(--border);
  border-radius: var(--radius-lg);
  background: var(--bg-secondary);
  display: flex;
  flex-direction: column;
  transition: border-color 0.2s, background 0.2s;
  overflow: hidden;
}

.dropzone.dragging {
  border-color: var(--accent);
  background: color-mix(in srgb, var(--accent) 8%, transparent);
}

.dropzone-inner {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 24px;
}

.hint {
  font-size: 14px;
  color: var(--text-secondary);
}

.actions {
  display: flex;
  gap: 8px;
  margin-top: 4px;
}

.btn {
  padding: 7px 18px;
  border-radius: var(--radius);
  font-size: 13px;
  font-weight: 500;
  transition: all 0.15s;
}

.btn-sm {
  padding: 5px 12px;
  font-size: 12px;
}

.btn-primary {
  background: var(--accent);
  color: #fff;
}
.btn-primary:hover {
  background: var(--accent-hover);
}

.btn-secondary {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}
.btn-secondary:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.bottom-section {
  padding: 12px 16px;
  border-top: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.output-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.output-row label {
  font-size: 11px;
  color: var(--text-muted);
  white-space: nowrap;
  flex-shrink: 0;
}

.path-row {
  display: flex;
  gap: 6px;
  flex: 1;
  min-width: 0;
}

.path-row input {
  flex: 1;
  min-width: 0;
  padding: 5px 8px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  font-size: 11px;
  font-family: var(--font-mono);
  transition: border-color 0.15s;
}

.btn-start {
  padding: 9px 0;
  border-radius: var(--radius);
  font-size: 13px;
  font-weight: 600;
  background: var(--success);
  color: #fff;
  text-align: center;
}
.btn-start:hover { filter: brightness(1.1); }

.btn-stop {
  padding: 9px 0;
  border-radius: var(--radius);
  font-size: 13px;
  font-weight: 600;
  background: var(--error);
  color: #fff;
  text-align: center;
}
.btn-stop:hover { filter: brightness(1.1); }
</style>
