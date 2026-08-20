// vendored_arnesia_test.go — fija los diffs del vendored ArnesIA contra upstream
// (vitalia-app/tools/cockpit). Cada test acá corresponde a una fila de la tabla de
// README-vendored.md: si un merge desde upstream borra una de estas adaptaciones, el
// test cae y el drift se ve, en vez de descubrirse con el board en 400.
package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── DA-2 · escrituras sobre el pseudo-sistema `platform` ────────────────────

// harness-studio tiene su docs/product en la RAÍZ: todo su contenido es `platform`. Con
// la regla upstream (escrituras solo a sistemas reales) el board sería de solo lectura.
func TestWritableSistemasHonraElFlag(t *testing.T) {
	setupPlatformEmpresa(t)

	t.Cleanup(func() { flagAllowPlatformWrites = false })

	flagAllowPlatformWrites = false
	if containsStr(writableSistemas(), platformKey) {
		t.Fatalf("sin flag, platform NO debe ser escribible (paridad upstream): %v", writableSistemas())
	}

	flagAllowPlatformWrites = true
	if !containsStr(writableSistemas(), platformKey) {
		t.Fatalf("con flag, platform debe ser escribible: %v", writableSistemas())
	}
}

// El env var es el camino que usa cockpit.ps1 cuando no puede pasar flags al hijo.
func TestWritableSistemasHonraElEnv(t *testing.T) {
	setupPlatformEmpresa(t)
	t.Setenv("COCKPIT_ALLOW_PLATFORM_WRITES", "1")

	if !containsStr(writableSistemas(), platformKey) {
		t.Fatalf("con COCKPIT_ALLOW_PLATFORM_WRITES=1, platform debe ser escribible: %v", writableSistemas())
	}
}

// ── R4 · las capabilities generadas no se editan por PATCH ──────────────────

