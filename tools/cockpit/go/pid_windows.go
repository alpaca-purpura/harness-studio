//go:build windows

package main

import "os"

// En Windows no hay signal-0 probe confiable; FindProcess + intento de probe.
func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	_, err := os.FindProcess(pid)
	return err == nil
}
