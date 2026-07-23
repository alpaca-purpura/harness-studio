package history_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/history"
	"github.com/alpacapurpura/arnesia/internal/domain"
)

// TestDirParaCwd fija el encoding observado del CLI real (todo no-alfanumérico → '-').
func TestDirParaCwd(t *testing.T) {
	if got := history.DirParaCwd("/home/chalreme/Proyectos/luana-vitalia/vitalia"); got != "-home-chalreme-Proyectos-luana-vitalia-vitalia" {
		t.Errorf("encoding = %q", got)
	}
	if got := history.DirParaCwd("/a/b.c_d"); got != "-a-b-c-d" {
		t.Errorf("encoding con . y _ = %q", got)
	}
}

// TestTurnosLeeCorpus (RF-201): user/assistant con texto entran; bloques sin texto
// (tool_use) y tipos ajenos (queue-operation) se saltan; JSONL ausente = error honesto.
func TestTurnosLeeCorpus(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, history.DirParaCwd("/proj/vitalia"))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	jsonl := `{"type":"queue-operation","operation":"enqueue"}
{"type":"user","message":{"role":"user","content":[{"type":"text","text":"arreglá la rule"}]}}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Write"}]}}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"listo, la arreglé"}]}}
{"type":"user","message":{"role":"user","content":"gracias"}}
`
	if err := os.WriteFile(filepath.Join(dir, "cc-1.jsonl"), []byte(jsonl), 0o600); err != nil {
		t.Fatal(err)
	}

	r, err := history.New(base)
	if err != nil {
		t.Fatal(err)
	}
	turnos, err := r.Turnos("/proj/vitalia", "cc-1")
	if err != nil {
		t.Fatal(err)
	}
	quiero := []domain.Turn{
		{Rol: domain.RolUser, Text: "arreglá la rule"},
		{Rol: domain.RolAssistant, Text: "listo, la arreglé"},
		{Rol: domain.RolUser, Text: "gracias"},
	}
	if len(turnos) != len(quiero) {
		t.Fatalf("turnos = %+v", turnos)
	}
	for i, q := range quiero {
		if turnos[i] != q {
			t.Errorf("turno %d = %+v, quiero %+v", i, turnos[i], q)
		}
	}

	if _, err := r.Turnos("/proj/vitalia", "cc-borrada"); err == nil {
		t.Error("JSONL ausente debe reportarse, jamás inventarse")
	}
	if r.Existe("/proj/vitalia", "cc-borrada") {
		t.Error("Existe debe decir la verdad")
	}
}
