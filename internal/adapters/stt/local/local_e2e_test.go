package local

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// EnvAudioE2E apunta a un archivo de audio real. Sin él, el test SKIPEA.
//
// Los demás tests de este paquete usan stubs del runner: prueban el contrato (qué argv sale,
// de dónde se lee el transcripto, que no queda audio en disco) sin depender de qué tenga
// instalada la máquina. Este test es el otro lado: corre **el motor de verdad** sobre **audio
// de verdad**, que es lo único que caza los desalineados entre lo que creemos que hace el CLI
// y lo que hace (así se destaparon los dos bugs de la primera versión del adaptador: leer
// stdout en vez del `.txt`, y no fijar `--device cpu`, que hace explotar al CLI con
// `Library libcublas.so.12 is not found` en máquinas sin NVIDIA).
//
// Correrlo:
//
//	PATH="$HOME/.local/share/arnesia-stt/bin:$PATH" \
//	ARNESIA_STT_E2E_AUDIO=/ruta/dictado.wav \
//	go test ./internal/adapters/stt/local/ -run E2E -v
//
// Skipea por defecto a propósito: exige un motor instalado + un modelo descargado, y CI no los
// tiene. Que skipee es honesto; fingir que pasó no lo sería.
const EnvAudioE2E = "ARNESIA_STT_E2E_AUDIO"

// EnvMimeE2E overrides the mime of the E2E audio (default `audio/wav`).
const EnvMimeE2E = "ARNESIA_STT_E2E_MIME"

// TestE2ETranscribeConMotorReal corre el adaptador REAL contra el motor REAL.
func TestE2ETranscribeConMotorReal(t *testing.T) {
	audioPath := os.Getenv(EnvAudioE2E)
	if audioPath == "" {
		t.Skipf("skip: fijá %s a un archivo de audio para correr el E2E", EnvAudioE2E)
	}

	a := New()
	d := a.Disponible(context.Background())
	if !d.Disponible {
		t.Skipf("skip: no hay motor usable — %s (instalar: %v)", d.Motivo, d.Instalar)
	}
	t.Logf("motor detectado: %s", d.Motor)

	audio, err := os.ReadFile(audioPath)
	if err != nil {
		t.Fatalf("leer audio: %v", err)
	}
	mime := os.Getenv(EnvMimeE2E)
	if mime == "" {
		mime = "audio/wav"
	}

	t0 := time.Now()
	texto, err := a.Transcribir(context.Background(), audio, mime)
	lat := time.Since(t0)
	if err != nil {
		t.Fatalf("Transcribir: %v", err)
	}
	t.Logf("latencia: %s · %d chars", lat.Round(time.Millisecond), len(texto))
	t.Logf("transcripto: %q", texto)

	if strings.TrimSpace(texto) == "" {
		t.Fatal("el transcripto vino vacío: el motor corrió pero no devolvió texto")
	}
	// La regresión concreta que este test existe para cazar: el ruido de stdout del CLI
	// colándose como si fuera el dictado del operador.
	for _, ruido := range []string{"Detected language", "written to", "directory"} {
		if strings.Contains(texto, ruido) {
			t.Errorf("el transcripto trae ruido del CLI (%q): se volvió a leer stdout en vez del .txt", ruido)
		}
	}
}
