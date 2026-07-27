package usecase_test

// Tests de la transición atómica (T15 · RF-307/308/310/311/312/316 · CV-D7/CV-D11).
// Lo que se fija acá es que crear y retomar son UNA transición: o pasó entera, o el estado
// anterior quedó intacto — nunca una sesión con cero activas, con dos, o con el conductor
// viejo todavía atado a un hilo que ya no recibe los turnos.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// pubGrabador guarda los frames publicados para poder afirmar sobre ellos. `stubPub` los
// tira, que alcanza para los tests que sólo miran el estado.
type pubGrabador struct {
	mu     sync.Mutex
	frames []map[string]any
}

func (p *pubGrabador) Publish(_ string, data []byte) {
	var m map[string]any
	if json.Unmarshal(data, &m) != nil {
		return
	}
	p.mu.Lock()
	p.frames = append(p.frames, m)
	p.mu.Unlock()
}

func (p *pubGrabador) snapshot() []map[string]any {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]map[string]any, len(p.frames))
	copy(out, p.frames)
	return out
}

// deKind devuelve los frames de un `kind` dado, en orden de publicación.
func (p *pubGrabador) deKind(kind string) []map[string]any {
	var out []map[string]any
	for _, f := range p.snapshot() {
		if f["kind"] == kind {
			out = append(out, f)
		}
	}
	return out
}

