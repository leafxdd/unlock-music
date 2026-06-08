# Bundled ffmpeg (custom minimal static build)

Unlock Music can embed its own `ffmpeg`/`ffprobe` so the GUI and CLI work without
the user installing ffmpeg. To keep the binaries small we build a **custom static
ffmpeg with only the components the code actually uses** — roughly 8–15 MB per
binary instead of the ~75 MB stock "full" build.

`build.sh` produces the binaries; the Go side embeds them (gzip-compressed) when
compiled with the `um_embed_ffmpeg` build tag and extracts them on first use. See
`internal/ffmpeg/resolve.go`.

## How resolution works at runtime

`ResolveBinary` (in `internal/ffmpeg/resolve.go`) picks the binary in this order:

1. **`UM_FFMPEG` / `UM_FFPROBE`** env var — an explicit path. Override / escape hatch.
2. **Embedded binary** — present only in `-tags um_embed_ffmpeg` builds; extracted
   to a temporary dir and removed on exit (the CLI defers `ffmpeg.Cleanup()`, the
   GUI calls it from `OnShutdown`), so nothing persists between runs.
3. **PATH** — the dev fallback (default `go build` embeds nothing) and the only
   option on platforms not yet bundled.

To prefer a system ffmpeg over the bundled one, set `UM_FFMPEG`.

## Building

`build.sh` builds one target at a time, inferring `GOOS`/`GOARCH` from the host;
override them (with a cross toolchain) to build any target.

### Native build

Run in a POSIX shell with a C toolchain. On Windows use the **MSYS2 MINGW64**
shell (a bare Git-Bash lacks `make`).

```sh
# Windows / MSYS2 MINGW64:
pacman -S --needed git make pkgconf nasm mingw-w64-x86_64-gcc mingw-w64-x86_64-zlib
# Debian/Ubuntu:  apt-get install -y git build-essential nasm pkg-config zlib1g-dev
# macOS:          brew install nasm pkg-config

build/ffmpeg/build.sh
```

Outputs land in `internal/ffmpeg/bin/<goos>_<goarch>/`:
`ffmpeg[.exe].gz`, `ffprobe[.exe].gz`, `version.txt` (no `.exe` on non-Windows).

### Cross-compiling the bundled targets from one Linux host

Set `GOOS`/`GOARCH` plus a toolchain; `build.sh` then compiles a **static zlib from
source** for the target automatically and passes the right
`--cross-prefix`/`--cc`/`--arch`/`--target-os` to ffmpeg.

```sh
# Debian/Ubuntu prerequisites:
apt-get install -y git build-essential nasm pkg-config \
    gcc-aarch64-linux-gnu gcc-mingw-w64-x86-64
# windows/arm64 also needs llvm-mingw (clang-based) on PATH:
#   https://github.com/mstorsjo/llvm-mingw/releases

build/ffmpeg/build.sh                                                       # linux/amd64 (native)
GOOS=linux   GOARCH=arm64 CROSS_PREFIX=aarch64-linux-gnu-  build/ffmpeg/build.sh
GOOS=windows GOARCH=amd64 CROSS_PREFIX=x86_64-w64-mingw32- build/ffmpeg/build.sh
GOOS=windows GOARCH=arm64 CROSS_PREFIX=aarch64-w64-mingw32- \
    CC=aarch64-w64-mingw32-clang build/ffmpeg/build.sh
```

`nasm` is only needed for amd64 targets (arm64 uses the integrated assembler).
macOS is intentionally not bundled — darwin falls back to a PATH ffmpeg
(`brew install ffmpeg`).

Then build the app with embedding enabled:

```sh
go build -tags um_embed_ffmpeg ./cmd/um
cd cmd/gui && wails build -tags um_embed_ffmpeg
```

The default build (no tag) ignores the embed files entirely, so a clean checkout
builds and tests without ever running `build.sh`.

`EXTRA_CFLAGS` / `EXTRA_LDFLAGS` are appended to ffmpeg's flags — use them to point
at a zlib or toolchain outside the default search path, e.g. on a bare MinGW
without a packaged zlib:

