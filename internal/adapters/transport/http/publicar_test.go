package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// publicar_test.go prueba el WIRE de POST /api/portafolio/arneses/{clave}/publicaciones
// (RF-B2.5): status por centinela, jamás parseando texto — espejo de TestTraerStatusPorCentinela.

// Cada centinela de Publicar mapea a SU status; ninguno cae en un 500 genérico.
func TestPublicarStatusPorCentinela(t *testing.T) {
	casos := []struct {
		nombre string
		err    error
		want   int
	}{
		{"clave desconocida ⇒ 404", fmt.Errorf("%w: x", usecase.ErrObservarClaveNoEncontrada), 404},
		{"sin canónico ⇒ 400", fmt.Errorf("%w: x", domain.ErrPublicarSinCanonico), 400},
		{"sin home ⇒ 400", fmt.Errorf("%w: x", domain.ErrPublicarSinHome), 400},
		{"no propio ⇒ 400", fmt.Errorf("%w: x", domain.ErrPublicarNoPropio), 400},
		{"versión inválida ⇒ 400", fmt.Errorf("%w: latest", domain.ErrPublicarVersionInvalida), 400},
		{"versión ya publicada ⇒ 409", fmt.Errorf("%w: 0.6.0", domain.ErrPublicarVersionYaPublicada), 409},
		{"push rechazado ⇒ 409", fmt.Errorf("%w: non-fast-forward", domain.ErrPublicarPushRechazado), 409},
		{"conformance rojo pelado ⇒ 409", fmt.Errorf("%w", domain.ErrPublicarConformanceRojo), 409},
		{"sin auth ⇒ 503", fmt.Errorf("%w: 403", domain.ErrPublicarSinAuth), 503},
		{"publisher no cableado ⇒ 503", usecase.ErrPublicarNoDisponible, 503},
		{"fallo local ⇒ 500", errors.New("publicar: ENOSPC"), 500},
	}
	vistos := map[int]bool{}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			rec := httptest.NewRecorder()
			escribirErrorPublicar(rec, c.err)
			if rec.Code != c.want {
				t.Fatalf("status = %d, want %d", rec.Code, c.want)
			}
			if rec.Body.Len() == 0 || !strings.Contains(rec.Body.String(), "error") {
				t.Fatalf("el body debe traer el motivo real: %s", rec.Body.String())
			}
			vistos[c.want] = true
		})
	}
	for _, code := range []int{400, 404, 409, 503, 500} {
		if !vistos[code] {
			t.Fatalf("la tabla no cubre el status %d", code)
		}
	}
}

// El 409 del gate rojo lleva el reporte ADJUNTO en `conformance` (body enriquecido, precedente
// conflictoTraerBody): la UI lista los checks FAIL, no solo «está rojo».
func TestPublicarConformanceRojoBody(t *testing.T) {
	err := &usecase.ErrorConformancePublicar{Reporte: domain.ConformanceReport{
		Results: []domain.CheckResult{{
			Check:     domain.Check{ID: "escritor-unico", Severidad: domain.SevError},
			Veredicto: domain.VeredictoFail,
			Detalle:   "dos cajas escriben el mismo artefacto",
		}},
	}}
	rec := httptest.NewRecorder()
	escribirErrorPublicar(rec, err)

	if rec.Code != 409 {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
	var body conflictoPublicarBody
	if derr := json.Unmarshal(rec.Body.Bytes(), &body); derr != nil {
		t.Fatalf("decode: %v — %s", derr, rec.Body.String())
	}
	if body.Error == "" || len(body.Conformance.Results) != 1 ||
		body.Conformance.Results[0].Check.ID != "escritor-unico" {
		t.Fatalf("body = %+v", body)
	}
}

// El handler con el servicio sin publisher cableado responde 503 honesto por la ruta real.
func TestPostPublicacionesSinPublisher503(t *testing.T) {
	e := armarHTTP(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), "POST", "/api/portafolio/arneses/x~y~/publicaciones", nil)
	req.SetPathValue("clave", "x~y~")
	postPublicar(e.svc)(rec, req)
	if rec.Code != 503 {
		t.Fatalf("status = %d, want 503: %s", rec.Code, rec.Body.String())
	}
}
