//go:build unix

package main

import "syscall"

// pidAlive: signal 0 = probe (EPERM = vivo de otro usuario · ESRCH = muerto).
func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	if err == nil {
		return true
	}
	return err == syscall.EPERM
}