```sh
EXTRA_CFLAGS="-I/path/zlib/include" EXTRA_LDFLAGS="-L/path/zlib/lib" \
  MAKE=mingw32-make build/ffmpeg/build.sh
```

### Git-Bash + pacman (no MSYS2 install)

If you only have Git-for-Windows (whose bundled `pacman` works but ships zlib
without dev files) plus a standalone MinGW (`C:\mingw64`), you can still build:

```sh
pacman -S --noconfirm --overwrite '*' mingw-w64-x86_64-zlib   # restores zlib.h + libz.a
ZI=$(cygpath -ms /mingw64/include); ZL=$(cygpath -ms /mingw64/lib)   # 8.3 paths, no spaces
EXTRA_CFLAGS="-I$ZI" EXTRA_LDFLAGS="-L$ZL" MAKE=mingw32-make build/ffmpeg/build.sh
```

`build.sh` already passes `--target-os=mingw32` on Windows, so it works from a
plain `MSYS` shell too (not only the `MINGW64` one). `cygpath -ms` is required
because the native `C:\mingw64` gcc cannot read POSIX paths or paths with spaces
(`C:\Program Files\Git`).

## Upgrading ffmpeg

Change **`FFMPEG_REF`** at the top of `build.sh` (a pinned release tag, e.g.
`n7.1.1`) and re-run. `./configure` errors loudly if any enabled component was
renamed or removed upstream, so breakage shows at build time. Re-run the project
tests (`go test ./internal/... ./algo/...`) against the new build to confirm every
command path still works, then commit the new `version.txt` value.

`BUILD_REV` bumps the cache-busting version when you change flags without changing
the ffmpeg tag.

## Why each component is enabled

Derived by tracing the real commands in `internal/ffmpeg` (`ExtractAlbumArt`,
`ProbeReader`, `updateMetaFFmpeg`) — **audio is always stream-copied (`-codec:a
copy`), never decoded or re-encoded.** FLAC metadata is written natively by
`go-flac` and never touches ffmpeg.

| Group | Components | Why |
|---|---|---|
| Protocols | `file`, `pipe` | probe & cover-extract use stdin/stdout pipes; cover input is a temp file |
| Demuxers (audio) | `mp3 mov wav ogg asf flac aac aiff ape dsf iff` | containers `sniff` detects / we probe & remux; `mov` covers m4a/mp4 |
| Demuxers (cover) | `image2 image2pipe jpeg_pipe png_pipe bmp_pipe webp_pipe gif` | cover input formats `sniff.ImageExtension` detects |
| Muxers (audio) | `mp3 mov mp4 ipod wav ogg asf` | output containers; **`.m4a` → `ipod` muxer** |
| Muxers (cover) | `image2 image2pipe` | `ExtractAlbumArt -f image2` |
| Encoders | `mjpeg` | cover re-encode for mp3/m4a |
| Decoders | `mjpeg png bmp gif webp` | decode the source cover before mjpeg re-encode (no audio decoders) |
| Parsers | `mpegaudio aac flac vorbis mjpeg` | frame the cover; harden audio remux |
| BSF | `aac_adtstoasc` | defensive for aac→mp4 |
| Filters | `scale format null aformat anull aresample` | auto-inserted pixfmt conversion (rgb→yuv) on cover re-encode |
| External | `zlib` | required by the png decoder |

**DSDIFF (`.dff`)**: there is no ffmpeg DSD muxer, so metadata cannot be written.
`internal/ffmpeg.SupportsMetadata` returns `false` for `.dff` and the processor
copies the decrypted stream verbatim. The `dsf`/`iff` demuxers are kept only so a
`.dff` can still be probed if a future path needs it.

## Adding another platform

Currently bundled: **windows/amd64, windows/arm64, linux/amd64, linux/arm64** —
each has an `internal/ffmpeg/embed_<goos>_<goarch>.go`. macOS is deliberately left
on the PATH fallback.

To bundle a new target: build its binaries (above), add an
`internal/ffmpeg/embed_<goos>_<goarch>.go` mirroring an existing one, and remove
the combo from the negation in `embed_unsupported.go` so it no longer falls
through to PATH.
