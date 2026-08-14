package local

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// pathCon builds a LookPath stub where only the named binaries exist.
func pathCon(bins ...string) func(string) (string, error) {
	set := make(map[string]bool, len(bins))
	for _, b := range bins {
		set[b] = true
	}
	return func(bin string) (string, error) {
		if set[bin] {
			return "/usr/bin/" + bin, nil
		}
		return "", errors.New("not found")
	}
}

// corrida records one invocation of the runner.
type corrida struct {
	bin  string
	args []string
}

// runner returns a stub runner that records calls and answers with salida on stdout.
func runner(reg *[]corrida, salida string) func(context.Context, string, ...string) ([]byte, error) {
	return func(_ context.Context, bin string, args ...string) ([]byte, error) {
		*reg = append(*reg, corrida{bin: filepath.Base(bin), args: args})
		return []byte(salida), nil
	}
}

// runnerQueEscribeArchivo imita a un motor de ARCHIVO (whisper-ctranslate2): imprime ruido en
// stdout y deja el transcripto en `<outDir>/<basename>.txt`. El ruido es literal el que
// devuelve el binario real — así el test falla si alguien vuelve a leer stdout.
func runnerQueEscribeArchivo(reg *[]corrida, transcripto string) func(context.Context, string, ...string) ([]byte, error) {
	return func(_ context.Context, bin string, args ...string) ([]byte, error) {
		*reg = append(*reg, corrida{bin: filepath.Base(bin), args: args})
		if filepath.Base(bin) == "ffmpeg" {
			return nil, nil
		}
		audio := args[len(args)-1]
		outDir := ""
		for i, a := range args {
			if a == "--output_dir" && i+1 < len(args) {
				outDir = args[i+1]
			}
		}
		base := strings.TrimSuffix(filepath.Base(audio), filepath.Ext(audio))
		if outDir != "" {
			if err := os.WriteFile(filepath.Join(outDir, base+".txt"), []byte(transcripto), 0o600); err != nil {
				return nil, err
			}
		}
		return []byte("Detected language 'Spanish' with probability 1.000000\n" +
			"Transcription results written to '" + outDir + "' directory\n"), nil
	}
}

// conModeloCpp fija el env que whisper.cpp necesita y lo restaura al salir.
func conModeloCpp(t *testing.T) {
	t.Helper()
	t.Setenv(EnvModeloWhisperCpp, "/opt/whisper/models/ggml-base.bin")
}

// TestAdapterLocalEligeElPrimerMotorDelPath — V-D1: el adaptador DETECTA, no hardcodea.
func TestAdapterLocalEligeElPrimerMotorDelPath(t *testing.T) {
	casos := []struct {
		nombre    string
		instalado []string
		quiero    string
	}{
		{"whisper.cpp gana por preferencia", []string{"whisper-cli", "whisper-ctranslate2", "ffmpeg"}, "whisper-cli"},
		{"sin ffmpeg cae al que lee contenedores", []string{"whisper-cli", "whisper-ctranslate2"}, "whisper-ctranslate2"},
		{"solo ctranslate2", []string{"whisper-ctranslate2"}, "whisper-ctranslate2"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			conModeloCpp(t)
			a := &Adapter{look: pathCon(c.instalado...), run: runner(&[]corrida{}, "")}
			e, _, ok := a.Detectar()
			if !ok {
				t.Fatal("no detectó ningún motor con uno instalado")
			}
			if e.bin != c.quiero {
				t.Errorf("motor = %q, quiero %q", e.bin, c.quiero)
			}
			// La ruta ABSOLUTA es parte del contrato: correr por nombre pelado volvería a
			// depender del $PATH del proceso, que es lo que falta en el launcher gráfico.
			if !filepath.IsAbs(e.ruta) {
				t.Errorf("ruta = %q, se esperaba absoluta", e.ruta)
			}
		})
	}
}

// TestDisponibleSinMotorDiceMotivoYQueInstalar — la cara visible de la degradación honesta.
func TestDisponibleSinMotorDiceMotivoYQueInstalar(t *testing.T) {
	a := &Adapter{look: pathCon(), run: runner(&[]corrida{}, "")}

	d := a.Disponible(context.Background())
	if d.Disponible {
		t.Fatal("sin motor no puede estar disponible")
	}
	if d.Motivo == "" {
		t.Error("un no-disponible sin motivo es un gap escondido")
	}
	if len(d.Instalar) == 0 {
		t.Error("hay que decir QUÉ instalar, no solo que falta algo")
	}
}

