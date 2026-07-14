package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// tuercaFake implementa ports.SelfUpdater para probar el WIRE (RF-106/107): mapeo de
// sentinels a HTTP y el contrato de que el request jamás parametriza nada.
type tuercaFake struct {
	version ports.VersionInfo
	huella  string
	// repoUsado registra que el build corre con el repo CONFIGURADO, no con nada del request.
	repoUsado string
}

func (f *tuercaFake) Version() ports.VersionInfo { return f.version }
func (f *tuercaFake) Verificar(context.Context) (string, error) {
	return "ok", nil
}

func (f *tuercaFake) Build(context.Context) (string, error) {
	f.repoUsado = f.version.Repo
	return "ok", nil
}

func (f *tuercaFake) VerificarBinario(context.Context) (string, string, error) {
	return f.huella, "ok", nil
}
func (f *tuercaFake) Instalar(context.Context) (string, error) { return "ok", nil }
func (f *tuercaFake) Reiniciar() error                         { return nil }
func (f *tuercaFake) ConfigurarRepo(context.Context, string) (string, error) {
	return "ok", nil
}

func TestGetVersionWire(t *testing.T) {
	svc := usecase.NewSelfUpdateService(&tuercaFake{version: ports.VersionInfo{
		Huella: "1c7443f", Fecha: "2026-07-07", InstaladoEn: "/home/x/.local/bin/arnesia",
		Repo: "/repo", Escribible: true,
	}}, nil)
	w := httptest.NewRecorder()
	getVersion(svc)(w, httptest.NewRequestWithContext(context.Background(), "GET", "/api/version", nil))
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	// Los campos del RF-107, con los nombres del spec.
	for _, k := range []string{"huella", "fecha", "instalado_en", "escribible", "repo", "sucio"} {
		if _, ok := body[k]; !ok {
			t.Fatalf("falta %q en el wire: %v", k, body)
		}
	}
	if body["huella"] != "1c7443f" || body["escribible"] != true {
		t.Fatalf("wire infiel: %v", body)
	}
}

func TestPostSelfUpdateIgnoraParametrosDelRequest(t *testing.T) {
	f := &tuercaFake{version: ports.VersionInfo{
		Huella: "aaaaaaa", Repo: "/repo-configurado", Escribible: true, InstaladoEn: "/x",
	}, huella: "bbbbbbb"}
	svc := usecase.NewSelfUpdateService(f, nil)
	// Un request hostil manda rutas: NO viajan a ningún lado (RF-106: cero params).
	r := httptest.NewRequestWithContext(context.Background(), "POST", "/api/self-update", strings.NewReader(`{"repo":"/tmp/evil"}`))
	w := httptest.NewRecorder()
	postSelfUpdate(svc)(w, r)
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	if f.repoUsado != "/repo-configurado" {
		t.Fatalf("el build debe correr con el repo CONFIGURADO, corrió con %q", f.repoUsado)
	}
	var rep usecase.SelfUpdateReport
	if err := json.Unmarshal(w.Body.Bytes(), &rep); err != nil {
		t.Fatal(err)
	}
	if rep.Resultado != usecase.ResultadoActualizado || len(rep.Pasos) != 5 {
		t.Fatalf("reporte infiel: %+v", rep)
	}
}

func TestPostSelfUpdate409EnVuelo(t *testing.T) {
	f := &tuercaFake{version: ports.VersionInfo{Huella: "aaaaaaa", Repo: "/r", Escribible: true, InstaladoEn: "/x"}, huella: "bbbbbbb"}
	svc := usecase.NewSelfUpdateService(f, nil)
	// Primer update termina «actualizado» → lock retenido (decisión #9) → segundo = 409.
	w1 := httptest.NewRecorder()
	postSelfUpdate(svc)(w1, httptest.NewRequestWithContext(context.Background(), "POST", "/api/self-update", nil))
	if w1.Code != 200 {
		t.Fatalf("primer POST: %d", w1.Code)
	}
	w2 := httptest.NewRecorder()
	postSelfUpdate(svc)(w2, httptest.NewRequestWithContext(context.Background(), "POST", "/api/self-update", nil))
	if w2.Code != 409 {
		t.Fatalf("segundo POST en la ventana de reinicio: %d, quiero 409", w2.Code)
	}
}

