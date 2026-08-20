package main

import (
	"os"
	"path/filepath"
	"testing"
)

// mkBoard crea {dir}/{sistema}/docs/product (un board navegable) y devuelve dir.
// Compartido por los tests de ledger/navconfig/workspace-externo (vivía en
// portfolio_test.go hasta que CK-02 Stage 2 movió ese archivo a la célula cockpit).
func mkBoard(t *testing.T, sistema string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, sistema, "docs", "product"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}
