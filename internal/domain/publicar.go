package domain

import (
	"errors"
	"path/filepath"
)

// publicar.go es el contrato PURO de `▲ Publicar` (B2 del paquete
// 2026-07-30-volverlo-de-arnesia-y-publicar, B-D5 FIRMADA 🧑‍⚖️): qué se le pide al publisher y
// qué devuelve, más los centinelas que el transporte clasifica por errors.Is. Sin un solo
// `os.`/`exec.` — el mecanismo git vive en `internal/adapters/publish`, la política (guardas +
// gate de conformance) en `internal/usecase/publicar.go`; ninguno de los dos puede importar al
// otro y por eso los centinelas viven acá (mismo criterio que los de Traer, C23 de design.md).

// SolicitudPublicacion es lo que el usecase le pide al PublishPort: publicar el árbol del
// canónico (OrigenDir) como la versión Version del arnés ID en el marketplace-home RepoHome
// (canonicalizado host/owner/repo, RN-IDENT-1).
type SolicitudPublicacion struct {
	RepoHome  string // repo del marketplace destino, canonicalizado (github.com/owner/repo).
	ID        string // id del arnés (la fila `name` del marketplace.json y el dir plugins/<id>/).
	Version   string // X.Y.Z del autor (SoT plugin.json.version, conventions/versionado-arnes.md).
	OrigenDir string // path del canónico a copiar (excluyendo .git/.in_use/.orphaned_at).
}

// ResultadoPublicacion es lo que el publisher reporta tras el push. Tag puede venir vacío con
// un aviso «publicado sin tag: …» (B-D5: fallo del tag = éxito parcial VISIBLE, no rollback
// fantasma — el branch ya está pusheado y deshacerlo mentiría).
type ResultadoPublicacion struct {
	Commit string   `json:"commit"`
	Tag    string   `json:"tag,omitempty"`
	Avisos []string `json:"avisos,omitempty"`
}

// Centinelas de Publicar. El handler HTTP los mapea por errors.Is (espejo de la tabla de
// Traer, §13.8): 400 = precondición del pedido · 409 = conflicto reintentable/no-se-pisa ·
// 503 = sin credencial · 500 = fallo nuestro.
var (
	// ErrPublicarSinAuth: sin `gh` autenticado ni PAT para pushear al repo → 503. ArnesIA no
	// pide, guarda ni proxya credenciales (BR-18, mismo contrato que Traer).
	ErrPublicarSinAuth = errors.New("publicar: sin credencial para escribir en el marketplace (gh no autenticado y sin PAT)")
	// ErrPublicarPushRechazado: el remoto rechazó el push (non-fast-forward) — alguien publicó
	// antes → 409 «reintentá». JAMÁS se resuelve con --force (innegociable del spec).
	ErrPublicarPushRechazado = errors.New("publicar: el push fue rechazado — alguien publicó antes; reintentá")
	// ErrPublicarVersionYaPublicada: `plugins/<id>/<version>/` ya está poblado en el
	// marketplace → 409 y NADA se toca (una versión publicada es inmutable).
	ErrPublicarVersionYaPublicada = errors.New("publicar: esa versión ya está publicada en el marketplace (una versión publicada no se pisa)")
	// ErrPublicarVersionInvalida: la versión del canónico no es semver X.Y.Z → 400. La SoT es
	// plugin.json.version (conventions/versionado-arnes.md) — no se fabrica una.
	ErrPublicarVersionInvalida = errors.New("publicar: la versión del canónico no es semver X.Y.Z (SoT: plugin.json.version)")
	// ErrPublicarSinCanonico: la entrada no tiene canónico → 400. Solo se publica la única
	// copia editable — jamás una instalación read-only (ley anti-drift).
	ErrPublicarSinCanonico = errors.New("publicar: la entrada no tiene canónico — no hay copia editable que publicar")
	// ErrPublicarSinHome: la identidad es provisional (sin home declarado) → 400. Sin
	// marketplace autor no hay estante destino; se resuelve con «Resolver origen» (S7).
	ErrPublicarSinHome = errors.New("publicar: la identidad no declara home — resolvé el origen del arnés primero")
	// ErrPublicarNoPropio: el marketplace del home no es de clase `propio` → 400. Es BR-1 del
	// lado write: jamás se escribe en el estante de un tercero.
	ErrPublicarNoPropio = errors.New("publicar: el marketplace del home no es propio — solo se publica en marketplaces propios")
	// ErrPublicarConformanceRojo: el gate `conformance` del canónico tiene fails bloqueantes →
	// 409 con el reporte adjunto. Innegociable del spec: no se publica un arnés en rojo.
	ErrPublicarConformanceRojo = errors.New("publicar: el gate de conformance está en rojo — no se publica un arnés que no pasa su contrato")
)

// RaizPublicaciones es el único territorio donde Publicar clona y escribe localmente:
// `<raizArnesia>/publicaciones` (B-D5: checkouts/ son canónicos de arneses, no se mezclan).
// Espejo de RaizCheckouts (traer.go).
func RaizPublicaciones(raizArnesia string) string {
	return filepath.Join(raizArnesia, "publicaciones")
}

// SlugRepo convierte un repo canonicalizado (host/owner/repo) en UN segmento de path seguro
// para el workdir de publicaciones — misma regla que los segmentos de RutaCanonico
// (segmentoDeDestino): slug legible + sufijo de huella cuando el slug pierde información, así
// dos repos distintos que colapsan al mismo slug jamás comparten workdir. ok=false solo con un
// repo vacío.
func SlugRepo(repo string) (string, bool) {
	return segmentoDeDestino(repo)
}
