// Package hooks proyecta el payload de un hook de Claude Code al evento canónico, con la
// **allowlist declarada** de arquitectura-modulo.md §6.1.
//
// 🔴 Por qué esto existe y no se reenvía el stdin: el payload del hook trae **contenido en
// claro** (ANEXO H4, medido, no supuesto):
//
//   - `UserPromptSubmit.prompt` = el prompt completo del usuario;
//   - `Stop.last_assistant_message` = la respuesta del asistente;
//   - `PostToolUse.tool_response` = lo que la herramienta leyó o escribió.
//
// Un hook que postea su stdin entero a `127.0.0.1` cumple «no egresa» y **aun así filtra la
// conversación al almacén local**. Por eso la proyección pasa ANTES de escribir a ningún
// lado, loopback incluido, y por eso hay dos checks de conformance y no uno.
//
// Vive acá y no en `cmd/` porque lo usan los DOS: el daemon (endpoint `/api/telemetria/proceso`)
// y el propio binario (subcomando `arnesia hook proceso`). Dos implementaciones de la misma
// allowlist es la forma segura de que una se olvide de un campo.
package hooks

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// AdaptadorVersionHook versiona el MAPEO del hook, igual que el del OTLP (D7.5).
const AdaptadorVersionHook = "cc-hook/1"

// allowlistHook es LA LISTA del camino 2. Default-deny: lo que no está acá no se lee.
//
// `cwd` está en la lista pero **NO se persiste crudo** (A14): se usa para resolver la
// instalación y después se tira; lo que queda es un id conocido o una huella.
var allowlistHook = map[string]bool{
	"session_id": true, "prompt_id": true, "hook_event_name": true,
	"tool_name": true, "duration_ms": true, "permission_mode": true,
	"reason": true, "cwd": true,
}

// negadosHook son los campos que se descartan NOMBRE POR NOMBRE. No hace falta enumerarlos
// —la allowlist ya los deja afuera— pero se listan para poder contarlos y para que
// `TestHookNoReenviaContenido` los busque por nombre en vez de por casualidad.
var negadosHook = []string{
	"prompt", "last_assistant_message", "tool_response", "tool_input",
	"transcript_path", "background_tasks", "session_crons", "stop_hook_active", "tool_use_id",
}

// NegadosHook expone la denylist para los tests de fitness y el check de conformance
// `hook-proyecta-campos`.
func NegadosHook() []string { return append([]string(nil), negadosHook...) }

// AllowlistHook expone la lista declarada.
func AllowlistHook() []string {
	out := make([]string, 0, len(allowlistHook))
	for k := range allowlistHook {
		out = append(out, k)
	}
	return out
}

// ResolverCWD mapea el `cwd` de una corrida a una instalación conocida del Portafolio
// (identidad `(home,id)`). Devuelve `ok=false` cuando no la reconoce — y entonces la ruta
// **igual se tira**: lo que se guarda es una huella, jamás el path del usuario.
type ResolverCWD func(cwd string) (arnesID, instalacionID string, ok bool)

// ErrHookFueraDeAlcance marca los eventos que el MVP no instrumenta a propósito
// (`PreToolUse`, `SessionStart`): no aportan al join y sí superficie.
var ErrHookFueraDeAlcance = fmt.Errorf("hooks: hook_event_name fuera del alcance del MVP")

// payload es el subconjunto DECLARADO del stdin del hook. Los campos de contenido **no están
// en este struct**: `encoding/json` los descarta al decodificar y no hay dónde ponerlos.
// Es la primera de las dos barreras; la segunda es que el evento se arma campo por campo.
type payload struct {
	SessionID      string       `json:"session_id"`
	PromptID       string       `json:"prompt_id"`
	HookEventName  string       `json:"hook_event_name"`
	ToolName       string       `json:"tool_name"`
	DuracionMs     *json.Number `json:"duration_ms"`
	PermissionMode string       `json:"permission_mode"`
	Reason         string       `json:"reason"`
	CWD            string       `json:"cwd"`
}

