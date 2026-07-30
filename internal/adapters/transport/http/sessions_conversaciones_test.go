package httpapi

// Tests del transporte del panel de conversaciones (T18 · RF-340…RF-345).
//
// Cierran el gap más grande del lado Go de este adaptador: hasta este ticket `sessions_test.go`
// ejercitaba UNO de los diez endpoints de sesión. Lo que se prueba acá no es el usecase (eso
// vive en su propio paquete) sino lo que sólo el transporte puede equivocar: el código de
// estado, la forma de la respuesta y a qué handler llega cada ruta.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// --- lo mínimo para tener un conductor que se quede hablando ---

type convSession struct{ events chan ports.AgentEvent }

func (s *convSession) Send(context.Context, string) error { return nil }
func (s *convSession) Events() <-chan ports.AgentEvent    { return s.events }
func (s *convSession) Close() error                       { return nil }
func (s *convSession) Interrupt(context.Context) error    { return nil }
func (s *convSession) RespondControl(context.Context, string, ports.ControlDecision) error {
	return nil
}

type convAgent struct {
	mu       sync.Mutex
	sessions []*convSession
}

func (a *convAgent) Spawn(context.Context, ports.SpawnOpts) (ports.AgentSession, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := &convSession{events: make(chan ports.AgentEvent, 8)}
	a.sessions = append(a.sessions, s)
	return s, nil
}

type convResolver struct{}

func (convResolver) Resolve(string) (string, bool, error) { return "/tmp/arnes-conv", true, nil }

func svcHTTP(t *testing.T) (*usecase.SessionService, string) {
	t.Helper()
	svc, err := usecase.NewSessionService(t.Context(), &convAgent{}, nilStore{}, nilPub{}, convResolver{}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "repro del bug de carga"})
	if err != nil {
		t.Fatal(err)
	}
	return svc, s.ID
}

func pedirConv(t *testing.T, h http.HandlerFunc, metodo, url, cuerpo string, vars map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var body *strings.Reader
	if cuerpo == "" {
		body = strings.NewReader("")
	} else {
		body = strings.NewReader(cuerpo)
	}
	req := httptest.NewRequestWithContext(t.Context(), metodo, url, body)
	for k, v := range vars {
		req.SetPathValue(k, v)
	}
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func cuerpoJSON(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("la respuesta no es un objeto JSON (%d): %s", rec.Code, rec.Body.String())
	}
	return m
}

// TestListarDevuelve404SiLaSesionNoExiste (E-35 · RF-340 CA-3 · BR-CV-10): jamás un 200 con
// lista vacía para una sesión que no existe. «Vacío» y «no existe» son dos cosas, y
// confundirlas deja al operador leyendo «no hay conversaciones» cuando el problema es otro.
func TestListarDevuelve404SiLaSesionNoExiste(t *testing.T) {
	svc, id := svcHTTP(t)
	h := listarConversaciones(svc)

	rec := pedirConv(t, h, http.MethodGet, "/api/sessions/x/conversaciones", "", map[string]string{"id": "no-existe"})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, quiero 404: %s", rec.Code, rec.Body.String())
	}

	rec = pedirConv(t, h, http.MethodGet, "/api/sessions/x/conversaciones", "", map[string]string{"id": id})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, quiero 200: %s", rec.Code, rec.Body.String())
	}
	m := cuerpoJSON(t, rec)
	if m["total"] != float64(1) {
		t.Errorf("total = %v, quiero 1", m["total"])
	}
	lista, _ := m["conversaciones"].([]any)
	if len(lista) != 1 {
		t.Fatalf("conversaciones = %v", m["conversaciones"])
	}
	if _, trae := lista[0].(map[string]any)["conv"]; trae {
		t.Error("la lista NO manda el transcript: son 12 KB por conversación que nadie pidió")
	}
}

