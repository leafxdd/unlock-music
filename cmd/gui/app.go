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
)

type App struct {
	ctx      context.Context
	logger   *zap.Logger
	settings Settings

	mu     sync.Mutex
	cancel context.CancelFunc
}

func NewApp() *App {
	a := &App{
		settings: loadSettings(),
	}
	// The logger tees to stdout (visible under `wails dev`) and to the frontend
	// log panel via a.emitLog.
	a.logger = newLogger(a.emitLog)
	return a
}

// emitLog forwards a log entry to the frontend log panel. It is a no-op until
// the Wails runtime context is ready (set in Startup) -- always the case by the
// time any processing runs.
func (a *App) emitLog(level, msg string) {
	if a.ctx == nil {
		return
	}
	wailsRuntime.EventsEmit(a.ctx, "log", map[string]string{"level": level, "msg": msg})
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

// DropTarget describes a dropped path so the frontend can persist the input
// directory when a folder is dropped.
type DropTarget struct {
	Dir   string `json:"dir"`   // dir to remember: the path itself if a dir, else its parent
	IsDir bool   `json:"isDir"` // whether the dropped path is a directory
}

// ResolveDrop classifies the first dropped path via os.Stat, so the frontend
// can reliably tell a folder drop from a file drop.
func (a *App) ResolveDrop(paths []string) DropTarget {
	if len(paths) == 0 {
		return DropTarget{}
	}
	p := paths[0]
	if info, err := os.Stat(p); err == nil && info.IsDir() {
		return DropTarget{Dir: p, IsDir: true}
	}
	return DropTarget{Dir: filepath.Dir(p), IsDir: false}
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
		// Logs reach the panel via the teed logger (see emitLog), not this hook.
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
