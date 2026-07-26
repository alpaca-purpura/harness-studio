package marketplace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// DetectorCC lee el lado DETECTADO del registro: `~/.claude/plugins/known_marketplaces.json`
// (AG-D9, eslabón `cc-known-marketplaces`). READ-ONLY sobre un árbol AJENO — se lee, nunca se
// escribe (boundary `marketplace-referencia-es-solo-procedencia` L2 punto 3).
type DetectorCC struct {
	// CCPluginsDir default ~/.claude/plugins — INYECTABLE para tests, igual que
	// portafolio.Scanner.CCPluginsDir.
	CCPluginsDir string
}

var _ ports.MarketplaceDetector = (*DetectorCC)(nil)

func (d *DetectorCC) ccPluginsDir() string {
	if d.CCPluginsDir != "" {
		return d.CCPluginsDir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "plugins")
}

// ccKnownMarketplace es el shape ajeno de UNA entrada de `known_marketplaces.json`.
//
// DEUDA CONSCIENTE (C15 de design.md): `internal/adapters/portafolio/scanner.go` ya tiene un
// decoder equivalente (`leerKnownMarketplaces`, privado) y `deriva.go` lo reusa — pero
// go-arch-lint prohíbe adapter→adapter (allow-list, default-deny), así que las ~20 líneas se
// duplican en vez de (a) extraer un componente al grafo cuyo único contenido sería un DTO o
// (b) tocar la firma FIJA de `ports.DerivaEvaluator`. Entrada propuesta en `BACKLOG.md`:
// «unificar el decoder de metadata CC detrás de un puerto».
type ccKnownMarketplace struct {
	Source struct {
		Source string `json:"source"`
		Repo   string `json:"repo"`
		URL    string `json:"url"`
	} `json:"source"`
	InstallLocation string `json:"installLocation"`
	LastUpdated     string `json:"lastUpdated"`
}

// Detectados devuelve una fila por marketplace que Claude Code conoce, con el repo ya
// canonicalizado (RN-IDENT-1) y su `installLocation` — la ruta contra la que `deriva` compara.
// Metadata ausente ⇒ (nil, nil): «CC no conoce ninguno» es una afirmación verdadera. Metadata
// ILEGIBLE ⇒ (nil, error): el usecase lo convierte en `aviso_detector` VISIBLE (E-68), nunca en
// una lista vacía silenciosa.
func (d *DetectorCC) Detectados() ([]domain.MarketplaceConocido, error) {
	ruta := filepath.Join(d.ccPluginsDir(), "known_marketplaces.json")
	b, err := os.ReadFile(ruta) //nolint:gosec // G304: ruta fija de metadata CC, inyectable en tests.
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("marketplace detector: leer %s: %w", ruta, err)
	}
	var crudo map[string]ccKnownMarketplace
	if uerr := json.Unmarshal(b, &crudo); uerr != nil {
		return nil, fmt.Errorf("marketplace detector: %s ilegible: %w", ruta, uerr)
	}

	nombres := make([]string, 0, len(crudo))
	for n := range crudo {
		nombres = append(nombres, n)
	}
	sort.Strings(nombres) // orden estable: el map de Go no lo garantiza.

	out := make([]domain.MarketplaceConocido, 0, len(nombres))
	for _, n := range nombres {
		m := crudo[n]
		fila := domain.MarketplaceConocido{
			Nombre:          n,
			Clase:           domain.ClaseReferencia, // CC no modela clase: fail-safe (regla 3 del merge).
			Eslabones:       []domain.EslabonMarketplace{domain.EslabonCCKnown},
			InstallLocation: m.InstallLocation,
			CCActualizado:   m.LastUpdated,
		}
		crudoRepo := primerNoVacio(m.Source.Repo, m.Source.URL)
		if canon, ok := domain.CanonicalizarRepo(crudoRepo); ok {
			fila.Repo = canon
		} else if crudoRepo != "" {
			// No canonicaliza: el crudo NO se descarta, queda visible como discrepancia (S1-D3).
			fila.Discrepancias = append(fila.Discrepancias,
				fmt.Sprintf("Claude Code declara un repo que no resuelve a host/owner/repo: %q", crudoRepo))
		}
		out = append(out, fila)
	}
	return out, nil
}
