# internal/ffmpeg/bin/

Staging area for the custom minimal ffmpeg/ffprobe binaries embedded into release
builds. Populated by `build/ffmpeg/build.sh` as `<goos>_<goarch>/{ffmpeg,ffprobe}
[.exe].gz` plus `version.txt`.

These artifacts are **gitignored** (see `.gitignore`) — a default `go build` does
not need them, and they are only referenced under the `um_embed_ffmpeg` build tag.
Run `build/ffmpeg/build.sh` before building with that tag.

See [build/ffmpeg/README.md](../../../build/ffmpeg/README.md) for details.
