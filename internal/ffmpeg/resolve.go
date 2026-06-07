package ffmpeg

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

// Binary resolution order for ffmpeg/ffprobe:
//
//  1. An explicit path from the environment (UM_FFMPEG / UM_FFPROBE). Lets a user
//     force a specific build and is the escape hatch if a bundled binary misbehaves.
//  2. The binary embedded into this build — release builds compiled with the
//     `um_embed_ffmpeg` tag after running build/ffmpeg/build.sh — extracted into a
//     per-process temp dir that Cleanup() removes on exit (nothing persists).
//  3. A binary found on PATH. This is the dev fallback (the default `go build`
//     embeds nothing) and the only option on platforms not yet bundled.
//
// To prefer a system ffmpeg over the embedded one, set UM_FFMPEG, or swap the
// embedded and PATH steps in locateBinary.

var (
	resolveMu    sync.Mutex
	resolveCache = map[string]string{}
	resolveErr   = map[string]error{}
)

var (
	extractMu   sync.Mutex
	extractDir  string // per-process temp dir for extracted binaries ("" until created)
	extractErr  error
	extractInit bool
)

// tempExtractDir lazily creates a private temp directory to hold the extracted
// binaries (shared by ffmpeg and ffprobe); Cleanup removes it.
func tempExtractDir() (string, error) {
	extractMu.Lock()
	defer extractMu.Unlock()
	if !extractInit {
		extractDir, extractErr = os.MkdirTemp("", "um-ffmpeg-"+embeddedVersion+"-")
		extractInit = true
	}
	return extractDir, extractErr
}

// Cleanup removes the binaries extracted from the embedded payload this run, so the
// bundled ffmpeg never persists on disk between runs. The CLI defers it; the GUI
// calls it from OnShutdown. It is a no-op when nothing was extracted (the UM_FFMPEG
// override or a PATH binary was used, or this build embeds nothing).
func Cleanup() {
	// Drop memoised paths so a later resolve re-extracts instead of handing back a
	// path under the directory we are about to delete.
	resolveMu.Lock()
	resolveCache = map[string]string{}
	resolveErr = map[string]error{}
	resolveMu.Unlock()

	extractMu.Lock()
	dir := extractDir
	extractDir, extractErr, extractInit = "", nil, false
	extractMu.Unlock()

	if dir != "" {
		_ = os.RemoveAll(dir)
	}
}

// ResolveBinary returns a usable path for the named ffmpeg-family binary
// ("ffmpeg" or "ffprobe"), memoised for the lifetime of the process. The first
// call for an embedded binary extracts it to disk.
func ResolveBinary(name string) (string, error) {
	resolveMu.Lock()
	defer resolveMu.Unlock()
	if p, ok := resolveCache[name]; ok {
		return p, resolveErr[name]
	}
	p, err := locateBinary(name)
	resolveCache[name] = p
	resolveErr[name] = err
	return p, err
}

func locateBinary(name string) (string, error) {
	// 1. explicit override
	if env := binEnvVar(name); env != "" {
		if p := os.Getenv(env); p != "" {
			if _, err := os.Stat(p); err != nil {
				return "", fmt.Errorf("%s=%q is not usable: %w", env, p, err)
			}
			return p, nil
		}
	}

	// 2. embedded binary (release builds compiled with -tags um_embed_ffmpeg)
	if gz := embeddedBinary(name); len(gz) > 0 {
		if p, err := extractEmbedded(name, gz); err == nil {
			return p, nil
		} else if onPath, lookErr := exec.LookPath(exeName(name)); lookErr == nil {
			return onPath, nil // extraction failed but PATH has a usable copy
		} else {
			return "", fmt.Errorf("extract embedded %s: %w", name, err)
		}
	}

	// 3. PATH
	if p, err := exec.LookPath(exeName(name)); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("%s not found: this build embeds no %s and none is on PATH (set %s to override)",
		name, name, binEnvVar(name))
}

// Available reports whether ffmpeg can be resolved without extracting it. The GUI
// uses this to decide whether metadata/cover updating is possible for the run.
func Available() bool {
	if env := os.Getenv(binEnvVar("ffmpeg")); env != "" {
		if _, err := os.Stat(env); err == nil {
			return true
		}
	}
	if len(embeddedBinary("ffmpeg")) > 0 {
		return true
	}
	_, err := exec.LookPath(exeName("ffmpeg"))
	return err == nil
}

func binEnvVar(name string) string {
	switch name {
	case "ffmpeg":
		return "UM_FFMPEG"
	case "ffprobe":
		return "UM_FFPROBE"
	default:
		return ""
	}
}

func embeddedBinary(name string) []byte {
	switch name {
	case "ffmpeg":
		return embeddedFFmpeg
	case "ffprobe":
		return embeddedFFprobe
	}
	return nil
}

func exeName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// extractEmbedded gunzips the embedded binary into the per-process temp dir
// (created lazily, shared by ffmpeg and ffprobe) and returns its path. Cleanup()
// removes the dir on exit, so nothing is left on disk between runs.
func extractEmbedded(name string, gz []byte) (string, error) {
	dir, err := tempExtractDir()
	if err != nil {
		return "", err
	}

	zr, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return "", err
	}
	defer zr.Close()
	raw, err := io.ReadAll(zr)
	if err != nil {
		return "", err
	}

	// The temp dir is private to this process (unique name) and extraction runs
	// under resolveMu, so a plain write is safe — no cross-instance race to guard.
	target := filepath.Join(dir, exeName(name))
	if err := os.WriteFile(target, raw, 0o755); err != nil {
		return "", err
	}
	return target, nil
}