// svcConv arma un servicio con el publicador grabador y el store dado.
func svcConv(t *testing.T, agent ports.AgentPort, store ports.SessionStore, pub usecase.EventPublisher) *usecase.SessionService {
	t.Helper()
	svc, err := usecase.NewSessionService(t.Context(), agent, store, pub, stubResolver{path: t.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return svc
}

// TestCrearConTurnoEnVueloEsErrBusy (E-07 · RF-312 CA-3): el guard se re-evalúa en el
// servidor, no se confía en que el FE haya deshabilitado el botón. Entre que el FE pintó el
// `＋` habilitado y llegó el pedido, el operador pudo mandar un turno (carrera CR-2).
func TestCrearConTurnoEnVueloEsErrBusy(t *testing.T) {
	agent := &stubAgent{}
	svc := svcConv(t, agent, stubStore{}, stubPub{})
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Turn(s.ID, "un pedido largo"); err != nil {
		t.Fatal(err)
	}
	// Sin result: la sesión quedó en streaming.
	if _, _, err := svc.CrearConversacion(s.ID); !errors.Is(err, usecase.ErrBusy) {
		t.Fatalf("crear con el turno en vuelo = %v, quiero ErrBusy (→ 409)", err)
	}
	m, _ := svc.Get(s.ID)
	if len(m.Conversaciones) != 1 {
		t.Errorf("el rechazo no puede haber creado nada: %d conversaciones", len(m.Conversaciones))
	}
}

// TestRetomarConTurnoEnVueloEsErrBusy (E-11): misma ley para retomar. Cambiar de hilo con un
// turno en vuelo dejaría dos conductores sobre el mismo stdin.
func TestRetomarConTurnoEnVueloEsErrBusy(t *testing.T) {
	agent := &stubAgent{}
	svc := svcConv(t, agent, stubStore{}, stubPub{})
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	primera := activaDe(t, s).ID
	if _, _, err := svc.CrearConversacion(s.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.Turn(s.ID, "un pedido largo"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ActivarConversacion(s.ID, primera); !errors.Is(err, usecase.ErrBusy) {
		t.Fatalf("retomar con el turno en vuelo = %v, quiero ErrBusy (→ 409)", err)
	}
	m, _ := svc.Get(s.ID)
	if c := activaDe(t, m); c.ID == primera {
		t.Error("el rechazo cambió la activa igual — la transición no fue atómica")
	}
}

// TestTransicionDeniegaPermisosPendientesConMotivo (E-08 · RF-311): el hilo que se desactiva
// no puede dejar tarjetas de permiso abiertas.
//
// El estado de partida es real, no fabricado: un `control_request` deja la sesión en `await`
// con la tarjeta pendiente, y el `result` del turno la devuelve a `idle` SIN limpiar el
// pendiente (`consume`, rama result). Ahí el guard de turno quieto deja pasar la transición y
// el pendiente sigue vivo — que es exactamente el guardrail que este test cubre.
func TestTransicionDeniegaPermisosPendientesConMotivo(t *testing.T) {
	agent := &stubAgent{}
	pub := &pubGrabador{}
	svc := svcConv(t, agent, stubStore{}, pub)
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Turn(s.ID, "editá el spec"); err != nil {
		t.Fatal(err)
	}
	agent.sessions[0].events <- ports.AgentEvent{
		Kind: ports.EventControlRequest, RequestID: "cr-1", Tool: "Write", Input: []byte(`{"file_path":"spec.md"}`),
	}
	espera(t, func() bool { m, _ := svc.Get(s.ID); return m.Status == domain.StatusAwait })
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "listo"}
	espera(t, func() bool { m, _ := svc.Get(s.ID); return m.Status == domain.StatusIdle })

	if _, _, err := svc.CrearConversacion(s.ID); err != nil {
		t.Fatalf("crear: %v", err)
	}

	var denegado map[string]any
	for _, f := range pub.deKind("permission_result") {
		if f["request_id"] == "cr-1" {
			denegado = f
		}
	}
	if denegado == nil {
		t.Fatal("la tarjeta pendiente quedó abierta: ningún permission_result para cr-1")
	}
	if denegado["decision"] != string(domain.DecisionDeny) {
		t.Errorf("decision = %v, quiero deny (deny-by-default)", denegado["decision"])
	}
	motivo, _ := denegado["text"].(string)
	if !strings.Contains(motivo, "desactivada") {
		t.Errorf("motivo = %q, quiero uno que nombre la desactivación", motivo)
	}
	if strings.Contains(motivo, "interrumpido") {
		t.Errorf("motivo = %q: tiene que ser DISTINGUIBLE del de Interrupt (RF-311 CA-1)", motivo)
	}
}

// TestFalloDePersistenciaDejaElEstadoAnterior (E-36, E-37 · RF-308 CA-2): si el disco no
// acepta la escritura, en memoria no queda lo que no se guardó — ni la conversación nueva,
// ni la desactivación de la anterior, ni la pérdida del conductor.
func TestFalloDePersistenciaDejaElEstadoAnterior(t *testing.T) {
	agent := &stubAgent{}
	st := &storeQueFalla{}
	svc := svcConv(t, agent, st, stubPub{})
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	if terr := svc.Turn(s.ID, "primer pedido"); terr != nil {
		t.Fatal(terr)
	}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventInit, ClaudeSessionID: "cc-uno"}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "hecho", CtxPct: 12}
	espera(t, func() bool { m, _ := svc.Get(s.ID); return m.Status == domain.StatusIdle })

	antes, _ := svc.Get(s.ID)
	activaAntes := activaDe(t, antes)

	st.roto = true
	if _, _, err := svc.CrearConversacion(s.ID); err == nil {
		t.Fatal("crear con el disco roto tiene que devolver el motivo del filesystem")
	}

	m, _ := svc.Get(s.ID)
	if len(m.Conversaciones) != 1 {
		t.Fatalf("quedaron %d conversaciones: el rollback no deshizo la creación", len(m.Conversaciones))
	}
	if c := activaDe(t, m); c.ID != activaAntes.ID {
		t.Errorf("la activa es %q, quiero la de antes (%q)", c.ID, activaAntes.ID)
	}

	// Y el conductor sigue siendo el mismo: un evento del proceso vivo todavía aterriza,
	// que es la prueba de que `r.live` volvió a su lugar y no se cerró nada.
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "sigo acá", CtxPct: 19}
	espera(t, func() bool { mm, _ := svc.Get(s.ID); return activaSinFallar(mm).CtxPct == 19 })
	if len(agent.spawns) != 1 {
		t.Errorf("spawns = %d: el rollback no puede haber dejado a la sesión sin conductor", len(agent.spawns))
	}
}

