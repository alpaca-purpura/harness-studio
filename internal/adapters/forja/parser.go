// Package forja materializa la semilla `.arnesia/` en el proyecto del usuario (contrato
// docs/architecture/contracts/semilla-arnesia.md, paquete 2026-07-30-arnesia-en-el-proyecto
// A-T2). Tres piezas: el parser del subset de `semilla/arnes.yaml` (este archivo), el
// scaffolder determinista (scaffolder.go) y el doctor de salud por presencia (doctor.go).
// La fuente es un fs.FS inyectado por cmd (la semilla embebida `//go:embed all:semilla`,
// via iofs.Sub) — el seam queda listo para leer el `arnes.yaml` DEL PROYECTO en Fase 2
// (flag `--arnes-yaml`, gap declarado en el contrato §5).
package forja

import (
	"errors"
	"fmt"
	"io/fs"

	"gopkg.in/yaml.v3"
)

// Territorio es una fila de `territorios` del arnes.yaml (label del rubro sobre el id
// canónico, regla E) — en ORDEN de documento (el render debe ser determinista, D12).
type Territorio struct {
	ID         string
	Label      string
	Naturaleza string
}

// Dimension es una fila de `dimensiones` (id canónico + territorio + zachman + label) —
// en ORDEN de documento. La necesita el render de `terreno/<territorio>/INDEX.md`
// (contrato §5: «territorios.<id> + sus dimensiones»).
type Dimension struct {
	ID         string
	Territorio string
	Zachman    string
	Label      string
}

// TipoPaquete es el lifecycle de un tipo de paquete (`gestion_trabajo.tipos_paquete`):
// jerarquía + estados + wip_caps + criterio de cierre (D19/D20).
type TipoPaquete struct {
	Jerarquia []string       `yaml:"jerarquia"`
	Estados   []string       `yaml:"estados"`
	WipCaps   map[string]int `yaml:"wip_caps"`
	Cierre    string         `yaml:"cierre"`
}

// PasoSpine es un paso de `proceso.spines.<tipo>`: la plantilla vacía que se siembra y el
// artefacto que se llena dentro del paquete (D18/D20). `caja` (opcional, MA-T1a/AUD-1)
// referencia la caja del grafo que ejecuta el paso — ausente = «paso sin caja aún»,
// visible en el foco del Mapa (E13), jamás un default inventado.
type PasoSpine struct {
	Paso      string `yaml:"paso"`
	Rol       string `yaml:"rol"`
	Plantilla string `yaml:"plantilla"`
	Artefacto string `yaml:"artefacto"`
	Cond      string `yaml:"cond"`
	Caja      string `yaml:"caja"`
}

// ArnesSemilla es el SUBSET parseado de `semilla/arnes.yaml` que la siembra necesita
// (spec RF-A.2): identidad para el lock + territorios/dimensiones para los INDEX +
// tipos de paquete y spines para las plantillas de proceso.
type ArnesSemilla struct {
	Schema       int // versión canónica pineada del lock (`version_pineada`, contrato §3)
	ArnesID      string
	Territorios  []Territorio
	Dimensiones  []Dimension
	TiposPaquete map[string]TipoPaquete
	Spines       map[string][]PasoSpine
}

// arnesYAML es la forma cruda del documento. `territorios` y `dimensiones` se capturan
// como yaml.Node para PRESERVAR el orden del documento (un map Go lo perdería y el
// render dejaría de ser determinista byte a byte).
type arnesYAML struct {
	Schema int `yaml:"schema"`
	Arnes  struct {
		ID string `yaml:"id"`
	} `yaml:"arnes"`
	Territorios    yaml.Node `yaml:"territorios"`
	Dimensiones    yaml.Node `yaml:"dimensiones"`
	GestionTrabajo struct {
		TiposPaquete map[string]TipoPaquete `yaml:"tipos_paquete"`
	} `yaml:"gestion_trabajo"`
	Proceso struct {
		Spines map[string][]PasoSpine `yaml:"spines"`
	} `yaml:"proceso"`
}

// ParseSemilla lee y valida el subset de `arnes.yaml` desde la raíz de fsys (la semilla
// embebida, o en Fase 2 la del proyecto). Ilegible o sin lo esencial ⇒ error honesto con
// el motivo — jamás un default inventado.
func ParseSemilla(fsys fs.FS) (ArnesSemilla, error) {
	b, err := fs.ReadFile(fsys, "arnes.yaml")
	if err != nil {
		return ArnesSemilla{}, fmt.Errorf("forja: semilla sin arnes.yaml: %w", err)
	}
	var raw arnesYAML
	if err := yaml.Unmarshal(b, &raw); err != nil {
		return ArnesSemilla{}, fmt.Errorf("forja: arnes.yaml ilegible: %w", err)
	}

	out := ArnesSemilla{
		Schema:       raw.Schema,
		ArnesID:      raw.Arnes.ID,
		TiposPaquete: raw.GestionTrabajo.TiposPaquete,
		Spines:       raw.Proceso.Spines,
	}
	if err := walkMapping(&raw.Territorios, func(id string, val *yaml.Node) error {
		var t struct {
			Label      string `yaml:"label"`
			Naturaleza string `yaml:"naturaleza"`
		}
		if err := val.Decode(&t); err != nil {
			return fmt.Errorf("territorio %q: %w", id, err)
		}
		out.Territorios = append(out.Territorios, Territorio{ID: id, Label: t.Label, Naturaleza: t.Naturaleza})
		return nil
	}); err != nil {
		return ArnesSemilla{}, fmt.Errorf("forja: arnes.yaml territorios: %w", err)
	}
	if err := walkMapping(&raw.Dimensiones, func(id string, val *yaml.Node) error {
		var d struct {
			Territorio string `yaml:"territorio"`
			Zachman    string `yaml:"zachman"`
			Label      string `yaml:"label"`
		}
		if err := val.Decode(&d); err != nil {
			return fmt.Errorf("dimensión %q: %w", id, err)
		}
		out.Dimensiones = append(out.Dimensiones, Dimension{ID: id, Territorio: d.Territorio, Zachman: d.Zachman, Label: d.Label})
		return nil
	}); err != nil {
		return ArnesSemilla{}, fmt.Errorf("forja: arnes.yaml dimensiones: %w", err)
	}

	switch {
	case out.ArnesID == "":
		return ArnesSemilla{}, errors.New("forja: arnes.yaml sin arnes.id — la semilla no identifica su fuente")
	case len(out.Territorios) == 0:
		return ArnesSemilla{}, errors.New("forja: arnes.yaml sin territorios — nada que sembrar")
	case len(out.Spines) == 0:
		return ArnesSemilla{}, errors.New("forja: arnes.yaml sin proceso.spines — nada que sembrar")
	}
	return out, nil
}

// walkMapping recorre un nodo mapping de yaml.v3 en ORDEN de documento (Content viene en
// pares clave/valor). Un nodo vacío (clave ausente) no es error: recorre cero filas.
func walkMapping(n *yaml.Node, fn func(key string, val *yaml.Node) error) error {
	if n == nil || n.Kind == 0 {
		return nil
	}
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("se esperaba un mapping, vino kind=%d", n.Kind)
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if err := fn(n.Content[i].Value, n.Content[i+1]); err != nil {
			return err
		}
	}
	return nil
}
