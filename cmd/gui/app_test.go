package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDrop(t *testing.T) {
	a := &App{}
	dir := t.TempDir()

	if got := a.ResolveDrop([]string{dir}); !got.IsDir || got.Dir != dir {
		t.Errorf("folder drop: got %+v, want Dir=%q IsDir=true", got, dir)
	}

	f := filepath.Join(dir, "song.qmc0")
	if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := a.ResolveDrop([]string{f}); got.IsDir || got.Dir != dir {
		t.Errorf("file drop: got %+v, want Dir=%q IsDir=false", got, dir)
	}

	if got := a.ResolveDrop(nil); got.Dir != "" || got.IsDir {
		t.Errorf("empty drop: got %+v, want zero value", got)
	}
}
