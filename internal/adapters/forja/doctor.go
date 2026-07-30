package forja

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// Chequear es el doctor v0 (contrato §4, A-D4): juzga la instalación por PRESENCIA de las
// piezas del árbol §2 — SIN hashes (la evaluación de deriva de semilla contra el lock es
// gap declarado Fase 2, §3). Estados visibles, jamás fabricados:
//
//	sana       — `.arnesia/` con TODAS las piezas
//	ausente    — no existe `.arnesia/`
//	incompleta — faltan piezas o el lock está malformado; Faltantes las lista
func (a *Adapter) Chequear(dir string) (domain.SaludSemilla, error) {
	raiz := filepath.Join(dir, ".arnesia")
	if fi, err := os.Stat(raiz); err != nil || !fi.IsDir() {
		// «no existe .arnesia/» es el VEREDICTO ausente (§4.4), no un fallo del doctor
		// — el error de Stat es la evidencia, no la excepción.
		return domain.SaludSemilla{ //nolint:nilerr // veredicto ausente, no error del doctor.
			Estado:  domain.SemillaAusente,
			Detalle: fmt.Sprintf("no existe %s — sembrar con `arnesia init`", raiz),
		}, nil
	}

	// La lista esperada se deriva de la MISMA semilla que siembra (plan compartido con
	// el scaffolder): contrato §2 con una sola fuente, ni doctor ni sembrador con su
	// propia copia del árbol.
	sem, err := ParseSemilla(a.fsys)
	if err != nil {
		return domain.SaludSemilla{}, err
	}
	piezas, err := a.plan(sem, datosPlantilla{ProyectoNombre: "chequeo", ArnesID: sem.ArnesID, Fecha: "0000-00-00"})
	if err != nil {
		return domain.SaludSemilla{}, err
	}

	salud := domain.SaludSemilla{Estado: domain.SemillaSana}
	for _, p := range piezas {
		if _, serr := os.Lstat(filepath.Join(raiz, filepath.FromSlash(p.Ruta))); serr != nil {
			salud.Faltantes = append(salud.Faltantes, path.Join(".arnesia", p.Ruta))
		}
	}

	// El lock: ausente es faltante; presente pero ilegible es MALFORMADO — visible con
	// motivo, jamás «sana» por no mirar.
	rutaLock := filepath.Join(raiz, nombreLock)
	if b, rerr := os.ReadFile(rutaLock); rerr != nil { //nolint:gosec // G304: ruta bajo el dir que el caller eligió chequear.
		salud.Faltantes = append(salud.Faltantes, path.Join(".arnesia", nombreLock))
	} else if jerr := json.Unmarshal(b, &lockSemilla{}); jerr != nil {
		salud.Faltantes = append(salud.Faltantes, path.Join(".arnesia", nombreLock)+" (malformado)")
		salud.Detalle = fmt.Sprintf("%s malformado: %v", nombreLock, jerr)
	}

	if len(salud.Faltantes) > 0 {
		salud.Estado = domain.SemillaIncompleta
		if salud.Detalle == "" {
			salud.Detalle = fmt.Sprintf("faltan %d pieza(s) del árbol del contrato semilla-arnesia.md §2", len(salud.Faltantes))
		}
	}
	return salud, nil
}
