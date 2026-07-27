package domain

// A Session is a *frente de trabajo* (work-front): un arnés, un cwd, una vista. Es la
// unidad central de la tesis de fábrica — la app corre muchas en paralelo (multisesión,
// HS-03 it.14). El daemon es dueño del registro de sesiones abiertas.
//
// La sesión ya NO es la conversación (CV-D3): CONTIENE N conversaciones y exactamente una
// está activa. El id de Claude Code, el uso de contexto, la cadena de rotaciones y el
// registro liviano de turnos viven en Conversacion (conversacion.go), no acá.

// SessionStatus is the live state of a session's conductor, mirrored on the rail pip.
type SessionStatus string

const (
	// StatusStreaming — the conductor is generating (pip pulses live-blue).
	StatusStreaming SessionStatus = "streaming"
	// StatusAwait — Claude Code needs the human (permission / decision); pip pulses warn.
	StatusAwait SessionStatus = "await"
	// StatusIdle — parked, waiting for the next turn (pip dim).
	StatusIdle SessionStatus = "idle"
)

// Valid reports whether s is a known status.
func (s SessionStatus) Valid() bool {
	switch s {
	case StatusStreaming, StatusAwait, StatusIdle:
		return true
	default:
		return false
	}
}

// Salud is the health of the arnés behind a session, shown as the rail health dot.
// Mirrors the mockup's ok|warn|crit|info scale.
type Salud string

// The four health grades painted on the rail dot, from healthy to newborn.
const (
	SaludOK   Salud = "ok"   // sano.
	SaludWarn Salud = "warn" // atención.
	SaludCrit Salud = "crit" // señales incompletas.
	SaludInfo Salud = "info" // naciendo.
)

// Rol is who authored a transcript turn. "sys" is a system breadcrumb rendered inline
// (skill_activated, hand-offs) — not a chat bubble.
type Rol string

// The three turn authors: the human, the conductor's assistant, and the inline
// system breadcrumb (see the Rol comment).
const (
	RolUser      Rol = "user"
	RolAssistant Rol = "assistant"
	RolSys       Rol = "sys"
	// RolAct — un paso de actividad del turno (CH-D2/D3), texto "<tool> <blanco>"; el FE
	// agrupa consecutivos en una tarjeta desplegable, jamás como burbuja de texto.
	RolAct Rol = "act"
)

// Turn is one entry of a session's lightweight transcript. It is NOT the source of
// truth for the conversation (Claude Code's own JSONL is); it is what the shell shows
// immediately when you switch back to a session, before/while the live stream resumes.
type Turn struct {
	Rol  Rol    `json:"rol"`
	Text string `json:"text"`
}

// Session is a work-front: an open Claude Code conversation over one arnés.
type Session struct {
	// ID is our stable session id (survives daemon restarts); distinct from the
	// Claude Code session id used to --resume the conversation.
	ID string `json:"id"`

	// Frente is the human name of the work-front, auto-derived from the first user
	// message and editable in the rail. It is what the rail card shows.
	Frente string `json:"frente"`

	// Arnes is the id of the harness this front operates on (N sessions may share one).
	Arnes   string `json:"arnes"`
	Empresa string `json:"empresa,omitempty"`
	Puesto  string `json:"puesto,omitempty"`
	Salud   Salud  `json:"salud,omitempty"`

	// Status is the live conductor state; View is the parked view (Mapa|Diag|…);
	// Parked is the free-text "quedaste en …" landing hint.
	Status SessionStatus `json:"status"`
	View   string        `json:"view"`
	Parked string        `json:"parked,omitempty"`

	// Reparacion marca la sesión abierta contra una INSTALACIÓN del Portafolio (ley A4,
	// RF-191 mejorar-arnes-conversando): edición legal con deriva visible + backport según
	// causa. El picker la setea al elegir la copia; el rail la pinta como chip.
	Reparacion bool `json:"reparacion,omitempty"`

	// Cwd es el directorio real del conductor (estampado al spawn) — el join hacia el
	// corpus JSONL nativo (~/.claude/projects/<dir-del-cwd>/) que el historial B2 lee.
	// Vive en la SESIÓN y no en la conversación: lo resuelve el arnés, así que las N
	// conversaciones de una sesión corren en el mismo directorio por construcción.
	Cwd string `json:"cwd,omitempty"`

	// CerradaEn (RFC3339) marca que esta SESIÓN se archivó (RF-306). Ya no describe una
	// conversación: las conversaciones no se cierran, se desactivan (CV-D12).
	CerradaEn string `json:"cerrada_en,omitempty"`

	// Conversaciones son los hilos de esta sesión (CV-D3): N ≥ 1, exactamente una activa.
	// Sin omitempty: una sesión con [] es una sesión rota, y tiene que verse.
	// La invariante y sus transiciones viven en conversacion.go.
	Conversaciones []Conversacion `json:"conversaciones"`
}
