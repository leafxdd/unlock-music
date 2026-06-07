package main

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newLogger builds the GUI logger. It writes to stdout (visible when running
// under `wails dev`) and tees Info-level and above to emit, which forwards each
// entry to the frontend log panel.
func newLogger(emit func(level, msg string)) *zap.Logger {
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encCfg.EncodeTime = zapcore.RFC3339TimeEncoder

	stdout := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encCfg),
		os.Stdout,
		zap.DebugLevel,
	)

	return zap.New(zapcore.NewTee(
		stdout,
		newEventLogCore(encCfg, zap.InfoLevel, emit),
	))
}

// eventLogCore is a zapcore.Core that forwards each entry's message (with its
// fields) to a sink, passing the level separately. It lets every existing
// processor zap log surface in the GUI log panel without touching call sites.
type eventLogCore struct {
	zapcore.LevelEnabler
	enc  zapcore.Encoder
	emit func(level, msg string)
}

func newEventLogCore(cfg zapcore.EncoderConfig, enab zapcore.LevelEnabler, emit func(level, msg string)) *eventLogCore {
	// The panel renders its own timestamp and receives the level separately, so
	// drop time/level/caller/name and keep just the message plus its fields.
	cfg.TimeKey = ""
	cfg.LevelKey = ""
	cfg.CallerKey = ""
	cfg.NameKey = ""
	return &eventLogCore{
		LevelEnabler: enab,
		enc:          zapcore.NewConsoleEncoder(cfg),
		emit:         emit,
	}
}

func (c *eventLogCore) With(fields []zapcore.Field) zapcore.Core {
	clone := &eventLogCore{
		LevelEnabler: c.LevelEnabler,
		enc:          c.enc.Clone(),
		emit:         c.emit,
	}
	for i := range fields {
		fields[i].AddTo(clone.enc)
	}
	return clone
}

func (c *eventLogCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *eventLogCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	buf, err := c.enc.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	msg := strings.TrimRight(buf.String(), "\n")
	buf.Free()
	c.emit(strings.ToUpper(ent.Level.String()), msg)
	return nil
}

func (c *eventLogCore) Sync() error { return nil }
