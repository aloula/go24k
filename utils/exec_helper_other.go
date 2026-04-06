//go:build !windows

package utils

import "os/exec"

func configureCommandForPlatform(cmd *exec.Cmd) {}
