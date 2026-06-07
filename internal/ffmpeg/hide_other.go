//go:build !windows

package ffmpeg

import "os/exec"

func hideWindow(_ *exec.Cmd) {}
