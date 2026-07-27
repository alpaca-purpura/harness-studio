package usecase

// Las tres operaciones que el operador tiene sobre las conversaciones de una sesión: crear,
// retomar y renombrar (CV-D2/CV-D7/CV-D11). Las dos primeras comparten un cuerpo —
// `transicionLocked`— porque son la MISMA transición: dejar de tener conductor en un hilo y
// pasar a tenerlo en otro. Escribirlas por separado sería tener dos versiones de la misma
// invariante, y una de las dos se iba a quedar atrás.
//
// La disciplina del servicio se respeta sin excepción: bajo `s.mu` sólo hay mutación pura y
// persistencia; el `Close()` del conductor viejo y los `publish` ocurren DESPUÉS de soltar el
// candado. Por eso `transicionLocked` no publica: DEVUELVE lo que hay que publicar.

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// kindConversacion es el `kind` del frame del panel de conversaciones.
const kindConversacion = "conversacion"

// Los cuatro eventos del frame `conversacion`.
const (
	eventoCreada     = "creada"
	eventoActivada   = "activada"
	eventoRenombrada = "renombrada"
	eventoRotada     = "rotada"
)

// motivoDesactivada es el motivo con el que se deniegan los permisos que quedaron pendientes
// en la conversación que se desactiva. Es DISTINGUIBLE del de `Interrupt` («interrumpido por
// el operador», RF-311 CA-1) a propósito: el operador tiene que poder leer en el rastro cuál
// de las dos cosas pasó.
const motivoDesactivada = "conversación desactivada por el operador"

// ConversacionResumen es lo que viaja en la LISTA. No tiene los turnos — y no por omisión,
// sino porque el tipo no los declara: así «no cargado» ni siquiera es expresable acá
// (boundary no-aplica-no-es-cero). Son 12 KB por conversación que la lista no necesita.
type ConversacionResumen struct {
	ID                string `json:"id"`
	Titulo            string `json:"titulo"`
	TituloEditado     bool   `json:"titulo_editado"`
	Activa            bool   `json:"activa"`
	Turnos            int    `json:"turnos"`             // derivado de los turnos, jamás persistido aparte.
	CtxPct            int    `json:"ctx_pct"`            // 0 es dato, no ausencia (BR-CV-9).
	RotacionPendiente bool   `json:"rotacion_pendiente"` // el `caliente` del chip sale de acá.
	UltimaInteraccion string `json:"ultima_interaccion,omitempty"`
	CreadaEn          string `json:"creada_en"`
	ClaudeSessionID   string `json:"claude_session_id,omitempty"`
	Model             string `json:"model,omitempty"`
	Fragmento         string `json:"fragmento,omitempty"` // sólo con ?q=
}

// ConversacionActiva es el resumen MÁS el registro liviano de turnos. Es la única forma que
// lo trae, y viaja en dos lugares: dentro de una sesión y en el frame de la transición, para
// que el FE repinte sin una segunda vuelta.
type ConversacionActiva struct {
	ConversacionResumen
	Conv []domain.Turn `json:"conv"` // SIN omitempty: `[]` y `null` no pueden significar lo mismo.
	// CadenaCC son los ids de Claude Code previos de este hilo (sus rotaciones).
	CadenaCC []string `json:"cadena_cc,omitempty"`
}

// resumenDe proyecta una conversación del dominio a lo que viaja en la lista.
// El `Checkpoint` NO viaja en ningún DTO: es el digest del system-prompt, ninguna superficie
// lo pinta, y mandarlo sería exportar contenido de la conversación por un cable que nadie lee.
func resumenDe(c domain.Conversacion) ConversacionResumen {
	return ConversacionResumen{
		ID:                c.ID,
		Titulo:            c.Titulo,
		TituloEditado:     c.TituloEditado,
		Activa:            c.Activa,
		Turnos:            c.NumTurnos(),
		CtxPct:            c.CtxPct,
		RotacionPendiente: c.RotacionPendiente,
		UltimaInteraccion: c.UltimaInteraccion,
		CreadaEn:          c.CreadaEn,
		ClaudeSessionID:   c.ClaudeSessionID,
		Model:             c.Model,
	}
}

// ConversacionActivaDe proyecta una conversación con sus turnos. `Conv` nunca sale `nil`: una
// conversación sin turnos tiene una lista vacía, que es un dato distinto de «no la cargué».
// Exportada porque el transporte la necesita para proyectar la sesión al wire — la sesión que
// viaja lleva SU activa, no sus N conversaciones con los transcripts completos.
func ConversacionActivaDe(c domain.Conversacion) ConversacionActiva {
	turnos := c.Conv
	if turnos == nil {
		turnos = []domain.Turn{}
	}
	return ConversacionActiva{ConversacionResumen: resumenDe(c), Conv: turnos, CadenaCC: c.CadenaCC}
}

