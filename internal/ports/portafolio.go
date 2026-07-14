package ports

import (
	"context"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// ArnesLoader es el puerto que PortafolioService usa para cargar un directorio a su
// grafo L0 — satisfecho por un wrapper sobre loader.LoadArnes (mismo patrón que
// loadArnesDir en cmd/arnesia/main.go). Un adapter aparte no importa a otro (go-arch-lint):
// el usecase orquesta cargar→resolver a través de puertos, no el scanner directamente.
type ArnesLoader interface {
	Load(dir string) (domain.Graph, error)
}

// PortafolioScanner es el puerto del walker físico READ-ONLY (satisfecho por
// portafolio.Scanner, D-DOM-3/6).
type PortafolioScanner interface {
	Escanear(ctx context.Context, root string) ([]domain.HallazgoInstalacion, error)
}

// DerivaEvaluator es el puerto del evaluador de deriva (satisfecho por
// `*portafolio.Referencias`, BR-4/S0-D7): sin version o sin referencia local accesible →
// domain.DerivaNoEvaluable con motivo, jamás semver-string ni `git status`.
type DerivaEvaluator interface {
	Evaluar(installDir, home, id, version string) (domain.EstadoDeriva, string)
}

// PortafolioStore es el puerto de persistencia del registro del Portafolio (satisfecho
// por portafolio.Store, S0-D5): degrada honesto — Listar jamás falla por una entrada
// corrupta, la reporta aparte (BR-11/C-N-4).
type PortafolioStore interface {
	Listar() ([]domain.EntradaPortafolio, []domain.EntradaCorrupta)
	Upsert(e domain.EntradaPortafolio) error
	Desvincular(clave string) (bool, error)
	// Checkouts devuelve los paths de canónicos conocidos (RN-IDENT-4): un path detectado
	// que ES un checkout conocido se clasifica CANÓNICO, jamás instalación.
	Checkouts() []string
}
