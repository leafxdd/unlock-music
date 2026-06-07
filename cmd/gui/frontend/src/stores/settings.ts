import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Settings } from '../types'
import { backend } from '../composables/useWails'

export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<Settings>({
    inputDir: '',
    outputDir: '',
    skipNoop: true,
    removeSource: false,
    updateMetadata: false,
    overwriteOutput: false,
    qmcMmkvPath: '',
    qmcMmkvKey: '',
    kggDbPath: '',
  })

  const ffmpegAvailable = ref(true)

  async function load() {
    const [s, ff] = await Promise.allSettled([
      backend.getSettings(),
      backend.checkFFmpeg(),
    ])
    if (s.status === 'fulfilled') settings.value = s.value
    ffmpegAvailable.value = ff.status === 'fulfilled' ? ff.value : false
  }

  async function save() {
    await backend.saveSettings(settings.value)
  }

  return { settings, ffmpegAvailable, load, save }
})
