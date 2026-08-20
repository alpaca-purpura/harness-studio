// gates.go — el DOMINIO de los gates SDD como funciones PURAS y testeables.
//
// Las reglas del ciclo de vida (10 estados, transiciones de operador G1-G2,
// terminalidad para el merge-gate G7-G8, política de status de capability) vivían
// inline dentro de los handlers HTTP — imposibles de probar sin levantar el server
// y un workspace en disco. Acá quedan separadas del transporte: entran valores
// planos, salen veredictos. Los handlers (handlers_stories/caps/releases.go) las
// invocan. Probadas en gates_test.go. Cambiar una regla = un solo lugar, con test.
package main

import "strings"

// ── ciclo de vida de stories — DERIVADO del descriptor de proceso (I-77/RN-31) ──
// Las tablas ya no son literales: se derivan del descriptor embebido (proceso.go).
// La instancia default (sdd-default, la que shipea el kit) reproduce EXACTO el
// hardcode pre-F4 — round-trip fijado en proceso_test.go; otro descriptor = otro
// proceso, cero cambio en este archivo.
var storyStates = proceso.estadoIDs()

func isStoryState(s string) bool { return containsStr(storyStates, s) }

// isTerminalState: una story cuenta como "cerrada" para el merge-gate (G7/G8).
// Por CATEGORÍA del contrato (completado/descartado), jamás por nombre de estado.
func isTerminalState(state string) bool { return proceso.esTerminal(state) }

// ── transiciones que el OPERADOR ejecuta desde el cockpit (G1-G2) ────────────
type operatorTransition struct {
	to             string
	requiresReason bool
}

var operatorAllowed = proceso.transicionesOperador()

// stateOwners: estados cuya transición NO la hace el operador (la ejecuta un rol/skill).
// Bindings nombrados del descriptor, rendidos a la string que el board ya muestra.
var stateOwners = proceso.duenosRender()

var transitionReasonMinLen = proceso.Parametros.RazonMinima

// transitionVerdict = decisión de dominio para una transición iniciada por el operador.
type transitionVerdict int

const (
	transitionOK transitionVerdict = iota
	transitionForbidden   // no es una transición permitida al operador desde fromState
	transitionNeedsReason // permitida, pero el reason no llega al mínimo
)

// evaluateOperatorTransition decide si el operador puede pasar fromState→targetState.
func evaluateOperatorTransition(fromState, targetState, reason string) transitionVerdict {
	var allowed *operatorTransition
	for _, t := range operatorAllowed[fromState] {
		if t.to == targetState {
			tt := t
			allowed = &tt
			break
		}
	}
	if allowed == nil {
		return transitionForbidden
	}
	if allowed.requiresReason && len(strings.TrimSpace(reason)) < transitionReasonMinLen {
		return transitionNeedsReason
	}
	return transitionOK
}

// transitionOwnerNote = el mensaje cuando la transición no la permite el cockpit.
func transitionOwnerNote(targetState string) string {
	if owner, ok := stateOwners[targetState]; ok {
		return "esta transition la ejecuta " + owner + ". Invoca la skill desde Claude Code."
	}
	return "transition no permitida desde el cockpit."
}

// ── política de edición de stories (PATCH operador) ──────────────────────────
var storyEditableFields = map[string]bool{
	"release": true, "priority": true, "goal": true, "anti": true, "reuse": true,
}

// forbiddenStoryFields = los campos del PATCH que el operador NO puede editar.
func forbiddenStoryFields(fields map[string]any) []string {
	var out []string
	for k := range fields {
		if !storyEditableFields[k] {
			out = append(out, k)
		}
	}
	return out
}

// ── política de status de capability (PATCH operador) ────────────────────────
var capForbiddenStructuralFields = map[string]bool{
	"scenarios": true, "change_log": true, "capability_id": true,
	"slug": true, "module": true, "license": true,
	"parent_cap": true, "derives_capabilities": true,
}

// forbiddenCapFields = campos estructurales que NO se editan directo (requieren story).
func forbiddenCapFields(body map[string]any) []string {
	var out []string
	for k := range body {
		if capForbiddenStructuralFields[k] {
			out = append(out, k)
		}
	}
	return out
}

var validCapStatuses = map[string]bool{
	"live": true, "beta": true, "deprecated": true, "sunset": true,
}

func isValidCapStatus(s string) bool { return validCapStatuses[s] }

const capStatusReasonMinLen = 10
