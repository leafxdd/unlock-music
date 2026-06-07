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
//     `um_embed_ffmpeg` tag after running build/ffmpeg/build.sh — extracted once
//     into the user cache dir.
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

// extractEmbedded gunzips the embedded binary into a per-version user cache dir
// and returns its path, reusing a prior extraction when the sizes match.
func extractEmbedded(name string, gz []byte) (string, error) {
	zr, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return "", err
	}
	defer zr.Close()
	raw, err := io.ReadAll(zr)
	if err != nil {
		return "", err
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cacheDir, "unlock-music", "ffmpeg", embeddedVersion)
	target := filepath.Join(dir, exeName(name))

	if fi, err := os.Stat(target); err == nil && fi.Size() == int64(len(raw)) {
		return target, nil // already extracted
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	// Write to a unique temp file then atomically rename, so concurrent CLI/GUI
	// instances cannot read a half-written binary.
	tmp, err := os.CreateTemp(dir, exeName(name)+".tmp-*")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once the rename succeeds

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		return "", err
	}
	if err := os.Rename(tmpName, target); err != nil {
		// A racing process may have created it first; accept a valid existing file.
		if fi, statErr := os.Stat(target); statErr == nil && fi.Size() == int64(len(raw)) {
			return target, nil
		}
		return "", err
	}
	return target, nil
}
