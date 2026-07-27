package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// storeQueFalla persiste hasta que se lo rompe. Modela el modo D: el filesystem deja de
// aceptar escrituras (disco lleno, permisos, montaje de sólo lectura).
type storeQueFalla struct {
	sesiones []domain.Session
	roto     bool
}

var errDisco = errors.New("no space left on device")

func (s *storeQueFalla) Load(context.Context) ([]domain.Session, error) {
	return append([]domain.Session{}, s.sesiones...), nil
}

func (s *storeQueFalla) Save(_ context.Context, ss []domain.Session) error {
	if s.roto {
		return errDisco
	}
	s.sesiones = append([]domain.Session{}, ss...)
	return nil
}

// storeBloqueado modela el modo E: el archivo en disco lo escribió un binario más nuevo,
// así que el registro se declara en solo-lectura y ninguna mutación se intenta siquiera.
type storeBloqueado struct{ storeQueFalla }

func (storeBloqueado) SoloLectura() (bool, string) {
	return true, "sesiones.json lo escribió un binario más nuevo (esquema 99, este entiende 2)"
}

func (storeBloqueado) Sembrable() bool { return false }

// TestPersistFallidoRevierteLaMutacion (E-45, RF-338) — cuando el disco falla, la mutación
// que el operador acaba de pedir NO queda a medias: devuelve el motivo del filesystem y el
// estado en memoria vuelve a lo que había. Estado en memoria y estado en disco no divergen
// en silencio.
func TestPersistFallidoRevierteLaMutacion(t *testing.T) {
	st := &storeQueFalla{}
	svc, err := usecase.NewSessionService(context.Background(), &stubAgent{}, st, stubPub{}, stubResolver{path: t.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	base, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "el frente de siempre"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetView(base.ID, "Mapa"); err != nil {
		t.Fatal(err)
	}
	antes := len(svc.List())

	st.roto = true

	t.Run("crear", func(t *testing.T) {
		_, cerr := svc.Create(domain.Session{Arnes: "vitalia", Frente: "no debería existir"})
		if !errors.Is(cerr, errDisco) {
			t.Fatalf("Create = %v, quiero el motivo del filesystem", cerr)
		}
		if n := len(svc.List()); n != antes {
			t.Errorf("quedaron %d sesiones, quiero %d — la que no se guardó no existe", n, antes)
		}
		for _, s := range svc.List() {
			if s.Frente == "no debería existir" {
				t.Error("la sesión que falló al guardarse quedó viva en memoria")
			}
		}
	})

	t.Run("renombrar", func(t *testing.T) {
		_, rerr := svc.Rename(base.ID, "nombre que no se guardó")
		if !errors.Is(rerr, errDisco) {
			t.Fatalf("Rename = %v, quiero el motivo del filesystem", rerr)
		}
		m, _ := svc.Get(base.ID)
		if m.Frente != "el frente de siempre" {
			t.Errorf("frente = %q — el renombre que falló no puede quedar en memoria", m.Frente)
		}
	})

	t.Run("cambiar de vista", func(t *testing.T) {
		_, verr := svc.SetView(base.ID, "Diag")
		if !errors.Is(verr, errDisco) {
			t.Fatalf("SetView = %v, quiero el motivo del filesystem", verr)
		}
		m, _ := svc.Get(base.ID)
		if m.View != "Mapa" {
			t.Errorf("view = %q — el cambio que falló no puede quedar en memoria", m.View)
		}
	})

	t.Run("cerrar", func(t *testing.T) {
		cerr := svc.Close(base.ID)
		if !errors.Is(cerr, errDisco) {
			t.Fatalf("Close = %v, quiero el motivo del filesystem", cerr)
		}
		if _, ok := svc.Get(base.ID); !ok {
			t.Fatal("la sesión desapareció del registro pese a que el cierre no se guardó")
		}
		if n := len(svc.List()); n != antes {
			t.Errorf("quedaron %d sesiones, quiero %d", n, antes)
		}
	})
}

// TestSoloLecturaRechazaAntesDeTocarMemoria (modo E) — con el registro bloqueado la
// mutación se rechaza ANTES de tocar nada, con el motivo adentro del error para que el
// operador lea por qué y no un «error interno».
func TestSoloLecturaRechazaAntesDeTocarMemoria(t *testing.T) {
	svc, err := usecase.NewSessionService(context.Background(), &stubAgent{}, &storeBloqueado{}, stubPub{}, stubResolver{path: t.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if bloqueado, motivo := svc.SoloLectura(); !bloqueado || motivo == "" {
		t.Fatalf("SoloLectura = %v %q, quiero bloqueado con motivo", bloqueado, motivo)
	}
	_, cerr := svc.Create(domain.Session{Arnes: "vitalia"})
	if !errors.Is(cerr, usecase.ErrSoloLectura) {
		t.Fatalf("Create con el registro bloqueado = %v, quiero ErrSoloLectura", cerr)
	}
	if !strings.Contains(cerr.Error(), "esquema 99") {
		t.Errorf("el error no dice por qué: %v", cerr)
	}
	if n := len(svc.List()); n != 0 {
		t.Errorf("se crearon %d sesiones con el registro bloqueado", n)
	}
}

// TestRegistroBloqueadoNoSeSiembra — un registro vacío por bloqueo NO es un primer
// arranque: sembrarlo mostraría tres sesiones de ejemplo sobre un archivo del futuro que
// el operador todavía tiene entero en disco.
func TestRegistroBloqueadoNoSeSiembra(t *testing.T) {
	svc, err := usecase.NewSessionService(context.Background(), &stubAgent{}, &storeBloqueado{}, stubPub{}, stubResolver{path: t.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(svc.List()); n != 0 {
		t.Errorf("el registro bloqueado se sembró con %d sesiones de ejemplo", n)
	}
}
