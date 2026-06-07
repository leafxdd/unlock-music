package processor

type FileStatus string

const (
	StatusQueued     FileStatus = "queued"
	StatusValidating FileStatus = "validating"
	StatusDecrypting FileStatus = "decrypting"
	StatusMetadata   FileStatus = "metadata"
	StatusWriting    FileStatus = "writing"
	StatusDone       FileStatus = "done"
	StatusSkipped    FileStatus = "skipped"
	StatusFailed     FileStatus = "failed"
)

type FileEvent struct {
	Path       string
	Status     FileStatus
	OutputPath string
	AudioExt   string
	Error      error
}

type ProgressEvent struct {
	Path    string
	Current int64
	Total   int64
}

type Hooks struct {
	OnFileEvent func(FileEvent)
	OnProgress  func(ProgressEvent)
	OnLog       func(level string, msg string)
}

var noopHooks = Hooks{
	OnFileEvent: func(FileEvent) {},
	OnProgress:  func(ProgressEvent) {},
	OnLog:       func(string, string) {},
}

func (h *Hooks) defaults() {
	if h.OnFileEvent == nil {
		h.OnFileEvent = noopHooks.OnFileEvent
	}
	if h.OnProgress == nil {
		h.OnProgress = noopHooks.OnProgress
	}
	if h.OnLog == nil {
		h.OnLog = noopHooks.OnLog
	}
}
