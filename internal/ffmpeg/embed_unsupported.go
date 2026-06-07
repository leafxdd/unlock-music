//go:build um_embed_ffmpeg && !(windows && amd64)

package ffmpeg

// Built with -tags um_embed_ffmpeg on a platform whose ffmpeg binaries have not
// been bundled yet: nothing is embedded, so ResolveBinary falls back to PATH. To
// bundle this platform, build its binaries (build/ffmpeg/build.sh) and add an
// embed_<goos>_<goarch>.go mirroring embed_windows_amd64.go.
var (
	embeddedFFmpeg  []byte
	embeddedFFprobe []byte
	embeddedVersion = ""
)