// TestElHintDeInstalarSoloOfreceBinariosQueExisten — `faster-whisper` es una LIBRERÍA: no
// expone ejecutable (verificado instalándola). Ofrecerla mandaría al operador a instalar algo
// que no serviría, y el botón seguiría gris sin que entienda por qué.
func TestElHintDeInstalarSoloOfreceBinariosQueExisten(t *testing.T) {
	a := &Adapter{look: pathCon(), run: runner(&[]corrida{}, "")}

	for _, b := range a.Disponible(context.Background()).Instalar {
		if b == "faster-whisper" {
			t.Error("el hint ofrece `faster-whisper`, que no tiene CLI — el binario real es whisper-ctranslate2")
		}
	}
}

// TestDisponibleAvisaQueFaltaFfmpeg — whisper.cpp solo come WAV. Si es el ÚNICO motor y falta
// el transcodificador, decirlo ahora es mejor que fallar tras 3 min de dictado.
func TestDisponibleAvisaQueFaltaFfmpeg(t *testing.T) {
	conModeloCpp(t)
	a := &Adapter{look: pathCon("whisper-cli"), run: runner(&[]corrida{}, "")}

	d := a.Disponible(context.Background())
	if d.Disponible {
		t.Fatal("whisper-cli sin ffmpeg no puede transcribir el audio/mp4 del WebView")
	}
	if !strings.Contains(d.Motivo, "ffmpeg") {
		t.Errorf("el motivo no nombra ffmpeg: %q", d.Motivo)
	}
}

// TestWhisperCppSinModeloNoEsUsableYLoDice — whisper.cpp no tiene ubicación por defecto para
// sus pesos: toma `-m <path>` y no hay convención que adivinar. Estar en $PATH no alcanza.
func TestWhisperCppSinModeloNoEsUsableYLoDice(t *testing.T) {
	t.Setenv(EnvModeloWhisperCpp, "")
	a := &Adapter{look: pathCon("whisper-cli", "ffmpeg"), run: runner(&[]corrida{}, "")}

	d := a.Disponible(context.Background())
	if d.Disponible {
		t.Fatal("whisper-cli sin modelo no es usable")
	}
	if !strings.Contains(d.Motivo, EnvModeloWhisperCpp) {
		t.Errorf("el motivo no dice qué variable fijar: %q", d.Motivo)
	}
}

// TestMotorInstaladoPeroIncompletoNoDiceQueNoHayNada — mandar a «instalar un motor» a quien ya
// lo tiene instalado es peor que no decir nada: lo hace reinstalar lo que ya está.
func TestMotorInstaladoPeroIncompletoNoDiceQueNoHayNada(t *testing.T) {
	t.Setenv(EnvModeloWhisperCpp, "")
	a := &Adapter{look: pathCon("whisper-cli", "ffmpeg"), run: runner(&[]corrida{}, "")}

	d := a.Disponible(context.Background())
	if strings.Contains(d.Motivo, "no hay motor de transcripción instalado") {
		t.Errorf("motivo genérico con un motor presente: %q", d.Motivo)
	}
	if len(d.Instalar) != 0 {
		t.Errorf("no hay que ofrecer instalar nada: falta config, no software (%v)", d.Instalar)
	}
}

// TestTranscribirSinMotorFallaExplicito — jamás un panic ni un string vacío que parezca voz.
func TestTranscribirSinMotorFallaExplicito(t *testing.T) {
	a := &Adapter{look: pathCon(), run: runner(&[]corrida{}, "")}

	if _, err := a.Transcribir(context.Background(), []byte("x"), "audio/mp4"); !errors.Is(err, ErrSinMotor) {
		t.Fatalf("err = %v, quiero ErrSinMotor", err)
	}
}

// TestTranscribirRechazaMimeNoSoportado — el formato lo fija el motor del WebView; lo que no
// sabemos manejar se dice, no se adivina.
func TestTranscribirRechazaMimeNoSoportado(t *testing.T) {
	a := &Adapter{look: pathCon("whisper-ctranslate2"), run: runner(&[]corrida{}, "")}

	if _, err := a.Transcribir(context.Background(), []byte("x"), "audio/flac"); !errors.Is(err, ErrMimeNoSoportado) {
		t.Fatalf("err = %v, quiero ErrMimeNoSoportado", err)
	}
}