// TestCrearDevuelve409ConTurnoEnVuelo (E-07, E-39 · RF-342 CA-2): el mismo código que un
// turno concurrente. No es un fallo del pedido, es un conflicto con algo que ya está pasando.
func TestCrearDevuelve409ConTurnoEnVuelo(t *testing.T) {
	svc, id := svcHTTP(t)
	h := crearConversacion(svc)

	rec := pedirConv(t, h, http.MethodPost, "/api/sessions/x/conversaciones", "", map[string]string{"id": id})
	if rec.Code != http.StatusCreated {
		t.Fatalf("crear con la sesión quieta: status = %d: %s", rec.Code, rec.Body.String())
	}
	m := cuerpoJSON(t, rec)
	if _, ok := m["nueva"].(map[string]any); !ok {
		t.Errorf("la respuesta no trae `nueva`: %s", rec.Body.String())
	}
	if m["desactivada"] == nil {
		t.Error("había una activa: `desactivada` tiene que nombrarla")
	}

	if err := svc.Turn(id, "un pedido largo"); err != nil {
		t.Fatal(err)
	}
	rec = pedirConv(t, h, http.MethodPost, "/api/sessions/x/conversaciones", "", map[string]string{"id": id})
	if rec.Code != http.StatusConflict {
		t.Fatalf("crear con el turno en vuelo: status = %d, quiero 409: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "streaming") {
		t.Errorf("el 409 tiene que decir POR QUÉ: %s", rec.Body.String())
	}
}

// TestCrearSinActivaDevuelveDesactivadaNula: `""` y «no había ninguna» no pueden viajar igual.
func TestCrearPrimeraConversacionNoInventaDesactivada(t *testing.T) {
	svc, err := usecase.NewSessionService(t.Context(), &convAgent{}, nilStore{}, nilPub{}, convResolver{}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Una sesión que llegó del disco SIN conversaciones (registro pre-migración): la
	// normalización de carga le crea una, así que la creación siguiente sí desactiva. Lo que
	// se fija acá es la forma del campo, no el caso.
	s, err := svc.Create(domain.Session{Arnes: "vitalia"})
	if err != nil {
		t.Fatal(err)
	}
	rec := pedirConv(t, crearConversacion(svc), http.MethodPost, "/api/sessions/x/conversaciones", "", map[string]string{"id": s.ID})
	m := cuerpoJSON(t, rec)
	if _, presente := m["desactivada"]; !presente {
		t.Error("`desactivada` tiene que estar SIEMPRE presente: su ausencia y su nulo dirían lo mismo")
	}
}

// TestActivarConvAjenaDevuelve404 (RF-343 CA-3 · BR-CV-2): una conversación cuelga de SU
// sesión. Un id que no es de esta sesión no existe para esta sesión — nunca se busca global.
func TestActivarConvAjenaDevuelve404(t *testing.T) {
	svc, id := svcHTTP(t)
	rec := pedirConv(t, activarConversacion(svc), http.MethodPost, "/api/sessions/x/conversaciones/y/activar", "",
		map[string]string{"id": id, "cid": "cv-de-otra-sesion"})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, quiero 404: %s", rec.Code, rec.Body.String())
	}
}

// TestRenombrarVacioDevuelve400YNoCambia (E-30 · RF-344 CA-1): el 400 no es sólo un código —
// el título anterior tiene que seguir ahí.
func TestRenombrarVacioDevuelve400YNoCambia(t *testing.T) {
	svc, id := svcHTTP(t)
	sess, _ := svc.Get(id)
	cid := sess.Conversaciones[0].ID
	h := renombrarConversacion(svc)

	rec := pedirConv(t, h, http.MethodPatch, "/api/sessions/x/conversaciones/y", `{"titulo":"el bug del índice"}`,
		map[string]string{"id": id, "cid": cid})
	if rec.Code != http.StatusOK {
		t.Fatalf("renombrar: status = %d: %s", rec.Code, rec.Body.String())
	}
	if m := cuerpoJSON(t, rec); m["titulo"] != "el bug del índice" || m["titulo_editado"] != true {
		t.Errorf("respuesta = %s", rec.Body.String())
	}

	rec = pedirConv(t, h, http.MethodPatch, "/api/sessions/x/conversaciones/y", `{"titulo":"   "}`,
		map[string]string{"id": id, "cid": cid})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("título vacío: status = %d, quiero 400: %s", rec.Code, rec.Body.String())
	}
	sess, _ = svc.Get(id)
	if got := sess.Conversaciones[0].Titulo; got != "el bug del índice" {
		t.Errorf("el título cambió con un renombrado rechazado: %q", got)
	}

	// Sin `titulo` en el cuerpo tampoco se adivina.
	rec = pedirConv(t, h, http.MethodPatch, "/api/sessions/x/conversaciones/y", `{}`,
		map[string]string{"id": id, "cid": cid})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("cuerpo sin `titulo`: status = %d, quiero 400", rec.Code)
	}
}

// TestCerradas1Devuelve400ConPuntero (RF-345 CA-1): el parámetro se retiró. Un 400 que dice
// a dónde ir es mejor que ignorarlo en silencio (el cliente cree que preguntó bien) o que
// devolver 200 con una lista vacía (el cliente cree que no hay nada).
func TestCerradas1Devuelve400ConPuntero(t *testing.T) {
	svc, _ := svcHTTP(t)
	rec := pedirConv(t, listSessions(svc), http.MethodGet, "/api/sessions?cerradas=1", "", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, quiero 400: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "/conversaciones") {
		t.Errorf("el 400 tiene que traer el puntero al endpoint nuevo: %s", rec.Body.String())
	}
}

// TestListSessionsSiempreEsArray (check `forma-de-respuesta-unica`): era la única respuesta
// bimorfa del daemon — la misma ruta devolvía una lista o un objeto según un query. Una
// operación así es indocumentable: no hay un `200` que describir, hay dos.
func TestListSessionsSiempreEsArray(t *testing.T) {
	svc, _ := svcHTTP(t)
	h := listSessions(svc)
	for _, url := range []string{"/api/sessions", "/api/sessions?arnes=vitalia", "/api/sessions?arnes=no-existe"} {
		rec := pedirConv(t, h, http.MethodGet, url, "", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d", url, rec.Code)
		}
		var arr []map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &arr); err != nil {
			t.Errorf("%s: la respuesta no es un array: %s", url, rec.Body.String())
		}
	}
}

