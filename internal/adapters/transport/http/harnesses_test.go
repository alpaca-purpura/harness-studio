package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// indiceConEntradas es un ports.IndexPort mínimo que devuelve un set fijo de entradas —
// lo único que listHarnesses consume. Query/Upsert/Rebuild no se llaman en este test.
type indiceConEntradas []ports.EntradaIndice

func (i indiceConEntradas) Rebuild(context.Context) error { return nil }

func (i indiceConEntradas) Query(context.Context, string) (domain.Graph, error) {
	return domain.Graph{}, errors.New("indiceConEntradas: Query no implementado")
}

func (i indiceConEntradas) List(context.Context) ([]ports.EntradaIndice, error) {
	return []ports.EntradaIndice(i), nil
}

func (i indiceConEntradas) Upsert(context.Context, string, domain.Graph) error { return nil }

// TestListarArnesesDevuelveLaClaveNoElArnesID cierra el cartel «no está en el índice del
// daemon»: la sesión guarda la clave calificada del Portafolio (CAP-142) y la usa para pedir
// el grafo, así que el listado tiene que hablar ESE espacio de llaves. Cuando devolvía
// `arnes.id` (el id interno del manifiesto) el cliente comparaba peras con manzanas y un
// arnés cargable se veía ausente.
func TestListarArnesesDevuelveLaClaveNoElArnesID(t *testing.T) {
	idx := indiceConEntradas{
		{
			Clave: "sin-home~vitalia~vitalia",
			Grafo: domain.Graph{Arnes: &domain.Arnes{
				ID:       "vitalia",
				Rol:      "Ingeniería · Desarrollo full-cycle",
				Proceso:  "desarrollo del producto Vitalia",
				Empresas: []string{"vitalia"},
			}},
		},
		{Clave: "dev-full-cycle", Grafo: domain.Graph{Arnes: &domain.Arnes{ID: "dev-full-cycle"}}},
		// Sin manifiesto no hay nada que mostrar: la fila se saltea, no se inventa.
		{Clave: "sin-manifiesto", Grafo: domain.Graph{}},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/harnesses", nil)
	listHarnesses(usecase.NewMapService(idx))(rec, req)
	if rec.Code != 200 {
		t.Fatalf("GET /api/harnesses = %d, want 200", rec.Code)
	}

	var got []harnessSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body: %v (%s)", err, rec.Body.String())
	}
	if len(got) != 2 {
		t.Fatalf("filas = %d (%v), want 2 — la entrada sin manifiesto no se lista", len(got), got)
	}
	if got[0].ID != "sin-home~vitalia~vitalia" {
		t.Errorf("id[0] = %q, want la clave del índice %q (no el arnes.id interno)",
			got[0].ID, "sin-home~vitalia~vitalia")
	}
	// Las facetas siguen saliendo del manifiesto: la clave cambia el `id`, nada más.
	if got[0].Rol == "" || got[0].Proceso == "" || len(got[0].Empresas) != 1 {
		t.Errorf("fila[0] perdió facetas del manifiesto: %+v", got[0])
	}
	if got[1].ID != "dev-full-cycle" {
		t.Errorf("id[1] = %q, want %q — una clave simple sigue igual", got[1].ID, "dev-full-cycle")
	}
}