// TestCrearDosVecesRapidoDejaUnaActiva (E-09 · CR-1): dos `＋` casi simultáneos. `s.mu` los
// serializa; la segunda encuentra la primera activa con 0 turnos y crea otra igual — eso es
// legal y declarado. Lo que NO puede pasar es que queden dos activas o ninguna.
func TestCrearDosVecesRapidoDejaUnaActiva(t *testing.T) {
	svc := svcConv(t, &stubAgent{}, stubStore{}, stubPub{})
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range errs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, errs[i] = svc.CrearConversacion(s.ID)
		}()
	}
	wg.Wait()
	for i, e := range errs {
		if e != nil {
			t.Fatalf("crear #%d: %v", i, e)
		}
	}
	m, _ := svc.Get(s.ID)
	if len(m.Conversaciones) != 3 {
		t.Errorf("conversaciones = %d, quiero 3 (la inicial + las dos creadas)", len(m.Conversaciones))
	}
	if err := domain.VerificarUnaActiva(m); err != nil {
		t.Errorf("la invariante se rompió con dos creaciones concurrentes: %v", err)
	}
}

// TestRetomarLaActivaEsNoOp (E-15 · RF-316): elegir la que ya está activa no es un error, es
// que no hay transición que hacer. No se cierra el conductor, no se resetea el vuelo y no se
// publica ningún frame — cobrarle al operador un reinicio por ese click sería un defecto.
func TestRetomarLaActivaEsNoOp(t *testing.T) {
	agent := &stubAgent{}
	pub := &pubGrabador{}
	svc := svcConv(t, agent, stubStore{}, pub)
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	if terr := svc.Turn(s.ID, "primer pedido"); terr != nil {
		t.Fatal(terr)
	}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "hecho", CtxPct: 7}
	espera(t, func() bool { m, _ := svc.Get(s.ID); return m.Status == domain.StatusIdle })
	activa := activaDe(t, mustGet(t, svc, s.ID))

	c, desactivada, err := svc.ActivarConversacion(s.ID, activa.ID)
	if err != nil {
		t.Fatalf("retomar la activa devolvió error: %v", err)
	}
	if c.ID != activa.ID {
		t.Errorf("devolvió %q, quiero la misma activa %q", c.ID, activa.ID)
	}
	if desactivada != "" {
		t.Errorf("desactivada = %q, quiero vacío: no se desactivó nada", desactivada)
	}
	if fs := pub.deKind("conversacion"); len(fs) != 0 {
		t.Errorf("%d frame(s) de conversación para un no-op", len(fs))
	}
	if len(agent.spawns) != 1 {
		t.Errorf("spawns = %d: un no-op no re-spawnea", len(agent.spawns))
	}
	// El conductor sigue vivo: su próximo evento aterriza.
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "sigo", CtxPct: 21}
	espera(t, func() bool { m, _ := svc.Get(s.ID); return activaSinFallar(m).CtxPct == 21 })
}

// TestFramesDelConductorViejoSeDescartan (CR-3): después de la transición, lo que emita el
// proceso anterior no puede escribirse en el hilo nuevo. Lo resuelve código que ya existía —
// `transicionLocked` deja `r.live` en nil y los siete tramos de `consume` comparan
// `r.live == live` antes de tocar nada.
//
// ⚠ Lo que este test NO afirma, y hay que decirlo: los `s.publish` de `consume` están FUERA
// de ese guard (preexistente, `session_service.go`), así que un evento tardío del proceso
// viejo igual sale por el SSE — con `run_id` vacío. No muta estado, que es lo que acá se
// fija; el frame huérfano queda declarado en PARIDAD, no tapado.
func TestFramesDelConductorViejoSeDescartan(t *testing.T) {
	agent := &stubAgent{}
	svc := svcConv(t, agent, stubStore{}, stubPub{})
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	if terr := svc.Turn(s.ID, "primer pedido"); terr != nil {
		t.Fatal(terr)
	}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventInit, ClaudeSessionID: "cc-viejo"}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "hecho", CtxPct: 12}
	espera(t, func() bool { m, _ := svc.Get(s.ID); return m.Status == domain.StatusIdle })

	nueva, _, err := svc.CrearConversacion(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if nueva.Turnos != 0 || nueva.CtxPct != 0 {
		t.Fatalf("la nueva nace con %d turnos y ctx %d, quiero 0 y 0", nueva.Turnos, nueva.CtxPct)
	}

	// El proceso viejo sigue hablando: nada de esto puede caer en el hilo nuevo.
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "respuesta del hilo viejo", CtxPct: 88}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventActivity, Tool: "Write", Text: "spec.md"}
	espera(t, func() bool { return len(agent.sessions) == 1 })

	m, _ := svc.Get(s.ID)
	c := activaDe(t, m)
	if c.ID != nueva.ID {
		t.Fatalf("la activa cambió sola: %q", c.ID)
	}
	if len(c.Conv) != 0 {
		t.Errorf("el hilo nuevo tiene %d turnos: los escribió el conductor viejo", len(c.Conv))
	}
	if c.CtxPct != 0 {
		t.Errorf("ctx del hilo nuevo = %d: lo movió el conductor viejo", c.CtxPct)
	}
	if m.Status != domain.StatusIdle {
		t.Errorf("status = %q: lo movió el conductor viejo", m.Status)
	}
}

