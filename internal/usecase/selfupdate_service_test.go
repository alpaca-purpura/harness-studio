package usecase

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// fakeUpdater implementa ports.SelfUpdater con desenlaces programables y registra el
// orden real de llamadas (el usecase es quien ordena y corta — eso es lo que se prueba).
type fakeUpdater struct {
	mu       sync.Mutex
	llamadas []string

	version      ports.VersionInfo
	verificarErr error
	buildErr     error
	verBinErr    error
	instalarErr  error
	huellaNueva  string
	// buildEmpezo/buildSigue permiten congelar el flujo DENTRO del build para probar
	// el 409 concurrente de verdad (no por timing).
	buildEmpezo chan struct{}
	buildSigue  chan struct{}
}

func (f *fakeUpdater) marca(paso string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.llamadas = append(f.llamadas, paso)
}

func (f *fakeUpdater) Version() ports.VersionInfo { return f.version }

func (f *fakeUpdater) Verificar(context.Context) (string, error) {
	f.marca("verificar")
	return "terreno OK", f.verificarErr
}

func (f *fakeUpdater) Build(context.Context) (string, error) {
	f.marca("build")
	if f.buildEmpezo != nil {
		close(f.buildEmpezo)
		<-f.buildSigue
	}
	if f.buildErr != nil {
		return "cola del stderr", f.buildErr
	}
	return "build OK", nil
}

func (f *fakeUpdater) VerificarBinario(context.Context) (string, string, error) {
	f.marca("verificar binario")
	return f.huellaNueva, "bin/arnesia inspeccionado", f.verBinErr
}

func (f *fakeUpdater) Instalar(context.Context) (string, error) {
	f.marca("instalar")
	return "instalado", f.instalarErr
}

func (f *fakeUpdater) Reiniciar() error {
	f.marca("reiniciar")
	return nil
}

func actualizable(huellaNueva string) *fakeUpdater {
	return &fakeUpdater{
		version:     ports.VersionInfo{Huella: "aaaaaaa", InstaladoEn: "/x/arnesia", Repo: "/repo", Escribible: true},
		huellaNueva: huellaNueva,
	}
}

func estadosPorPaso(t *testing.T, rep SelfUpdateReport) map[string]string {
	t.Helper()
	if len(rep.Pasos) != 5 {
		t.Fatalf("el reporte SIEMPRE lleva los 5 pasos, tengo %d: %+v", len(rep.Pasos), rep.Pasos)
	}
	m := make(map[string]string, 5)
	for _, p := range rep.Pasos {
		m[p.Paso] = p.Estado
	}
	return m
}

func TestActualizarCaminoFeliz(t *testing.T) {
	f := actualizable("bbbbbbb")
	s := NewSelfUpdateService(f)
	rep, err := s.Actualizar(context.Background())
	if err != nil {
		t.Fatalf("camino feliz: %v", err)
	}
	if rep.Resultado != ResultadoActualizado || rep.HuellaNueva != "bbbbbbb" {
		t.Fatalf("resultado %q huella %q", rep.Resultado, rep.HuellaNueva)
	}
	m := estadosPorPaso(t, rep)
	for _, p := range []string{PasoVerificar, PasoBuild, PasoVerificarBinario, PasoInstalar} {
		if m[p] != EstadoOK {
			t.Fatalf("paso %q = %q, quiero ok", p, m[p])
		}
	}
	if m[PasoReiniciar] != EstadoAgendado {
		t.Fatalf("reiniciar = %q, quiero agendado (la respuesta sale ANTES del re-exec)", m[PasoReiniciar])
	}
}

func TestCorteAlPrimerFallo(t *testing.T) {
	f := actualizable("bbbbbbb")
	f.buildErr = errors.New("exit 1")
	s := NewSelfUpdateService(f)
	rep, err := s.Actualizar(context.Background())
	if err != nil {
		t.Fatalf("un paso fallido es desenlace normal (200), no error: %v", err)
	}
	if rep.Resultado != ResultadoFallo {
		t.Fatalf("resultado = %q", rep.Resultado)
	}
	m := estadosPorPaso(t, rep)
	if m[PasoVerificar] != EstadoOK || m[PasoBuild] != EstadoFallo {
		t.Fatalf("verificar=%q build=%q", m[PasoVerificar], m[PasoBuild])
	}
	for _, p := range []string{PasoVerificarBinario, PasoInstalar, PasoReiniciar} {
		if m[p] != EstadoNoCorrido {
			t.Fatalf("paso %q = %q, quiero no-corrido", p, m[p])
		}
	}
	for _, l := range f.llamadas {
		if l == "verificar binario" || l == "instalar" {
			t.Fatalf("tras el fallo del build NADA más corre; corrió %q", l)
		}
	}
	// El detalle del fallo lleva el stderr del adapter (RF-104).
	for _, p := range rep.Pasos {
		if p.Paso == PasoBuild && !strings.Contains(p.Detalle, "cola del stderr") {
			t.Fatalf("detalle del build sin stderr: %q", p.Detalle)
		}
	}
}

