package utils

import (
	"errors"
	"os"
	"strings"
	"testing"
)

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("boom") }

func TestWriteTempFileSuccess(t *testing.T) {
	content := "hello world"
	path, err := WriteTempFile(strings.NewReader(content), ".mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	if !strings.HasSuffix(path, ".mp3") {
		t.Errorf("expected .mp3 suffix, got %q", path)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Errorf("content = %q, want %q", got, content)
	}
}

// TestWriteTempFileErrorReturnsNoPath confirms a copy failure reports an error
// and returns no path (the temp file is cleaned up internally rather than leaked).
func TestWriteTempFileErrorReturnsNoPath(t *testing.T) {
	path, err := WriteTempFile(errReader{}, ".mp3")
	if err == nil {
		t.Fatal("expected error from failing reader")
	}
	if path != "" {
		t.Errorf("expected empty path on error, got %q", path)
	}
}