// TestLaSesionQueViajaLlevaSuActivaYNoSusNConversaciones (RF-300 CA-2 · §6.1): la sesión del
// wire no es el agregado serializado. Manda la conversación ACTIVA con su transcript y NO las
// N con los suyos — eso costaría doce kilobytes por conversación en cada listado del rail.
func TestLaSesionQueViajaLlevaSuActivaYNoSusNConversaciones(t *testing.T) {
	svc, id := svcHTTP(t)
	if _, _, err := svc.CrearConversacion(id); err != nil {
		t.Fatal(err)
	}
	// `?arnes=vitalia` deja fuera las tres sesiones de la semilla ilustrativa del primer
	// arranque, que tienen sus propios arneses.
	rec := pedirConv(t, listSessions(svc), http.MethodGet, "/api/sessions?arnes=vitalia", "", nil)
	var arr []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &arr); err != nil || len(arr) != 1 {
		t.Fatalf("respuesta = %s", rec.Body.String())
	}
	s := arr[0]
	if _, trae := s["conversaciones"]; trae {
		t.Error("la sesión del wire NO lleva sus N conversaciones: la lista vive en su propio endpoint")
	}
	activa, ok := s["activa"].(map[string]any)
	if !ok {
		t.Fatalf("la sesión del wire no trae `activa`: %s", rec.Body.String())
	}
	if activa["titulo"] != domain.TituloConversacionNueva {
		t.Errorf("titulo de la activa = %v", activa["titulo"])
	}
	if activa["turnos"] != float64(0) || activa["ctx_pct"] != float64(0) {
		t.Errorf("una conversación nueva tiene 0 turnos y ctx 0, y los DOS son datos: %v", activa)
	}
	conv, presente := activa["conv"]
	if !presente {
		t.Error("`conv` sin `omitempty`: `[]` y «no la cargué» no pueden serializarse igual")
	}
	if lista, _ := conv.([]any); len(lista) != 0 {
		t.Errorf("conv = %v, quiero []", conv)
	}
	// Los campos que BAJARON a la conversación ya no están en la sesión (RF-300 CA-2).
	for _, campo := range []string{"claude_session_id", "model", "ctx_pct", "conv", "turnos", "cadena_cc", "checkpoint"} {
		if _, trae := s[campo]; trae {
			t.Errorf("la sesión del wire todavía trae %q: bajó a la conversación", campo)
		}
	}
}

// TestElMuxNoAmbiguaEntreSesionYConversaciones: la ruta específica gana. Se verifica sobre el
// router REAL, no sobre un mux armado en el test: lo que puede romperse es el registro, y un
// mux de mentira no lo tocaría.
func TestElMuxNoAmbiguaEntreSesionYConversaciones(t *testing.T) {
	svc, id := svcHTTP(t)
	// `events` no puede ser nil (el mux lo registra sin guarda) y `ui` sí: sin UI embebida
	// el catch-all responde 404 con motivo, que es justo lo que la ruta retirada tiene que
	// devolver — o sirve, o 404; nunca 200 vacío.
	sinEventos := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := NewHandler(nil, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, sinEventos,
		AuthConfigFor("127.0.0.1:4200", ""))

	casos := []struct {
		metodo, ruta string
		quiero       int
	}{
		{http.MethodGet, "/api/sessions/" + id, http.StatusOK},
		{http.MethodGet, "/api/sessions/" + id + "/conversaciones", http.StatusOK},
		{http.MethodPost, "/api/sessions/" + id + "/conversaciones", http.StatusCreated},
		{http.MethodPatch, "/api/sessions/" + id + "/conversaciones/no-existe", http.StatusNotFound},
		{http.MethodPost, "/api/sessions/" + id + "/conversaciones/no-existe/activar", http.StatusNotFound},
		// La ruta retirada (RF-345 CA-2): ya no la sirve nadie. Cae en el catch-all de la
		// SPA, que sin UI embebida responde 404 — o sirve, o 404 con motivo; nunca 200 vacío.
		{http.MethodGet, "/api/sessions/cerradas/" + id + "/historial", http.StatusNotFound},
	}
	for _, c := range casos {
		// Cuerpo válido para el PATCH: lo que se mide acá es a qué handler LLEGA cada
		// ruta, y un 400 por el cuerpo taparía el 404 del handler correcto.
		req := httptest.NewRequestWithContext(t.Context(), c.metodo, c.ruta, strings.NewReader(`{"titulo":"x"}`))
		req.Host = "127.0.0.1:4200"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != c.quiero {
			t.Errorf("%s %s → %d, quiero %d: %s", c.metodo, c.ruta, rec.Code, c.quiero, rec.Body.String())
		}
	}
}
