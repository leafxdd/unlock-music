package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Settings struct {
	InputDir        string `json:"inputDir"`
	OutputDir       string `json:"outputDir"`
	SkipNoop        bool   `json:"skipNoop"`
	RemoveSource    bool   `json:"removeSource"`
	UpdateMetadata  bool   `json:"updateMetadata"`
	OverwriteOutput bool   `json:"overwriteOutput"`
	QmcMMKVPath     string `json:"qmcMmkvPath"`
	QmcMMKVKey      string `json:"qmcMmkvKey"`
	KggDbPath       string `json:"kggDbPath"`
}

var (
	settingsMu   sync.Mutex
	settingsPath string
)

func defaultSettings() Settings {
	return Settings{
		SkipNoop: true,
	}
}

func settingsFilePath() string {
	settingsMu.Lock()
	defer settingsMu.Unlock()
	if settingsPath != "" {
		return settingsPath
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	settingsPath = filepath.Join(dir, "unlock-music-gui", "settings.json")
	return settingsPath
}

func loadSettings() Settings {
	s := defaultSettings()
	data, err := os.ReadFile(settingsFilePath())
	if err != nil {
		return s
	}
	_ = json.Unmarshal(data, &s)
	return s
}

func saveSettings(s Settings) error {
	p := settingsFilePath()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}
