import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { LogEntry } from '../types'

export const useLogsStore = defineStore('logs', () => {
  const entries = ref<Array<LogEntry & { ts: number }>>([])
  const maxEntries = 500

  function add(entry: LogEntry) {
    entries.value.push({ ...entry, ts: Date.now() })
    if (entries.value.length > maxEntries) {
      entries.value = entries.value.slice(-maxEntries)
    }
  }

  function clear() {
    entries.value = []
  }

  const recent = computed(() => entries.value.slice(-100))

  return { entries, recent, add, clear }
})
