package usecase_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

func entradaVitalia() domain.EntradaPortafolio {
	return domain.EntradaPortafolio{
		Identidad: domain.IdentidadArnes{ID: "vitalia", Scope: "."},
		Nombre:    "Vitalia",
		Instalaciones: []domain.Instalacion{{
			ProyectoPath:  "/proj/luana-vitalia/vitalia",
			InstallPath:   "/proj/luana-vitalia/vitalia",
			Deriva:        domain.DerivaNoEvaluable,
			DerivaDetalle: "sin version instalada conocida",
		}},
	}
}

// TestTarjetaInstalacionEsReparacion (RF-189/190): cwd que matchea una instalación → la
// tarjeta declara INSTALACIÓN + deriva + el loop de reparación A4.
func TestTarjetaInstalacionEsReparacion(t *testing.T) {
	tarjeta := usecase.TarjetaIdentidad([]domain.EntradaPortafolio{entradaVitalia()},
		"vitalia", "/proj/luana-vitalia/vitalia", "Ingeniería · Desarrollo full-cycle")
	for _, quiero := range []string{"INSTALACIÓN", "REPARACIÓN", "deriva", "backport", "Ingeniería · Desarrollo full-cycle", "causa"} {
		if !strings.Contains(tarjeta, quiero) {
			t.Errorf("tarjeta sin %q:\n%s", quiero, tarjeta)
		}
	}
}

// TestTarjetaCanonico: cwd que matchea el canónico → lo dice, sin loop de reparación.
func TestTarjetaCanonico(t *testing.T) {
	e := entradaVitalia()
	e.Canonico = &domain.Canonico{Path: "/canon/vitalia"}
	tarjeta := usecase.TarjetaIdentidad([]domain.EntradaPortafolio{e}, "vitalia", "/canon/vitalia", "reviewer")
	if !strings.Contains(tarjeta, "CANÓNICO") || strings.Contains(tarjeta, "REPARACIÓN") {
		t.Errorf("tarjeta canónico mal armada:\n%s", tarjeta)
	}
}

// TestTarjetaSinRegistro: cwd fuera del Portafolio → tarjeta mínima honesta; y sin rol,
// la tarjeta lo DICE — y dice que SÍ se puede escribir vía la tarjeta de permiso del
// panel (DD-1). El texto previo («sin rol no hay escritura», aprendizaje T4) hacía que
// el modelo se auto-negara antes de intentar el Write y el HITL jamás se ejercía.
func TestTarjetaSinRegistro(t *testing.T) {
	tarjeta := usecase.TarjetaIdentidad(nil, "suelto", "/tmp/x", "")
	if !strings.Contains(tarjeta, "NO registrado") || !strings.Contains(tarjeta, "sin rol declarado") {
		t.Errorf("tarjeta suelta mal armada:\n%s", tarjeta)
	}
	if !strings.Contains(tarjeta, "podés escribir") || strings.Contains(tarjeta, "no puede escribir") {
		t.Errorf("la tarjeta sin rol debe decir que la escritura pasa por el panel (DD-1), no negarla:\n%s", tarjeta)
	}
}

// stubInjector implementa InjectionProvisioner registrando las llamadas.
type stubInjector struct {
	sessions []string
	extras   []string
}

func (s *stubInjector) Provision(context.Context) (ports.Injection, error) {
	return ports.Injection{SystemPromptFile: "/base/doctrine.md"}, nil
}

func (s *stubInjector) ProvisionSession(_ context.Context, sessionID, extra string) (ports.Injection, error) {
	s.sessions = append(s.sessions, sessionID)
	s.extras = append(s.extras, extra)
	return ports.Injection{SystemPromptFile: "/sessions/" + sessionID + "/system.md"}, nil
}

// TestSpawnUsaTarjetaPorSesion (RF-189): el spawn provisiona POR SESIÓN con la tarjeta del
// grounding cableado.
func TestSpawnUsaTarjetaPorSesion(t *testing.T) {
	agent := &stubAgent{}
	inj := &stubInjector{}
	svc, err := usecase.NewSessionService(context.Background(), agent, stubStore{}, stubPub{}, stubResolver{path: t.TempDir()}, 40, inj, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetGrounding(func(_ context.Context, arnesID, cwd string) string {
		return "tarjeta-de-" + arnesID
	})
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Turn(s.ID, "hola"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for len(inj.sessions) == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if len(inj.sessions) != 1 || inj.sessions[0] != s.ID || inj.extras[0] != "tarjeta-de-vitalia" {
		t.Fatalf("ProvisionSession llamado con %v / %v", inj.sessions, inj.extras)
	}
}
