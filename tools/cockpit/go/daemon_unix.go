//go:build unix

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func spawnDaemon(self string, args []string, logFile *os.File) (int, error) {
	child := exec.Command(self, args...)
	child.Stdout = logFile
	child.Stderr = logFile
	child.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := child.Start(); err != nil {
		return 0, err
	}
	return child.Process.Pid, nil
}

func terminatePID(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}
