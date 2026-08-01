package forja_test

import (
	iofs "io/fs"
	"testing"
	"testing/fstest"

	doctrina "github.com/alpacapurpura/arnesia"
	"github.com/alpacapurpura/arnesia/internal/adapters/forja"
)

// semillaReal devuelve la semilla EMBEBIDA en el binario (la misma que siembra la app
// instalada) — el golden de estos tests es el `semilla/arnes.yaml` real, no un fixture.
func semillaReal(t *testing.T) iofs.FS {
	t.Helper()
	sub, err := iofs.Sub(doctrina.Semilla, "semilla")
	if err != nil {
		t.Fatalf("iofs.Sub(semilla): %v", err)
	}
	return sub
}

// TestParseSemillaGoldenContraArnesYAMLReal fija el subset parseado contra la copia
// GRADUADA firmada (contrato §5, check «parser golden» del §7).
func TestParseSemillaGoldenContraArnesYAMLReal(t *testing.T) {
	sem, err := forja.ParseSemilla(semillaReal(t))
	if err != nil {
		t.Fatalf("ParseSemilla: %v", err)
	}

	if sem.ArnesID != "dev" {
		t.Errorf("ArnesID = %q, quiero \"dev\"", sem.ArnesID)
	}
	if sem.Schema != 0 {
		t.Errorf("Schema = %d, quiero 0 (version_pineada del lock)", sem.Schema)
	}

	// Territorios: 4, en ORDEN de documento (el render es determinista, D12).
	quieroTerr := []struct{ id, label, naturaleza string }{
		{"proposito", "Propósito", "definicion"},
		{"producto", "Producto", "definicion"},
		{"organizacion", "Organización", "definicion"},
		{"wip", "WIP", "instancia"},
	}
	if len(sem.Territorios) != len(quieroTerr) {
		t.Fatalf("territorios = %d, quiero %d", len(sem.Territorios), len(quieroTerr))
	}
	for i, q := range quieroTerr {
		got := sem.Territorios[i]
		if got.ID != q.id || got.Label != q.label || got.Naturaleza != q.naturaleza {
			t.Errorf("territorio[%d] = %+v, quiero %+v", i, got, q)
		}
	}

	// Dimensiones: las 11 canónicas, primera y última en orden de documento.
	if len(sem.Dimensiones) != 11 {
		t.Fatalf("dimensiones = %d, quiero 11 (canónicas D19)", len(sem.Dimensiones))
	}
	if sem.Dimensiones[0].ID != "encargo" || sem.Dimensiones[0].Territorio != "proposito" || sem.Dimensiones[0].Zachman != "Why" {
		t.Errorf("dimensión[0] = %+v, quiero encargo/proposito/Why", sem.Dimensiones[0])
	}
	if ult := sem.Dimensiones[10]; ult.ID != "gestion-trabajo" || ult.Territorio != "organizacion" {
		t.Errorf("dimensión[10] = %+v, quiero gestion-trabajo/organizacion", ult)
	}

	// Tipos de paquete: historia (lifecycle completo) + spike.
	historia, ok := sem.TiposPaquete["historia"]
	if !ok {
		t.Fatal("falta tipos_paquete.historia")
	}
	if len(historia.Estados) != 6 || historia.Estados[0] != "idea" || historia.Estados[5] != "done" {
		t.Errorf("historia.Estados = %v, quiero idea…done (6)", historia.Estados)
	}
	if historia.WipCaps["refinando"] != 3 || historia.WipCaps["en-curso"] != 2 {
		t.Errorf("historia.WipCaps = %v, quiero refinando:3 en-curso:2", historia.WipCaps)
	}
	if historia.Cierre != "ratifica_capability" {
		t.Errorf("historia.Cierre = %q", historia.Cierre)
	}
	if spike, ok := sem.TiposPaquete["spike"]; !ok || spike.Cierre != "documenta_decision" {
		t.Errorf("tipos_paquete.spike = %+v, quiero cierre documenta_decision", spike)
	}

	// Spines: historia 6 pasos (ux/arquitectura condicionales), spike 2.
	sh := sem.Spines["historia"]
	if len(sh) != 6 {
		t.Fatalf("spines.historia = %d pasos, quiero 6", len(sh))
	}
	if sh[0].Paso != "research" || sh[0].Rol != "pm" || sh[0].Plantilla != "proceso/historia/00-research.md" {
		t.Errorf("historia[0] = %+v", sh[0])
	}
	if sh[2].Cond != "tiene_ui" || sh[3].Cond != "toca_arquitectura" {
		t.Errorf("conds = %q/%q, quiero tiene_ui/toca_arquitectura", sh[2].Cond, sh[3].Cond)
	}
	if ss := sem.Spines["spike"]; len(ss) != 2 || ss[1].Artefacto != "decision.md" {
		t.Errorf("spines.spike = %+v", ss)
	}
}

