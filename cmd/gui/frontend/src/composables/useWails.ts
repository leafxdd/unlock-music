import { onMounted, onUnmounted } from 'vue'

declare global {
  interface Window {
    go: {
      main: {
        App: {
          GetSettings(): Promise<any>
          SaveSettings(s: any): Promise<void>
          SelectInputDir(): Promise<string>
          SelectOutputDir(): Promise<string>
          SelectInputFiles(): Promise<string[]>
          StartProcessing(inputPath: string): Promise<void>
          StartProcessingBatch(paths: string[]): Promise<void>
          StopProcessing(): Promise<void>
          IsProcessing(): Promise<boolean>
          ListFiles(paths: string[]): Promise<string[]>
          CheckFFmpeg(): Promise<boolean>
        }
      }
    }
    runtime: {
      EventsOn(event: string, callback: (...args: any[]) => void): () => void
      EventsOff(event: string): void
      OnFileDrop(callback: (x: number, y: number, paths: string[]) => void, useDropTarget: boolean): void
      OnFileDropOff(): void
    }
  }
}

export function useWailsEvent(event: string, handler: (...args: any[]) => void) {
  let cleanup: (() => void) | null = null
  onMounted(() => {
    cleanup = window.runtime.EventsOn(event, handler)
  })
  onUnmounted(() => {
    cleanup?.()
  })
}

export const backend = {
  getSettings: () => window.go.main.App.GetSettings(),
  saveSettings: (s: any) => window.go.main.App.SaveSettings(s),
  selectInputDir: () => window.go.main.App.SelectInputDir(),
  selectOutputDir: () => window.go.main.App.SelectOutputDir(),
  selectInputFiles: () => window.go.main.App.SelectInputFiles(),
  startProcessing: (path: string) => window.go.main.App.StartProcessing(path),
  startProcessingBatch: (paths: string[]) => window.go.main.App.StartProcessingBatch(paths),
  stopProcessing: () => window.go.main.App.StopProcessing(),
  isProcessing: () => window.go.main.App.IsProcessing(),
  listFiles: (paths: string[]) => window.go.main.App.ListFiles(paths),
  checkFFmpeg: () => window.go.main.App.CheckFFmpeg(),
}
