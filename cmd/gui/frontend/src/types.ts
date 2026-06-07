export interface FileEvent {
  Path: string
  Status: 'queued' | 'validating' | 'decrypting' | 'metadata' | 'writing' | 'done' | 'skipped' | 'failed'
  OutputPath: string
  AudioExt: string
  Error: string | null
}

export interface ProgressEvent {
  Path: string
  Current: number
  Total: number
}

export interface LogEntry {
  level: string
  msg: string
}

export interface Settings {
  inputDir: string
  outputDir: string
  skipNoop: boolean
  removeSource: boolean
  updateMetadata: boolean
  overwriteOutput: boolean
  qmcMmkvPath: string
  qmcMmkvKey: string
  kggDbPath: string
}
