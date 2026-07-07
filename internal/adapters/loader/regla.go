package loader

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// reconocerRegla emite el nodo `rule` de banda Base desde el CLAUDE.md de la raíz del arnés
// (§3, fila rule: la celda es CLAUDE.md en AMBAS formas físicas). Devuelve ok=false si el
// arnés no tiene CLAUDE.md — legal, cero nodos.
//
// Convención de identidad (definida aquí, HS-11 — el CLAUDE.md de un arnés DEBE cumplirla
// para que la regla tenga id estable en el grafo; el dogfood dev-full-cycle la cumple):
//  1. frontmatter YAML con `id:` (+ `nombre:` opcional, cae al id si falta); o si no,
//  2. primera línea heading `# <id> — <nombre>` (separador « — », em dash con espacios) —
//     p.ej. `# std-spec — estándar de spec` → id `std-spec`, nombre `estándar de spec`.
//
// Si el archivo existe pero ninguna de las dos aplica, el reconocedor NO lo entendió →
// nodo no-reconocido visible (§4.5), jamás descarte silencioso.
func reconocerRegla(dir string) (domain.Box, bool, error) {
	ruta := filepath.Join(dir, "CLAUDE.md")
	b, err := os.ReadFile(ruta) //nolint:gosec // G304: path bajo el dir del arnés que el caller eligió cargar (local-first).
	if errors.Is(err, os.ErrNotExist) {
		return domain.Box{}, false, nil
	}
	if err != nil {
		return domain.Box{}, false, fmt.Errorf("leer %s: %w", ruta, err)
	}

	fuente := rutaEstampada(dir, "CLAUDE.md")
	id, nombre := identidadRegla(b)
	if id == "" {
		return nodoNoReconocido("claude-md", fuente), true, nil
	}
	return domain.Box{
		ID:          id,
		Clase:       domain.ClaseRule,
		Nombre:      nombre,
		Banda:       domain.BandaBase,
		FuentePath:  fuente,
		Procedencia: domain.ProcDeclarado,
	}, true, nil
}

// identidadRegla aplica la convención de identidad del CLAUDE.md (ver reconocerRegla):
// frontmatter `id:`/`nombre:` primero; heading `# <id> — <nombre>` como segunda vía.
// Devuelve id vacío si ninguna aplica.
func identidadRegla(b []byte) (id, nombre string) {
	if fm, err := frontmatter(b); err == nil {
		if v, ok := fm["id"].(string); ok && v != "" {
			if n, ok := fm["nombre"].(string); ok && n != "" {
				return v, n
			}
			return v, v
		}
	}
	for _, l := range strings.Split(string(b), "\n") {
		resto, ok := strings.CutPrefix(strings.TrimSpace(l), "# ")
		if !ok {
			continue
		}
		// Solo el PRIMER heading decide: si no trae el separador « — », la convención
		// no se cumple y la identidad queda vacía (→ no-reconocido).
		izq, der, hay := strings.Cut(resto, " — ")
		if !hay || strings.TrimSpace(izq) == "" || strings.TrimSpace(der) == "" {
			return "", ""
		}
		return strings.TrimSpace(izq), strings.TrimSpace(der)
	}
	return "", ""
}
