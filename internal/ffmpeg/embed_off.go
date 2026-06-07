//go:build !um_embed_ffmpeg

package ffmpeg

// This build embeds no ffmpeg binaries (the default for `go build`/`go test`).
// ResolveBinary falls back to the UM_FFMPEG/UM_FFPROBE overrides or PATH. Release
// builds embed binaries by running build/ffmpeg/build.sh and compiling with
// `-tags um_embed_ffmpeg` (see build/ffmpeg/README.md).
var (
	embeddedFFmpeg  []byte
	embeddedFFprobe []byte
	embeddedVersion = ""
)
