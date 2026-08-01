package loader_test

// Tests de MA-T1b (spec v2 2026-07-30-definicion-de-arnes §1): derivación del catálogo
// Arnes.Actividades + faceta Box.Actividades desde el arnes.yaml de la raíz del arnés,
// AL INDEXAR. Cubren el contrato de honestidad completo: ausente=legal · ilegible=aviso
// visible sin abortar · degradado-sin-sello=sin catálogo · paso sin caja preservado (E13)
// · caja no referenciada con faceta vacía (E6).

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/loader"
)

const manifiestoMin = `{"id":"dv","rol":"developer","proceso":"desarrollo","reporta_a":null,"empresas":["vitalia"]}`

const skillCaja = "---\nname: %s\ncontract:\n  why: ejecuta el paso\n  clase: skill\n  arquetipo: pipeline\n  perfil_harness: T1\n  caja: true\n  fase: construccion\n  estado: \"entra -> sale\"\n  gate: {tipo: none}\n---\n# x\n"

func TestLoaderActividadesDerivadas(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude-plugin", "plugin.json"), `{"name":"dv"}`)
	escribir(t, filepath.Join(dir, "arnes.l0.json"), manifiestoMin)
	escribir(t, filepath.Join(dir, "skills", "dev-team", "SKILL.md"), strings.ReplaceAll(skillCaja, "%s", "dev-team"))
	escribir(t, filepath.Join(dir, "skills", "commit-push", "SKILL.md"), strings.ReplaceAll(skillCaja, "%s", "commit-push"))
	escribir(t, filepath.Join(dir, "skills", "suelta", "SKILL.md"), strings.ReplaceAll(skillCaja, "%s", "suelta"))
	escribir(t, filepath.Join(dir, "arnes.yaml"), `
gestion_trabajo:
  tipos_paquete:
    historia:
      estados: [idea, en-curso, done]
      cierre: ratifica_capability
    spike:
      estados: [abierto, cerrado]
      cierre: documenta_decision
proceso:
  spines:
    historia:
      - { paso: tomar-spec, rol: dev, artefacto: plan.md, caja: dev-team }
      - { paso: entregar,   rol: dev, artefacto: pr,      caja: commit-push }
    spike:
      - { paso: investigar, rol: dev, artefacto: hallazgos.md, caja: dev-team }
      - { paso: decidir,    rol: dev, artefacto: decision.md }
`)

	g, err := loader.LoadArnes(dir)
	if err != nil {
		t.Fatalf("LoadArnes: %v", err)
	}
	if g.Arnes == nil {
		t.Fatal("Arnes nil con manifiesto presente")
	}
	acts := g.Arnes.Actividades
	if len(acts) != 2 {
		t.Fatalf("actividades = %d, quiero 2 (historia, spike en orden de documento)", len(acts))
	}
	if acts[0].ID != "historia" || acts[1].ID != "spike" {
		t.Errorf("orden = %s,%s — quiero orden de documento historia,spike", acts[0].ID, acts[1].ID)
	}
	if acts[0].Cierre != "ratifica_capability" || len(acts[0].Estados) != 3 {
		t.Errorf("historia = %+v, quiero cierre ratifica_capability + 3 estados", acts[0])
	}
	if len(acts[0].Pasos) != 2 || acts[0].Pasos[0].Caja != "dev-team" || acts[0].Pasos[1].Caja != "commit-push" {
		t.Errorf("historia.pasos = %+v", acts[0].Pasos)
	}
	// E13: el paso sin caja se PRESERVA — el hueco es dato, no se filtra.
	if len(acts[1].Pasos) != 2 || acts[1].Pasos[1].Paso != "decidir" || acts[1].Pasos[1].Caja != "" {
		t.Errorf("spike.pasos = %+v, quiero decidir con caja vacía (E13)", acts[1].Pasos)
	}

	// Faceta por caja: dev-team ∈ historia+spike (dedup, orden de catálogo);
	// commit-push ∈ historia; `suelta` sin faceta (grupo sin-actividad, E6).
	dt, _ := g.NodeByID("dev-team")
	if len(dt.Actividades) != 2 || dt.Actividades[0] != "historia" || dt.Actividades[1] != "spike" {
		t.Errorf("dev-team.actividades = %v, quiero [historia spike]", dt.Actividades)
	}
	cp, _ := g.NodeByID("commit-push")
	if len(cp.Actividades) != 1 || cp.Actividades[0] != "historia" {
		t.Errorf("commit-push.actividades = %v", cp.Actividades)
	}
	sl, _ := g.NodeByID("suelta")
	if len(sl.Actividades) != 0 {
		t.Errorf("suelta.actividades = %v, quiero vacía (sin-actividad E6)", sl.Actividades)
	}

	// §8 de la spec: el grafo NUEVO (con arnes.actividades[] + facetas) valida contra
	// graph.l0.schema.json — la enmienda es aditiva de verdad, no de palabra.
	validarContraSchema(t, repoRoot(t), g)
}

