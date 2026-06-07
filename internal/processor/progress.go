package processor

import (
	"io"
	"sync"
	"time"
)

type progressReader struct {
	r       io.Reader
	path    string
	current int64
	total   int64
	onProg  func(ProgressEvent)

	mu       sync.Mutex
	lastEmit time.Time
}

func newProgressReader(r io.Reader, path string, total int64, onProg func(ProgressEvent)) *progressReader {
	return &progressReader{
		r:      r,
		path:   path,
		total:  total,
		onProg: onProg,
	}
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	if n > 0 {
		pr.current += int64(n)
		pr.maybeEmit()
	}
	return n, err
}

func (pr *progressReader) maybeEmit() {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	now := time.Now()
	if now.Sub(pr.lastEmit) < 100*time.Millisecond {
		return
	}
	pr.lastEmit = now
	pr.onProg(ProgressEvent{
		Path:    pr.path,
		Current: pr.current,
		Total:   pr.total,
	})
}
