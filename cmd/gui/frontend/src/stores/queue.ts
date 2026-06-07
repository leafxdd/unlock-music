import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { FileEvent, ProgressEvent } from '../types'
import { backend } from '../composables/useWails'

export interface QueueItem {
  path: string
  status: FileEvent['Status']
  outputPath: string
  audioExt: string
  error: string | null
  progress: number
}

export const useQueueStore = defineStore('queue', () => {
  const items = ref<Map<string, QueueItem>>(new Map())
  const processing = ref(false)
  const pendingPaths = ref<string[]>([])

  const list = computed(() => Array.from(items.value.values()))
  const doneCount = computed(() => list.value.filter(i => i.status === 'done' || i.status === 'skipped').length)
  const failedCount = computed(() => list.value.filter(i => i.status === 'failed').length)
  const totalCount = computed(() => items.value.size)

  async function addPaths(paths: string[]) {
    for (const p of paths) {
      if (!pendingPaths.value.includes(p)) {
        pendingPaths.value.push(p)
      }
    }
    const files = await backend.listFiles(paths)
    if (files) {
      for (const f of files) {
        if (!items.value.has(f)) {
          items.value.set(f, {
            path: f,
            status: 'queued',
            outputPath: '',
            audioExt: '',
            error: null,
            progress: 0,
          })
        }
      }
    }
  }

  function handleFileEvent(e: FileEvent) {
    const existing = items.value.get(e.Path)
    items.value.set(e.Path, {
      path: e.Path,
      status: e.Status,
      outputPath: e.OutputPath || existing?.outputPath || '',
      audioExt: e.AudioExt || existing?.audioExt || '',
      error: e.Error,
      progress: e.Status === 'done' ? 100 : (existing?.progress ?? 0),
    })
  }

  function handleProgress(e: ProgressEvent) {
    const item = items.value.get(e.Path)
    if (item && e.Total > 0) {
      item.progress = Math.round((e.Current / e.Total) * 100)
    }
  }

  function clear() {
    items.value.clear()
    pendingPaths.value = []
  }

  return { items, list, processing, pendingPaths, doneCount, failedCount, totalCount, addPaths, handleFileEvent, handleProgress, clear }
})