func TestYaAlDia(t *testing.T) {
	f := actualizable("aaaaaaa") // misma huella que el binario corriendo.
	s := NewSelfUpdateService(f)
	rep, err := s.Actualizar(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rep.Resultado != ResultadoYaAlDia {
		t.Fatalf("resultado = %q, quiero ya-al-dia", rep.Resultado)
	}
	m := estadosPorPaso(t, rep)
	if m[PasoInstalar] != EstadoNoCorrido || m[PasoReiniciar] != EstadoNoCorrido {
		t.Fatalf("ya-al-día NO reinstala ni reinicia: instalar=%q reiniciar=%q", m[PasoInstalar], m[PasoReiniciar])
	}
	for _, l := range f.llamadas {
		if l == "instalar" {
			t.Fatal("ya-al-día jamás llama Instalar")
		}
	}
	// El flujo terminó: el lock se libera y otro update puede correr.
	if _, err := s.Actualizar(context.Background()); err != nil {
		t.Fatalf("tras ya-al-día el lock debe estar libre: %v", err)
	}
}

func TestNoActualizable(t *testing.T) {
	t.Run("no escribible", func(t *testing.T) {
		f := actualizable("bbbbbbb")
		f.version.Escribible = false
		s := NewSelfUpdateService(f)
		if _, err := s.Actualizar(context.Background()); !errors.Is(err, ErrNoActualizable) {
			t.Fatalf("quiero ErrNoActualizable, tengo %v", err)
		}
		if len(f.llamadas) != 0 {
			t.Fatalf("no-actualizable corta ANTES de cualquier paso; corrió %v", f.llamadas)
		}
	})
	t.Run("sin repo", func(t *testing.T) {
		f := actualizable("bbbbbbb")
		f.version.Repo = ""
		s := NewSelfUpdateService(f)
		if _, err := s.Actualizar(context.Background()); !errors.Is(err, ErrNoActualizable) {
			t.Fatalf("quiero ErrNoActualizable, tengo %v", err)
		}
		// El error temprano libera el lock.
		f.version.Repo = "/repo"
		if _, err := s.Actualizar(context.Background()); err != nil {
			t.Fatalf("el lock debe quedar libre tras un 503: %v", err)
		}
	})
}

func Test409Concurrente(t *testing.T) {
	f := actualizable("bbbbbbb")
	f.buildEmpezo = make(chan struct{})
	f.buildSigue = make(chan struct{})
	s := NewSelfUpdateService(f)

	done := make(chan SelfUpdateReport, 1)
	go func() {
		rep, _ := s.Actualizar(context.Background())
		done <- rep
	}()
	<-f.buildEmpezo // el primero está DENTRO del build.

	if _, err := s.Actualizar(context.Background()); !errors.Is(err, ErrActualizacionEnCurso) {
		t.Fatalf("segundo POST en vuelo debe dar ErrActualizacionEnCurso, tengo %v", err)
	}
	close(f.buildSigue)
	if rep := <-done; rep.Resultado != ResultadoActualizado {
		t.Fatalf("el primero debió terminar actualizado, tengo %q", rep.Resultado)
	}
}

func TestLockRetenidoTrasActualizado(t *testing.T) {
	f := actualizable("bbbbbbb")
	s := NewSelfUpdateService(f)
	rep, err := s.Actualizar(context.Background())
	if err != nil || rep.Resultado != ResultadoActualizado {
		t.Fatalf("setup: %v %q", err, rep.Resultado)
	}
	// Decisión #9: el proceso es terminal (re-exec inminente); un segundo POST en la
	// ventana recibe 409 honesto, jamás un rebuild fantasma.
	if _, err := s.Actualizar(context.Background()); !errors.Is(err, ErrActualizacionEnCurso) {
		t.Fatalf("tras «actualizado» el lock se retiene: quiero 409, tengo %v", err)
	}
}
