package usecase_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// fakeForjaPort registra las llamadas — el usecase solo valida el dir y delega (A-T3).
type fakeForjaPort struct {
	sembrados  []string
	chequeados []string
}

func (f *fakeForjaPort) Sembrar(dir string) (domain.InformeSemilla, error) {
	f.sembrados = append(f.sembrados, dir)
	return domain.InformeSemilla{Creados: []string{".arnesia/terreno/INDEX.md"}}, nil
}

func (f *fakeForjaPort) Chequear(dir string) (domain.SaludSemilla, error) {
	f.chequeados = append(f.chequeados, dir)
	return domain.SaludSemilla{Estado: domain.SemillaSana}, nil
}

// TestForjaServiceValidaDirProtegido: la política de path protegido del Portafolio aplica
// a la siembra — una semilla JAMÁS se escribe bajo ~/.claude, $HOME, relativos o
// inexistentes; el puerto NI SE TOCA.
func TestForjaServiceValidaDirProtegido(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0o750); err != nil {
		t.Fatal(err)
	}

	puerto := &fakeForjaPort{}
	svc := usecase.NewForjaService(puerto)
	ctx := context.Background()

	casos := map[string]string{
		"protegido ~/.claude": claudeDir,
		"HOME entero":         home,
		"relativo":            "proyecto-relativo",
		"inexistente":         filepath.Join(home, "no-existe"),
	}
	for nombre, dir := range casos {
		if _, err := svc.Sembrar(ctx, dir); err == nil {
			t.Errorf("Sembrar(%s): quiero error, vino nil", nombre)
		}
		if _, err := svc.Chequear(ctx, dir); err == nil {
			t.Errorf("Chequear(%s): quiero error, vino nil", nombre)
		}
	}
	if len(puerto.sembrados)+len(puerto.chequeados) != 0 {
		t.Errorf("el puerto se llamó con un dir inválido: sembrados=%v chequeados=%v", puerto.sembrados, puerto.chequeados)
	}
}

// TestForjaServiceDelegaConDirValido: dir legal ⇒ delega al puerto tal cual (limpio).
func TestForjaServiceDelegaConDirValido(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	proyecto := filepath.Join(home, "proyectos", "demo")
	if err := os.MkdirAll(proyecto, 0o750); err != nil {
		t.Fatal(err)
	}

	puerto := &fakeForjaPort{}
	svc := usecase.NewForjaService(puerto)
	inf, err := svc.Sembrar(context.Background(), proyecto)
	if err != nil {
		t.Fatalf("Sembrar: %v", err)
	}
	if len(inf.Creados) != 1 {
		t.Errorf("informe no viajó del puerto: %+v", inf)
	}
	salud, err := svc.Chequear(context.Background(), proyecto)
	if err != nil || salud.Estado != domain.SemillaSana {
		t.Fatalf("Chequear: salud=%+v err=%v", salud, err)
	}
	if len(puerto.sembrados) != 1 || len(puerto.chequeados) != 1 {
		t.Errorf("delegación = sembrados %v chequeados %v, quiero 1 y 1", puerto.sembrados, puerto.chequeados)
	}
}