// ErrSinConversacionActiva — la sesión quedó sin activa después de una operación que tenía
// que dejar exactamente una. Es la invariante rota, y por eso la transición revierte.
var ErrSinConversacionActiva = errors.New("la sesión quedó sin conversación activa")

// CrearConversacion abre una conversación nueva en la sesión y desactiva la anterior, en UNA
// transición (RF-307/RF-308, BR-CV-4). Devuelve la nueva y el id de la que quedó inactiva
// ("" si no había ninguna). No spawnea: el conductor arranca perezoso, en el primer turno.
func (s *SessionService) CrearConversacion(id string) (ConversacionActiva, string, error) {
	var desactivada string
	return s.operarConversacion(id, func(sess *domain.Session) error {
		_, d := sess.CrearConversacion(domain.NuevoConvID(), time.Now().UTC())
		desactivada = d
		return nil
	}, &desactivada)
}

// ActivarConversacion retoma una conversación inactiva de ESTA sesión (RF-310, CV-D11):
// pasa a activa, la anterior a inactiva, y su transcript queda listo para repintarse. Sobre
// la que ya está activa es un no-op declarado (RF-316/E-15): ni cierra el conductor, ni
// resetea el estado de vuelo, ni persiste — no hay transición que hacer.
func (s *SessionService) ActivarConversacion(id, cid string) (ConversacionActiva, string, error) {
	var desactivada string
	return s.operarConversacion(id, func(sess *domain.Session) error {
		d, err := sess.ActivarConversacion(cid)
		desactivada = d
		return err
	}, &desactivada)
}

// operarConversacion es el envoltorio de bloqueo compartido por crear y retomar: toma el
// candado, corre la transición, lo suelta, y recién entonces cierra el conductor viejo y
// publica. `desactivada` lo escribe el propio `aplicar`; se pasa por puntero para que el
// llamador lo lea después de que la transición haya corrido.
func (s *SessionService) operarConversacion(id string, aplicar func(*domain.Session) error, desactivada *string) (ConversacionActiva, string, error) {
	s.mu.Lock()
	r := s.rt[id]
	if r == nil {
		s.mu.Unlock()
		return ConversacionActiva{}, "", errNotFound(id)
	}
	viejo, frames, err := s.transicionLocked(r, aplicar)
	if err != nil {
		s.mu.Unlock()
		return ConversacionActiva{}, "", err
	}
	activa, ok := r.meta.Activa()
	if !ok {
		s.mu.Unlock()
		return ConversacionActiva{}, "", fmt.Errorf("%w: %q", ErrSinConversacionActiva, id)
	}
	salida := ConversacionActivaDe(*activa)
	s.mu.Unlock()

	// Fuera del candado, y en este orden: el proceso viejo muere DESPUÉS de que el disco
	// confirmó (paso 7), así que un fallo de persistencia jamás deja al operador sin
	// conductor. Y los frames salen en el orden en que se armaron.
	if viejo != nil {
		_ = viejo.Close()
	}
	for _, f := range frames {
		s.publish(f)
	}
	return salida, *desactivada, nil
}

// RenombrarConversacion aplica un título escrito por el operador (RF-344). NO es una
// transición: no cambia quién tiene el conductor, así que no exige turno quieto, no cierra
// nada y no resetea el estado de vuelo. Un título vacío devuelve ErrTituloVacio y el título
// anterior queda intacto (E-30).
func (s *SessionService) RenombrarConversacion(id, cid, titulo string) (ConversacionResumen, error) {
	s.mu.Lock()
	r := s.rt[id]
	if r == nil {
		s.mu.Unlock()
		return ConversacionResumen{}, errNotFound(id)
	}
	if bloqueado, motivo := s.soloLecturaLocked(); bloqueado {
		s.mu.Unlock()
		return ConversacionResumen{}, fmt.Errorf("%w: %s", ErrSoloLectura, motivo)
	}
	previo := r.meta.Instantanea()
	if err := r.meta.RenombrarConversacion(cid, titulo); err != nil {
		s.mu.Unlock()
		return ConversacionResumen{}, err
	}
	if err := s.persistLocked(); err != nil {
		*r.meta = previo
		s.mu.Unlock()
		return ConversacionResumen{}, err
	}
	c, _ := r.meta.BuscarConversacion(cid)
	salida := ConversacionActivaDe(*c)
	s.mu.Unlock()

	s.publish(dockFrame{
		SessionID:          id,
		Kind:               kindConversacion,
		ConversacionID:     cid,
		ConversacionEvento: eventoRenombrada,
		Conversacion:       &salida,
	})
	return salida.ConversacionResumen, nil
}

