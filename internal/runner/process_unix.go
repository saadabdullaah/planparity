//go:build !windows

package runner

import (
	"os/exec"
	"syscall"
)

type processTree struct{}

func configureProcess(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func attachProcessTree(*exec.Cmd) (*processTree, error) { return &processTree{}, nil }

func terminateProcessTree(command *exec.Cmd, _ *processTree) error {
	if command.Process == nil {
		return nil
	}
	return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
}

func closeProcessTree(*processTree) {}
