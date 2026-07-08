package artifact

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFixture drops an artifact file under dir and returns its relative ref.
func writeFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return name
}

func TestStatusPresente(t *testing.T) {
	dir := t.TempDir()
	r := NewReader()

	cases := map[string]struct {
		doc  string
		want string
	}{
		"plano":        {"---\nstatus: done\n---\ncuerpo\n", "done"},
		"con comillas": {"---\nstatus: \"blocked\"\n---\n", "blocked"},
		"entre claves": {"---\ntitulo: x\nstatus: working\notros: y\n---\n", "working"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ref := writeFixture(t, dir, strings.ReplaceAll(name, " ", "-")+".md", tc.doc)
			status, exists, err := r.Status(context.Background(), dir, ref)
			if err != nil {
				t.Fatalf("status: %v", err)
			}
			if !exists {
				t.Error("exists = false, el artefacto existe")
			}
			if status != tc.want {
				t.Errorf("status = %q, want %q", status, tc.want)
			}
		})
	}
}

func TestStatusAusente(t *testing.T) {
	dir := t.TempDir()
	r := NewReader()

	t.Run("artefacto no existe aún", func(t *testing.T) {
		status, exists, err := r.Status(context.Background(), dir, "todavia-no.md")
		if err != nil {
			t.Fatalf("un artefacto ausente no es error (la caja aún no lo produjo): %v", err)
		}
		if exists || status != "" {
			t.Errorf("got (%q, %v), want (\"\", false)", status, exists)
		}
	})

	t.Run("sin frontmatter", func(t *testing.T) {
		ref := writeFixture(t, dir, "sin-fm.md", "# solo cuerpo\n")
		status, exists, err := r.Status(context.Background(), dir, ref)
		if err != nil || !exists || status != "" {
			t.Errorf("got (%q, %v, %v), want (\"\", true, nil) — sin status declarado, honesto", status, exists, err)
		}
	})

	t.Run("frontmatter sin status", func(t *testing.T) {
		ref := writeFixture(t, dir, "fm-sin-status.md", "---\ntitulo: x\n---\ncuerpo\n")
		status, exists, err := r.Status(context.Background(), dir, ref)
		if err != nil || !exists || status != "" {
			t.Errorf("got (%q, %v, %v), want (\"\", true, nil)", status, exists, err)
		}
	})
}

func TestFrontmatterRoto(t *testing.T) {
	dir := t.TempDir()
	r := NewReader()
	ref := writeFixture(t, dir, "roto.md", "---\nstatus: done\nsin cierre\n")
	_, exists, err := r.Status(context.Background(), dir, ref)
	if err == nil {
		t.Error("un frontmatter sin cerrar debe ser error, no leerse como estado")
	}
	if !exists {
		t.Error("exists = false; el archivo existe aunque su frontmatter esté roto")
	}
}

func TestPathQueIntentaEscapar(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "fuera.md")
	if err := os.WriteFile(outside, []byte("---\nstatus: done\n---\n"), 0o600); err != nil {
		t.Fatalf("write outside: %v", err)
	}
	r := NewReader()

	for _, bad := range []string{
		"../fuera.md",
		"sub/../../fuera.md",
		outside, // absoluta
		"..",
	} {
		if _, _, err := r.Status(context.Background(), dir, bad); err == nil {
			t.Errorf("Status(%q) debió rechazarse — escapa del directorio del arnés", bad)
		}
	}

	// Un ref con dir vacío también se rechaza: sin confinamiento no hay lectura.
	if _, _, err := r.Status(context.Background(), "", "x.md"); err == nil {
		t.Error("Status con dir vacío debió rechazarse")
	}

	// Y un subdirectorio legítimo SÍ resuelve (el join no es paranoia, es contención).
	ref := writeFixture(t, dir, filepath.Join("sub", "ok.md"), "---\nstatus: done\n---\n")
	status, exists, err := r.Status(context.Background(), dir, ref)
	if err != nil || !exists || status != "done" {
		t.Errorf("subdir legítimo: got (%q, %v, %v), want (done, true, nil)", status, exists, err)
	}
}

func TestResumen(t *testing.T) {
	dir := t.TempDir()
	r := NewReader()
	ctx := context.Background()

	// Artefacto ausente → "" sin error (no hay nada que citar).
	if s, err := r.Resumen(ctx, dir, "spec.md"); err != nil || s != "" {
		t.Errorf("ausente: want vacío sin error, got %q err=%v", s, err)
	}

	// Con artefacto: cita SOLO el frontmatter, jamás el cuerpo.
	doc := "---\nstatus: done\nwhy: probar\n---\n\n# Cuerpo\n\nSECRETO-DEL-CUERPO\n"
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(doc), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := r.Resumen(ctx, dir, "spec.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "status: done") {
		t.Errorf("resumen sin frontmatter: %q", s)
	}
	if strings.Contains(s, "SECRETO-DEL-CUERPO") {
		t.Errorf("el cuerpo del documento viajó al resumen (rompe p11): %q", s)
	}

	// El digest sidecar gana sobre el frontmatter.
	if err := os.WriteFile(filepath.Join(dir, "spec.md.digest.md"), []byte("digest determinista ≤200 tok"), 0o600); err != nil {
		t.Fatal(err)
	}
	if s, _ := r.Resumen(ctx, dir, "spec.md"); !strings.Contains(s, "digest determinista") {
		t.Errorf("digest sidecar no ganó: %q", s)
	}

	// Confinamiento: escapar del árbol es error, igual que Status.
	if _, err := r.Resumen(ctx, dir, "../fuera.md"); err == nil {
		t.Error("ref que escapa del arnés debe fallar")
	}
}