// TestRenombrarVacioNoTocaElTitulo (E-30 · RF-344 CA-1) y su contraparte feliz: el renombrado
// no es una transición y no exige turno quieto — no cambia quién tiene el conductor.
func TestRenombrarNoEsUnaTransicion(t *testing.T) {
	agent := &stubAgent{}
	pub := &pubGrabador{}
	svc := svcConv(t, agent, stubStore{}, pub)
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	cid := activaDe(t, s).ID
	if terr := svc.Turn(s.ID, "un pedido largo"); terr != nil {
		t.Fatal(terr)
	}
	// Con la sesión en streaming: renombrar SÍ se puede.
	c, err := svc.RenombrarConversacion(s.ID, cid, "  el bug del índice  ")
	if err != nil {
		t.Fatalf("renombrar con un turno en vuelo: %v", err)
	}
	if c.Titulo != "el bug del índice" || !c.TituloEditado {
		t.Errorf("titulo = %q editado = %v", c.Titulo, c.TituloEditado)
	}
	if _, err := svc.RenombrarConversacion(s.ID, cid, "   "); !errors.Is(err, domain.ErrTituloVacio) {
		t.Fatalf("renombrar a vacío = %v, quiero ErrTituloVacio (→ 400)", err)
	}
	m, _ := svc.Get(s.ID)
	if got := activaDe(t, m).Titulo; got != "el bug del índice" {
		t.Errorf("el título cambió con un renombrado rechazado: %q", got)
	}
	if _, err := svc.RenombrarConversacion(s.ID, "cv-ajena", "otra"); !errors.Is(err, domain.ErrConvNoEncontrada) {
		t.Errorf("renombrar una conversación ajena tiene que ser ErrConvNoEncontrada (→ 404)")
	}
	if fs := pub.deKind("conversacion"); len(fs) != 1 || fs[0]["conversacion_evento"] != "renombrada" {
		t.Errorf("frames de conversación = %+v, quiero exactamente uno `renombrada`", fs)
	}
}

// TestSoloLecturaBloqueaLasTresOperaciones (RF-338 · T18 lo mapea a 503): con el registro en
// solo-lectura —el archivo en disco lo escribió un binario más nuevo— ninguna de las tres
// mutaciones se intenta siquiera. No es culpa del pedido: es que el daemon no puede escribir,
// y decirlo es mejor que aceptar un cambio que se va a perder al salir.
func TestSoloLecturaBloqueaLasTresOperaciones(t *testing.T) {
	st := &storeBloqueado{}
	st.sesiones = []domain.Session{{
		ID: "s1", Arnes: "vitalia", Frente: "el frente de siempre",
		Conversaciones: []domain.Conversacion{
			{ID: "cv1", Titulo: "una", Activa: true, Conv: []domain.Turn{}},
			{ID: "cv2", Titulo: "otra", Conv: []domain.Turn{}},
		},
	}}
	svc := svcConv(t, &stubAgent{}, st, stubPub{})
	if bloqueado, motivo := svc.SoloLectura(); !bloqueado || motivo == "" {
		t.Fatalf("SoloLectura() = %v %q, quiero bloqueado con motivo", bloqueado, motivo)
	}
	if _, _, err := svc.CrearConversacion("s1"); !errors.Is(err, usecase.ErrSoloLectura) {
		t.Errorf("crear = %v, quiero ErrSoloLectura", err)
	}
	if _, _, err := svc.ActivarConversacion("s1", "cv2"); !errors.Is(err, usecase.ErrSoloLectura) {
		t.Errorf("retomar = %v, quiero ErrSoloLectura", err)
	}
	if _, err := svc.RenombrarConversacion("s1", "cv1", "otro nombre"); !errors.Is(err, usecase.ErrSoloLectura) {
		t.Errorf("renombrar = %v, quiero ErrSoloLectura", err)
	}
	m, _ := svc.Get("s1")
	if len(m.Conversaciones) != 2 || activaDe(t, m).ID != "cv1" {
		t.Errorf("el registro bloqueado mutó igual: %+v", m.Conversaciones)
	}
}

