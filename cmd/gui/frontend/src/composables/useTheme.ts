import { ref } from 'vue'

// Theme handling lives entirely in the frontend (localStorage), independent of
// the Go-side settings.json: it must apply synchronously before mount to avoid a
// dark→light flash, and it is a pure UI preference orthogonal to decrypt options.

export type ThemePref = 'system' | 'light' | 'dark'
export type ResolvedTheme = 'light' | 'dark'

const STORAGE_KEY = 'um-theme'
// Cycle order for the toolbar toggle. Keep in sync with the inline boot script
// in index.html (which only needs the resolve half).
const CYCLE: ThemePref[] = ['system', 'light', 'dark']

const media = window.matchMedia('(prefers-color-scheme: light)')

function readStored(): ThemePref {
  const v = localStorage.getItem(STORAGE_KEY)
  return v === 'light' || v === 'dark' || v === 'system' ? v : 'system'
}

const pref = ref<ThemePref>(readStored())

function resolve(p: ThemePref): ResolvedTheme {
  if (p === 'system') return media.matches ? 'light' : 'dark'
  return p
}

function apply() {
  document.documentElement.setAttribute('data-theme', resolve(pref.value))
}

function setPref(p: ThemePref) {
  pref.value = p
  localStorage.setItem(STORAGE_KEY, p)
  apply()
}

function cycle() {
  setPref(CYCLE[(CYCLE.indexOf(pref.value) + 1) % CYCLE.length])
}

// While "follow system" is active, react to OS appearance changes live.
media.addEventListener('change', () => {
  if (pref.value === 'system') apply()
})

// Apply once at module load so the attribute matches the stored preference even
// before the first component mounts (the inline boot script handles the very
// first paint; this keeps the reactive ref and the attribute in agreement).
apply()

export function useTheme() {
  return { pref, resolve, setPref, cycle }
}
