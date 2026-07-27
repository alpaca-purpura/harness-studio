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
// checkpoint y deja el breadcrumb RolSys. Caller holds s.mu; el spawn fresco ocurre
// después, por el camino normal (live==nil → spawnLocked SIN --resume).
//
// Todo esto pasa DENTRO de la conversación activa: la rotación es invisible y NO corta el
// hilo (CV-D10) — es la misma conversación, con una marca inline.
//
// DEVUELVE el frame en vez de publicarlo, y esa es la única diferencia con la versión
// anterior. El motivo no es estilo: se la llama desde `Turn` **con `s.mu` tomado**, y en este
// servicio nada publica bajo el candado. Que la marca no llegara en vivo (el operador no la
// veía hasta recargar) no era un olvido — era que el lugar donde vive esta función no puede
// publicar. Se corrige respetando la regla, no rompiéndola.
//
// El texto de la marca es la constante de arriba y NO el del dibujo del mockup: ese string ya
// está en el registro de conversaciones vivas, y reescribirlo dejaría los transcripts ya
// persistidos con la marca vieja (contradicción C-5 del diseño, resuelta a favor del código).
func (s *SessionService) rotarLocked(r *sessionRuntime) dockFrame {
	if r.live != nil {
		_ = r.live.Close()
		r.live = nil
	}
	conv := r.activa()
	if conv.ClaudeSessionID != "" {
		conv.CadenaCC = append(conv.CadenaCC, conv.ClaudeSessionID)
	}
	conv.Checkpoint = CheckpointMecanico(conv.Conv)
	conv.ClaudeSessionID = ""
	conv.RotacionPendiente = false
	conv.Conv = append(conv.Conv, domain.Turn{Rol: domain.RolSys, Text: breadcrumbRotacion})

	// El índice donde quedó la marca. Es lo que hace idempotente a este frame: no lleva
	// run_id —no pertenece a un turno— así que el FE appendea sólo si su copia del
	// transcript tiene exactamente esa longitud. Un replay por Last-Event-ID llega con la
	// copia más larga y se descarta solo.
	idx := len(conv.Conv) - 1
	return dockFrame{
		SessionID:          r.meta.ID,
		Kind:               kindConversacion,
		ConversacionID:     conv.ID,
		ConversacionEvento: eventoRotada,
		TurnoIdx:           &idx,
		Text:               breadcrumbRotacion,
		// Sin CtxPct: el uso real del hilo fresco llega con el result del turno nuevo.
		// Mandar el viejo pintaría un número que ya no describe nada.
	}
}
