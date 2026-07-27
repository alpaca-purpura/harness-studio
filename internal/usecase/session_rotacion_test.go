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

// TestCtxHistYUmbralRotacion (RF-194/195): el histórico crece por turno y cruzar el umbral
// marca RotacionPendiente — después de responder, jamás a mitad del turno.
func TestCtxHistYUmbralRotacion(t *testing.T) {
	agent := &stubAgent{}
	svc, err := usecase.NewSessionService(context.Background(), agent, stubStore{}, stubPub{}, stubResolver{path: t.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetUmbralRotacion(40)

	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	turno := func(ctx int) {
		if err := svc.Turn(s.ID, "hola"); err != nil {
			t.Fatal(err)
		}
		agent.sessions[len(agent.sessions)-1].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "ok", CtxPct: ctx}
		espera(t, func() bool {
			m, _ := svc.Get(s.ID)
			return m.Status == domain.StatusIdle && activaSinFallar(m).CtxPct == ctx
		})
	}

	turno(10)
	m, _ := svc.Get(s.ID)
	c := activaDe(t, m)
	if len(c.CtxHist) != 1 || c.CtxHist[0] != 10 || c.RotacionPendiente {
		t.Fatalf("tras turno 1: hist=%v pendiente=%v", c.CtxHist, c.RotacionPendiente)
	}

	turno(45)
	m, _ = svc.Get(s.ID)
	c = activaDe(t, m)
	if len(c.CtxHist) != 2 || c.CtxHist[1] != 45 {
		t.Errorf("hist=%v, quiero [10 45]", c.CtxHist)
	}
	if !c.RotacionPendiente {
		t.Error("45 >= umbral 40 debe marcar RotacionPendiente")
	}
}

// activaDe devuelve la conversación activa de una sesión leída del servicio. Falla el test
// si no hay ninguna: bajo CV-D3 eso es la invariante rota, no un caso a tolerar.
func activaDe(t *testing.T, s domain.Session) domain.Conversacion {
	t.Helper()
	c, ok := s.Activa()
	if !ok {
		t.Fatalf("la sesión %q no tiene conversación activa (invariante CV-D3): %+v", s.ID, s.Conversaciones)
	}
	return *c
}

// activaSinFallar es la variante para usar DENTRO de una condición de espera, donde un
// t.Fatal desde otra goroutine sería ilegal. Un registro sin activa devuelve el cero.
func activaSinFallar(s domain.Session) domain.Conversacion {
	if c, ok := s.Activa(); ok {
		return *c
	}
	return domain.Conversacion{}
}

func espera(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("timeout esperando condición")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestRotacionInvisible (RF-196/197/198): con RotacionPendiente, el próximo Turn cierra el
// proceso viejo, encadena el ClaudeSessionID, spawnea FRESCO (sin --resume) con el
// checkpoint en el system-prompt por sesión, deja breadcrumb RolSys y el Conv no pierde
// ningún turno. Session.ID no cambia.
func TestRotacionInvisible(t *testing.T) {
	agent := &stubAgent{}
	inj := &stubInjector{}
	svc, err := usecase.NewSessionService(context.Background(), agent, stubStore{}, stubPub{}, stubResolver{path: t.TempDir()}, 40, inj, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetUmbralRotacion(40)
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}

	// Turno 1: init con ClaudeSessionID real + result que cruza el umbral.
	if err := svc.Turn(s.ID, "primer pedido"); err != nil {
		t.Fatal(err)
	}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventInit, ClaudeSessionID: "cc-viejo"}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "hecho", CtxPct: 45}
	espera(t, func() bool { m, _ := svc.Get(s.ID); return activaSinFallar(m).RotacionPendiente })

	// Turno 2: rota por detrás y sigue.
	if err := svc.Turn(s.ID, "segundo pedido"); err != nil {
		t.Fatal(err)
	}
	if len(agent.spawns) != 2 {
		t.Fatalf("spawns = %d, quiero 2 (proceso fresco)", len(agent.spawns))
	}
	if agent.spawns[1].Resume != "" {
		t.Errorf("el spawn post-rotación debe ir SIN --resume, got %q", agent.spawns[1].Resume)
	}
	m, _ := svc.Get(s.ID)
	if m.ID != s.ID {
		t.Error("Session.ID no debe cambiar")
	}
	c := activaDe(t, m)
	if len(m.Conversaciones) != 1 {
		t.Errorf("la rotación es INVISIBLE (CV-D10): sigue siendo UNA conversación, hay %d", len(m.Conversaciones))
	}
	if len(c.CadenaCC) != 1 || c.CadenaCC[0] != "cc-viejo" {
		t.Errorf("CadenaCC = %v, quiero [cc-viejo]", c.CadenaCC)
	}
	if c.RotacionPendiente {
		t.Error("la rotación debe consumir la marca")
	}
	if c.Checkpoint == "" || !strings.Contains(c.Checkpoint, "primer pedido") {
		t.Errorf("checkpoint mecánico sin los últimos turnos: %q", c.Checkpoint)
	}
	// El checkpoint viaja en el system-prompt por sesión del spawn 2.
	if len(inj.extras) != 2 || !strings.Contains(inj.extras[1], "Checkpoint de rotación") || !strings.Contains(inj.extras[1], "primer pedido") {
		t.Errorf("extras del injector: %d, último sin checkpoint: %.120q", len(inj.extras), inj.extras[len(inj.extras)-1])
	}
	// Conv íntegro + breadcrumb: user1, assistant1, sys, user2.
	roles := []string{}
	for _, tu := range c.Conv {
		roles = append(roles, string(tu.Rol))
	}
	quiero := []string{"user", "assistant", "sys", "user"}
	if len(roles) != len(quiero) {
		t.Fatalf("roles del Conv = %v, quiero %v", roles, quiero)
	}
	for i := range quiero {
		if roles[i] != quiero[i] {
			t.Fatalf("roles del Conv = %v, quiero %v", roles, quiero)
		}
	}
}