// estadoDeVuelo es la foto del estado por-turno del runtime. Existe para que el rollback del
// paso 7 sea de verdad total: revertir el agregado y dejar los buffers a medio resetear
// sería «el estado anterior quedó intacto» mentido a medias.
type estadoDeVuelo struct {
	pendingTurn   string
	wasResume     bool
	sawInit       bool
	resumeRetried bool
	msgFlushed    bool
	ensamblado    string
	pendingPerm   map[string]pendingPermission
	grants        map[string]domain.Grant
	convActiva    string
	live          ports.AgentSession
}

func capturarVuelo(r *sessionRuntime) estadoDeVuelo {
	return estadoDeVuelo{
		pendingTurn:   r.pendingTurn,
		wasResume:     r.wasResume,
		sawInit:       r.sawInit,
		resumeRetried: r.resumeRetried,
		msgFlushed:    r.msgFlushed,
		ensamblado:    r.assembling.String(),
		pendingPerm:   r.pendingPerm,
		grants:        r.grants,
		convActiva:    r.convActiva,
		live:          r.live,
	}
}

func restaurarVuelo(r *sessionRuntime, e estadoDeVuelo) {
	r.pendingTurn = e.pendingTurn
	r.wasResume = e.wasResume
	r.sawInit = e.sawInit
	r.resumeRetried = e.resumeRetried
	r.msgFlushed = e.msgFlushed
	r.assembling.Reset()
	r.assembling.WriteString(e.ensamblado)
	r.pendingPerm = e.pendingPerm
	r.grants = e.grants
	r.convActiva = e.convActiva
	r.live = e.live
}

