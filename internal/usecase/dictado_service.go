package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// EstadoDictado says whether the returned text was ordered or is the raw transcript.
//
// La distinción es OBLIGATORIA en el wire (RF-222): pasar un crudo por limpio sería
// exactamente el pass fabricado que la doctrina de la casa prohíbe. El FE la pinta.
type EstadoDictado string

const (
	// DictadoLimpio — the cleanup ran and the text is the ordered request.
	DictadoLimpio EstadoDictado = "limpio"
	// DictadoCrudo — the cleanup failed or timed out; the text is the raw transcript and
	// the UI must mark it as unordered (V-D4's escape, RF-227).
	DictadoCrudo EstadoDictado = "crudo"
)

// Dictado is the result of one dictation: the text the composer gets, plus how honest we
// are being about it.
type Dictado struct {
	Texto  string        `json:"texto"`
	Estado EstadoDictado `json:"estado"`
	// Motivo explains why the text came back raw. Empty when Estado is limpio.
	Motivo string `json:"motivo,omitempty"`
	// Motor is the STT engine that produced the transcript (observability, not decoration:
	// which engine ran changes what errors to expect).
	Motor string `json:"motor,omitempty"`
}

// ErrDictadoVacio is returned when transcription produced no text at all. Silencio y falla
// no son lo mismo: el FE avisa y NO toca el composer (RF-227).
var ErrDictadoVacio = errors.New("dictado: la transcripción vino vacía")

// ErrSesionDesconocida is returned when the session id does not exist.
var ErrSesionDesconocida = errors.New("dictado: la sesión no existe")

// ErrDictadoNoDisponible is returned when there is no usable engine. Es un sentinel propio
// —y no un error genérico— porque el transporte lo mapea a **503, no a 500**: «no puedo
// transcribir» es una capacidad ausente, no una falla interna, y confundirlas le haría al
// operador buscar un bug donde falta un `apt install`. Mismo criterio que el 503 de
// `validarMarketplace` («no puedo mirar» ≠ «tu url está mal»).
var ErrDictadoNoDisponible = errors.New("dictado: no disponible")

// límites del paso de limpieza. topeLimpieza sale de lo medido (13-17 s, §1.8) con margen:
// más allá de eso, esperar cuesta más de lo que el orden vale y conviene caer al crudo —
// que es el mismo escalón de V-D4, no un camino aparte.
const topeLimpieza = 40 * time.Second

// turnosDeContexto is how many trailing transcript turns feed the cleanup (V-D2 FIRMADA:
// "últimos 2-3"). Se toma el techo del rango.
const turnosDeContexto = 3

// maxCharsPorTurno recorta cada turno del contexto. Lo que desenreda es contexto CHICO y de
// dominio (§1.5): un turno gigante pegado entero diluye el glosario y encima paga latencia
// sobre el paso que ya es el 87 % del costo.
const maxCharsPorTurno = 400

// glosarioGlobal is the static domain glossary the cleanup gets on every dictation.
//
// V-D2 lo firmó **global** (no por arnés) para arrancar. Es corto a propósito: son los
// términos que el STT escribe mal o que el operador nombra de oído. Crece por evidencia —
// una palabra que el dictado erró de verdad— no por completitud.
const glosarioGlobal = `Vocabulario del proyecto (corregí la transcripción contra esta lista):
ArnesIA, arnés/arneses, daemon, Claude Code, Tauri, WebKitGTK, Go, React, Zustand,
Storybook, Vitest, Biome, SSE, capability/capabilities, boundary/boundaries, conformance,
dogfood, Mapa, Portafolio, marketplace, composer, chat dock, PermissionCard, sidecar,
self-update, deriva, procedencia, canal, banda, spine, caja, frente de trabajo.`

// SessionLookup is the slice of the session registry the dictation needs: the transcript of
// one session, to build the cleanup context.
type SessionLookup interface {
	Get(id string) (domain.Session, bool)
}

// DictadoService turns recorded audio into text for the composer: transcribe, then order it
// with domain context.
//
// Los dos pasos cuestan MUY distinto (medido, §1.8, sobre 47 s de voz real): transcribir
// 2.2 s, ordenar 13-17 s. La limpieza es el 87 % del gasto — por eso tiene su propio tope y
// su propia salida de emergencia, y por eso el FE la muestra como etapa aparte.
type DictadoService struct {
	stt      ports.TranscriptionPort
	limpieza ports.LimpiezaPort
	sesiones SessionLookup
	// tope caps the cleanup step; zero uses topeLimpieza.
	tope time.Duration
}

// NewDictadoService wires the dictation use case. limpieza may be nil: sin limpiador el
// servicio devuelve crudo marcado, que es degradación honesta y no una falla.
func NewDictadoService(stt ports.TranscriptionPort, limpieza ports.LimpiezaPort, sesiones SessionLookup) *DictadoService {
	return &DictadoService{stt: stt, limpieza: limpieza, sesiones: sesiones, tope: topeLimpieza}
}

// SetTopeLimpieza overrides how long the cleanup step may take before the dictation falls
// back to raw. Zero restores the default.
func (s *DictadoService) SetTopeLimpieza(d time.Duration) { s.tope = d }

// Disponibilidad reports whether dictation can run, with the reason when it cannot.
func (s *DictadoService) Disponibilidad(ctx context.Context) ports.Disponibilidad {
	if s.stt == nil {
		return ports.Disponibilidad{Motivo: "este build no trae transcripción"}
	}
	return s.stt.Disponible(ctx)
}

