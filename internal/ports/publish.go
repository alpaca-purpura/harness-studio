package ports

import (
	"context"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// PublishPort publica un arnés en su marketplace-home (write-side prenter-marketplace, B2 del
// paquete 2026-07-30-volverlo-de-arnesia-y-publicar). El adapter (internal/adapters/publish) es
// MECANISMO puro: clone/pull del repo del marketplace, copia del árbol, read-modify-write de
// marketplace.json/catalogo.json, commit+push sin force y tag tras el push. La POLÍTICA (guardas
// de canónico/home/clase/semver + el gate de conformance) vive en el usecase — el adapter asume
// una solicitud ya sancionada.
type PublishPort interface {
	// Publicar materializa sol en el marketplace remoto. Errores por centinela del dominio:
	// ErrPublicarVersionYaPublicada (idempotencia, nada se toca), ErrPublicarPushRechazado
	// (non-fast-forward: reintentá), ErrPublicarSinAuth. Un fallo del tag NO es error: viaja
	// como aviso en ResultadoPublicacion.Avisos (B-D5).
	Publicar(ctx context.Context, sol domain.SolicitudPublicacion) (domain.ResultadoPublicacion, error)
}