func mustGet(t *testing.T, svc *usecase.SessionService, id string) domain.Session {
	t.Helper()
	m, ok := svc.Get(id)
	if !ok {
		t.Fatalf("la sesión %q desapareció", id)
	}
	return m
}

// TestRecalibracionLlegaAlRegistroVivo (N-24): el re-key del arranque escribe el disco
// DESPUÉS de que el servicio cargó el registro; sin puente, la memoria se queda con las
// llaves viejas y la API sigue escondiendo las sesiones hasta el arranque siguiente.
//
// Medido contra el binario antes de existir este test: 1er arranque servía 2 de 5
// sesiones del operador, 2º arranque las 5. El operador estrenaba la función viendo el
// bug que la función arregla.
func TestRecalibracionLlegaAlRegistroVivo(t *testing.T) {
	previas := &memSessionStore{sesiones: []domain.Session{
		{ID: "s1", Arnes: "vitalia", Conversaciones: []domain.Conversacion{{ID: "cv1", Activa: true, Conv: []domain.Turn{}}}},
		{ID: "s2", Arnes: "sin-home~vitalia~vitalia", Conversaciones: []domain.Conversacion{{ID: "cv2", Activa: true, Conv: []domain.Turn{}}}},
		{ID: "s3", Arnes: "vitalia", Conversaciones: []domain.Conversacion{{ID: "cv3", Activa: true, Conv: []domain.Turn{}}}},
	}}
	svc, err := usecase.NewSessionService(context.Background(), &stubAgent{}, previas, stubPub{}, stubResolver{path: t.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	porArnes := func(clave string) int {
		n := 0
		for _, s := range svc.List() {
			if s.Arnes == clave {
				n++
			}
		}
		return n
	}
	if got := porArnes("sin-home~vitalia~vitalia"); got != 1 {
		t.Fatalf("antes de recalibrar la clave calificada trae %d, want 1", got)
	}

	// Lo que el store ya escribió en disco, aplicado a memoria.
	movidas := svc.AplicarRecalibracion(map[string]string{
		"s1": "sin-home~vitalia~vitalia",
		"s3": "sin-home~vitalia~vitalia",
		"s2": "sin-home~vitalia~vitalia", // ya calificada: no cuenta como movida.
	})
	if movidas != 2 {
		t.Errorf("movidas = %d, want 2 (s2 ya estaba calificada)", movidas)
	}
	if got := porArnes("sin-home~vitalia~vitalia"); got != 3 {
		t.Errorf("tras recalibrar la clave calificada trae %d, want 3 — el registro vivo no se enteró", got)
	}
	if got := porArnes("vitalia"); got != 0 {
		t.Errorf("quedan %d sesiones con la llave pelada, want 0", got)
	}

	// Idempotente: correrla de nuevo no mueve nada.
	if again := svc.AplicarRecalibracion(map[string]string{"s1": "sin-home~vitalia~vitalia"}); again != 0 {
		t.Errorf("segunda corrida movió %d, want 0 (idempotencia)", again)
	}
	// Una sesión que no existe no explota ni inventa.
	if ghost := svc.AplicarRecalibracion(map[string]string{"s-fantasma": "x"}); ghost != 0 {
		t.Errorf("sesión inexistente movió %d, want 0", ghost)
	}
}
