package usecase

import (
	"path/filepath"
	"testing"
)

// TestRefiereAlgunoNoMatcheaArnesiaDeProyecto FIJA el runtime de la enmienda A-D3 (A-T5,
// contrato semilla-arnesia.md §1): el guardrail del paquete cerrado (CH-D6) matchea
// `~/.arnesia` — el paquete propio de la app instalada — y NO `<proyecto>/.arnesia`, la
// semilla process-as-code del proyecto del usuario. Si alguien «endurece» las marcas a
// `.arnesia` pelado, este test se pone rojo antes de que el chat le niegue al usuario
// escribir su propio proceso.
func TestRefiereAlgunoNoMatcheaArnesiaDeProyecto(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	s := &SessionService{}
	s.ProtegerPaqueteCerrado(filepath.Join(home, ".arnesia", "kit"))
	marcas := s.cerrado

	// El guardrail SÍ atrapa el paquete cerrado, en sus dos formas (abs y ~/).
	absKit := []byte(`{"file_path":"` + filepath.Join(home, ".arnesia", "kit", "doctrine.md") + `"}`)
	if !refiereAlguno(absKit, marcas) {
		t.Error("el paquete cerrado (forma abs) debe seguir denegándose — CH-D6 intacto")
	}
	if !refiereAlguno([]byte(`{"command":"cat ~/.arnesia/kit/doctrine.md"}`), marcas) {
		t.Error("el paquete cerrado (forma ~/) debe seguir denegándose — CH-D6 intacto")
	}

	// Y NO atrapa la semilla del proyecto: `.arnesia/` de un proyecto cualquiera es zona
	// de escritura sancionada del PROCESO (A-D3) — el runtime queda INTACTO.
	proyecto := []byte(`{"file_path":"/home/x/proyecto/.arnesia/terreno/INDEX.md"}`)
	if refiereAlguno(proyecto, marcas) {
		t.Error("la semilla del proyecto (<proyecto>/.arnesia/) NO debe bloquearse — A-D3, runtime intacto")
	}
	bash := []byte(`{"command":"ls /home/x/proyecto/.arnesia/wip/activo"}`)
	if refiereAlguno(bash, marcas) {
		t.Error("un comando sobre <proyecto>/.arnesia/ NO debe bloquearse — A-D3, runtime intacto")
	}
}