// TestRotacionEmiteFrameConTurnoIdx (E-19 · RF-313 CA-4 · H-8): la marca de rotación tiene
// que llegar EN VIVO. Antes no llegaba: `rotarLocked` no publicaba nada y el espejo del FE se
// arma con frames, así que el operador no veía la marca hasta recargar la app.
func TestRotacionEmiteFrameConTurnoIdx(t *testing.T) {
	agent := &stubAgent{}
	pub := &pubGrabador{}
	svc, err := usecase.NewSessionService(t.Context(), agent, stubStore{}, pub, stubResolver{path: t.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetUmbralRotacion(40)
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	if terr := svc.Turn(s.ID, "primer pedido"); terr != nil {
		t.Fatal(terr)
	}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventInit, ClaudeSessionID: "cc-viejo"}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "hecho", CtxPct: 45}
	espera(t, func() bool { m, _ := svc.Get(s.ID); return activaSinFallar(m).RotacionPendiente })

	if terr := svc.Turn(s.ID, "segundo pedido"); terr != nil {
		t.Fatal(terr)
	}
	frames := pub.deKind("conversacion")
	if len(frames) != 1 {
		t.Fatalf("frames de conversación = %d, quiero exactamente 1 (la rotación)", len(frames))
	}
	f := frames[0]
	if f["conversacion_evento"] != "rotada" {
		t.Errorf("evento = %v, quiero rotada", f["conversacion_evento"])
	}
	if f["text"] != "— contexto rotado, seguimos —" {
		t.Errorf("text = %q: el texto es el que YA se persiste, no el del dibujo (C-5)", f["text"])
	}
	if _, tiene := f["run_id"]; tiene {
		t.Error("el frame de rotación no pertenece a un turno: no puede llevar run_id")
	}
	if _, tiene := f["ctx_pct"]; tiene {
		t.Error("el ctx del hilo fresco llega con el result del turno nuevo, no acá")
	}
	// El índice es el del breadcrumb en el transcript: es lo que hace idempotente al frame.
	m, _ := svc.Get(s.ID)
	c := activaDe(t, m)
	idx, ok := f["turno_idx"].(float64)
	if !ok {
		t.Fatalf("turno_idx ausente o no numérico: %#v", f["turno_idx"])
	}
	if int(idx) >= len(c.Conv) || c.Conv[int(idx)].Rol != domain.RolSys {
		t.Errorf("turno_idx = %v no apunta al breadcrumb del transcript (%d turnos)", idx, len(c.Conv))
	}
	if got := f["conversacion_id"]; got != c.ID {
		t.Errorf("conversacion_id = %v, quiero la MISMA conversación %q (la rotación es invisible)", got, c.ID)
	}
}