// TestParseSemillaCajaOpcionalEnPaso (MA-T1a/AUD-1): `caja` referencia la caja del grafo
// que ejecuta el paso; ausente queda "" — «paso sin caja aún» (E13), jamás un default.
func TestParseSemillaCajaOpcionalEnPaso(t *testing.T) {
	fsys := fstest.MapFS{"arnes.yaml": &fstest.MapFile{Data: []byte(
		"schema: 0\narnes: {id: developer-vitalia}\n" +
			"territorios:\n  producto: {label: Producto, naturaleza: definicion}\n" +
			"proceso:\n  spines:\n    bugfix:\n" +
			"      - { paso: reproducir, rol: dev, plantilla: p/00.md, artefacto: repro.md, caja: dev-team }\n" +
			"      - { paso: decidir,    rol: dev, plantilla: p/01.md, artefacto: decision.md }\n")}}
	sem, err := forja.ParseSemilla(fsys)
	if err != nil {
		t.Fatalf("ParseSemilla: %v", err)
	}
	sb := sem.Spines["bugfix"]
	if len(sb) != 2 || sb[0].Caja != "dev-team" {
		t.Errorf("bugfix[0].Caja = %q, quiero dev-team", sb[0].Caja)
	}
	if sb[1].Caja != "" {
		t.Errorf("bugfix[1].Caja = %q, quiero \"\" (paso sin caja aún, E13)", sb[1].Caja)
	}
}

// TestParseSemillaRealSinCajas: la semilla embebida HOY no declara cajas (las cajas son
// del arnés que la opera, no de la plantilla genérica) — si esto cambia, el golden avisa.
func TestParseSemillaRealSinCajas(t *testing.T) {
	sem, err := forja.ParseSemilla(semillaReal(t))
	if err != nil {
		t.Fatalf("ParseSemilla: %v", err)
	}
	for tipo, pasos := range sem.Spines {
		for _, p := range pasos {
			if p.Caja != "" {
				t.Errorf("semilla %s/%s declara caja %q — actualizar este golden a propósito", tipo, p.Paso, p.Caja)
			}
		}
	}
}

// TestParseSemillaIlegibleErrorHonesto: YAML roto o incompleto ⇒ error con motivo, jamás
// defaults inventados (spec RF-A.2).
func TestParseSemillaIlegibleErrorHonesto(t *testing.T) {
	casos := map[string]fstest.MapFS{
		"sin arnes.yaml": {},
		"yaml roto":      {"arnes.yaml": &fstest.MapFile{Data: []byte(":\n  - [")}},
		"sin arnes.id":   {"arnes.yaml": &fstest.MapFile{Data: []byte("schema: 0\nterritorios:\n  a: {label: A}\nproceso:\n  spines:\n    x: []\n")}},
		"sin spines":     {"arnes.yaml": &fstest.MapFile{Data: []byte("schema: 0\narnes: {id: x}\nterritorios:\n  a: {label: A}\n")}},
	}
	for nombre, fsys := range casos {
		if _, err := forja.ParseSemilla(fsys); err == nil {
			t.Errorf("%s: quiero error honesto, vino nil", nombre)
		}
	}
}
