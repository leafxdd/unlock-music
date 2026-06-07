package processor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.um-react.app/um/cli/algo/common"
	"git.um-react.app/um/cli/internal/ffmpeg"
	"git.um-react.app/um/cli/internal/sniff"
	"git.um-react.app/um/cli/internal/utils"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

type Processor struct {
	config Config
	logger *zap.Logger
	hooks  Hooks
}

func New(cfg Config, logger *zap.Logger, hooks Hooks) *Processor {
	hooks.defaults()
	return &Processor{
		config: cfg,
		logger: logger,
		hooks:  hooks,
	}
}

func (p *Processor) ProcessDir(ctx context.Context, dir string) error {
	items, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var lastError error
	for _, item := range items {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		filePath := filepath.Join(dir, item.Name())
		if item.IsDir() {
			if err = p.ProcessDir(ctx, filePath); err != nil {
				lastError = err
			}
			continue
		}

		if err := p.ProcessFile(ctx, filePath); err != nil {
			lastError = err
			p.logger.Error("conversion failed", zap.String("source", item.Name()), zap.Error(err))
		}
	}
	if lastError != nil {
		return fmt.Errorf("last error: %w", lastError)
	}
	return nil
}

func (p *Processor) ProcessFile(ctx context.Context, filePath string) error {
	p.logger.Debug("processFile", zap.String("file", filePath), zap.String("inputDir", p.config.InputDir))

	allDec := common.GetDecoder(filePath, p.config.SkipNoop)
	if len(allDec) == 0 {
		p.hooks.OnFileEvent(FileEvent{Path: filePath, Status: StatusSkipped})
		return errors.New("skipping while no suitable decoder")
	}

	p.hooks.OnFileEvent(FileEvent{Path: filePath, Status: StatusValidating})

	if err := p.safeProcess(ctx, filePath, allDec); err != nil {
		p.hooks.OnFileEvent(FileEvent{Path: filePath, Status: StatusFailed, Error: err.Error()})
		return err
	}

	if p.config.RemoveSource {
		if err := os.RemoveAll(filePath); err != nil {
			return err
		}
		p.logger.Info("source file removed after success conversion", zap.String("source", filePath))
	}
	return nil
}

func (p *Processor) WatchDir(ctx context.Context, dir string) error {
	if err := p.ProcessDir(ctx, dir); err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
					f, err := os.OpenFile(event.Name, os.O_RDONLY, os.ModeExclusive)
					if err != nil {
						p.logger.Debug("failed to open file exclusively", zap.String("path", event.Name), zap.Error(err))
						time.Sleep(1 * time.Second)
						continue
					}
					_ = f.Close()

					if err := p.ProcessFile(ctx, event.Name); err != nil {
						p.logger.Warn("failed to process file", zap.String("path", event.Name), zap.Error(err))
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				p.logger.Error("file watcher got error", zap.Error(err))
			}
		}
	}()

	if err = watcher.Add(dir); err != nil {
		return fmt.Errorf("failed to watch dir %s: %w", dir, err)
	}

	<-ctx.Done()
	return nil
}

func (p *Processor) findDecoder(decoders []common.DecoderFactory, params *common.DecoderParams) (*common.Decoder, *common.DecoderFactory, error) {
	for _, factory := range decoders {
		dec := factory.Create(params)
		if err := dec.Validate(); err == nil {
			return &dec, &factory, nil
		} else {
			p.logger.Warn("try decode failed", zap.Error(err))
		}
	}
	return nil, nil, errors.New("no any decoder can resolve the file")
}

// safeProcess runs process and converts any panic into an error. The per-file
// decoders parse untrusted, potentially crafted input; a panic there (out-of-range
// slice, divide-by-zero, nil deref) would otherwise abort the whole CLI run and
// crash the GUI process. Recovering here downgrades it to a single failed file.
func (p *Processor) safeProcess(ctx context.Context, inputFile string, allDec []common.DecoderFactory) (err error) {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error("recovered from panic while processing file",
				zap.String("source", inputFile),
				zap.Any("panic", r),
				zap.Stack("stack"))
			err = fmt.Errorf("panic while processing %q: %v", inputFile, r)
		}
	}()
	return p.process(ctx, inputFile, allDec)
}

