# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Optional bundled ffmpeg/ffprobe. Release builds compiled with the `um_embed_ffmpeg` tag embed a custom **minimal static** ffmpeg (only the demuxers/muxers/encoders the app actually exercises — audio is always stream-copied), so the GUI and CLI work without a system ffmpeg. At runtime the binary is resolved from `UM_FFMPEG`/`UM_FFPROBE`, then the embedded copy (extracted to a temp dir and removed on exit), then PATH. Build it with `build/ffmpeg/build.sh`; currently bundled for windows/amd64.

### Changed
- Raised the minimum Go version to 1.26.
- DSDIFF (`.dff`) files are now copied without metadata instead of failing the file, because ffmpeg has no DSD muxer to write tags into.

### Fixed
- Preserve existing FLAC tags when writing metadata (an inverted check previously dropped them on success) and no longer write a duplicate front-cover image.
- Crafted or corrupt input files no longer crash the program: the per-file pipeline recovers from panics, and the QMC/NCM/KGM/KWM/Ximalaya decoders and the AES/PKCS7 helpers validate lengths and padding instead of panicking.
- GUI: failed files now show the real error message in the queue instead of `[object Object]`.
- Temporary files are cleaned up on error, and the reopened temp file is closed so it can be removed on Windows.
- Metadata/cover updating is disabled (with a warning) when ffmpeg is missing, even if the saved setting is on.
- Ctrl-C now cancels non-watch CLI runs cleanly.

### Security
- Bound attacker-controlled length fields by the remaining file size to prevent huge allocations (DoS) from crafted files.
- Detect WebP by its full `RIFF????WEBP` signature instead of a bare `RIFF` prefix (which also matches WAV/AVI).
- Anchor ffmpeg input/output path arguments so a filename beginning with `-` cannot be interpreted as a flag.

## [v0.2.19] - 2025-11-26

### Fixed
- MMKV parsing: Handle cases where MMKV value is empty.

## [v0.2.18] - 2025-11-16

### Changed
- QMC2: Fix `musicex\0` tag parsing.
- MMKV: Improved tolerance for corrupted MMKV file parsing.
- Updated project dependencies.

## [v0.2.17] - 2025-09-09 ⚠️ **(Broken Release)**

### Changed
- Update RegEx used to extract UDID in plist.

## [v0.2.16] - 2025-09-09 ⚠️ **(Broken Release)**

### Changed
- Update RegEx used to extract UDID in plist.

## [v0.2.15] - 2025-09-09 ⚠️ **(Broken Release)**

### Added
- Support MMKV dump in QQMusic Mac 10.x (AppStore version).

## [v0.2.14] - 2025-09-08 ⚠️ **(Broken Release)**

### Added
- Support MMKV dump in QQMusic Mac 10.x.

## [v0.2.13] - 2025-09-06 ⚠️ **(Broken Release)**

### Changed
- Updated project namespace and repository URLs to new url
- Upgraded Go version requirement to 1.25
- Restricted KGG database support to Windows platform only
- Enhanced MMKV key extraction logic with improved reliability

### Fixed
- Fixed NCM metadata parsing to properly handle mixed-type artist arrays
- Drop i386 targets in CI build

## [v0.2.12] - 2025-05-07

### Added
- KGG (KGMv5) file format support
- Support for `.mflacm` file extension

### Changed
- Updated default version identifier to "custom" for development builds
- Upgraded GoLang version

## [v0.2.11] - 2024-11-05

### Fixed
- Resolved relative path resolution issues on Windows platforms (#108)
- Improved cross-platform compatibility for file path handling

---

## Historical Versions

**Note**: This changelog was created starting from v0.2.11. For changes in earlier versions (v0.2.10 and below), please refer to the project's git commit history:

```bash
git log --oneline --before="2024-11-05"
```

Or view the complete commit history on the project repository for detailed information about features, fixes, and improvements in previous releases.
