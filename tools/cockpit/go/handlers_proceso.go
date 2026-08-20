// handlers_proceso.go — GET /api/proceso (I-77 · RN-32): el descriptor de proceso que
// gobierna el motor, completo + derivados, para que la consola rinda TODO desde el dato
// (la vista /proceso no lleva un solo literal de estado — RN-33).
//
// Sin ?sistema= en v1: el binario embebe UN descriptor (el default que shipea el kit);
// descriptor por workspace = extensión F5+.
package main

import "net/http"

func handleProceso(w http.ResponseWriter, r *http.Request) {
	estados := make([]map[string]any, 0, len(proceso.Estados))
	for _, e := range proceso.Estados {
		estados = append(estados, map[string]any{
			"id":          e.ID,
			"categoria":   e.Categoria,
			"nombre":      e.Nombre,
			"descripcion": e.Descripcion,
			"inicial":     e.Inicial,
			"wip":         e.Wip,                    // límite WIP del estado (RN-48; 0 = sin límite)
			"terminal":    proceso.esTerminal(e.ID), // derivado por categoría (RN-31)
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"descriptor":   proceso.Descriptor,
		"categorias":   categoriasOrden,
		"estados":      estados,
		"transiciones": proceso.Transiciones,
		"gates":        proceso.Gates,
		"duenos":       proceso.Duenos,
		"parametros":   proceso.Parametros,
		"fuente":       "process/sdd-default.yaml (SSoT: el kit — espejo gated embebido, I-77)",
	})
}
