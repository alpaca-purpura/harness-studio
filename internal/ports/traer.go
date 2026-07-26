package ports

import (
	"context"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// Materializador materializa un PlanTraer en un directorio de STAGING que el usecase le da
// (AG-D17, design.md §13.3). NO mueve al destino, NO registra nada, NO limpia: el usecase es
// dueño del ciclo de vida del temporal (BR-15) — así la atomicidad se testea en UN lugar y los
// dos adapters quedan tontos.
//
// `avisos` son problemas del DATO o de la copia que NO impiden traer y que se muestran (BR-8):
// un symlink que escapa del árbol y no se copió, un symlink roto, un reintento por `ref`. El
// contrato de design.md §13.3 no los devolvía; sin ellos los hallazgos de E-92 se perderían
// (C21 de design.md).
type Materializador interface {
	// Materializar deja el CONTENIDO FINAL del canónico en `staging` (ya con la subruta
	// extraída y con el set de exclusión aplicado). ctx cancelable: un fetch en vuelo se corta.
	// Devuelve el sha efectivo materializado ("" en camino local: no hay commit que reportar).
	Materializar(ctx context.Context, plan domain.PlanTraer, staging string) (shaEfectivo string, avisos []string, err error)
	Camino() domain.CaminoTraer
}
