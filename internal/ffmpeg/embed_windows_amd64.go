//go:build um_embed_ffmpeg && windows && amd64

package ffmpeg

import (
	_ "embed"
	"strings"
)

// These files are produced by build/ffmpeg/build.sh (run in MSYS2) and are only
// referenced when compiling with `-tags um_embed_ffmpeg`. The default build never
// touches them, so they need not exist for `go build`/`go test` to succeed.

//go:embed bin/windows_amd64/ffmpeg.exe.gz
var embeddedFFmpeg []byte

//go:embed bin/windows_amd64/ffprobe.exe.gz
var embeddedFFprobe []byte

//go:embed bin/windows_amd64/version.txt
var embeddedVersionRaw string

// embeddedVersion names the cache subdirectory for the extracted binaries; it must
// change whenever the embedded binaries change so stale copies are not reused.
var embeddedVersion = strings.TrimSpace(embeddedVersionRaw)
