package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"git.um-react.app/um/cli/algo/common"
	_ "git.um-react.app/um/cli/algo/kgm"
	_ "git.um-react.app/um/cli/algo/kwm"
	_ "git.um-react.app/um/cli/algo/ncm"
	"git.um-react.app/um/cli/algo/qmc"
	_ "git.um-react.app/um/cli/algo/tm"
	_ "git.um-react.app/um/cli/algo/xiami"
	_ "git.um-react.app/um/cli/algo/ximalaya"
	"git.um-react.app/um/cli/internal/processor"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var AppVersion = "custom"

var logger = setupLogger(false) // TODO: inject logger to application, instead of using global logger

func main() {
	module, ok := debug.ReadBuildInfo()
	if ok && module.Main.Version != "(devel)" {
		AppVersion = module.Main.Version
	}
	app := cli.App{
		Name:     "Unlock Music CLI",
		HelpName: "um",
		Usage:    "Unlock your encrypted music file https://git.um-react.app/um/cli",
		Version:  fmt.Sprintf("%s (%s,%s/%s)", AppVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH),
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "input", Aliases: []string{"i"}, Usage: "path to input file or dir", Required: false},
			&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "path to output dir", Required: false},
			&cli.StringFlag{Name: "qmc-mmkv", Aliases: []string{"db"}, Usage: "path to QQMusic mmkv path", Required: false},
			&cli.StringFlag{Name: "qmc-mmkv-key", Aliases: []string{"key"}, Usage: "QQMusic mmkv password (16 ascii chars)", Required: false},
			&cli.StringFlag{Name: "kgg-db", Usage: "path to kgg db (win32 kugou v11)", Required: false},
			&cli.BoolFlag{Name: "remove-source", Aliases: []string{"rs"}, Usage: "remove source file", Required: false, Value: false},
			&cli.BoolFlag{Name: "skip-noop", Aliases: []string{"n"}, Usage: "skip noop decoder", Required: false, Value: true},
			&cli.BoolFlag{Name: "verbose", Aliases: []string{"V"}, Usage: "verbose logging", Required: false, Value: false},
			&cli.BoolFlag{Name: "update-metadata", Usage: "update metadata & album art from network", Required: false, Value: false},
			&cli.BoolFlag{Name: "overwrite", Usage: "overwrite output file without asking", Required: false, Value: false},
			&cli.BoolFlag{Name: "watch", Usage: "watch the input dir and process new files", Required: false, Value: false},

			&cli.BoolFlag{Name: "supported-ext", Usage: "show supported file extensions and exit", Required: false, Value: false},
		},

		Action:          appMain,
		Copyright:       fmt.Sprintf("Copyright (c) 2020 - %d Unlock Music https://git.um-react.app/um/cli/src/branch/main/LICENSE", time.Now().Year()),
		HideHelpCommand: true,
		UsageText:       "um [-o /path/to/output/dir] [--extra-flags] [-i] /path/to/input",
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal("run app failed", zap.Error(err))
	}
}

func printSupportedExtensions() {
	var exts []string
	extSet := make(map[string]int)
	for _, factory := range common.DecoderRegistry {
		ext := strings.TrimPrefix(factory.Suffix, ".")
		if n, ok := extSet[ext]; ok {
			extSet[ext] = n + 1
		} else {
			extSet[ext] = 1
		}
	}
	for ext := range extSet {
		exts = append(exts, ext)
	}
	slices.Sort(exts)
	for _, ext := range exts {
		fmt.Printf("%s: %d\n", ext, extSet[ext])
	}
}

func setupLogger(verbose bool) *zap.Logger {
	logConfig := zap.NewProductionEncoderConfig()
	logConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	enabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		if verbose {
			return true
		}
		return level >= zapcore.InfoLevel
	})

	return zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(logConfig),
		os.Stdout,
		enabler,
	))
}

func appMain(c *cli.Context) (err error) {
	logger = setupLogger(c.Bool("verbose"))

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if c.Bool("supported-ext") {
		printSupportedExtensions()
		return nil
	}
	input := c.String("input")
	if input == "" {
		switch c.Args().Len() {
		case 0:
			input = cwd
		case 1:
			input = c.Args().Get(0)
		default:
			return errors.New("please specify input file (or directory)")
		}
	}

	input, absErr := filepath.Abs(input)
	if absErr != nil {
		return fmt.Errorf("get abs path failed: %w", absErr)
	}

	output := c.String("output")
	inputStat, err := os.Stat(input)
	if err != nil {
		return err
	}

	var inputDir string
	if inputStat.IsDir() {
		inputDir = input
	} else {
		inputDir = filepath.Dir(input)
	}
	inputDir, absErr = filepath.Abs(inputDir)
	if absErr != nil {
		return fmt.Errorf("get abs path (inputDir) failed: %w", absErr)
	}

	if output == "" {
		output = inputDir
	}
	logger.Debug("resolve input/output path", zap.String("inputDir", inputDir), zap.String("input", input), zap.String("output", output))

	outputStat, err := os.Stat(output)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(output, 0755)
		}
		if err != nil {
			return err
		}
	} else if !outputStat.IsDir() {
		return errors.New("output should be a writable directory")
	}

	qmcKeys, err := qmc.LoadMMKVOrDefault(c.String("qmc-mmkv"), c.String("qmc-mmkv-key"), logger)
	if err != nil {
		return err
	}

	kggDbPath := c.String("kgg-db")
	if kggDbPath == "" && runtime.GOOS == "windows" {
		// KGG (KGMv5) decoding reads keys from the Kugou SQLite DB, which only
		// exists on Windows; the default path is meaningless elsewhere.
		kggDbPath = filepath.Join(os.Getenv("APPDATA"), "Kugou8", "KGMusicV3.db")
	}

	proc := processor.New(processor.Config{
		InputDir:        inputDir,
		OutputDir:       output,
		SkipNoop:        c.Bool("skip-noop"),
		RemoveSource:    c.Bool("remove-source"),
		UpdateMetadata:  c.Bool("update-metadata"),
		OverwriteOutput: c.Bool("overwrite"),
		Crypto: common.CryptoParams{
			KggDbPath: kggDbPath,
			QmcKeys:   qmcKeys,
		},
	}, logger, processor.Hooks{})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if inputStat.IsDir() {
		if c.Bool("watch") {
			return proc.WatchDir(ctx, input)
		}
		return proc.ProcessDir(ctx, input)
	}
	return proc.ProcessFile(ctx, input)
}
