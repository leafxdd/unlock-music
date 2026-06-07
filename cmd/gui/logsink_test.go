package main

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestNewLoggerForwardsInfoAndAbove(t *testing.T) {
	type rec struct{ level, msg string }
	var got []rec
	logger := newLogger(func(level, msg string) {
		got = append(got, rec{level, msg})
	})

	logger.Debug("debug line")                            // below Info -> dropped
	logger.Info("converted", zap.String("dst", "x.flac")) // forwarded
	logger.Warn("skipped")                                // forwarded

	if len(got) != 2 {
		t.Fatalf("forwarded %d entries, want 2: %+v", len(got), got)
	}
	if got[0].level != "INFO" {
		t.Errorf("entry0 level = %q, want INFO", got[0].level)
	}
	if !strings.Contains(got[0].msg, "converted") || !strings.Contains(got[0].msg, "x.flac") {
		t.Errorf("entry0 msg = %q, want message + field", got[0].msg)
	}
	if got[1].level != "WARN" {
		t.Errorf("entry1 level = %q, want WARN", got[1].level)
	}
}