// transicionLocked es el cuerpo compartido de crear y retomar. Caller holds s.mu.
//
// Devuelve el conductor a cerrar (el llamador lo cierra DESPUÉS de soltar el candado) y los
// frames a publicar (ídem). Un fallo nunca deja media transición: los únicos pasos que
// pueden fallar son el guard —antes de mutar nada—, la operación pura —rollback trivial— y
// la persistencia —rollback del snapshot—.
func (s *SessionService) transicionLocked(
	r *sessionRuntime, aplicar func(*domain.Session) error,
) (viejo ports.AgentSession, frames []dockFrame, err error) {
	// 1 · GUARD, re-evaluado acá y no en el cliente: entre que el FE pintó el botón y
	// llegó el pedido, el operador pudo mandar un turno (CR-2).
	if r.meta.Status == domain.StatusStreaming || r.meta.Status == domain.StatusAwait {
		return nil, nil, ErrBusy
	}
	if bloqueado, motivo := s.soloLecturaLocked(); bloqueado {
		return nil, nil, fmt.Errorf("%w: %s", ErrSoloLectura, motivo)
	}

	// 2 · SNAPSHOT de rollback: el agregado y el estado de vuelo.
	previo := r.meta.Instantanea()
	vuelo := capturarVuelo(r)
	antes := map[string]bool{}
	antesID := ""
	for i := range r.meta.Conversaciones {
		antes[r.meta.Conversaciones[i].ID] = true
		if r.meta.Conversaciones[i].Activa {
			antesID = r.meta.Conversaciones[i].ID
		}
	}

	// 3 · la operación PURA del dominio. Es la única que decide si el pedido es legal.
	if aerr := aplicar(r.meta); aerr != nil {
		*r.meta = previo
		return nil, nil, aerr
	}
	nueva, ok := r.meta.Activa()
	if !ok {
		*r.meta = previo
		return nil, nil, fmt.Errorf("%w: %q", ErrSinConversacionActiva, r.meta.ID)
	}
	if nueva.ID == antesID {
		// No hubo transición (RF-316/E-15): retomar la que ya estaba activa. No se cierra
		// el conductor ni se resetea nada — hacerlo sería cobrarle al operador un
		// reinicio por un click que no pidió cambiar de hilo.
		return nil, nil, nil
	}

	// 4 · los permisos que quedaron pendientes en la que se desactiva se deniegan CON
	// MOTIVO. Precedente literal: Interrupt (session_service.go, sus permission_result con
	// decision deny). La diferencia con Interrupt: allá el proceso sobrevive y hay que
	// contestarle por el canal de control; acá el proceso se cierra al soltar el candado,
	// que es la forma más fuerte de no dejarlo esperando una respuesta que nadie va a dar.
	denegados := make([]dockFrame, 0, len(r.pendingPerm))
	for reqID, p := range r.pendingPerm {
		denegados = append(denegados, dockFrame{
			SessionID: r.meta.ID,
			Kind:      "permission_result",
			RequestID: reqID,
			Tool:      p.Tool,
			Decision:  string(domain.DecisionDeny),
			Text:      motivoDesactivada,
		})
	}
	// El orden del mapa es aleatorio en Go: se ordena para que dos corridas iguales emitan
	// lo mismo, que es lo que hace verificable el frame.
	sort.Slice(denegados, func(i, j int) bool { return denegados[i].RequestID < denegados[j].RequestID })
	r.pendingPerm = nil

	// 5 · el estado de vuelo se resetea: el hilo nuevo no hereda el turno a medio mandar,
	// ni el buffer de ensamblado, ni el resume del anterior. Los `grants` se DESCARTAN
	// (deny-by-default): un grant es una aprobación que el operador dio dentro de un hilo,
	// y heredarla sería aprobar algo que nunca vio acá.
	// `runSeq` NO se toca: es monótono por sesión a propósito — reiniciarlo colisionaría
	// con los run_id que el FE ya tiene marcados como terminados.
	r.pendingTurn = ""
	r.wasResume = false
	r.sawInit = false
	r.resumeRetried = false
	r.msgFlushed = false
	r.assembling.Reset()
	r.grants = nil

	// 6 · el runtime pasa a pertenecer al hilo nuevo y suelta el conductor viejo. Con
	// `live` en nil, los frames que lleguen del proceso anterior los descarta el guard que
	// `consume` ya tiene (`r.live == live`) — cero código nuevo para esa carrera.
	r.convActiva = nueva.ID
	viejo = vuelo.live
	r.live = nil

	// 7 · el disco manda. Si no se pudo guardar, en memoria no queda lo que no se guardó.
	if perr := s.persistLocked(); perr != nil {
		*r.meta = previo
		restaurarVuelo(r, vuelo)
		return nil, nil, perr
	}

	// 8 · los frames. Primero el de la conversación (trae el estado final, así que
	// aplicarlo dos veces da el mismo resultado), después los permisos denegados.
	evento := eventoActivada
	if !antes[nueva.ID] {
		evento = eventoCreada
	}
	resultado := ConversacionActivaDe(*nueva)
	frames = append(frames, dockFrame{
		SessionID:          r.meta.ID,
		Kind:               kindConversacion,
		ConversacionID:     nueva.ID,
		ConversacionEvento: evento,
		Conversacion:       &resultado,
	})
	return viejo, append(frames, denegados...), nil
}

// ─── La búsqueda (T17 · RF-321/322/324/341 · CV-D8) ──────────────────────────────────────
//
// Corre en el daemon y no en el navegador, y el motivo es de volumen: la lista NO manda los
// turnos por el cable (son 12 KB por conversación), así que mandarlos sólo para que el FE
// filtre serían más de un megabyte por cada apertura del panel. El daemon ya los tiene en
// memoria. Sin índice de texto, sin base de datos, sin tocar el índice del Mapa: un barrido
// lineal sobre lo que ya está cargado.

// contextoFragmento es cuántos caracteres se muestran a cada lado de la coincidencia.
const contextoFragmento = 40

// Conversaciones devuelve las conversaciones de UNA sesión (CV-D4 — jamás las de otra) y el
// total de la sesión. Con `q` no vacío sólo viajan las que coinciden, cada una con su
// fragmento, y el total SIGUE SIENDO EL TOTAL: es el denominador del «N de M coinciden», no
// la cuenta de coincidencias.
func (s *SessionService) Conversaciones(id, q string) ([]ConversacionResumen, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.rt[id]
	if r == nil {
		return nil, 0, errNotFound(id)
	}
	total := len(r.meta.Conversaciones)
	// Un `q` de puros espacios es «sin filtro», no «buscá el vacío» (E-26).
	nq := normalizar(strings.TrimSpace(q))
	out := make([]ConversacionResumen, 0, total)
	for i := range r.meta.Conversaciones {
		c := r.meta.Conversaciones[i]
		if nq == "" {
			out = append(out, resumenDe(c))
			continue
		}
		enTitulo := strings.Contains(normalizar(c.Titulo), nq)
		frag := fragmentoDe(c.Conv, nq)
		if !enTitulo && frag == "" {
			continue
		}
		res := resumenDe(c)
		// Coincidencia SÓLO en el título ⇒ sin fragmento: no hay nada que explicar
		// (RF-322 CA-3). Un fragmento vacío ahí sería una línea en blanco con sangría.
		res.Fragmento = frag
		out = append(out, res)
	}
	return out, total, nil
}

