// proceso.go — el descriptor de proceso como DATO embebido (I-77 · DH-07 · RN-31).
//
// El ciclo de vida ya no vive hardcodeado en gates.go: se DERIVA del descriptor que
// shipea el kit (espejo generado+gated en process/sdd-default.yaml, go:embed — la
// fábrica lo emite desde products/kit/core-harness/process/, gen_proceso_descriptor.py).
// Contrato = process-descriptor.schema.yaml (L0): categorías semánticas FIJAS
// (propuesto·en-progreso·completado·descartado·pausado); el motor se ata SIEMPRE a la
// categoría, jamás al nombre del estado — otro descriptor = otro proceso, cero cambio
// de código. Descriptor inválido → panic al boot (fail-fast: un motor sin proceso
// válido no arranca; jamás un proceso a medias).
package main

import (
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v3"
)

//go:embed process/sdd-default.yaml
var procesoRaw []byte

// categorías del contrato L0, en su orden canónico — FIJAS e inmutables desde v1
// (regla Azure/D4): el motor y la consola se atan a esto, jamás al nombre del estado.
var categoriasOrden = []string{"propuesto", "en-progreso", "completado", "descartado", "pausado"}

// categorías TERMINALES del contrato L0 (derivadas de la categoría, nunca del nombre).
var categoriasTerminales = map[string]bool{"completado": true, "descartado": true}

var categoriasValidas = func() map[string]bool {
	m := make(map[string]bool, len(categoriasOrden))
	for _, c := range categoriasOrden {
		m[c] = true
	}
	return m
}()

type procesoMeta struct {
	ID          string `yaml:"id" json:"id"`
	Nombre      string `yaml:"nombre" json:"nombre"`
	Version     int    `yaml:"version" json:"version"`
	Descripcion string `yaml:"descripcion" json:"descripcion,omitempty"`
}

type procesoEstado struct {
	ID          string `yaml:"id" json:"id"`
	Categoria   string `yaml:"categoria" json:"categoria"`
	Nombre      string `yaml:"nombre" json:"nombre,omitempty"`
	Descripcion string `yaml:"descripcion" json:"descripcion,omitempty"`
	Inicial     bool   `yaml:"inicial" json:"inicial,omitempty"`
	Wip         int    `yaml:"wip" json:"wip,omitempty"` // límite WIP; 0 = sin límite (RN-48)
}

type procesoTransicion struct {
	De            string `yaml:"de" json:"de"`
	A             string `yaml:"a" json:"a"`
	Ejecutor      string `yaml:"ejecutor" json:"ejecutor"`
	RequiereRazon bool   `yaml:"requiere_razon" json:"requiere_razon,omitempty"`
	Verbo         string `yaml:"verbo" json:"verbo,omitempty"`   // evento CDEvents <sujeto>.<predicado> — NO único (RN-45)
	Nombre        string `yaml:"nombre" json:"nombre,omitempty"` // etiqueta humana de la acción (RN-48)
}

type procesoBinding struct {
	Rol   string `yaml:"rol" json:"rol"`
	Arnes string `yaml:"arnes" json:"arnes,omitempty"`
	Nota  string `yaml:"nota" json:"nota,omitempty"`
}

// momentoLista acepta escalar o lista en YAML (un gate puede aplicar a N momentos).
type momentoLista []string

func (m *momentoLista) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		*m = momentoLista{value.Value}
		return nil
	}
	var lista []string
	if err := value.Decode(&lista); err != nil {
		return err
	}
	*m = lista
	return nil
}

type procesoGate struct {
	ID        string         `yaml:"id" json:"id"`
	Nombre    string         `yaml:"nombre" json:"nombre"`
	Momento   momentoLista   `yaml:"momento" json:"momento"`
	Autoridad procesoBinding `yaml:"autoridad" json:"autoridad"`
	Checklist []string       `yaml:"checklist" json:"checklist"`
}

type procesoParametros struct {
	RazonMinima int `yaml:"razon_minima" json:"razon_minima"`
}

type procesoDescriptor struct {
	Descriptor   procesoMeta                 `yaml:"descriptor" json:"descriptor"`
	Estados      []procesoEstado             `yaml:"estados" json:"estados"`
	Transiciones []procesoTransicion         `yaml:"transiciones" json:"transiciones"`
	Gates        []procesoGate               `yaml:"gates" json:"gates"`
	Duenos       map[string][]procesoBinding `yaml:"duenos" json:"duenos"`
	Parametros   procesoParametros           `yaml:"parametros" json:"parametros"`

	categoriaPorEstado map[string]string
}