// TestTranscribirLeeElArchivoNoStdout — **el bug que destapó correr el binario de verdad.**
// `whisper-ctranslate2` imprime «Detected language…» / «Transcription results written to…» en
// stdout y deja el transcripto en un `.txt`. Leer stdout le poblaba el composer al operador
// con la charla del CLI en vez de con su dictado.
func TestTranscribirLeeElArchivoNoStdout(t *testing.T) {
	var reg []corrida
	a := &Adapter{
		look: pathCon("whisper-ctranslate2"),
		run:  runnerQueEscribeArchivo(&reg, " Cambiá el color del pip. "),
		tmp:  t.TempDir(),
	}

	got, err := a.Transcribir(context.Background(), []byte("audio"), "audio/mp4")
	if err != nil {
		t.Fatalf("Transcribir: %v", err)
	}
	if got != "Cambiá el color del pip." {
		t.Errorf("texto = %q — si trae «Detected language» se volvió a leer stdout", got)
	}
	if strings.Contains(got, "Detected language") || strings.Contains(got, "written to") {
		t.Error("el ruido de stdout del CLI se colaría como dictado")
	}
}

// TestTranscribirFijaCpuYInt8 — sin `--device cpu` el CLI intenta CUDA y **explota**
// (`Library libcublas.so.12 is not found`) en cualquier máquina sin toolkit de NVIDIA.
// Probado en vivo. `int8` es además la config medida en §1.8.
func TestTranscribirFijaCpuYInt8(t *testing.T) {
	var reg []corrida
	a := &Adapter{
		look: pathCon("whisper-ctranslate2"),
		run:  runnerQueEscribeArchivo(&reg, "texto"),
		tmp:  t.TempDir(),
	}
	if _, err := a.Transcribir(context.Background(), []byte("audio"), "audio/mp4"); err != nil {
		t.Fatalf("Transcribir: %v", err)
	}
	args := strings.Join(reg[0].args, " ")
	for _, quiero := range []string{"--device cpu", "--compute_type int8", "--model " + modelo} {
		if !strings.Contains(args, quiero) {
			t.Errorf("falta %q en el argv: %s", quiero, args)
		}
	}
	if strings.Contains(args, "--output_dir -") {
		t.Error("`--output_dir -` no es stdout: crearía un directorio llamado `-`")
	}
}

// TestAdapterLocalBorraElTemporal — V-D5: el audio NO se persiste. El temporal es la única
// excepción y tiene que morir en el mismo llamado — igual que el dir de salida del motor.
func TestAdapterLocalBorraElTemporal(t *testing.T) {
	dir := t.TempDir()
	var reg []corrida
	a := &Adapter{
		look: pathCon("whisper-ctranslate2"),
		run:  runnerQueEscribeArchivo(&reg, "texto"),
		tmp:  dir,
	}

	if _, err := a.Transcribir(context.Background(), []byte("audio-real"), "audio/mp4"); err != nil {
		t.Fatalf("Transcribir: %v", err)
	}
	quedaron, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(quedaron) != 0 {
		nombres := make([]string, 0, len(quedaron))
		for _, e := range quedaron {
			nombres = append(nombres, e.Name())
		}
		t.Errorf("quedó rastro en disco: %v — V-D5 dice que el audio no se persiste", nombres)
	}
}

// TestTranscribirTranscodificaParaMotoresQueSoloComenWav — responsabilidad del ADAPTADOR,
// no del dominio ni del FE (RF-223).
func TestTranscribirTranscodificaParaMotoresQueSoloComenWav(t *testing.T) {
	conModeloCpp(t)
	dir := t.TempDir()
	var reg []corrida
	a := &Adapter{look: pathCon("whisper-cli", "ffmpeg"), run: runner(&reg, "texto"), tmp: dir}

	if _, err := a.Transcribir(context.Background(), []byte("audio"), "audio/mp4"); err != nil {
		t.Fatalf("Transcribir: %v", err)
	}
	if len(reg) != 2 {
		t.Fatalf("corridas = %d, quiero 2 (ffmpeg + motor): %+v", len(reg), reg)
	}
	if reg[0].bin != "ffmpeg" {
		t.Errorf("la primera corrida fue %q, quiero ffmpeg", reg[0].bin)
	}
	// 16 kHz mono es lo que whisper.cpp exige; si eso se pierde, el motor devuelve basura.
	args := strings.Join(reg[0].args, " ")
	for _, quiero := range []string{"-ar 16000", "-ac 1"} {
		if !strings.Contains(args, quiero) {
			t.Errorf("falta %q en el transcodificado: %s", quiero, args)
		}
	}
	if reg[1].bin != "whisper-cli" {
		t.Errorf("la segunda corrida fue %q, quiero whisper-cli", reg[1].bin)
	}
	if !strings.HasSuffix(reg[1].args[len(reg[1].args)-1], ".wav") {
		t.Errorf("al motor no le llegó el WAV: %v", reg[1].args)
	}
	// El modelo viaja por `-m`, tomado del env: sin eso whisper.cpp no arranca.
	if !strings.Contains(strings.Join(reg[1].args, " "), "ggml-base.bin") {
		t.Errorf("el modelo no llegó al motor: %v", reg[1].args)
	}
	quedaron, _ := os.ReadDir(dir)
	if len(quedaron) != 0 {
		t.Errorf("quedaron %d archivos en disco", len(quedaron))
	}
}

