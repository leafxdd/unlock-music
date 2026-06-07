package ffmpeg

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// clearResolveCache resets the process-wide memo so each test resolves freshly.
func clearResolveCache() {
	resolveMu.Lock()
	defer resolveMu.Unlock()
	resolveCache = map[string]string{}
	resolveErr = map[string]error{}
}

func TestSupportsMetadata(t *testing.T) {
	cases := map[string]bool{
		".dff":  false, // DSDIFF has no ffmpeg muxer
		".DFF":  false, // case-insensitive
		".mp3":  true,
		".flac": true, // native go-flac path
		".m4a":  true,
		".wav":  true,
		"":      true,
	}
	for ext, want := range cases {
		if got := SupportsMetadata(ext); got != want {
			t.Errorf("SupportsMetadata(%q) = %v, want %v", ext, got, want)
		}
	}
}

func TestResolveBinaryEnvOverride(t *testing.T) {
	clearResolveCache()
	t.Cleanup(clearResolveCache)

	fake := filepath.Join(t.TempDir(), exeName("ffmpeg"))
	if err := os.WriteFile(fake, []byte("#!stub"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("UM_FFMPEG", fake)

	got, err := ResolveBinary("ffmpeg")
	if err != nil {
		t.Fatalf("ResolveBinary: %v", err)
	}
	if got != fake {
		t.Errorf("ResolveBinary = %q, want override %q", got, fake)
	}
	if !Available() {
		t.Error("Available() = false with a valid UM_FFMPEG override")
	}
}

func TestResolveBinaryEnvOverrideMissing(t *testing.T) {
	clearResolveCache()
	t.Cleanup(clearResolveCache)

	t.Setenv("UM_FFMPEG", filepath.Join(t.TempDir(), "does-not-exist"))
	if _, err := ResolveBinary("ffmpeg"); err == nil {
		t.Error("expected an error for a non-existent UM_FFMPEG path, got nil")
	}
}

// TestResolveBinaryPATHFallback checks that with no override and no embedded
// binary (the default test build) resolution defers to PATH. ffmpeg is not
// guaranteed in CI, so the test only asserts internal consistency.
func TestResolveBinaryPATHFallback(t *testing.T) {
	clearResolveCache()
	t.Cleanup(clearResolveCache)

	t.Setenv("UM_FFMPEG", "")
	got, err := ResolveBinary("ffmpeg")
	if err != nil {
		t.Skipf("ffmpeg not on PATH and not embedded: %v", err)
	}
	if got == "" {
		t.Error("ResolveBinary returned an empty path with a nil error")
	}
}

// TestEmbeddedExtractRuns validates the embed -> extract -> exec path against the
// real bundled binary. It only runs in `-tags um_embed_ffmpeg` builds (after
// build/ffmpeg/build.sh has staged the binaries); otherwise it skips.
func TestEmbeddedExtractRuns(t *testing.T) {
	if len(embeddedBinary("ffmpeg")) == 0 {
		t.Skip("no embedded ffmpeg in this build (run with -tags um_embed_ffmpeg)")
	}
	clearResolveCache()
	t.Cleanup(Cleanup)        // also removes the temp extraction
	t.Setenv("UM_FFMPEG", "") // force the embedded path, not an env override

	p, err := ResolveBinary("ffmpeg")
	if err != nil {
		t.Fatalf("ResolveBinary: %v", err)
	}

	if base := filepath.Base(filepath.Dir(p)); !strings.HasPrefix(base, "um-ffmpeg-") {
		t.Errorf("resolved %q, expected extraction under a um-ffmpeg-* temp dir", p)
	}

	out, err := exec.Command(p, "-version").CombinedOutput()
	if err != nil {
		t.Fatalf("run extracted ffmpeg: %v\n%s", err, out)
	}
	if !bytes.Contains(out, []byte("ffmpeg version")) {
		t.Errorf("unexpected -version output: %s", out)
	}
}

// TestEmbeddedCleanupRemovesExtraction confirms Cleanup deletes the extracted
// binary so nothing persists on disk. Embedded builds only.
func TestEmbeddedCleanupRemovesExtraction(t *testing.T) {
	if len(embeddedBinary("ffmpeg")) == 0 {
		t.Skip("no embedded ffmpeg in this build (run with -tags um_embed_ffmpeg)")
	}
	clearResolveCache()
	t.Cleanup(Cleanup)
	t.Setenv("UM_FFMPEG", "")

	p, err := ResolveBinary("ffmpeg")
	if err != nil {
		t.Fatalf("ResolveBinary: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("extracted binary not found: %v", err)
	}

	Cleanup()
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Errorf("Cleanup left the extraction behind: stat err = %v", err)
	}
}
