[Root](../../../CLAUDE.md) > [cmd](../../) > [gui](../CLAUDE.md) > **frontend**

# cmd/gui/frontend/ -- Vue 3 Frontend

> Updated: 2026-05-04

## Module Purpose

Vue 3 + Pinia + TypeScript frontend for the Unlock Music desktop GUI. Rendered inside Wails' WebView. Provides drag-and-drop file input, real-time file queue with progress, settings panel, and log viewer. Dark theme, left-right split layout.

## Tech Stack

| Technology | Version | Role |
|-----------|---------|------|
| Vue | 3.5 | UI framework (Composition API, `<script setup>`) |
| Pinia | 2.3 | State management |
| TypeScript | 5.8 | Type safety |
| Vite | 6.3 | Build tool + dev server |
| CSS Variables | -- | Custom design tokens, no component library |

## UI Layout

```
+---------------------------------------------------+
|  Header (logo + "Unlock Music" + tab navigation)  |
|  Tabs: [文件队列] [设置] [日志]                       |
+------------------------+--------------------------+
|                        |                          |
|   DropZoneCard         |   FileQueueTable         |
|   (drag-drop area      |   (file list with        |
|    + file/dir buttons   |    status badges &       |
|    + output dir picker  |    progress bars)        |
|    + start/stop btn)    |                          |
|                        |                          |
+------------------------+--------------------------+
|   ProgressPanel (overall progress bar)             |
+---------------------------------------------------+
```

Settings and Log panels are shown when their respective tabs are active.

## Component Tree

```
App.vue
  +-- AppShell.vue (main layout, tab switching, event wiring)
        +-- DropZoneCard.vue (left panel: drag-drop, file/dir picker, output dir, start/stop)
        +-- FileQueueTable.vue (right panel: file list with status, progress, clear)
        +-- ProgressPanel.vue (bottom: overall progress bar)
        +-- SettingsPanel.vue (tab: directory, processing options, advanced)
        +-- LogPanel.vue (tab: scrollable log with level coloring)
```

## Pinia Stores

### `stores/queue.ts` -- File Queue

- **State**: `Map<string, QueueItem>` keyed by file path, `processing` flag, `pendingPaths`
- **Computed**: `list`, `doneCount`, `failedCount`, `totalCount`
- **Actions**: `addPaths(paths)` (calls `backend.listFiles` to filter), `handleFileEvent(e)`, `handleProgress(e)`, `clear()`

### `stores/settings.ts` -- App Settings

- **State**: `Settings` object (mirrors Go `Settings` struct), `ffmpegAvailable` flag
- **Actions**: `load()` (calls `backend.getSettings()` + `backend.checkFFmpeg()`), `save()`

### `stores/logs.ts` -- Log Entries

- **State**: Array of `{level, msg, ts}`, max 500 entries
- **Computed**: `recent` (last 100)
- **Actions**: `add(entry)`, `clear()`

## Wails Integration (`composables/useWails.ts`)

### Backend Proxy

`backend` object wraps all `window.go.main.App.*` calls:

```typescript
backend.getSettings()            // -> App.GetSettings()
backend.saveSettings(s)          // -> App.SaveSettings(s)
backend.selectInputDir()         // -> App.SelectInputDir()
backend.selectOutputDir()        // -> App.SelectOutputDir()
backend.selectInputFiles()       // -> App.SelectInputFiles()
backend.startProcessing(path)    // -> App.StartProcessing(path)
backend.startProcessingBatch(ps) // -> App.StartProcessingBatch(ps)
backend.stopProcessing()         // -> App.StopProcessing()
backend.isProcessing()           // -> App.IsProcessing()
backend.listFiles(paths)         // -> App.ListFiles(paths)
backend.checkFFmpeg()            // -> App.CheckFFmpeg()
```

### Event Listener

`useWailsEvent(event, handler)` -- composable that registers `EventsOn` in `onMounted` and cleans up in `onUnmounted`.

### Drag and Drop

Uses Wails native drag-and-drop (`window.runtime.OnFileDrop`). CSS property `--wails-drop-target: drop` marks the drop zone. File paths are absolute native paths.

## TypeScript Types (`types.ts`)

```typescript
interface FileEvent {
  Path: string
  Status: 'queued' | 'validating' | 'decrypting' | 'metadata' | 'writing' | 'done' | 'skipped' | 'failed'
  OutputPath: string
  AudioExt: string
  Error: string | null
}

interface ProgressEvent { Path: string; Current: number; Total: number }
interface LogEntry { level: string; msg: string }
interface Settings { inputDir, outputDir, skipNoop, removeSource, updateMetadata, overwriteOutput, qmcMmkvPath, qmcMmkvKey, kggDbPath }
```

## Design System (`assets/tokens.css`)

Dark theme only. Key tokens:

| Token | Value | Usage |
|-------|-------|-------|
| `--bg-primary` | `#0f1117` | Page background |
| `--bg-secondary` | `#1a1d27` | Card/panel background |
| `--accent` | `#6c8cff` | Primary actions, active tabs |
| `--success` | `#4ade80` | Done status |
| `--error` | `#f87171` | Failed status, errors |
| `--warning` | `#fbbf24` | Skipped status |
| `--font-sans` | IBM Plex Sans | UI text |
| `--font-mono` | JetBrains Mono | Paths, logs, code |

## Key Features

- **Drag-and-drop**: Native Wails file drop with visual feedback (border color change)
- **Batch processing**: Multiple files/directories queued and processed sequentially
- **Real-time progress**: Per-file progress bars (byte-level) + overall progress (file count)
- **FFmpeg detection**: Settings panel disables "update metadata" toggle when ffmpeg not found, shows warning
- **Status tracking**: Each file shows status badge (queued/validating/decrypting/metadata/writing/done/skipped/failed) with colored indicators
- **Title bar drag**: Header has `--wails-draggable: drag` for native window dragging
- **Auto-scroll logs**: Log panel auto-scrolls to bottom on new entries

## Build

```bash
npm install          # Install dependencies
npm run dev          # Vite dev server (used by wails dev)
npm run build        # Production build (vue-tsc + vite build)
```

Output: `frontend/dist/` (embedded into Go binary via `//go:embed`).

## Related Files

| File | Purpose |
|------|---------|
| `src/App.vue` | Root component, imports AppShell |
| `src/main.ts` | Vue + Pinia initialization |
| `src/types.ts` | Shared TypeScript interfaces |
| `src/composables/useWails.ts` | Wails backend proxy + event composable |
| `src/stores/queue.ts` | File queue state management |
| `src/stores/settings.ts` | Settings state + ffmpeg detection |
| `src/stores/logs.ts` | Log buffer (500 max) |
| `src/components/AppShell.vue` | Main layout, tab navigation, event wiring |
| `src/components/DropZoneCard.vue` | Drag-drop zone, file/dir pickers, start/stop |
| `src/components/FileQueueTable.vue` | File list with status badges + progress |
| `src/components/ProgressPanel.vue` | Overall progress bar |
| `src/components/SettingsPanel.vue` | Directory, processing options, advanced settings |
| `src/components/LogPanel.vue` | Scrollable log viewer |
| `src/assets/tokens.css` | CSS design tokens (dark theme) |
| `index.html` | HTML entry point |
| `vite.config.ts` | Vite + Vue plugin + `@` alias |
| `package.json` | NPM dependencies and scripts |

## Changelog

| Date | Change |
|------|--------|
| 2026-05-04 | Initial CLAUDE.md for frontend module |