// parseProceso valida lo mínimo que el motor necesita para operar sin sorpresas
// (la validación completa contra el L0 la hace la fábrica ANTES de emitir el espejo).
func parseProceso(raw []byte) (*procesoDescriptor, error) {
	var p procesoDescriptor
	if err := yaml.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("YAML ilegible: %w", err)
	}
	if len(p.Estados) == 0 {
		return nil, fmt.Errorf("descriptor sin estados")
	}
	p.categoriaPorEstado = make(map[string]string, len(p.Estados))
	iniciales := 0
	for _, e := range p.Estados {
		if e.ID == "" {
			return nil, fmt.Errorf("estado sin id")
		}
		if _, dup := p.categoriaPorEstado[e.ID]; dup {
			return nil, fmt.Errorf("estado duplicado: %s", e.ID)
		}
		if !categoriasValidas[e.Categoria] {
			return nil, fmt.Errorf("estado %s: categoria %q fuera del contrato", e.ID, e.Categoria)
		}
		p.categoriaPorEstado[e.ID] = e.Categoria
		if e.Inicial {
			iniciales++
		}
	}
	if iniciales != 1 {
		return nil, fmt.Errorf("exactamente UN estado inicial requerido, hay %d", iniciales)
	}
	for _, t := range p.Transiciones {
		if _, ok := p.categoriaPorEstado[t.De]; !ok {
			return nil, fmt.Errorf("transición %s→%s: 'de' no es un estado", t.De, t.A)
		}
		if _, ok := p.categoriaPorEstado[t.A]; !ok {
			return nil, fmt.Errorf("transición %s→%s: 'a' no es un estado", t.De, t.A)
		}
	}
	for estado := range p.Duenos {
		if _, ok := p.categoriaPorEstado[estado]; !ok {
			return nil, fmt.Errorf("duenos: %q no es un estado", estado)
		}
	}
	if p.Parametros.RazonMinima <= 0 {
		return nil, fmt.Errorf("parametros.razon_minima debe ser > 0")
	}
	return &p, nil
}

// ── derivaciones que alimentan las tablas de gates.go ────────────────────────

func (p *procesoDescriptor) estadoIDs() []string {
	out := make([]string, len(p.Estados))
	for i, e := range p.Estados {
		out[i] = e.ID
	}
	return out
}

// esTerminal: por CATEGORÍA (contrato), jamás por nombre. Estado desconocido → false.
func (p *procesoDescriptor) esTerminal(estado string) bool {
	return categoriasTerminales[p.categoriaPorEstado[estado]]
}

func (p *procesoDescriptor) transicionesOperador() map[string][]operatorTransition {
	out := make(map[string][]operatorTransition)
	for _, t := range p.Transiciones {
		if t.Ejecutor != "operador" {
			continue
		}
		out[t.De] = append(out[t.De], operatorTransition{to: t.A, requiresReason: t.RequiereRazon})
	}
	return out
}

// verboDe: el EVENTO que una transición emite ("qué acaba de pasar" — principio 5).
// Sin verbo declarado → "" (el campo es opcional en el contrato).
func (p *procesoDescriptor) verboDe(de, a string) string {
	for _, t := range p.Transiciones {
		if t.De == de && t.A == a {
			return t.Verbo
		}
	}
	return ""
}

// duenosRender: los bindings como la string de owner que el board ya muestra
// ("{arnes} ({nota})" unidos por " o " — round-trip byte-a-byte con el hardcode viejo).
func (p *procesoDescriptor) duenosRender() map[string]string {
	out := make(map[string]string, len(p.Duenos))
	for estado, bindings := range p.Duenos {
		s := ""
		for i, b := range bindings {
			if i > 0 {
				s += " o "
			}
			quien := b.Arnes
			if quien == "" {
				quien = b.Rol
			}
			s += quien
			if b.Nota != "" {
				s += " (" + b.Nota + ")"
			}
		}
		if s != "" {
			out[estado] = s
		}
	}
	return out
}

// ── el descriptor vivo del binario (fail-fast al boot) ───────────────────────

var proceso = mustProceso()

func mustProceso() *procesoDescriptor {
	p, err := parseProceso(procesoRaw)
	if err != nil {
		panic("descriptor de proceso embebido inválido (process/sdd-default.yaml): " + err.Error())
	}
	return p
}
