//go:build um_embed_ffmpeg && linux && arm64

package ffmpeg

import (
	_ "embed"
	"strings"
)

// These files are produced by build/ffmpeg/build.sh (GOOS=linux GOARCH=arm64,
// cross-compiled via aarch64-linux-gnu) and are only referenced when compiling
// with `-tags um_embed_ffmpeg`. The default build never touches them, so they need
// not exist for `go build`/`go test` to succeed.

//go:embed bin/linux_arm64/ffmpeg.gz
var embeddedFFmpeg []byte

//go:embed bin/linux_arm64/ffprobe.gz
var embeddedFFprobe []byte

//go:embed bin/linux_arm64/version.txt
var embeddedVersionRaw string

// embeddedVersion names the temp-dir prefix for the extracted binaries; it must
// change whenever the embedded binaries change so stale copies are not reused.
var embeddedVersion = strings.TrimSpace(embeddedVersionRaw)
