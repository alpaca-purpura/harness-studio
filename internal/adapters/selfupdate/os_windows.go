//go:build windows

package selfupdate

// os_windows.go — la mitad Windows de la mecánica por-OS del self-update (paquete
// 2026-08-13-compilacion-windows). Diferencias REALES del OS, no cosméticas: NTFS no
// tiene bit de ejecución (el veredicto es la extensión), la resolución en PATH pasa
// por PATHEXT, el bundle corre con python (el `bash` del PATH de sistema suele ser el
// relay roto de WSL) y no existe el kill-de-grupo de unix.

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// nombreBinNuevo — nombre del binario que el bundle deja en <repo>/bin/.
func nombreBinNuevo() string { return "arnesia.exe" }

// scriptDeBundle — en Windows el bundle es scripts/bundle.py (portable, mismo sello
// RF-231 que bundle.sh); bundle.sh exigiría un bash confiable que el PATH de sistema
// no garantiza.
func scriptDeBundle() string { return filepath.Join("scripts", "bundle.py") }

// descripcionBundle — cómo se nombra el paso de build en los detalles humanos.
const descripcionBundle = "scripts/bundle.py --solo-daemon"

// toolchainRequerida — herramientas que Verificar exige resolubles en PATH.
func toolchainRequerida() []string { return []string{"go", "pnpm", "python"} }

// esEjecutable — NTFS no tiene bit de ejecución (os.Stat reporta 0666 para cualquier
// archivo): el veredicto honesto es la extensión del binario.
func esEjecutable(_ fs.FileMode, path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".exe")
}

// ejecutableEnDir — semántica exec.LookPath de Windows: el tool pelado se resuelve
// probando las extensiones de PATHEXT (go.exe, pnpm.cmd, python.exe…), sin chequear
// bits de modo (no existen en NTFS).
func ejecutableEnDir(dir, tool string) (string, bool) {
	exts := []string{".exe", ".cmd", ".bat", ".com"}
	if pe := os.Getenv("PATHEXT"); pe != "" {
		exts = exts[:0]
		for _, e := range strings.Split(pe, ";") {
			if e = strings.TrimSpace(e); e != "" {
				exts = append(exts, strings.ToLower(e))
			}
		}
	}
	if filepath.Ext(tool) != "" {
		exts = append([]string{""}, exts...) // ya trae extensión: probarlo tal cual primero
	}
	for _, ext := range exts {
		candidato := filepath.Join(dir, tool+ext)
		if st, err := os.Stat(candidato); err == nil && !st.IsDir() {
			return candidato, true
		}
	}
	return "", false
}

// candidatosPATHOS — ubicaciones estándar Windows que un proceso lanzado desde el
// Explorador/acceso directo puede no heredar: go install (USERPROFILE\go\bin), el
// instalador oficial de Go (Program Files\Go\bin), los shims globales de npm
// (APPDATA\npm — ahí vive pnpm.cmd) y pnpm standalone (LOCALAPPDATA\pnpm).
// NO se agrega WindowsApps: su python3.exe es un stub falso que abre la Store.
func candidatosPATHOS(home string) []string {
	cands := []string{
		filepath.Join(home, "go", "bin"),
		filepath.Join(`C:\`, "Program Files", "Go", "bin"),
	}
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		cands = append(cands, filepath.Join(appdata, "npm"))
	}
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		cands = append(cands, filepath.Join(local, "pnpm"))
	}
	return cands
}

// comandoBundle arma el subproceso del paso ② (RF-104): python scripts/bundle.py
// --solo-daemon con cwd=repo. Sin grupos de proceso: la cancelación de CommandContext
// mata al python y WaitDelay corta el pipe si un hijo (go build) sobrevive —
// degradación aceptada frente al kill-de-grupo de unix, documentada acá.
func comandoBundle(ctx context.Context, repo string) *exec.Cmd {
	python := "python"
	if p, err := lookPathEn("python", pathAumentado()); err == nil {
		python = p
	}
	//nolint:gosec // G204: es EL diseño (RF-104 ②): correr el bundle del repo CONFIGURADO
	// al daemon (flag/env, jamás del request — RF-106); Verificar ya ancló el árbol al módulo esperado.
	cmd := exec.CommandContext(ctx, python, scriptDeBundle(), "--solo-daemon")
	cmd.Dir = repo
	cmd.WaitDelay = 2 * time.Second
	return cmd
}
