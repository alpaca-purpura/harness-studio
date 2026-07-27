package usecase_test

// Tests de A-1 y A-2 (auditoría 2026-07-26): las dos decisiones FIRMADAS que se habían
// escrito en el dominio y nunca se habían cableado al camino real del turno.
//
//   - CV-D9 / RF-303 — el título de la conversación se DERIVA del primer mensaje del
//     usuario. `domain.DerivarTitulo` existía, estaba probada, y no la llamaba nadie en
//     producción: toda conversación se llamaba «nueva conversación» para siempre.
//   - CV-D13 / RF-304 — `UltimaInteraccion` se estampa EN CADA TURNO, no al desactivar.
//     Su única asignación de producción la ponía vacía (la migración, que no inventa
//     fechas). Toda fila de la lista decía «sin fecha» y RF-320 CA-1 caía en silencio al
//     orden de creación.
//
// Por qué se escaparon a tres capas de verificación verde: los tests de dominio prueban la
// LEY (quién gana si ambos escriben), no el CABLEADO. Estos tests corren contra el
// SessionService real —el mismo camino que usa el daemon— y por eso se ponen rojos si la
// llamada se quita.

import (
	"context"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// TestPrimerTurnoDerivaElTituloDeLaConversacion — RF-303 escenario «derivación».
// Cablea `DerivarTitulo` al turno: sin la llamada en `Turn`, el título se queda en el
// default y este test falla.
func TestPrimerTurnoDerivaElTituloDeLaConversacion(t *testing.T) {
	agent := &stubAgent{}
	svc := newSvc(t, agent, func(context.Context, string) string { return "" })
	id := svc.List()[0].ID

	s, _ := svc.Get(id)
	if c := activaDe(t, s); c.Titulo != domain.TituloConversacionNueva {
		t.Fatalf("precondición: la conversación nace con %q, no con %q", domain.TituloConversacionNueva, c.Titulo)
	}

	const primer = "no veo nada en el mapa, ¿podés repararlo?"
	if err := svc.Turn(id, primer); err != nil {
		t.Fatalf("turn: %v", err)
	}

	s, _ = svc.Get(id)
	c := activaDe(t, s)
	if c.Titulo == domain.TituloConversacionNueva {
		t.Errorf("el título NO se derivó: sigue en %q — CV-D9/RF-303 sin construir", c.Titulo)
	}
	if want := domain.RecorteDeTitulo(primer); c.Titulo != want {
		t.Errorf("título = %q, quiero %q (el derivador es el que ya existe, RF-303 CA-2)", c.Titulo, want)
	}
	if c.TituloEditado {
		t.Error("derivar automáticamente NO es editar: `titulo_editado` tiene que seguir en false")
	}

	// RF-303 CA-3: el Frente de la SESIÓN conserva su propia derivación. Son dos nombres
	// distintos y ambos siguen existiendo.
	if s.Frente == "" || s.Frente == "nuevo frente" {
		t.Errorf("el frente de la sesión dejó de derivarse: %q", s.Frente)
	}
}

// TestElSegundoTurnoNoRebautizaLaConversacion — RF-303: se deriva del PRIMER mensaje. El
// segundo turno encuentra un título que ya no es el default y no lo toca.
func TestElSegundoTurnoNoRebautizaLaConversacion(t *testing.T) {
	agent := &stubAgent{}
	svc := newSvc(t, agent, func(context.Context, string) string { return "" })
	id := svc.List()[0].ID

	if err := svc.Turn(id, "el bug del índice"); err != nil {
		t.Fatalf("turn: %v", err)
	}
	s, _ := svc.Get(id)
	primero := activaDe(t, s).Titulo

	cerrarTurno(t, svc, agent, id)
	if err := svc.Turn(id, "ahora mirá otra cosa completamente distinta"); err != nil {
		t.Fatalf("turn 2: %v", err)
	}
	s, _ = svc.Get(id)
	if c := activaDe(t, s); c.Titulo != primero {
		t.Errorf("el segundo turno rebautizó la conversación: %q → %q", primero, c.Titulo)
	}
}

// TestTituloEditadoPorElOperadorGanaContraLaDerivacion — RF-303 escenario «no se
// re-deriva», pero por el camino REAL: renombrar por el servicio y después mandar turnos.
// El test de dominio prueba la ley; este prueba que el cableado la respeta.
func TestTituloEditadoPorElOperadorGanaContraLaDerivacion(t *testing.T) {
	agent := &stubAgent{}
	svc := newSvc(t, agent, func(context.Context, string) string { return "" })
	id := svc.List()[0].ID
	s, _ := svc.Get(id)
	conv := activaDe(t, s)

	if _, err := svc.RenombrarConversacion(id, conv.ID, "el bug del índice"); err != nil {
		t.Fatalf("renombrar: %v", err)
	}
	if err := svc.Turn(id, "no veo nada en el mapa, ¿podés repararlo?"); err != nil {
		t.Fatalf("turn: %v", err)
	}
	s, _ = svc.Get(id)
	if c := activaDe(t, s); c.Titulo != "el bug del índice" {
		t.Errorf("la derivación pisó al operador: título = %q", c.Titulo)
	}
}

// TestCadaTurnoEstampaLaUltimaInteraccion — RF-304 CA-1. Sin el estampado en `Turn` el
// campo queda vacío y este test falla; sin el estampado al cerrar el turno del assistant,
// falla la segunda mitad.
func TestCadaTurnoEstampaLaUltimaInteraccion(t *testing.T) {
	agent := &stubAgent{}
	svc := newSvc(t, agent, func(context.Context, string) string { return "" })
	id := svc.List()[0].ID

	// RF-304 CA-2 / BR-CV-14: con 0 turnos el campo va VACÍO. Vacío ⟺ 0 turnos.
	s, _ := svc.Get(id)
	if c := activaDe(t, s); c.UltimaInteraccion != "" {
		t.Errorf("una conversación de 0 turnos no tiene fecha que estampar, tiene %q", c.UltimaInteraccion)
	}

	antes := time.Now().UTC().Add(-time.Second)
	if err := svc.Turn(id, "hola"); err != nil {
		t.Fatalf("turn: %v", err)
	}
	s, _ = svc.Get(id)
	c := activaDe(t, s)
	if c.UltimaInteraccion == "" {
		t.Fatal("`ultima_interaccion` sigue vacía después de un turno del usuario — CV-D13/RF-304 sin construir")
	}
	delUsuario, err := time.Parse(time.RFC3339, c.UltimaInteraccion)
	if err != nil {
		t.Fatalf("`ultima_interaccion` = %q no es RFC3339: %v", c.UltimaInteraccion, err)
	}
	if delUsuario.Before(antes) {
		t.Errorf("la fecha estampada (%s) es anterior al turno (%s)", delUsuario, antes)
	}
	if delUsuario.Location() != time.UTC {
		t.Errorf("`ultima_interaccion` tiene que ser UTC, es %s", delUsuario.Location())
	}

	// El turno del assistant también la mueve: «última interacción» es la del último
	// mensaje del hilo, no la del último que escribió el operador (RF-304 CA-1).
	time.Sleep(1100 * time.Millisecond) // RFC3339 tiene resolución de segundo.
	cerrarTurno(t, svc, agent, id)
	s, _ = svc.Get(id)
	c = activaDe(t, s)
	delAsistente, err := time.Parse(time.RFC3339, c.UltimaInteraccion)
	if err != nil {
		t.Fatalf("`ultima_interaccion` = %q no es RFC3339: %v", c.UltimaInteraccion, err)
	}
	if !delAsistente.After(delUsuario) {
		t.Errorf("el turno del assistant no movió la fecha: %s ≤ %s", delAsistente, delUsuario)
	}
}

// TestDesactivarNoTocaLaUltimaInteraccion — RF-304 escenario «se estampa al turno, no al
// desactivar»: la conversación que se deja atrás conserva la fecha de SU último turno.
// Es la mitad del requisito que hace que una conversación inactiva no mienta.
func TestDesactivarNoTocaLaUltimaInteraccion(t *testing.T) {
	agent := &stubAgent{}
	svc := newSvc(t, agent, func(context.Context, string) string { return "" })
	id := svc.List()[0].ID

	if err := svc.Turn(id, "el primer tema"); err != nil {
		t.Fatalf("turn: %v", err)
	}
	cerrarTurno(t, svc, agent, id)
	s, _ := svc.Get(id)
	vieja := activaDe(t, s)
	fechaVieja := vieja.UltimaInteraccion
	if fechaVieja == "" {
		t.Fatal("precondición: la conversación con turnos tiene que tener fecha")
	}

	time.Sleep(1100 * time.Millisecond)
	if _, _, err := svc.CrearConversacion(id); err != nil {
		t.Fatalf("crear: %v", err)
	}
	s, _ = svc.Get(id)
	for _, c := range s.Conversaciones {
		if c.ID != vieja.ID {
			continue
		}
		if c.Activa {
			t.Fatal("la vieja tenía que quedar inactiva")
		}
		if c.UltimaInteraccion != fechaVieja {
			t.Errorf("desactivar movió la fecha: %q → %q — mentiría cuándo fue el último mensaje", fechaVieja, c.UltimaInteraccion)
		}
		return
	}
	t.Fatalf("la conversación vieja %q desapareció del registro", vieja.ID)
}

// TestUnaConversacionNuevaNaceSinFechaYConElTituloDefault — el borde que la fila pinta
// como «sin turnos todavía» (RF-304 CA-2, BR-CV-14): crear no es interactuar.
func TestUnaConversacionNuevaNaceSinFechaYConElTituloDefault(t *testing.T) {
	agent := &stubAgent{}
	svc := newSvc(t, agent, func(context.Context, string) string { return "" })
	id := svc.List()[0].ID
	if err := svc.Turn(id, "el primer tema"); err != nil {
		t.Fatalf("turn: %v", err)
	}
	cerrarTurno(t, svc, agent, id)

	nueva, _, err := svc.CrearConversacion(id)
	if err != nil {
		t.Fatalf("crear: %v", err)
	}
	if nueva.UltimaInteraccion != "" {
		t.Errorf("crear no es interactuar: `ultima_interaccion` = %q", nueva.UltimaInteraccion)
	}
	if nueva.Titulo != domain.TituloConversacionNueva {
		t.Errorf("una conversación sin turnos se llama %q, no %q", domain.TituloConversacionNueva, nueva.Titulo)
	}
}

// cerrarTurno cierra el turno en vuelo con un EventResult y espera a que el servicio
// vuelva a idle. No toca el uso de contexto: ninguno de estos tests rota.
func cerrarTurno(t *testing.T, svc interface {
	Get(string) (domain.Session, bool)
}, agent *stubAgent, id string,
) {
	t.Helper()
	sess := agent.sessions[len(agent.sessions)-1]
	sess.events <- ports.AgentEvent{Kind: ports.EventResult, Text: "listo"}
	waitUntil(t, func() bool {
		s, _ := svc.Get(id)
		return s.Status == domain.StatusIdle
	})
}
