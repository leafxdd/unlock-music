[Root](../../CLAUDE.md) > [cmd](..) > **gui**

# cmd/gui/ -- Wails v2 Desktop Application

> Updated: 2026-06-08

## Module Purpose

Wails v2 GUI desktop application for Unlock Music. Provides a native window with drag-and-drop file decryption, real-time progress, settings management, and log viewing. Shares the same decryption pipeline as the CLI via `internal/processor/`.

## Entry and Startup

- **`main.go`**: Wails entry point. Embeds `frontend/dist` via `//go:embed`. Configures window (960x640), drag-and-drop (`DragAndDrop.EnableFileDrop`), `OnStartup`/`OnShutdown` lifecycle hooks (`shutdown` calls `ffmpeg.Cleanup()` to remove any extracted embedded ffmpeg), and `WebviewGpuIsDisabled` (avoids a DWM hardware-cursor stall on startup); binds the `App` struct.
- **`app.go`**: `App` struct with all methods exposed to the frontend via Wails bindings. Manages processing lifecycle with mutex-protected cancel function.
- **`settings.go`**: Settings persistence to `os.UserConfigDir()/unlock-music-gui/settings.json`.

## Exposed Bindings (Go -> JS)

All methods on `App` are callable from JavaScript via `window.go.main.App.*`:

| Method | Signature | Description |
|--------|-----------|-------------|
| `CheckFFmpeg` | `() bool` | Reports whether ffmpeg is usable via `ffmpeg.Available()` (`UM_FFMPEG` env -> embedded -> PATH) |
| `GetSettings` | `() Settings` | Returns current settings |
| `SaveSettings` | `(Settings) error` | Persists settings to disk |
| `SelectInputDir` | `() (string, error)` | Native directory picker dialog |
| `SelectOutputDir` | `() (string, error)` | Native directory picker dialog |
| `SelectInputFiles` | `() ([]string, error)` | Native multi-file picker dialog |
| `ListFiles` | `([]string) ([]string, error)` | Filters paths to supported encrypted files (walks directories) |
| `StartProcessing` | `(string) error` | Start single-path processing (file or dir) |
| `StartProcessingBatch` | `([]string) error` | Start batch processing of multiple paths |
| `StopProcessing` | `()` | Cancel current processing via context |
| `IsProcessing` | `() bool` | Check if processing is in progress |

## Emitted Events (Go -> JS)

| Event | Payload | Description |
|-------|---------|-------------|
| `file:event` | `processor.FileEvent` | File status change (queued, validating, decrypting, metadata, writing, done, skipped, failed) |
| `file:progress` | `processor.ProgressEvent` | Byte-level progress (path, current, total) -- throttled to 100ms |
| `processing:done` | none | All files processed |
| `processing:error` | `string` | Processing error message |
| `log` | `{level, msg}` | Log entry from processor |

## Settings Schema

```go
type Settings struct {
    InputDir        string `json:"inputDir"`
    OutputDir       string `json:"outputDir"`
    SkipNoop        bool   `json:"skipNoop"`         // default: true
    RemoveSource    bool   `json:"removeSource"`
    UpdateMetadata  bool   `json:"updateMetadata"`
    OverwriteOutput bool   `json:"overwriteOutput"`
    QmcMMKVPath     string `json:"qmcMmkvPath"`
    QmcMMKVKey      string `json:"qmcMmkvKey"`
    KggDbPath       string `json:"kggDbPath"`
}
```

Storage: `%APPDATA%/unlock-music-gui/settings.json` (Windows), `~/Library/Application Support/unlock-music-gui/settings.json` (macOS).

## Key Dependencies

| Package | Role |
|---------|------|
| `wailsapp/wails/v2` | Desktop framework, native dialogs, drag-and-drop, event system |
| `internal/processor` | Shared decryption pipeline |
| `algo/qmc` | QMC key loading (MMKV) |
| `algo/common` | Decoder registry for file filtering (`ListFiles`) |
| `go.uber.org/zap` | Structured logging |

## Processing Flow

1. User drops files or selects via dialog -> frontend calls `ListFiles()` to filter supported files
2. Frontend calls `StartProcessingBatch(paths)` -> Go creates cancel context, spawns goroutine
3. For each path: `processor.ProcessFile()` -> emits `file:event` and `file:progress` via Wails events
4. On completion: emits `processing:done`, resets cancel function
5. User can call `StopProcessing()` at any time to cancel via context

## Build Commands

```bash
cd cmd/gui
wails dev                                                  # Development with hot reload
wails build                                                # Production build (ffmpeg from PATH)
wails build -tags um_embed_ffmpeg                          # Bundle the embedded minimal ffmpeg
wails build -platform windows/arm64 -tags um_embed_ffmpeg  # Cross-build (CC=aarch64-w64-mingw32-*, see build/ffmpeg/)
```

## Related Files

| File | Purpose |
|------|---------|
| `main.go` | Wails entry, embed assets, window config |
| `app.go` | App struct, bindings, processing orchestration |
| `settings.go` | Settings load/save to JSON |
| `wails.json` | Wails project configuration |
| `frontend/` | Vue 3 frontend (see [frontend/CLAUDE.md](frontend/CLAUDE.md)) |

## Changelog

| Date | Change |
|------|--------|
| 2026-06-08 | Documented bundled ffmpeg (`um_embed_ffmpeg`; win/amd64+arm64 GUI), `OnShutdown` -> `ffmpeg.Cleanup()`, and corrected `CheckFFmpeg` (now `ffmpeg.Available()`, not PATH-only) |
| 2026-05-04 | Initial CLAUDE.md for GUI module |
