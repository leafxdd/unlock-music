package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"git.um-react.app/um/cli/algo/common"
	"git.um-react.app/um/cli/algo/qmc"
	"git.um-react.app/um/cli/internal/processor"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type App struct {
	ctx      context.Context
	logger   *zap.Logger
	settings Settings

	mu     sync.Mutex
	cancel context.CancelFunc
}

func NewApp() *App {
	logConfig := zap.NewProductionEncoderConfig()
	logConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	logConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	l := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(logConfig),
		os.Stdout,
		zap.DebugLevel,
	))
	return &App{
		logger:   l,
		settings: loadSettings(),
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) CheckFFmpeg() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

func (a *App) GetSettings() Settings {
	return a.settings
}

func (a *App) SaveSettings(s Settings) error {
	a.settings = s
	return saveSettings(s)
}

func (a *App) SelectInputDir() (string, error) {
	return wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择输入目录",
	})
}

func (a *App) SelectOutputDir() (string, error) {
	return wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择输出目录",
	})
}

func (a *App) SelectInputFiles() ([]string, error) {
	return wailsRuntime.OpenMultipleFilesDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择加密音乐文件",
	})
}

func (a *App) IsProcessing() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cancel != nil
}

func (a *App) StartProcessing(inputPath string) error {
	a.mu.Lock()
	if a.cancel != nil {
		a.mu.Unlock()
		return nil
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	a.mu.Unlock()

	go func() {
		defer func() {
			a.mu.Lock()
			a.cancel = nil
			a.mu.Unlock()
			wailsRuntime.EventsEmit(a.ctx, "processing:done")
		}()

		if err := a.runProcessor(ctx, inputPath); err != nil {
			a.logger.Error("processing failed", zap.Error(err))
			wailsRuntime.EventsEmit(a.ctx, "processing:error", err.Error())
		}
	}()
	return nil
}

func (a *App) StartProcessingBatch(paths []string) error {
	a.mu.Lock()
	if a.cancel != nil {
		a.mu.Unlock()
		return nil
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	a.mu.Unlock()

	go func() {
		defer func() {
			a.mu.Lock()
			a.cancel = nil
			a.mu.Unlock()
			wailsRuntime.EventsEmit(a.ctx, "processing:done")
		}()

		s := a.settings

		qmcKeys, err := qmc.LoadMMKVOrDefault(s.QmcMMKVPath, s.QmcMMKVKey, a.logger)
		if err != nil {
			a.logger.Warn("load QMC keys failed, continuing without keys", zap.Error(err))
			qmcKeys = nil
		}

		kggDbPath := s.KggDbPath
		if kggDbPath == "" {
			kggDbPath = filepath.Join(os.Getenv("APPDATA"), "Kugou8", "KGMusicV3.db")
		}

		for _, p := range paths {
			if ctx.Err() != nil {
				return
			}
			if err := a.runProcessorWithCrypto(ctx, p, qmcKeys, kggDbPath); err != nil {
				a.logger.Error("processing failed", zap.String("path", p), zap.Error(err))
				wailsRuntime.EventsEmit(a.ctx, "processing:error", err.Error())
			}
		}
	}()
	return nil
}

func (a *App) StopProcessing() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
	}
}

func (a *App) ListFiles(paths []string) ([]string, error) {
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			if len(common.GetDecoder(p, false)) > 0 {
				files = append(files, p)
			}
			continue
		}
		filepath.WalkDir(p, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if len(common.GetDecoder(path, false)) > 0 {
				files = append(files, path)
			}
			return nil
		})
	}
	return files, nil
}

func (a *App) runProcessor(ctx context.Context, inputPath string) error {
	s := a.settings

	qmcKeys, err := qmc.LoadMMKVOrDefault(s.QmcMMKVPath, s.QmcMMKVKey, a.logger)
	if err != nil {
		a.logger.Warn("load QMC keys failed, continuing without keys", zap.Error(err))
		qmcKeys = nil
	}

	kggDbPath := s.KggDbPath
	if kggDbPath == "" {
		kggDbPath = filepath.Join(os.Getenv("APPDATA"), "Kugou8", "KGMusicV3.db")
	}

	return a.runProcessorWithCrypto(ctx, inputPath, qmcKeys, kggDbPath)
}

func (a *App) runProcessorWithCrypto(ctx context.Context, inputPath string, qmcKeys common.QMCKeys, kggDbPath string) error {
	s := a.settings

	inputStat, err := os.Stat(inputPath)
	if err != nil {
		return err
	}

	outputDir := s.OutputDir
	if outputDir == "" {
		if inputStat.IsDir() {
			outputDir = inputPath
		} else {
			outputDir = filepath.Dir(inputPath)
		}
	}

	inputDir := inputPath
	if !inputStat.IsDir() {
		inputDir = filepath.Dir(inputPath)
	}

	hooks := processor.Hooks{
		OnFileEvent: func(e processor.FileEvent) {
			wailsRuntime.EventsEmit(a.ctx, "file:event", e)
		},
		OnProgress: func(e processor.ProgressEvent) {
			wailsRuntime.EventsEmit(a.ctx, "file:progress", e)
		},
		OnLog: func(level, msg string) {
			wailsRuntime.EventsEmit(a.ctx, "log", map[string]string{"level": level, "msg": msg})
		},
	}

	proc := processor.New(processor.Config{
		InputDir:        inputDir,
		OutputDir:       outputDir,
		SkipNoop:        s.SkipNoop,
		RemoveSource:    s.RemoveSource,
		UpdateMetadata:  s.UpdateMetadata,
		OverwriteOutput: s.OverwriteOutput,
		Crypto: common.CryptoParams{
			KggDbPath: kggDbPath,
			QmcKeys:   qmcKeys,
		},
	}, a.logger, hooks)

	if inputStat.IsDir() {
		return proc.ProcessDir(ctx, inputPath)
	}
	return proc.ProcessFile(ctx, inputPath)
}