// normalizar deja un texto comparable: minúsculas y sin diacríticos. Vive en UN solo lugar y
// se aplica a los dos lados de la comparación; el FE no normaliza nada, manda el texto crudo.
//
// ⚠ Es un plegado RUNA A RUNA, no la descomposición canónica de Unicode, y es deliberado por
// dos razones. La primera es que el fragmento se recorta del texto ORIGINAL: si la
// normalización cambiara la cantidad de runas, las posiciones dejarían de corresponderse y el
// recorte saldría movido. La segunda es que la descomposición canónica no está en la
// biblioteca estándar y traerla sumaría tablas Unicode al binario del daemon por un caso que
// el plegado cubre — el texto que llega del conductor y del operador viene precompuesto.
// Lo que NO cubre, dicho para que nadie lo descubra a los golpes: un texto ya descompuesto
// (letra + marca combinante suelta) no se pliega.
func normalizar(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		b.WriteRune(unicode.ToLower(plegarDiacritico(r)))
	}
	return b.String()
}

// diacriticos es la tabla de plegado: cada runa de la primera fila se reemplaza por la de la
// misma posición en la segunda. Cubre Latin-1 Supplement y lo usado de Latin Extended-A —
// castellano, portugués, francés, catalán y alemán. Uno a uno, sin excepciones: `ß` NO se
// expande a `ss` justamente porque cambiaría el largo.
var diacriticos = map[rune]rune{}

func init() {
	const (
		con = "ÀÁÂÃÄÅàáâãäåÇçÈÉÊËèéêëÌÍÎÏìíîïÑñÒÓÔÕÖØòóôõöøÙÚÛÜùúûüÝýÿŠšŽžĀāĒēĪīŌōŪū"
		sin = "AAAAAAaaaaaaCcEEEEeeeeIIIIiiiiNnOOOOOOooooooUUUUuuuuYyySsZzAaEeIiOoUu"
	)
	base := []rune(sin)
	for i, r := range []rune(con) {
		diacriticos[r] = base[i]
	}
}

func plegarDiacritico(r rune) rune {
	if b, ok := diacriticos[r]; ok {
		return b
	}
	return r
}

// fragmentoDe devuelve el PRIMER pedazo de transcript que contiene `nq` (ya normalizado),
// con elipsis a los dos lados cuando se recortó. Vacío = no coincidió en ningún turno.
//
// El texto que sale es el ORIGINAL, con sus acentos y sus mayúsculas: se busca sobre el
// plegado y se recorta sobre el crudo. Y sale en texto plano — el resaltado del término lo
// hace el FE. Mandar marcado desde acá sería mandar marcado que el navegador interpreta.
func fragmentoDe(conv []domain.Turn, nq string) string {
	if nq == "" {
		return ""
	}
	aguja := []rune(nq)
	for _, t := range conv {
		crudo := []rune(t.Text)
		plegado := []rune(normalizar(t.Text))
		i := indiceDeRunas(plegado, aguja)
		if i < 0 {
			continue
		}
		desde := max(i-contextoFragmento, 0)
		hasta := min(i+len(aguja)+contextoFragmento, len(crudo))
		var b strings.Builder
		if desde > 0 {
			b.WriteString("…")
		}
		b.WriteString(strings.TrimSpace(string(crudo[desde:hasta])))
		if hasta < len(crudo) {
			b.WriteString("…")
		}
		return b.String()
	}
	return ""
}

// indiceDeRunas es un `strings.Index` que cuenta en runas, no en bytes: el recorte del
// fragmento indexa el texto original y las dos mitades tienen que hablar la misma unidad.
// Barrido cuadrático a propósito — la conversación más larga medida son 12 KB, y un algoritmo
// astuto acá sería complejidad comprada sin nadie que la pague.
func indiceDeRunas(pajar, aguja []rune) int {
	if len(aguja) == 0 || len(aguja) > len(pajar) {
		return -1
	}
	for i := 0; i+len(aguja) <= len(pajar); i++ {
		if slices.Equal(pajar[i:i+len(aguja)], aguja) {
			return i
		}
	}
	return -1
}
