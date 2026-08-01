package ports

import (
	"context"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// marketplace.go son los puertos del ESTANTE (paquete 2026-07-23-portafolio-agregar-marketplace,
// design.md §3.4). Ningún nombre colisiona con los existentes (ArnesRegistry = arnés→cwd de
// sesión · PortafolioStore = identidades (home,id) · RepoConfigStore = repo del self-update).

// MarketplaceStore persiste el lado DECLARADO del registro (los que el operador registró por el
// wizard). Degrada honesto igual que PortafolioStore: Listar jamás falla por una fila corrupta.
type MarketplaceStore interface {
	Listar() ([]domain.MarketplaceConocido, []domain.EntradaCorrupta)
	Upsert(m domain.MarketplaceConocido) error
	Olvidar(nombre string) (bool, error)
}

// MarketplaceDetector expone el lado DETECTADO: `~/.claude/plugins/known_marketplaces.json`.
// READ-ONLY sobre un árbol ajeno. Metadata ilegible ⇒ (nil, error) — el usecase lo convierte
// en una discrepancia VISIBLE, nunca en una lista vacía silenciosa.
type MarketplaceDetector interface {
	Detectados() ([]domain.MarketplaceConocido, error)
}

// CatalogoReader lee el `marketplace.json` de UN marketplace. Hay dos implementaciones y el
// usecase decide el orden (política en el usecase, mecanismo en el adapter): local primero
// (checkout de CC, cero red, offline) y remoto como fallback (gh api). `Fuente` se anota en
// EstadoLectura.Fuente para trazabilidad (BR-3).
type CatalogoReader interface {
	Leer(ctx context.Context, m domain.MarketplaceConocido) (domain.Catalogo, error)
	Fuente() string // "local" | "remoto"
}

// CatalogoValidador prueba una URL ANTES de registrar (S6, BR-5/G3): busca
// `.claude-plugin/marketplace.json` legible y devuelve lo que LEYÓ. Nunca persiste nada.
// Es una interfaz aparte de CatalogoReader porque el insumo es una URL, no un
// MarketplaceConocido que todavía no existe.
type CatalogoValidador interface {
	Validar(ctx context.Context, url string) (domain.Catalogo, error)
}

// CatalogoSync sincroniza el checkout local de un marketplace PROPIO con su remoto
// (DD-2/E-bis, «↻ Refrescar» = fetch/pull real). El usecase lo invoca SOLO en refresco
// EXPLÍCITO y SOLO para clase propio — un marketplace de referencia JAMÁS se sincroniza
// (invariante 3 de marketplace-referencia-es-solo-procedencia: sobre lo ajeno solo se
// lee). Devuelve un detalle legible («ya al día» / «avanzó a <sha>») o error; el usecase
// arrastra el error a Lectura.Motivo y la lectura SIGUE — sincronizar es best-effort,
// jamás bloquea el catálogo.
type CatalogoSync interface {
	Sincronizar(ctx context.Context, m domain.MarketplaceConocido) (detalle string, err error)
}

// CatalogoCache guarda la última lectura EXITOSA por marketplace + su timestamp (AG-D8
// decisión 3, BR-3). Leer devuelve ok=false tanto si no hay caché como si el archivo está
// corrupto — con `motivo` poblado en el segundo caso, para que el degradado sea visible.
type CatalogoCache interface {
	Leer(nombre string) (cat domain.Catalogo, motivo string, ok bool)
	Guardar(nombre string, cat domain.Catalogo) error
	Olvidar(nombre string) error
}