// Dictar transcribes audio for the given session and returns the text for the composer.
//
// El contrato con el FE: **o devuelve texto, o devuelve error — nunca un texto inventado ni
// un cuelgue.** Si la transcripción falla o vuelve vacía es un error (el composer no se
// toca). Si falla la limpieza NO es un error: es un crudo marcado, porque el dictado no se
// puede perder por una falla de un paso opcional.
func (s *DictadoService) Dictar(ctx context.Context, sesionID string, audio []byte, mime string) (Dictado, error) {
	if s.stt == nil {
		return Dictado{}, fmt.Errorf("%w: este build no trae transcripción", ErrDictadoNoDisponible)
	}
	sess, ok := s.sesiones.Get(sesionID)
	if !ok {
		return Dictado{}, fmt.Errorf("%w: %s", ErrSesionDesconocida, sesionID)
	}

	disp := s.stt.Disponible(ctx)
	if !disp.Disponible {
		// No se acepta un dictado que después no se va a poder transcribir (RF-223).
		return Dictado{}, fmt.Errorf("%w: %s", ErrDictadoNoDisponible, disp.Motivo)
	}

	crudo, err := s.stt.Transcribir(ctx, audio, mime)
	if err != nil {
		return Dictado{}, fmt.Errorf("dictado: transcribir: %w", err)
	}
	if strings.TrimSpace(crudo) == "" {
		return Dictado{}, ErrDictadoVacio
	}
	crudo = strings.TrimSpace(crudo)

	if s.limpieza == nil {
		return Dictado{
			Texto: crudo, Estado: DictadoCrudo, Motor: disp.Motor,
			Motivo: "este build no trae el paso de limpieza",
		}, nil
	}

	tope := s.tope
	if tope <= 0 {
		tope = topeLimpieza
	}
	ctxLimpieza, cancel := context.WithTimeout(ctx, tope)
	defer cancel()

	// El contexto de limpieza sale de la conversación ACTIVA: el dictado se dicta sobre
	// el hilo que está abierto, no sobre la sesión entera (CV-D4/CV-D7). Una sesión sin
	// activa (registro roto) limpia sin contexto — degradación honesta, jamás un panic.
	var turnos []domain.Turn
	if conv, hay := sess.Activa(); hay {
		turnos = conv.Conv
	}
	limpio, err := s.limpieza.Ordenar(ctxLimpieza, crudo, contextoDeLimpieza(turnos))
	if err != nil {
		// V-D4: el escape a crudo. Mismo camino que el fallback de RF-227 — un solo
		// código, dos motivos para entrar.
		return Dictado{
			Texto: crudo, Estado: DictadoCrudo, Motor: disp.Motor,
			Motivo: motivoDeCrudo(err),
		}, nil
	}
	if strings.TrimSpace(limpio) == "" {
		return Dictado{
			Texto: crudo, Estado: DictadoCrudo, Motor: disp.Motor,
			Motivo: "la limpieza no devolvió texto",
		}, nil
	}
	return Dictado{Texto: strings.TrimSpace(limpio), Estado: DictadoLimpio, Motor: disp.Motor}, nil
}

// motivoDeCrudo turns the cleanup failure into one line the operator can read. Distinguir
// "tardó" de "falló" importa: con lo primero, reintentar tiene sentido.
func motivoDeCrudo(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "la limpieza tardó de más y se cortó"
	}
	return "falló el paso de limpieza"
}

// contextoDeLimpieza builds the short domain block the cleanup gets: the static global
// glossary plus the last few transcript turns.
//
// **Lo que NO manda** (V-D2 FIRMADA): la conversación entera, ni el CLAUDE.md completo. El
// hallazgo del spike fue que lo que desenreda es contexto CHICO y de dominio; más volumen
// empeora el resultado y encima paga latencia sobre el paso más caro de la cadena.
func contextoDeLimpieza(conv []domain.Turn) string {
	var b strings.Builder
	b.WriteString(glosarioGlobal)

	ultimos := turnosRelevantes(conv)
	if len(ultimos) == 0 {
		return b.String()
	}
	b.WriteString("\n\nÚltimo de la conversación (para resolver a qué se refiere):\n")
	for _, t := range ultimos {
		b.WriteString("- ")
		b.WriteString(etiquetaDeRol(t.Rol))
		b.WriteString(": ")
		b.WriteString(recortar(t.Text, maxCharsPorTurno))
		b.WriteString("\n")
	}
	return b.String()
}

// turnosRelevantes returns the last turnosDeContexto user/assistant turns.
//
// Se filtran `act` y `sys`: son migas de actividad («Read foo.go»), no conversación — meten
// ruido en un bloque que vale justamente por ser chico.
func turnosRelevantes(conv []domain.Turn) []domain.Turn {
	out := make([]domain.Turn, 0, turnosDeContexto)
	for i := len(conv) - 1; i >= 0 && len(out) < turnosDeContexto; i-- {
		switch conv[i].Rol {
		case domain.RolUser, domain.RolAssistant:
			if strings.TrimSpace(conv[i].Text) == "" {
				continue
			}
			out = append(out, conv[i])
		case domain.RolAct, domain.RolSys:
			continue
		}
	}
	// Se recolectó de atrás para adelante; el bloque se lee en orden cronológico.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// etiquetaDeRol names the speaker in the context block.
func etiquetaDeRol(r domain.Rol) string {
	if r == domain.RolUser {
		return "el operador dijo"
	}
	return "la respuesta fue"
}

// recortar caps a turn at n characters on a rune boundary, marking the cut.
func recortar(s string, n int) string {
	s = strings.TrimSpace(s)
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}
