package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// seamSistemaSlugs lee la lista de slugs del seam aceptando AMBAS keys (I-52 shim):
// la nueva `sistemas:` y la legacy `brands:` (seams pre-rename). El loader python hace
// lo mismo vía _SEAM_ALIASES — un seam sin migrar sigue cargando para siempre.
func TestSeamSistemaSlugsDualRead(t *testing.T) {
	for _, key := range []string{"sistemas", "brands"} {
		t.Run("key "+key, func(t *testing.T) {
			root := t.TempDir()
			seam := fmt.Sprintf("meta:\n  product: acme\n%s:\n  active:\n    - slug: uno\n    - slug: dos\n", key)
			if err := os.WriteFile(filepath.Join(root, "project.config.yaml"), []byte(seam), 0o644); err != nil {
				t.Fatal(err)
			}
			got := seamSistemaSlugs(root)
			want := []string{"uno", "dos"}
			if len(got) != len(want) {
				t.Fatalf("%s: esperaba %v, obtuve %v", key, want, got)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("%s: slug[%d]=%q, esperaba %q", key, i, got[i], want[i])
				}
			}
		})
	}
}