// tuercaConfigurable extiende tuercaFake con ConfigurarRepo programable (bugfix
// fix-repo-self-update, RF-109) para probar el wire de PUT /api/self-update/repo.
type tuercaConfigurable struct {
	tuercaFake
	configurarErr error
	pathRecibido  string
}

func (f *tuercaConfigurable) ConfigurarRepo(_ context.Context, path string) (string, error) {
	f.pathRecibido = path
	if f.configurarErr != nil {
		return "", f.configurarErr
	}
	return "repo válido", nil
}

func TestPutSelfUpdateRepoValido(t *testing.T) {
	f := &tuercaConfigurable{}
	svc := usecase.NewSelfUpdateService(f, nil)
	r := httptest.NewRequestWithContext(context.Background(), "PUT", "/api/self-update/repo", strings.NewReader(`{"path":"/repo/candidato"}`))
	w := httptest.NewRecorder()
	putSelfUpdateRepo(svc)(w, r)
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	if f.pathRecibido != "/repo/candidato" {
		t.Fatalf("el path del body debe llegar EXACTO al puerto, tengo %q", f.pathRecibido)
	}
}

func TestPutSelfUpdateRepoInvalido400(t *testing.T) {
	f := &tuercaConfigurable{configurarErr: errors.New("módulo ajeno")}
	svc := usecase.NewSelfUpdateService(f, nil)
	r := httptest.NewRequestWithContext(context.Background(), "PUT", "/api/self-update/repo", strings.NewReader(`{"path":"/repo/malo"}`))
	w := httptest.NewRecorder()
	putSelfUpdateRepo(svc)(w, r)
	if w.Code != 400 {
		t.Fatalf("status %d, quiero 400 (cuerpo %s)", w.Code, w.Body)
	}
	if !strings.Contains(w.Body.String(), "módulo ajeno") {
		t.Fatalf("el 400 debe llevar el motivo exacto: %s", w.Body)
	}
}

func TestPutSelfUpdateRepoSinPath400(t *testing.T) {
	f := &tuercaConfigurable{}
	svc := usecase.NewSelfUpdateService(f, nil)
	r := httptest.NewRequestWithContext(context.Background(), "PUT", "/api/self-update/repo", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	putSelfUpdateRepo(svc)(w, r)
	if w.Code != 400 {
		t.Fatalf("status %d, quiero 400 (path vacío)", w.Code)
	}
	if f.pathRecibido != "" {
		t.Fatal("sin path el puerto NUNCA debe llamarse")
	}
}

func TestPostSelfUpdate503NoActualizable(t *testing.T) {
	casos := []ports.VersionInfo{
		{Huella: "a", Repo: "/r", Escribible: false, InstaladoEn: "/usr/bin/arnesia"}, // RF-102
		{Huella: "a", Repo: "", Escribible: true, InstaladoEn: "/x"},                  // RF-103
	}
	for _, v := range casos {
		svc := usecase.NewSelfUpdateService(&tuercaFake{version: v}, nil)
		w := httptest.NewRecorder()
		postSelfUpdate(svc)(w, httptest.NewRequestWithContext(context.Background(), "POST", "/api/self-update", nil))
		if w.Code != 503 {
			t.Fatalf("no-actualizable %+v: %d, quiero 503 (cuerpo %s)", v, w.Code, w.Body)
		}
		if !strings.Contains(w.Body.String(), "self-update no puede correr") {
			t.Fatalf("el 503 debe llevar el motivo: %s", w.Body)
		}
	}
}
