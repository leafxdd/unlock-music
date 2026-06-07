package processor

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"git.um-react.app/um/cli/algo/common"
	"go.uber.org/zap"
)

func TestProgressReader(t *testing.T) {
	data := strings.Repeat("x", 1024)
	rd := strings.NewReader(data)

	var events []ProgressEvent
	pr := newProgressReader(rd, "/test/file.qmc0", int64(len(data)), func(e ProgressEvent) {
		events = append(events, e)
	})

	buf := make([]byte, 256)
	total := 0
	for {
		n, err := pr.Read(buf)
		total += n
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if total != len(data) {
		t.Errorf("read %d bytes, want %d", total, len(data))
	}
	if pr.current != int64(len(data)) {
		t.Errorf("current = %d, want %d", pr.current, len(data))
	}
	if len(events) == 0 {
		t.Error("expected at least one progress event")
	}
	for _, e := range events {
		if e.Path != "/test/file.qmc0" {
			t.Errorf("event path = %q, want /test/file.qmc0", e.Path)
		}
		if e.Total != int64(len(data)) {
			t.Errorf("event total = %d, want %d", e.Total, len(data))
		}
	}
}

func TestProgressReaderThrottle(t *testing.T) {
	data := bytes.Repeat([]byte{0}, 10000)
	rd := bytes.NewReader(data)

	var count int
	pr := newProgressReader(rd, "/test", int64(len(data)), func(e ProgressEvent) {
		count++
	})

	buf := make([]byte, 1)
	for {
		_, err := pr.Read(buf)
		if err == io.EOF {
			break
		}
	}

	if count > 200 {
		t.Errorf("too many events (%d) for 10000 byte-at-a-time reads with 100ms throttle", count)
	}
}

func TestHooksDefaults(t *testing.T) {
	h := Hooks{}
	h.defaults()

	if h.OnFileEvent == nil {
		t.Error("OnFileEvent should not be nil after defaults()")
	}
	if h.OnProgress == nil {
		t.Error("OnProgress should not be nil after defaults()")
	}
	if h.OnLog == nil {
		t.Error("OnLog should not be nil after defaults()")
	}

	h.OnFileEvent(FileEvent{})
	h.OnProgress(ProgressEvent{})
	h.OnLog("INFO", "test")
}

func TestHooksPreserveCustom(t *testing.T) {
	var called bool
	h := Hooks{
		OnFileEvent: func(e FileEvent) { called = true },
	}
	h.defaults()

	h.OnFileEvent(FileEvent{})
	if !called {
		t.Error("custom OnFileEvent should have been preserved")
	}
}

func TestProgressReaderEmptyReader(t *testing.T) {
	rd := strings.NewReader("")
	var events []ProgressEvent
	pr := newProgressReader(rd, "/empty", 0, func(e ProgressEvent) {
		events = append(events, e)
	})

	buf := make([]byte, 16)
	n, err := pr.Read(buf)
	if n != 0 || err != io.EOF {
		t.Errorf("expected (0, EOF), got (%d, %v)", n, err)
	}
}

func TestProgressEventFields(t *testing.T) {
	_ = time.Millisecond
	e := ProgressEvent{Path: "/a/b.qmc0", Current: 50, Total: 100}
	if e.Path != "/a/b.qmc0" || e.Current != 50 || e.Total != 100 {
		t.Error("ProgressEvent fields incorrect")
	}
}

func TestFileStatusConstants(t *testing.T) {
	statuses := []FileStatus{
		StatusQueued, StatusValidating, StatusDecrypting,
		StatusMetadata, StatusWriting, StatusDone,
		StatusSkipped, StatusFailed,
	}
	seen := make(map[FileStatus]bool)
	for _, s := range statuses {
		if seen[s] {
			t.Errorf("duplicate status: %s", s)
		}
		seen[s] = true
		if s == "" {
			t.Error("empty status constant")
		}
	}
}

// panicDecoder simulates a decoder that panics on crafted input, as several real
// decoders can when fed malformed files (e.g. an out-of-range slice in Validate).
// The processor must downgrade this to a failed file rather than crash the whole
// CLI run or the GUI process.
type panicDecoder struct{}

func (panicDecoder) Validate() error            { panic("crafted file: index out of range") }
func (panicDecoder) Read(_ []byte) (int, error) { return 0, io.EOF }

func init() {
	common.RegisterDecoder("panictest", false, func(*common.DecoderParams) common.Decoder {
		return panicDecoder{}
	})
}

func TestProcessFileRecoversFromPanic(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "evil.panictest")
	if err := os.WriteFile(src, []byte("malformed"), 0644); err != nil {
		t.Fatal(err)
	}

	var failed *FileEvent
	p := New(
		Config{InputDir: dir, OutputDir: dir},
		zap.NewNop(),
		Hooks{OnFileEvent: func(e FileEvent) {
			if e.Status == StatusFailed {
				ev := e
				failed = &ev
			}
		}},
	)

	// A panic inside the decoder must surface as an error, never crash the process.
	err := p.ProcessFile(context.Background(), src)
	if err == nil {
		t.Fatal("expected an error from a panicking decoder, got nil")
	}
	if !strings.Contains(err.Error(), "panic") {
		t.Errorf("error should mention panic, got: %v", err)
	}
	if failed == nil {
		t.Fatal("expected a StatusFailed file event")
	}
	if failed.Error == nil {
		t.Error("StatusFailed event should carry the error")
	}
}