// TestRotacionFrameLlegaAntesDelStatus: la marca se agrega al transcript ANTES del turno del
// usuario, así que su frame tiene que salir antes del `status` del turno. Al revés, la marca
// aparecería debajo del mensaje que la disparó.
func TestRotacionFrameLlegaAntesDelStatus(t *testing.T) {
	agent := &stubAgent{}
	pub := &pubGrabador{}
	svc, err := usecase.NewSessionService(t.Context(), agent, stubStore{}, pub, stubResolver{path: t.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetUmbralRotacion(40)
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	if terr := svc.Turn(s.ID, "primer pedido"); terr != nil {
		t.Fatal(terr)
	}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "hecho", CtxPct: 45}
	espera(t, func() bool { m, _ := svc.Get(s.ID); return activaSinFallar(m).RotacionPendiente })

	antes := len(pub.snapshot())
	if terr := svc.Turn(s.ID, "segundo pedido"); terr != nil {
		t.Fatal(terr)
	}
	posRotacion, posStatus := -1, -1
	for i, f := range pub.snapshot()[antes:] {
		switch {
		case f["kind"] == "conversacion" && posRotacion < 0:
			posRotacion = i
		case f["kind"] == "status" && posStatus < 0:
			posStatus = i
		}
	}
	if posRotacion < 0 || posStatus < 0 {
		t.Fatalf("faltó alguno de los dos frames: rotación=%d status=%d", posRotacion, posStatus)
	}
	if posRotacion > posStatus {
		t.Errorf("la rotación salió DESPUÉS del status (%d > %d): la marca quedaría debajo del mensaje que la disparó", posRotacion, posStatus)
	}
}

// TestRotacionNoCreaConversacionNueva (E-19, E-20 · CV-D10): rotar NO parte el hilo. La lista
// del panel muestra las mismas N conversaciones antes y después, y la activa es la misma.
func TestRotacionNoCreaConversacionNueva(t *testing.T) {
	agent := &stubAgent{}
	svc, err := usecase.NewSessionService(t.Context(), agent, stubStore{}, stubPub{}, stubResolver{path: t.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetUmbralRotacion(40)
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, cerr := svc.CrearConversacion(s.ID); cerr != nil {
		t.Fatal(cerr)
	}
	antes, _ := svc.Get(s.ID)
	activaAntes := activaDe(t, antes)

	if terr := svc.Turn(s.ID, "primer pedido"); terr != nil {
		t.Fatal(terr)
	}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "hecho", CtxPct: 45}
	espera(t, func() bool { m, _ := svc.Get(s.ID); return activaSinFallar(m).RotacionPendiente })

	// El título se toma ACÁ, no antes del primer turno: el primer turno lo DERIVA (CV-D9/
	// RF-303) y eso es otro sujeto. Lo que este test afirma es que la ROTACIÓN no lo toca,
	// así que su línea base tiene que ser el título ya derivado, inmediatamente antes de
	// rotar. Con la línea base vieja el test afirmaba «el título no cambia nunca», que era
	// verdad sólo mientras RF-303 estuviera sin construir.
	previo, _ := svc.Get(s.ID)
	tituloAntesDeRotar := activaDe(t, previo).Titulo
	if tituloAntesDeRotar == domain.TituloConversacionNueva {
		t.Fatalf("precondición: el primer turno tenía que derivar el título, sigue en %q", tituloAntesDeRotar)
	}

	if terr := svc.Turn(s.ID, "segundo pedido"); terr != nil {
		t.Fatal(terr)
	}

	m, _ := svc.Get(s.ID)
	if len(m.Conversaciones) != len(antes.Conversaciones) {
		t.Errorf("conversaciones = %d, antes %d: rotar NO parte el hilo", len(m.Conversaciones), len(antes.Conversaciones))
	}
	c := activaDe(t, m)
	if c.ID != activaAntes.ID {
		t.Errorf("la activa cambió al rotar: %q → %q", activaAntes.ID, c.ID)
	}
	if c.Titulo != tituloAntesDeRotar {
		t.Errorf("el título cambió al rotar: %q → %q", tituloAntesDeRotar, c.Titulo)
	}
}