// TestTranscribirNoTranscodificaSiYaEsWav — no se paga un ffmpeg de gusto.
func TestTranscribirNoTranscodificaSiYaEsWav(t *testing.T) {
	conModeloCpp(t)
	var reg []corrida
	a := &Adapter{look: pathCon("whisper-cli", "ffmpeg"), run: runner(&reg, "texto"), tmp: t.TempDir()}

	if _, err := a.Transcribir(context.Background(), []byte("audio"), "audio/wav"); err != nil {
		t.Fatalf("Transcribir: %v", err)
	}
	if len(reg) != 1 || reg[0].bin != "whisper-cli" {
		t.Errorf("se transcodificó un WAV al pedo: %+v", reg)
	}
}

// TestTranscribirAceptaElMimeConParametros — los navegadores mandan
// "audio/mp4;codecs=mp4a.40.2"; fallar por eso sería romper por cosmética.
func TestTranscribirAceptaElMimeConParametros(t *testing.T) {
	var reg []corrida
	a := &Adapter{
		look: pathCon("whisper-ctranslate2"),
		run:  runnerQueEscribeArchivo(&reg, "  hola  "),
		tmp:  t.TempDir(),
	}

	got, err := a.Transcribir(context.Background(), []byte("x"), "AUDIO/MP4; codecs=mp4a.40.2")
	if err != nil {
		t.Fatalf("Transcribir: %v", err)
	}
	if got != "hola" {
		t.Errorf("texto = %q (se esperaba recortado)", got)
	}
}

// TestElModeloEsElFirmado — V-D1 firmó `base`. Si alguien lo sube a `small` sin reabrir la
// decisión, este test lo dice: se paga 2.5× de latencia por gramática que el contexto ya da.
func TestElModeloEsElFirmado(t *testing.T) {
	if modelo != "base" {
		t.Errorf("modelo = %q; V-D1 firmó `base` — reabrí la decisión antes de cambiarlo", modelo)
	}
}

// TestPathAumentadoEncuentraElMotorFueraDelPathHeredado — **bug real, mismo root cause que ya
// arregló `selfupdate.pathAumentado`:** un proceso lanzado desde el launcher gráfico (el
// `.desktop` del `.deb`) NO hereda `~/.profile` ni `~/.bashrc`. Sin este fallback, el motor que
// el operador instaló —y que CUALQUIER terminal ve— queda invisible para la app instalada, y el
// botón se queda gris diciendo «falta el motor» con el motor ahí puesto.
func TestPathAumentadoEncuentraElMotorFueraDelPathHeredado(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows: os.UserHomeDir lee USERPROFILE, no HOME
	// PATH heredado VACÍO: es el escenario del launcher gráfico.
	t.Setenv("PATH", t.TempDir())

	venv := filepath.Join(home, ".local", "share", "arnesia-stt", "bin")
	if err := os.MkdirAll(venv, 0o755); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(venv, "whisper-ctranslate2")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil { //nolint:gosec // G306: es un stub ejecutable de test.
		t.Fatal(err)
	}

	got, err := lookPathAumentado("whisper-ctranslate2")
	if err != nil {
		t.Fatalf("el motor del venv quedó invisible: %v", err)
	}
	if got != bin {
		t.Errorf("ruta = %q, quiero %q", got, bin)
	}
	// Y el adaptador real lo tiene que ver como disponible.
	if d := New().Disponible(context.Background()); !d.Disponible {
		t.Errorf("Disponible dice que no con el motor instalado en el venv: %+v", d)
	}
}

// TestPathAumentadoNoInventaLoQueNoEsta — el fallback no puede volverse un falso positivo.
func TestPathAumentadoNoInventaLoQueNoEsta(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows: os.UserHomeDir lee USERPROFILE, no HOME
	t.Setenv("PATH", t.TempDir())

	if _, err := lookPathAumentado("whisper-ctranslate2"); err == nil {
		t.Fatal("sin motor en ningún lado, lookPathAumentado no puede devolver una ruta")
	}
}
