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
# Cross-compiling (build several targets from one Linux host): set GOOS/GOARCH and
# a toolchain. Examples (Debian/Ubuntu package in parentheses):
#   linux/arm64    CROSS_PREFIX=aarch64-linux-gnu-     (gcc-aarch64-linux-gnu)
#   windows/amd64  CROSS_PREFIX=x86_64-w64-mingw32-    (gcc-mingw-w64-x86-64)
#   windows/arm64  CROSS_PREFIX=aarch64-w64-mingw32- CC=aarch64-w64-mingw32-clang
#                  (llvm-mingw: github.com/mstorsjo/llvm-mingw/releases)
# Cross builds compile a static zlib from source automatically (ZLIB_FROM_SOURCE);
# nasm is only needed for amd64 targets (arm64 builds with --disable-asm, see below).
#
set -euo pipefail

# ---- configuration ---------------------------------------------------------
FFMPEG_REF="${FFMPEG_REF:-n8.1.2}"   # pinned ffmpeg release tag — the line to bump
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

# ---- cross-compilation toolchain -------------------------------------------
# Native build by default. To cross-compile, set CROSS_PREFIX (the binutils
# prefix, e.g. aarch64-linux-gnu-, x86_64-w64-mingw32-, aarch64-w64-mingw32-) and,
# for clang-based toolchains such as llvm-mingw, CC (e.g. aarch64-w64-mingw32-clang).
CROSS_PREFIX="${CROSS_PREFIX:-}"
CC="${CC:-}"

CROSS=0
if [ -n "$CROSS_PREFIX" ] || [ -n "$CC" ] \
   || [ "$GOOS" != "$(host_goos)" ] || [ "$GOARCH" != "$(host_goarch)" ]; then
  CROSS=1
fi

# ffmpeg --arch / --target-os for the target.
case "$GOARCH" in
  amd64) FF_ARCH=x86_64 ;;
  arm64) FF_ARCH=aarch64 ;;
  *)     FF_ARCH="$GOARCH" ;;
esac
case "$GOOS" in
  windows) FF_TARGET_OS=mingw32 ;;
  darwin)  FF_TARGET_OS=darwin ;;
  *)       FF_TARGET_OS="$GOOS" ;;
esac

# Strip for the target — the host strip cannot strip foreign binaries.
STRIP_BIN="${STRIP:-${CROSS_PREFIX}strip}"

# ---- static zlib (required by the png decoder) -----------------------------
# Cross builds compile zlib from source into a per-target prefix: uniform across
# targets, and the only option for windows/arm64 (no packaged cross zlib). Native
# builds use the system zlib; set ZLIB_FROM_SOURCE=1 to force from-source there too.
ZLIB_FROM_SOURCE="${ZLIB_FROM_SOURCE:-$CROSS}"

build_zlib() {
  local zprefix="$1"
  local zsrc="$WORK/zlib-src"
  local zcc="${CC:-${CROSS_PREFIX}gcc}"
  [ -f "$zprefix/lib/libz.a" ] && return 0          # already built (cached)
  if [ ! -d "$zsrc/.git" ]; then
    git clone --depth 1 --branch v1.3.1 https://github.com/madler/zlib.git "$zsrc"
  fi
  (
    cd "$zsrc"
    "$MAKE" distclean >/dev/null 2>&1 || true
    CC="$zcc" AR="${CROSS_PREFIX}ar" RANLIB="${CROSS_PREFIX}ranlib" \
      ./configure --static --prefix="$zprefix"
    "$MAKE" -j"$JOBS" libz.a
    "$MAKE" install
  )
}

if [ "$ZLIB_FROM_SOURCE" = 1 ]; then
  ZPREFIX="$WORK/zlib-${GOOS}_${GOARCH}"
  echo ">> building static zlib for $GOOS/$GOARCH -> $ZPREFIX"
  build_zlib "$ZPREFIX"
  EXTRA_CFLAGS="-I$ZPREFIX/include ${EXTRA_CFLAGS:-}"
  EXTRA_LDFLAGS="-L$ZPREFIX/lib ${EXTRA_LDFLAGS:-}"
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

# Target arch/OS. Required when cross-compiling (alongside the cross-prefix / cc);
# also forced on a native Windows build so configure does not refuse a "native"
# MSYS build regardless of the MSYS2 shell flavour.
if [ "$CROSS" = 1 ]; then
  configure_flags+=(--enable-cross-compile --arch="$FF_ARCH" --target-os="$FF_TARGET_OS")
  [ -n "$CROSS_PREFIX" ] && configure_flags+=(--cross-prefix="$CROSS_PREFIX")
  [ -n "$CC" ] && configure_flags+=(--cc="$CC")
elif [ "$GOOS" = windows ]; then
  configure_flags+=(--target-os="$FF_TARGET_OS" --arch="$FF_ARCH")
fi

# FFmpeg's hand-written aarch64 NEON (.S) fails to assemble with some llvm-mingw
# clang versions (unrecognized mnemonics / "instruction requires: dotprod"), and
# Unlock Music never runs SIMD-optimized codecs — cover re-encode + stream-copy
# remux only — so the C fallbacks are equivalent. Drop the hand-written asm on arm64.
if [ "$FF_ARCH" = aarch64 ]; then
  configure_flags+=(--disable-asm)
fi

# ---- configure + build -----------------------------------------------------
cd "$SRC"
# The ffmpeg source tree is reused across targets, so wipe any previous configure
# first — a stale config/objects from another GOOS/GOARCH must not leak into this
# build (it manifests as undefined symbols or arch-mismatch errors at link time).
"$MAKE" distclean >/dev/null 2>&1 || true
./configure "${configure_flags[@]}"
"$MAKE" -j"$JOBS" ffmpeg$EXE ffprobe$EXE

# ---- strip, compress, stage ------------------------------------------------
mkdir -p "$OUTDIR"
"$STRIP_BIN" "ffmpeg$EXE" "ffprobe$EXE"
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
