// actividades.go — MA-T1b (paquete 2026-07-30-definicion-de-arnes, spec v2 §1 AUD-1..4):
// deriva el catálogo Arnes.Actividades + la faceta Box.Actividades desde el `arnes.yaml`
// que vive en la RAÍZ del arnés (misma base que arnes.l0.json). La derivación corre AL
// INDEXAR (patrón de derivarEdges) — jamás en el request, jamás al sello.
//
// Este lector NO reusa forja.ParseSemilla a propósito: (1) el grafo de imports prohíbe
// loader→forja (allow-list .go-arch-lint.yml); (2) las validaciones de siembra (territorios
// obligatorios, spines obligatorios) NO aplican acá — un arnes.yaml solo-con-tipos es legal
// para el Mapa aunque no alcance para sembrar. Subset propio, mínimo, con el vocabulario
// REAL de la semilla (`cierre`, id = clave del mapa — AUD-5).

package loader

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// tipoPaqueteYAML es el subset de `gestion_trabajo.tipos_paquete.<id>` que el Mapa
// necesita (estados + cierre; jerarquia/wip_caps/appetite son de la siembra, no del Mapa).
type tipoPaqueteYAML struct {
	Estados []string `yaml:"estados"`
	Cierre  string   `yaml:"cierre"`
}

// pasoSpineYAML es el subset de un paso de `proceso.spines.<tipo>` que el Mapa necesita.
// `caja` es el campo nuevo de MA-T1a (AUD-1); ausente = «paso sin caja aún» (E13).
type pasoSpineYAML struct {
	Paso      string `yaml:"paso"`
	Rol       string `yaml:"rol"`
	Artefacto string `yaml:"artefacto"`
	Caja      string `yaml:"caja"`
}

// actividadesYAML es la forma cruda del subset. `tipos_paquete` se captura como yaml.Node
// para PRESERVAR el orden del documento (mismo requisito determinista que forja/parser.go:
// un map Go perdería el orden y el catálogo cambiaría entre corridas).
type actividadesYAML struct {
	GestionTrabajo struct {
		TiposPaquete yaml.Node `yaml:"tipos_paquete"`
	} `yaml:"gestion_trabajo"`
	Proceso struct {
		Spines map[string][]pasoSpineYAML `yaml:"spines"`
	} `yaml:"proceso"`
}

// leerActividades lee el `arnes.yaml` de la raíz del arnés y deriva el catálogo de
// actividades. Contrato de honestidad (mismo espíritu que leerManifiesto):
//   - archivo ausente → (nil, "", nil): arnés sin tipos declarados, legal (MA-L5/E8);
//   - archivo ilegible → (nil, aviso, nil): degradado VISIBLE, jamás aborta el grafo
//     (el grafo existe sin actividades; el aviso viaja en Info.Aviso);
//   - sin tipos NI spines → (nil, "", nil): nada que derivar, cero invento.
//
// Orden del catálogo: los tipos de `tipos_paquete` en ORDEN de documento; después los
// tipos que SOLO existen en `proceso.spines` (declarados a medias — visibles igual),
// en orden alfabético de clave para que el render sea determinista.
func leerActividades(dir string) ([]domain.Actividad, string, error) {
	ruta := filepath.Join(dir, "arnes.yaml")
	b, err := os.ReadFile(ruta) //nolint:gosec // G304: ruta bajo el dir del arnés que el caller eligió cargar.
	if errors.Is(err, os.ErrNotExist) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("leer %s: %w", ruta, err)
	}

	var raw actividadesYAML
	if uerr := yaml.Unmarshal(b, &raw); uerr != nil {
		return nil, fmt.Sprintf("arnes.yaml ilegible — actividades no derivadas: %v", uerr), nil
	}

	var acts []domain.Actividad
	vistos := map[string]bool{}
	if werr := walkMappingActividades(&raw.GestionTrabajo.TiposPaquete, func(id string, val *yaml.Node) error {
		var tp tipoPaqueteYAML
		if derr := val.Decode(&tp); derr != nil {
			return fmt.Errorf("tipos_paquete.%s: %w", id, derr)
		}
		vistos[id] = true
		acts = append(acts, domain.Actividad{
			ID:      id,
			Estados: tp.Estados,
			Cierre:  tp.Cierre,
			Pasos:   pasosDe(raw.Proceso.Spines[id]),
		})
		return nil
	}); werr != nil {
		return nil, fmt.Sprintf("arnes.yaml tipos_paquete ilegible — actividades no derivadas: %v", werr), nil
	}

	// Spines de tipos NO declarados en tipos_paquete: declarados a medias, visibles igual
	// (honestidad — el Mapa los muestra; el gate de completitud D19 es de la forja).
	var huerfanos []string
	for tipo := range raw.Proceso.Spines {
		if !vistos[tipo] {
			huerfanos = append(huerfanos, tipo)
		}
	}
	sort.Strings(huerfanos)
	for _, tipo := range huerfanos {
		acts = append(acts, domain.Actividad{ID: tipo, Pasos: pasosDe(raw.Proceso.Spines[tipo])})
	}

	if len(acts) == 0 {
		return nil, "", nil
	}
	return acts, "", nil
}

// pasosDe traduce los pasos crudos del YAML al dominio (subset del Mapa).
func pasosDe(crudos []pasoSpineYAML) []domain.PasoActividad {
	if len(crudos) == 0 {
		return nil
	}
	pasos := make([]domain.PasoActividad, 0, len(crudos))
	for _, p := range crudos {
		pasos = append(pasos, domain.PasoActividad{Paso: p.Paso, Rol: p.Rol, Artefacto: p.Artefacto, Caja: p.Caja})
	}
	return pasos
}

// estamparFacetaActividades estampa en cada caja la faceta `actividades` (qué
// procedimientos la referencian vía pasos[].caja) — derivada, dedupeada, en el orden del
// catálogo. Una caja que ningún procedimiento referencia queda con la faceta vacía: ese
// silencio ES el dato (grupo `sin-actividad` del Mapa, E6). Un `caja:` que no matchea
// ningún nodo NO estampa nada ni inventa nodos — el hueco lo destapa el gate de
// conformance, no el loader.
func estamparFacetaActividades(acts []domain.Actividad, nodos []domain.Box) {
	porCaja := map[string][]string{}
	for _, a := range acts {
		for _, p := range a.Pasos {
			if p.Caja == "" {
				continue
			}
			if !contieneString(porCaja[p.Caja], a.ID) {
				porCaja[p.Caja] = append(porCaja[p.Caja], a.ID)
			}
		}
	}
	for i := range nodos {
		if ids, ok := porCaja[nodos[i].ID]; ok {
			nodos[i].Actividades = ids
		}
	}
}

// walkMappingActividades recorre un mapping yaml.v3 en orden de documento — copia local
// de forja.walkMapping (el grafo de imports prohíbe loader→forja; 12 líneas duplicadas
// valen menos que abrir ese boundary).
func walkMappingActividades(n *yaml.Node, fn func(key string, val *yaml.Node) error) error {
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

// contieneString reporta si xs contiene v (dedup de la faceta sin traer una dep).
func contieneString(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