// El workspace de ArnesIA deriva sus caps de scripts/capabilities_to_yaml.py. Un PATCH
// escribiría a mano un archivo que la próxima regeneración pisa sin avisar — y su enum
// de status (`vivo`/`parcial`/`stub`) ni siquiera es el de gates.go. Se bloquea con un
// 409 que EXPLICA, no con el 400 genérico de "sistema desconocido".
func TestCapabilityPatchBloqueadoCuandoLasCapsSonGeneradas(t *testing.T) {
	platDir := setupPlatformEmpresa(t)
	escribirGenerador(t, platDir)

	req := httptest.NewRequest(http.MethodPatch, "/api/capabilities/m/c?sistema="+platformKey,
		strings.NewReader(`{"status":"deprecated","reason":"motivo suficientemente largo"}`))
	req.SetPathValue("module", "m")
	req.SetPathValue("cap", "c")
	rec := httptest.NewRecorder()
	handleCapabilitySingle(rec, req)

	if rec.Code != 409 {
		t.Fatalf("status = %d, want 409 (artefacto generado) · body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "GENERADOS") {
		t.Errorf("el 409 debe decir POR QUÉ está bloqueado, body=%s", rec.Body.String())
	}
}

// Sin generador en el workspace (caso upstream), el PATCH conserva su comportamiento.
func TestCapabilityPatchIntactoSinGenerador(t *testing.T) {
	setupPlatformEmpresa(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/capabilities/m/c?sistema="+platformKey,
		strings.NewReader(`{"status":"deprecated","reason":"motivo suficientemente largo"}`))
	req.SetPathValue("module", "m")
	req.SetPathValue("cap", "c")
	rec := httptest.NewRecorder()
	handleCapabilitySingle(rec, req)

	if rec.Code == 409 {
		t.Fatalf("sin generador NO debe bloquearse por R4 (regresión sobre upstream): %s", rec.Body.String())
	}
}

// ── lectura · el título legible sale de `name` cuando no hay user_facing_name ──

func TestReadCapabilityUsaNameComoNombreLegible(t *testing.T) {
	dir := t.TempDir()
	capPath := filepath.Join(dir, "abrir-shell.yaml")
	// Forma real de una capability de ArnesIA (generada): `name`, sin user_facing_name,
	// y con comentario inline en el status — que el parser YAML debe descartar.
	contenido := "---\n" +
		"capability_id: arnesia.cli-daemon.abrir-shell\n" +
		"slug: abrir-shell\n" +
		"name: \"Abrir shell (`open`)\"\n" +
		"status: vivo          # GENERADO (R4) — no teclear\n" +
		"module: cli-daemon\n" +
		"---\n\n# cuerpo\n"
	if err := os.WriteFile(capPath, []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}

	cap, err := readCapability(capPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := cap["user_facing_name"]; got != "Abrir shell (`open`)" {
		t.Errorf("user_facing_name = %v, want el `name` de la cap (si no, la UI pinta el slug)", got)
	}
	// El comentario inline NO debe contaminar el valor.
	if got := cap["status"]; got != "vivo" {
		t.Errorf("status = %q, want \"vivo\" (comentario inline descartado)", got)
	}
}

// Tres capabilities del repo (forja/*, portafolio/identificar) son DOCUMENTOS YAML: abren
// `---` y no cierran. Antes de este fix el reader devolvía el mapa vacío sin error y los
// defaults defensivos fabricaban una capability inexistente (`capability_id: ""` +
// `status: "live"`, un status que ni siquiera pertenece al enum de este repo).
func TestReadCapabilityLeeDocumentoYamlSinDelimitadorDeCierre(t *testing.T) {
	dir := t.TempDir()
	capPath := filepath.Join(dir, "chequear-semilla.yaml")
	contenido := "---\n" +
		"capability_id: arnesia.forja.chequear-semilla\n" +
		"slug: chequear-semilla\n" +
		"name: \"Chequear salud de la semilla\"\n" +
		"status: vivo\n" +
		"module: forja\n"
	if err := os.WriteFile(capPath, []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}

	cap, err := readCapability(capPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := cap["capability_id"]; got != "arnesia.forja.chequear-semilla" {
		t.Errorf("capability_id = %q, want el del archivo (no el default vacío)", got)
	}
	if got := cap["status"]; got != "vivo" {
		t.Errorf("status = %q, want \"vivo\" — un default \"live\" fabricado es peor que fallar", got)
	}
}

// user_facing_name explícito sigue ganando sobre name (coalesce, no rename).
func TestReadCapabilityPrefiereUserFacingNameExplicito(t *testing.T) {
	dir := t.TempDir()
	capPath := filepath.Join(dir, "x.yaml")
	contenido := "---\nslug: x\nname: interno\nuser_facing_name: El de cara al usuario\n---\n"
	if err := os.WriteFile(capPath, []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}

	cap, err := readCapability(capPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := cap["user_facing_name"]; got != "El de cara al usuario" {
		t.Errorf("user_facing_name = %v, want el explícito", got)
	}
}

// ── portabilidad Windows del guardia de path ────────────────────────────────

// El veredicto de isTraversalSafe debe ser IDÉNTICO en Windows y POSIX. Estos casos
// son los que divergían: filepath.Clean traducía a `\` (rompía el whitelist aguas
// abajo) y filepath.IsAbs no reconocía los absolutos POSIX ni las formas con volumen.
func TestIsTraversalSafeEsIndependienteDelOS(t *testing.T) {
	cases := []struct {
		in        string
		wantClean string
		wantOK    bool
		porQue    string
	}{
		{"docs/x.md", "docs/x.md", true, "el clean sale en forma slash o segmentWhitelisted no matchea"},
		{"docs/sub/y.md", "docs/sub/y.md", true, "ídem, anidado"},
		{`docs\x.md`, "docs/x.md", true, "un separador Windows se normaliza, no se rechaza"},
		{"/etc/passwd", "", false, "absoluto POSIX: filepath.IsAbs no lo ve en Windows"},
		{`\etc\passwd`, "", false, "rooted Windows sin volumen"},
		{`C:\Windows\win.ini`, "", false, "absoluto con volumen"},
		{`C:algo`, "", false, "relativo-al-drive: sigue siendo una forma con volumen"},
		{`\\servidor\share\x`, "", false, "UNC"},
		{"../secrets", "", false, "parent traversal"},
		{`..\secrets`, "", false, "parent traversal con separador Windows"},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			clean, ok := isTraversalSafe(c.in)
			if ok != c.wantOK {
				t.Fatalf("isTraversalSafe(%q) ok = %v, want %v — %s", c.in, ok, c.wantOK, c.porQue)
			}
			if ok && clean != c.wantClean {
				t.Errorf("isTraversalSafe(%q) clean = %q, want %q — %s", c.in, clean, c.wantClean, c.porQue)
			}
		})
	}
}

// docTypeFor clasifica por el path; en Windows fsnotify entrega backslashes y TODO
// caía en "other" (la UI no refrescaba la tarjeta correcta).
func TestDocTypeForReconoceSeparadoresNativos(t *testing.T) {
	cases := []struct{ in, want string }{
		{filepath.Join("docs", "product", "capabilities", "m", "c.yaml"), "capability"},
		{filepath.Join("docs", "product", "releases", "v1.yaml"), "release"},
		{filepath.Join("docs", "learnings", "x.md"), "learning"},
		{filepath.Join("docs", "product", "stories", "s", "checkpoint.md"), "checkpoint"},
	}
	for _, c := range cases {
		if got := docTypeFor(c.in); got != c.want {
			t.Errorf("docTypeFor(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ── helpers ─────────────────────────────────────────────────────────────────

func escribirGenerador(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, "scripts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "capabilities_to_yaml.py"), []byte("# generador\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