// §8: un grafo degradado (campo `degradado:true` en el wire desde S1-D27) también valida —
// la regularización del drift destapado en AUD-3 quedó en la misma enmienda.
func TestSchemaAdmiteDegradado(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude", "skills", "hola", "SKILL.md"), "---\nname: hola\n---\n# hola\n")
	g, err := loader.LoadArnes(dir)
	if err != nil {
		t.Fatalf("LoadArnes: %v", err)
	}
	if !g.Degradado {
		t.Fatal("fixture debía salir degradado")
	}
	validarContraSchema(t, repoRoot(t), g)
}

func TestLoaderActividadesSinArnesYaml(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude-plugin", "plugin.json"), `{"name":"dv"}`)
	escribir(t, filepath.Join(dir, "arnes.l0.json"), manifiestoMin)

	g, err := loader.LoadArnes(dir)
	if err != nil {
		t.Fatalf("LoadArnes: %v", err)
	}
	if len(g.Arnes.Actividades) != 0 {
		t.Errorf("sin arnes.yaml el catálogo debe quedar ausente (MA-L5/E8), got %+v", g.Arnes.Actividades)
	}
}

func TestLoaderActividadesIlegibleAvisoVisible(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude-plugin", "plugin.json"), `{"name":"dv"}`)
	escribir(t, filepath.Join(dir, "arnes.l0.json"), manifiestoMin)
	escribir(t, filepath.Join(dir, "arnes.yaml"), ":\n  - [")

	g, info, err := loader.LoadArnesInfo(dir)
	if err != nil {
		t.Fatalf("un arnes.yaml roto no aborta el grafo: %v", err)
	}
	if len(g.Arnes.Actividades) != 0 {
		t.Errorf("roto ⇒ sin catálogo, got %+v", g.Arnes.Actividades)
	}
	if !strings.Contains(info.Aviso, "arnes.yaml ilegible") {
		t.Errorf("aviso = %q, quiero que nombre `arnes.yaml ilegible`", info.Aviso)
	}
}

// Sin sello NO hay catálogo (el bloque `arnes` del wire no existe): primero se sella
// (S1-D28), después se agrupa — el degradado sigue siendo el degradado de siempre.
func TestLoaderActividadesSinManifiestoNoDeriva(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude", "skills", "hola", "SKILL.md"), "---\nname: hola\n---\n# hola\n")
	escribir(t, filepath.Join(dir, "arnes.yaml"), "gestion_trabajo:\n  tipos_paquete:\n    historia: {estados: [a], cierre: x}\n")

	g, err := loader.LoadArnes(dir)
	if err != nil {
		t.Fatalf("LoadArnes: %v", err)
	}
	if g.Arnes != nil || !g.Degradado {
		t.Fatalf("degradado esperado, got arnes=%+v degradado=%v", g.Arnes, g.Degradado)
	}
}

// Un spine cuyo tipo no está en tipos_paquete = actividad declarada a medias: visible
// igual (después de las declaradas, en orden alfabético determinista).
func TestLoaderActividadesSpineHuerfano(t *testing.T) {
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude-plugin", "plugin.json"), `{"name":"dv"}`)
	escribir(t, filepath.Join(dir, "arnes.l0.json"), manifiestoMin)
	escribir(t, filepath.Join(dir, "arnes.yaml"), `
gestion_trabajo:
  tipos_paquete:
    historia: {estados: [a, b], cierre: x}
proceso:
  spines:
    zeta:   [{ paso: p1 }]
    bugfix: [{ paso: reproducir }]
`)

	g, err := loader.LoadArnes(dir)
	if err != nil {
		t.Fatalf("LoadArnes: %v", err)
	}
	acts := g.Arnes.Actividades
	if len(acts) != 3 || acts[0].ID != "historia" || acts[1].ID != "bugfix" || acts[2].ID != "zeta" {
		ids := make([]string, len(acts))
		for i, a := range acts {
			ids[i] = a.ID
		}
		t.Errorf("orden = %v, quiero [historia bugfix zeta] (declaradas primero, huérfanos alfabéticos)", ids)
	}
}
