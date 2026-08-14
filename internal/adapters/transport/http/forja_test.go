package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// fakeForjaPort — el puerto de mentira: el handler traduce request↔usecase y nada más
// (dominio-independiente-de-transporte).
type fakeForjaPort struct {
	informe domain.InformeSemilla
	salud   domain.SaludSemilla
}

func (f *fakeForjaPort) Sembrar(string) (domain.InformeSemilla, error) { return f.informe, nil }
func (f *fakeForjaPort) Chequear(string) (domain.SaludSemilla, error)  { return f.salud, nil }

// proyectoDePrueba: un dir real y NO protegido (la validación del usecase corre de verdad).
func proyectoDePrueba(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows: os.UserHomeDir lee USERPROFILE, no HOME
	dir := filepath.Join(home, "proyecto")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestPostSembrarSemilla(t *testing.T) {
	dir := proyectoDePrueba(t)
	svc := usecase.NewForjaService(&fakeForjaPort{
		informe: domain.InformeSemilla{Creados: []string{".arnesia/terreno/INDEX.md"}, YaExistian: []string{".arnesia/product/backlog.md"}},
		salud:   domain.SaludSemilla{Estado: domain.SemillaSana},
	})

	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/forja/semillas", strings.NewReader(`{"path":"`+dir+`"}`))
	w := httptest.NewRecorder()
	postSembrarSemilla(svc)(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, quiero 200 (body: %s)", w.Code, w.Body)
	}
	var resp struct {
		Informe domain.InformeSemilla `json:"informe"`
		Salud   domain.SaludSemilla   `json:"salud"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("body ilegible: %v", err)
	}
	if len(resp.Informe.Creados) != 1 || len(resp.Informe.YaExistian) != 1 || resp.Salud.Estado != domain.SemillaSana {
		t.Errorf("respuesta = %+v, no refleja el informe+salud del usecase", resp)
	}
}

// TestPostSembrarSemillaPathProtegido: 400 con motivo, jamás un 200 que oculte el path
// inválido — y el body usa `path`, el MISMO campo que POST /api/portafolio/escaneos.
func TestPostSembrarSemillaPathProtegido(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows: os.UserHomeDir lee USERPROFILE, no HOME
	svc := usecase.NewForjaService(&fakeForjaPort{})

	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/forja/semillas", strings.NewReader(`{"path":"`+home+`"}`))
	w := httptest.NewRecorder()
	postSembrarSemilla(svc)(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, quiero 400 (HOME es raíz protegida)", w.Code)
	}
	var e errorBody
	if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil || e.Error == "" {
		t.Errorf("el 400 debe traer el motivo: body=%s err=%v", w.Body, err)
	}
}

func TestPostChequearSemilla(t *testing.T) {
	dir := proyectoDePrueba(t)
	svc := usecase.NewForjaService(&fakeForjaPort{
		salud: domain.SaludSemilla{Estado: domain.SemillaIncompleta, Faltantes: []string{".arnesia/wip/INDEX.md"}},
	})

	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/forja/semillas/chequeos", strings.NewReader(`{"path":"`+dir+`"}`))
	w := httptest.NewRecorder()
	postChequearSemilla(svc)(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, quiero 200 — una salud incompleta es un veredicto, no un error HTTP", w.Code)
	}
	var salud domain.SaludSemilla
	if err := json.Unmarshal(w.Body.Bytes(), &salud); err != nil {
		t.Fatalf("body ilegible: %v", err)
	}
	if salud.Estado != domain.SemillaIncompleta || len(salud.Faltantes) != 1 {
		t.Errorf("salud = %+v, el veredicto debe viajar TAL CUAL", salud)
	}
}