func (p *Processor) process(ctx context.Context, inputFile string, allDec []common.DecoderFactory) error {
	file, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer file.Close()
	logger := p.logger.With(zap.String("source", inputFile))

	pDec, decoderFactory, err := p.findDecoder(allDec, &common.DecoderParams{
		Reader:       file,
		Extension:    filepath.Ext(inputFile),
		FilePath:     inputFile,
		Logger:       logger,
		CryptoParams: p.config.Crypto,
	})
	if err != nil {
		return err
	}
	dec := *pDec

	p.hooks.OnFileEvent(FileEvent{Path: inputFile, Status: StatusDecrypting})

	params := &ffmpeg.UpdateMetadataParams{}

	// wrap decoder with progress reader
	fi, _ := file.Stat()
	var fileSize int64
	if fi != nil {
		fileSize = fi.Size()
	}
	progReader := newProgressReader(dec, inputFile, fileSize, p.hooks.OnProgress)

	header := bytes.NewBuffer(nil)
	_, err = io.CopyN(header, progReader, 64)
	if err != nil {
		return fmt.Errorf("read header failed: %w", err)
	}
	audio := io.MultiReader(header, progReader)
	params.AudioExt = sniff.AudioExtensionWithFallback(header.Bytes(), ".mp3")

	// DSDIFF (.dff) has no ffmpeg muxer, so metadata cannot be written; copy the
	// decrypted stream verbatim instead of failing the file.
	wantMeta := p.config.UpdateMetadata && ffmpeg.SupportsMetadata(params.AudioExt)
	if p.config.UpdateMetadata && !wantMeta {
		logger.Warn("metadata writing not supported for this format; copying without metadata",
			zap.String("ext", params.AudioExt))
	}

	if wantMeta {
		p.hooks.OnFileEvent(FileEvent{Path: inputFile, Status: StatusMetadata})

		if audioMetaGetter, ok := dec.(common.AudioMetaGetter); ok {
			metaCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			params.Audio, err = utils.WriteTempFile(audio, params.AudioExt)
			if err != nil {
				return fmt.Errorf("updateAudioMeta write temp file: %w", err)
			}
			defer os.Remove(params.Audio)

			params.Meta, err = audioMetaGetter.GetAudioMeta(metaCtx)
			if err != nil {
				logger.Warn("get audio meta failed", zap.Error(err))
			}

			if params.Meta == nil {
				// Reopen the temp file as the audio source. Close it on return so
				// the deferred os.Remove(params.Audio) above can succeed on Windows,
				// where removing a file with an open handle fails.
				tmpAudio, err := os.Open(params.Audio)
				if err != nil {
					return fmt.Errorf("updateAudioMeta open temp file: %w", err)
				}
				defer tmpAudio.Close()
				audio = tmpAudio
			}
		}
	}

	if wantMeta && params.Meta != nil {
		if coverGetter, ok := dec.(common.CoverImageGetter); ok {
			coverCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			if cover, err := coverGetter.GetCoverImage(coverCtx); err != nil {
				logger.Warn("get cover image failed", zap.Error(err))
			} else if imgExt, ok := sniff.ImageExtension(cover); !ok {
				logger.Warn("sniff cover image type failed", zap.Error(err))
			} else {
				params.AlbumArtExt = imgExt
				params.AlbumArt = cover
			}
		}
	}

	p.hooks.OnFileEvent(FileEvent{Path: inputFile, Status: StatusWriting})

	inputRelDir, err := filepath.Rel(p.config.InputDir, filepath.Dir(inputFile))
	if err != nil {
		return fmt.Errorf("get relative dir failed: %w", err)
	}

	inFilename := strings.TrimSuffix(filepath.Base(inputFile), decoderFactory.Suffix)
	outPath := filepath.Join(p.config.OutputDir, inputRelDir, inFilename+params.AudioExt)

	if !p.config.OverwriteOutput {
		_, err := os.Stat(outPath)
		if err == nil {
			logger.Warn("output file already exist, skip", zap.String("destination", outPath))
			p.hooks.OnFileEvent(FileEvent{Path: inputFile, Status: StatusSkipped})
			return nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat output file failed: %w", err)
		}
	}

	outDir := filepath.Dir(outPath)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("create output dir failed: %w", err)
	}

	if params.Meta == nil {
		outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, audio); err != nil {
			return err
		}
	} else {
		writeCtx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		if err := ffmpeg.UpdateMeta(writeCtx, outPath, params, logger); err != nil {
			return err
		}
	}

	logger.Info("successfully converted", zap.String("source", inputFile), zap.String("destination", outPath))
	p.hooks.OnFileEvent(FileEvent{
		Path:       inputFile,
		Status:     StatusDone,
		OutputPath: outPath,
		AudioExt:   params.AudioExt,
	})
	return nil
}
