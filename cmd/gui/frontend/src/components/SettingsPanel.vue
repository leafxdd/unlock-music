<script setup lang="ts">
import { useSettingsStore } from '@/stores/settings'
import { backend } from '@/composables/useWails'

const settingsStore = useSettingsStore()

async function pickInput() {
  const dir = await backend.selectInputDir()
  if (dir) {
    settingsStore.settings.inputDir = dir
    await settingsStore.save()
  }
}

async function toggle(key: 'skipNoop' | 'removeSource' | 'updateMetadata' | 'overwriteOutput') {
  settingsStore.settings[key] = !settingsStore.settings[key]
  await settingsStore.save()
}
</script>

<template>
  <div class="settings">
    <h3>目录设置</h3>
    <div class="field">
      <label>输入目录</label>
      <div class="path-row">
        <input readonly :value="settingsStore.settings.inputDir" placeholder="未设置" />
        <button @click="pickInput">浏览</button>
      </div>
    </div>

    <h3>处理选项</h3>
    <div class="toggle-group">
      <label class="toggle-row" @click="toggle('skipNoop')">
        <span class="toggle-box" :class="{ on: settingsStore.settings.skipNoop }" />
        <span>跳过无需解密的文件</span>
      </label>
      <label class="toggle-row" @click="toggle('removeSource')">
        <span class="toggle-box" :class="{ on: settingsStore.settings.removeSource }" />
        <span>解密后删除源文件</span>
      </label>
      <label class="toggle-row" :class="{ disabled: !settingsStore.ffmpegAvailable }" @click="settingsStore.ffmpegAvailable && toggle('updateMetadata')">
        <span class="toggle-box" :class="{ on: settingsStore.settings.updateMetadata && settingsStore.ffmpegAvailable, off: !settingsStore.ffmpegAvailable }" />
        <span>更新元数据和封面</span>
        <span v-if="!settingsStore.ffmpegAvailable" class="ffmpeg-warn">未检测到 ffmpeg</span>
      </label>
      <label class="toggle-row" @click="toggle('overwriteOutput')">
        <span class="toggle-box" :class="{ on: settingsStore.settings.overwriteOutput }" />
        <span>覆盖已存在的输出文件</span>
      </label>
    </div>

    <h3>高级设置</h3>
    <div class="field">
      <label>QMC MMKV 路径</label>
      <input v-model="settingsStore.settings.qmcMmkvPath" placeholder="自动检测" @change="settingsStore.save()" />
    </div>
    <div class="field">
      <label>QMC MMKV 密钥</label>
      <input v-model="settingsStore.settings.qmcMmkvKey" placeholder="16 位 ASCII" @change="settingsStore.save()" />
    </div>
    <div class="field">
      <label>KGG 数据库路径</label>
      <input v-model="settingsStore.settings.kggDbPath" placeholder="自动检测" @change="settingsStore.save()" />
    </div>
  </div>
</template>

<style scoped>
.settings {
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  padding: 20px;
}

h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 16px 0 10px;
}
h3:first-child { margin-top: 0; }

.field {
  margin-bottom: 12px;
}

.field label {
  display: block;
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 4px;
}

.field input {
  width: 100%;
  padding: 8px 12px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  font-size: 13px;
  font-family: var(--font-mono);
  transition: border-color 0.15s;
}

.path-row {
  display: flex;
  gap: 8px;
}
.path-row input { flex: 1; }
.path-row button {
  padding: 8px 14px;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  border-radius: var(--radius);
  font-size: 13px;
}
.path-row button:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.toggle-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.toggle-row {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-secondary);
}
.toggle-row:hover { color: var(--text-primary); }
.toggle-row.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.toggle-row.disabled:hover { color: var(--text-secondary); }

.toggle-box.off {
  background: var(--bg-tertiary);
  opacity: 0.4;
}

.ffmpeg-warn {
  font-size: 11px;
  color: var(--error);
  margin-left: 4px;
}

.toggle-box {
  width: 36px;
  height: 20px;
  border-radius: 10px;
  background: var(--bg-tertiary);
  position: relative;
  transition: background 0.2s ease-out;
  flex-shrink: 0;
}
.toggle-box::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--text-muted);
  transition: left 0.2s ease-out, background 0.2s ease-out;
}
.toggle-box.on {
  background: var(--accent);
}
.toggle-box.on::after {
  left: 18px;
  background: #fff;
}
</style>
