#!/usr/bin/env bash
#
# build.sh — build a minimal static ffmpeg + ffprobe for Unlock Music and stage
# them for embedding into internal/ffmpeg/bin/<goos>_<goarch>/.
#
# Only the components Unlock Music actually exercises are enabled. The list was
# derived empirically by tracing the three real command paths in internal/ffmpeg
# (ExtractAlbumArt, ProbeReader, updateMetaFFmpeg) against representative samples
# for every supported container and cover type. See README.md for the full
# component -> code-path rationale.
#
# Upgrading ffmpeg: change FFMPEG_REF below and re-run. ./configure errors loudly
# on any component renamed or removed upstream, so breakage surfaces at build time
# rather than at runtime; the integration tests then confirm the new build works.
#
# Prerequisites (Windows, MSYS2 MINGW64 shell):
#   pacman -S --needed git make pkgconf nasm \
#            mingw-w64-x86_64-gcc mingw-w64-x86_64-zlib
# Prerequisites (Debian/Ubuntu):
#   apt-get install -y git build-essential nasm pkg-config zlib1g-dev
# Prerequisites (macOS, Homebrew):
#   brew install nasm pkg-config   # zlib ships with the SDK
#
set -euo pipefail

# ---- configuration ---------------------------------------------------------
FFMPEG_REF="${FFMPEG_REF:-n7.1.1}"   # pinned ffmpeg release tag — the line to bump
BUILD_REV="${BUILD_REV:-1}"          # bump when flags change but ffmpeg does not
MAKE="${MAKE:-make}"
JOBS="${JOBS:-$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)}"

# ---- paths -----------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
WORK="${WORK:-$SCRIPT_DIR/.work}"
SRC="$WORK/ffmpeg"

# Resolve the target GOOS/GOARCH from the host (override via the environment).
host_goos() { case "$(uname -s)" in MINGW*|MSYS*|CYGWIN*) echo windows;; Darwin) echo darwin;; *) echo linux;; esac; }
host_goarch() { case "$(uname -m)" in x86_64|amd64) echo amd64;; aarch64|arm64) echo arm64;; *) uname -m;; esac; }
GOOS="${GOOS:-$(host_goos)}"
GOARCH="${GOARCH:-$(host_goarch)}"
EXE=""; [ "$GOOS" = windows ] && EXE=".exe"
OUTDIR="$REPO_ROOT/internal/ffmpeg/bin/${GOOS}_${GOARCH}"

echo ">> building ffmpeg $FFMPEG_REF (rev $BUILD_REV) for $GOOS/$GOARCH"
echo ">> output: $OUTDIR"

# ---- fetch source (pinned tag) ---------------------------------------------
mkdir -p "$WORK"
if [ ! -d "$SRC/.git" ]; then
  git clone --depth 1 --branch "$FFMPEG_REF" https://github.com/FFmpeg/FFmpeg.git "$SRC"
else
  git -C "$SRC" fetch --depth 1 origin "$FFMPEG_REF"
  git -C "$SRC" checkout -f "$FFMPEG_REF"
fi

# ---- minimal component set (comments allowed inside the array) -------------
# Each group maps to a code path in internal/ffmpeg; do not prune without
# re-tracing — see README.md.
configure_flags=(
  --pkg-config-flags=--static
  --extra-cflags="-Os -ffunction-sections -fdata-sections ${EXTRA_CFLAGS:-}"
  --extra-ldflags="-static -Wl,--gc-sections ${EXTRA_LDFLAGS:-}"
  --enable-static --disable-shared
  --disable-everything --disable-autodetect --disable-network
  --disable-debug --enable-small
  --disable-doc --disable-htmlpages --disable-manpages --disable-podpages --disable-txtpages
  --disable-ffplay --enable-ffmpeg --enable-ffprobe
  --enable-zlib   # required by the png decoder (cover re-encode)

  # protocols: stdin/stdout pipes (probe + cover extract) and the temp cover file
  --enable-protocol=file,pipe

  # demuxers — audio containers we probe and remux from
  --enable-demuxer=mp3,mov,wav,ogg,asf,flac,aac,aiff,ape,dsf,iff
  # demuxers — cover images we decode to re-encode (jpg/png/bmp/webp use *_pipe; gif is plain)
  --enable-demuxer=image2,image2pipe,jpeg_pipe,png_pipe,bmp_pipe,webp_pipe,gif

  # muxers — output containers we write (note: .m4a uses the 'ipod' muxer)
  --enable-muxer=mp3,mov,mp4,ipod,wav,ogg,asf
  # muxers — cover extraction writes "-f image2" to a pipe
  --enable-muxer=image2,image2pipe

  # decoders — ONLY to re-encode the cover to mjpeg; audio is always stream-copied
  --enable-decoder=mjpeg,png,bmp,gif,webp
  # encoder — the cover for mp3/m4a
  --enable-encoder=mjpeg

  # parsers — frame the cover (mjpeg) and harden same-container audio remux
  --enable-parser=mpegaudio,aac,flac,vorbis,mjpeg
  # bitstream filter — aac into mp4/m4a, defensive (not observed but cheap)
  --enable-bsf=aac_adtstoasc
  # filters — auto-inserted pixel-format conversion when re-encoding the cover
  --enable-filter=scale,format,null,aformat,anull,aresample
)

# Force the mingw target on Windows so the build works from any MSYS2 shell
# flavour (a plain "MSYS" shell makes configure refuse a native build); derived
# from GOARCH so it stays correct for arm64. The toolchain (gcc) is native, so no
# cross-prefix is needed.
if [ "$GOOS" = windows ]; then
  case "$GOARCH" in
    amd64) FF_ARCH=x86_64 ;;
    arm64) FF_ARCH=aarch64 ;;
    *)     FF_ARCH="$GOARCH" ;;
  esac
  configure_flags+=(--target-os=mingw32 --arch="$FF_ARCH")
fi

# ---- configure + build -----------------------------------------------------
cd "$SRC"
./configure "${configure_flags[@]}"
"$MAKE" -j"$JOBS" ffmpeg$EXE ffprobe$EXE

# ---- strip, compress, stage ------------------------------------------------
mkdir -p "$OUTDIR"
strip "ffmpeg$EXE" "ffprobe$EXE"
gzip -9 -c "ffmpeg$EXE"  > "$OUTDIR/ffmpeg$EXE.gz"
gzip -9 -c "ffprobe$EXE" > "$OUTDIR/ffprobe$EXE.gz"
printf '%s+um%s\n' "$FFMPEG_REF" "$BUILD_REV" > "$OUTDIR/version.txt"

echo
echo ">> staged binaries:"
ls -la "$OUTDIR"
echo
echo ">> raw vs compressed:"
du -h "ffmpeg$EXE" "ffprobe$EXE" "$OUTDIR"/*.gz 2>/dev/null || true
echo
echo ">> now build with embedding enabled, e.g.:"
echo "   go build -tags um_embed_ffmpeg ./cmd/um"
echo "   (cmd/gui) wails build -tags um_embed_ffmpeg"
