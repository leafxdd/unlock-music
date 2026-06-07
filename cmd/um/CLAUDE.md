[Root](../../CLAUDE.md) > [cmd](..) > **um**

# cmd/um/ -- CLI Entry Point

> Updated: 2026-05-04

## Module Purpose

Command-line interface for Unlock Music. Parses CLI flags via `urfave/cli/v2`, initializes the shared `internal/processor` pipeline, and processes files or watches directories.

## Entry: `main.go`

Single-file entry point. Contains:

1. **CLI setup**: `urfave/cli/v2` app with flags (`-i`, `-o`, `--watch`, `--update-metadata`, `--overwrite`, etc.)
2. **Logger setup**: Zap production encoder with color levels and RFC3339 timestamps. Verbose mode toggleable.
3. **Processor integration**: Constructs `processor.New(cfg, logger, hooks)` from CLI flags, delegates to `ProcessFile`, `ProcessDir`, or `WatchDir`.
4. **MMKV loading**: Reads QQ Music key database at startup via `qmc.LoadMMKVOrDefault()`
5. **Supported extensions**: `--supported-ext` flag prints registered decoder extensions

## CLI Flags

| Flag | Aliases | Default | Description |
|------|---------|---------|-------------|
| `--input` | `-i` | cwd or first arg | Input file or directory |
| `--output` | `-o` | same as input dir | Output directory |
| `--qmc-mmkv` | `--db` | auto-detect | QQ Music MMKV path |
| `--qmc-mmkv-key` | `--key` | (none) | QQ Music MMKV password (16 ASCII chars) |
| `--kgg-db` | | `%APPDATA%/Kugou8/KGMusicV3.db` | Kugou v5 database path |
| `--remove-source` | `-rs` | `false` | Delete source file after success |
| `--skip-noop` | `-n` | `true` | Skip noop decoders |
| `--verbose` | `-V` | `false` | Enable debug logging |
| `--update-metadata` | | `false` | Fetch and write metadata/cover art |
| `--overwrite` | | `false` | Overwrite existing output files |
| `--watch` | | `false` | Watch input directory for new files |
| `--supported-ext` | | `false` | Print supported extensions and exit |

## Processing Flow

1. Parse flags, resolve absolute input/output paths
2. Load QMC MMKV keys (optional, warns on failure)
3. Resolve KGG database path
4. Create `processor.New()` with config and empty hooks (CLI uses direct zap logging)
5. If input is directory and `--watch`: `proc.WatchDir(ctx, input)` with `SIGINT` context
6. If input is directory: `proc.ProcessDir(ctx, input)` (recursive)
7. If input is file: `proc.ProcessFile(ctx, input)`

## Key Dependencies

| Package | Role |
|---------|------|
| `urfave/cli/v2` | CLI framework |
| `internal/processor` | Shared processing pipeline |
| `algo/qmc` | QMC key loading |
| `go.uber.org/zap` | Logging |

## Related Files

| File | Purpose |
|------|---------|
| `main.go` | CLI entry, flag parsing, processor setup |

## Changelog

| Date | Change |
|------|--------|
| 2026-05-04 | Updated: reflects processor extraction, current CLI flags, removed inline processor logic |
| 2026-04-21 | Initial CLAUDE.md |
