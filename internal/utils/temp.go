package utils

import (
	"fmt"
	"io"
	"os"
)

func WriteTempFile(rd io.Reader, ext string) (string, error) {
	audioFile, err := os.CreateTemp("", "*"+ext)
	if err != nil {
		return "", fmt.Errorf("ffmpeg create temp file: %w", err)
	}
	name := audioFile.Name()

	if _, err := io.Copy(audioFile, rd); err != nil {
		_ = audioFile.Close()
		_ = os.Remove(name)
		return "", fmt.Errorf("ffmpeg write temp file: %w", err)
	}

	if err := audioFile.Close(); err != nil {
		_ = os.Remove(name)
		return "", fmt.Errorf("ffmpeg close temp file: %w", err)
	}

	return name, nil
}
