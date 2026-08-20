//go:build windows

package main

import (
	"fmt"
	"os"
)

func spawnDaemon(self string, args []string, logFile *os.File) (int, error) {
	return 0, fmt.Errorf("modo daemon (-d) no soportado en Windows · corré en foreground: cockpit start")
}

func terminatePID(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