// Proyectar reduce el stdin del hook a un evento canónico.
//
// `resolver` puede ser nil (el hook standalone no tiene el Portafolio a mano): en ese caso el
// `cwd` se guarda como HUELLA y la atribución queda `sin-dato` — que es la verdad, y el
// daemon la mejora al recibir, porque él sí tiene la tabla.
func Proyectar(stdin []byte, resolver ResolverCWD, ahora time.Time) (domain.EventoTelemetria, error) {
	var p payload
	if err := json.Unmarshal(stdin, &p); err != nil {
		return domain.EventoTelemetria{}, fmt.Errorf("%w: %v", domain.ErrPayloadInvalido, err)
	}
	if p.SessionID == "" {
		return domain.EventoTelemetria{}, domain.ErrEventoSinSesion
	}

	tipo, ok := tipoDeHook(p.HookEventName)
	if !ok {
		return domain.EventoTelemetria{}, ErrHookFueraDeAlcance
	}

	ev := domain.EventoTelemetria{
		LlaveJoin: domain.LlaveJoin{
			SesionID: p.SessionID,
			// La otra mitad de la llave del join (ANEXO H1): el MISMO identificador que
			// `prompt.id` trae por OTel. Por eso el join es una igualdad y no una
			// correlación por tiempo.
			TurnoID: p.PromptID,
		},
		Emisor:           domain.EmisorHook,
		Runtime:          "claude-code",
		AdaptadorVersion: AdaptadorVersionHook,
		TSRecibido:       ahora.UTC(),
		TipoEvento:       tipo,
		// El hook JAMÁS trae dinero (ANEXO H2). Los buckets quedan nil, que es «no aplica»,
		// no cero.
		Motivo: p.Reason,
		// `escenario` se deriva: un evento que llega SOLO por hook es s2-degradado. Si el
		// mismo turno trajo api_request, el servicio lo sube a s2-instrumentado o s1 — el
		// escenario es de la VENTANA, y esto es el default honesto de un evento suelto.
		Escenario: domain.EscenarioS2Degradado,
	}
	if p.ToolName != "" {
		ev.Herramienta = p.ToolName
	}
	if p.DuracionMs != nil {
		if n, err := p.DuracionMs.Int64(); err == nil {
			ev.DuracionMs = &n
		}
	}
	if p.PermissionMode != "" {
		ev.Gate = p.PermissionMode
	}

	// ── A14: el `cwd` se usa y se tira ──
	if p.CWD != "" {
		if resolver != nil {
			if arnes, inst, ok := resolver(p.CWD); ok {
				ev.ArnesID, ev.InstalacionID = arnes, inst
				ev.Atribucion = domain.ConfianzaPorProceso
			}
		}
		if ev.Atribucion == "" {
			// No se reconoció. Se guarda una HUELLA —no reversible a la ruta— para poder
			// agrupar corridas del mismo lugar sin saber cuál es ese lugar.
			ev.CWDHuella = HuellaCWD(p.CWD)
			ev.Atribucion = domain.ConfianzaSinDato
		}
	} else {
		ev.Atribucion = domain.ConfianzaSinDato
	}

	return ev, nil
}

// HuellaCWD reduce una ruta a una huella corta y estable. **No es reversible** y no lleva el
// nombre del usuario ni el del proyecto: es solo para agrupar corridas del mismo lugar.
//
// Se normaliza antes de hashear (`filepath.Clean`) para que `/a/b` y `/a/b/` den la misma
// huella; si no, la misma instalación se contaría como dos.
func HuellaCWD(cwd string) string {
	limpio := filepath.Clean(strings.TrimSpace(cwd))
	if limpio == "" || limpio == "." {
		return ""
	}
	suma := sha256.Sum256([]byte(limpio))
	return "h:" + hex.EncodeToString(suma[:])[:16]
}

// tipoDeHook mapea `hook_event_name` al tipo de evento canónico.
//
// `PreToolUse` y `SessionStart` **no se instrumentan en el MVP**: no aportan al join
// (`PostToolUse` ya trae la herramienta y su duración) y sí superficie.
func tipoDeHook(n string) (domain.TipoEvento, bool) {
	switch n {
	case "UserPromptSubmit":
		return domain.EventoTurnoInicio, true
	case "PostToolUse":
		return domain.EventoHerramienta, true
	case "Stop":
		return domain.EventoTurnoFin, true
	case "SessionEnd":
		return domain.EventoSesionFin, true
	default:
		return "", false
	}
}
