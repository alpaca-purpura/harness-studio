package usecase

import (
	"fmt"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// breadcrumbRotacion es el rastro RolSys que queda en el Conv al rotar (RF-198): si el
// usuario mira el historial entiende que pasó algo, sin interrupción real.
const breadcrumbRotacion = "— contexto rotado, seguimos —"

// checkpointTurnos es cuántos turnos recientes viajan al checkpoint mecánico (RF-196).
const checkpointTurnos = 6

// checkpointMaxTurno acota cada turno citado en el checkpoint — el punto es continuidad
// barata (espíritu Franja Artefactos: digest chico, −90 % contexto), no re-inyectar todo.
const checkpointMaxTurno = 700

// CheckpointMecanico arma el digest determinístico que un proceso fresco necesita para
// continuar la conversación como si nada (RF-196): los últimos N turnos (recortados) + la
// instrucción de retomar. Sin LLM — mecánico primero (MC-D5); si en la práctica no
// alcanza, una pasada de resumen es extensión, no rediseño.
func CheckpointMecanico(conv []domain.Turn) string {
	if len(conv) == 0 {
		return ""
	}
	desde := len(conv) - checkpointTurnos
	if desde < 0 {
		desde = 0
	}
	var b strings.Builder
	b.WriteString("Esta conversación venía en curso y el contexto rotó por detrás (mecanismo de ArnesIA).\n")
	b.WriteString("NO saludes de nuevo ni pidas que repitan nada — continuá donde quedó.\n\n")
	b.WriteString("### Últimos turnos\n\n")
	for _, t := range conv[desde:] {
		texto := strings.TrimSpace(t.Text)
		if len(texto) > checkpointMaxTurno {
			texto = texto[:checkpointMaxTurno] + "…"
		}
		fmt.Fprintf(&b, "- **[%s]** %s\n", t.Rol, texto)
	}
	b.WriteString("\n### Cómo seguir\n\nRetomá el último pedido del usuario desde donde quedó; el árbol del arnés en el cwd ya contiene todo lo aplicado hasta acá.\n")
	return b.String()
}

// rotarLocked ejecuta la rotación invisible (RF-197): cierra el proceso viejo, encadena su
// ClaudeSessionID (RF-198 — el join para el historial), limpia el resume, escribe el
// checkpoint en la sesión y deja el breadcrumb RolSys. Caller holds s.mu; el spawn fresco
// ocurre después, por el camino normal (live==nil → spawnLocked SIN --resume).
func (s *SessionService) rotarLocked(r *sessionRuntime) {
	if r.live != nil {
		_ = r.live.Close()
		r.live = nil
	}
	if r.meta.ClaudeSessionID != "" {
		r.meta.CadenaCC = append(r.meta.CadenaCC, r.meta.ClaudeSessionID)
	}
	r.meta.Checkpoint = CheckpointMecanico(r.meta.Conv)
	r.meta.ClaudeSessionID = ""
	r.meta.RotacionPendiente = false
	r.meta.Conv = append(r.meta.Conv, domain.Turn{Rol: domain.RolSys, Text: breadcrumbRotacion})
}
