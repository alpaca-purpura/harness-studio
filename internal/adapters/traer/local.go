package traer

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// CopiadorLocal es el camino A (§13.6): copia la subcarpeta del checkout que Claude Code ya
// mantiene. **Cero red, cero auth, cero `git`** — verificable por inspección de imports: este
// archivo no importa `net/http` ni `os/exec`.
type CopiadorLocal struct{}

var _ ports.Materializador = (*CopiadorLocal)(nil)

// Camino identifica la vía.
func (c *CopiadorLocal) Camino() domain.CaminoTraer { return domain.CaminoLocal }

// Materializar deja el contenido final del canónico en `staging`. El sha efectivo es "" a
// propósito: en el camino local NO hay commit que reportar (el catálogo declara una ruta, no un
// pin), y fabricar uno sería inventar procedencia.
func (c *CopiadorLocal) Materializar(_ context.Context, plan domain.PlanTraer, staging string) (string, []string, error) {
	if plan.Camino != domain.CaminoLocal {
		return "", nil, fmt.Errorf("traer local: el plan pide el camino %q", plan.Camino)
	}
	if plan.OrigenLocal == "" {
		return "", nil, errors.New("traer local: el plan no declara origen")
	}
	if fi, err := os.Stat(plan.OrigenLocal); err != nil || !fi.IsDir() {
		return "", nil, fmt.Errorf("traer local: el origen %s no existe o no es un directorio", plan.OrigenLocal)
	}
	avisos, err := CopiarArbol(plan.OrigenLocal, staging)
	if err != nil {
		return "", avisos, err
	}
	return "", avisos, nil
}
