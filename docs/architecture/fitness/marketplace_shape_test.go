package fitness

// Enforcer del check `dominio-no-adopta-shape-ajeno` del boundary
// `marketplace-referencia-es-solo-procedencia` (L2 punto 4, ANTI-CORRUPTION LAYER).
//
// Es un SOURCE-SCAN y no un test de comportamiento a propósito: el invariante no es «qué hace el
// dominio» sino «qué FORMA tiene» — que no crezca un campo por cada extra del formato de Claude
// Code. La traducción del shape ajeno vive en `internal/adapters/marketplace/parse.go`, que es el
// único lugar del árbol autorizado a nombrar esas claves.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// clavesAjenas son las claves CRUDAS del `marketplace.json` / metadata de Claude Code que el
// dominio NO debe adoptar como campo. Lo NORMALIZADO sí es del modelo propio y no está acá:
// `TipoSource("git-subdir")` es un valor del enum propio y `install_location` es un campo con
// doc-comment que explica que es la ruta contra la que `deriva` compara (C20 de design.md).
var clavesAjenas = []string{
	"$schema", "lspServers", "displayName", "strict", "keywords",
	"category", "tags", "homepage", "renames", "metadata",
	"installLocation", "enabledPlugins", "projectPath",
}

// reCampoJSON captura el nombre de la clave de un tag `json:"..."`.
var reCampoJSON = regexp.MustCompile("json:\"([^\",]+)")

func TestDominioNoAdoptaShapeAjeno(t *testing.T) {
	root := repoRoot()
	patron := filepath.Join(root, "internal", "domain", "marketplace*.go")
	archivos, err := filepath.Glob(patron)
	if err != nil {
		t.Fatalf("glob %s: %v", patron, err)
	}
	if len(archivos) == 0 {
		t.Fatalf("no hay archivos de dominio de marketplace en %s (¿se renombraron?)", patron)
	}

	var malos []string
	for _, f := range archivos {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		b, rerr := os.ReadFile(f) //nolint:gosec // rutas del propio árbol del repo
		if rerr != nil {
			t.Fatalf("leer %s: %v", f, rerr)
		}
		rel, _ := filepath.Rel(root, f)
		for _, m := range reCampoJSON.FindAllStringSubmatch(string(b), -1) {
			campo := m[1]
			for _, ajena := range clavesAjenas {
				if campo == ajena {
					malos = append(malos, rel+": json:\""+campo+"\"")
				}
			}
		}
	}
	if len(malos) > 0 {
		t.Fatalf("el dominio adoptó %d clave(s) CRUDA(S) del formato ajeno (traducilas en internal/adapters/marketplace/parse.go):\n  %s",
			len(malos), strings.Join(malos, "\n  "))
	}
}
