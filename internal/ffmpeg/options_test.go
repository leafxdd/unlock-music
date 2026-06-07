package ffmpeg

import (
	"slices"
	"testing"
)

// TestOutputBuilderArgsDeterministic guards against the previous map-iteration
// nondeterminism in the argument builder.
func TestOutputBuilderArgsDeterministic(t *testing.T) {
	build := func() []string {
		out := newOutputBuilder("out.mp3")
		out.AddOption("map", "0:a")
		out.AddOption("map", "1:v")
		out.AddOption("codec:a", "copy")
		out.AddOption("codec:v", "mjpeg")
		out.AddMetadata("", "title", "T")
		return out.Args()
	}
	first := build()
	for range 20 {
		if got := build(); !slices.Equal(got, first) {
			t.Fatalf("non-deterministic args:\n %v\nvs\n %v", first, got)
		}
	}

	// -map values must keep their insertion order (0:a before 1:v).
	iMap0 := slices.Index(first, "0:a")
	iMap1 := slices.Index(first, "1:v")
	if iMap0 < 0 || iMap1 < 0 || iMap0 > iMap1 {
		t.Errorf("expected -map 0:a before -map 1:v, got %v", first)
	}
	if first[len(first)-1] != "out.mp3" {
		t.Errorf("expected output path last, got %v", first)
	}
}

func TestSafeArgPath(t *testing.T) {
	tests := []struct{ in, want string }{
		{"out.mp3", "out.mp3"},
		{"-rf.mp3", "./-rf.mp3"},
		{"/abs/path.mp3", "/abs/path.mp3"},
		{`C:\abs\path.mp3`, `C:\abs\path.mp3`},
	}
	for _, tt := range tests {
		if got := safeArgPath(tt.in); got != tt.want {
			t.Errorf("safeArgPath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// TestInputBuilderArgsSafePath confirms a filename starting with '-' can't be
// parsed as an ffmpeg flag.
func TestInputBuilderArgsSafePath(t *testing.T) {
	in := newInputBuilder("-evil.flac")
	args := in.Args()
	if len(args) < 2 || args[len(args)-2] != "-i" || args[len(args)-1] != "./-evil.flac" {
		t.Errorf("expected '-i ./-evil.flac', got %v", args)
	}
}
