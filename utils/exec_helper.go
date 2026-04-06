package utils

import "os/exec"

func newExecCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	configureCommandForPlatform(cmd)
	return cmd
}
