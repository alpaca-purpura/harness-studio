// Package local is the local speech-to-text adapter: it transcribes with whatever engine
// the operator has installed, found on $PATH.
//
// # Por qué "el que haya" y no un binario fijo
//
// V-D1 se firmó como **A2 local, modelo `base`** — el rendimiento dejó de ser el argumento
// (2.2 s para 47 s de voz). Lo que quedó abierto es el EMPAQUETADO, y esa sub-decisión
// (`whisper.cpp` vs `faster-whisper`) **no se puede cerrar con lo medido**: se midió
// `faster-whisper` sobre WAV con voz sintética, no `whisper.cpp` y no el `audio/mp4` real
// que produce WebKitGTK. Elegir el binario hoy sería fabricar una decisión sobre evidencia
// que no existe.
//
// Así que el adaptador **detecta**: busca los motores que conoce en `$PATH` y usa el primero.
// Si no hay ninguno, lo DICE (Disponibilidad.Motivo) en vez de romper o fingir. La deuda de
// qué se bundlea en los instaladores queda registrada como T7 en el paquete
// docs/product/stories/2026-07-25-spike-voz-dictado/, con su criterio de cierre escrito.
package local

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// modelo is the Whisper model size the adapter asks for. `base` es lo firmado (V-D1): subir
// a `small` compra gramática pero NO vocabulario de dominio —los tres modelos escribieron
// "locita" por `lógica`— y eso lo repara el contexto del paso siguiente, no el tamaño del
// modelo. Pagar 2.5× de latencia por eso es plata tirada.
const modelo = "base"

// EnvModeloWhisperCpp is the env var that points at the whisper.cpp model file (a `.bin`
// ggml weights file). whisper.cpp has **no default model location** — it takes `-m <path>`
// and there is no convention to guess. Sin esta variable el motor queda no-usable **y se
// dice** (Disponibilidad.Motivo), en vez de adivinar una ruta y fallar recién cuando el
// operador ya habló tres minutos.
const EnvModeloWhisperCpp = "ARNESIA_WHISPER_MODEL"

// dondeSale says where a engine puts its transcript. **Verificado corriendo los binarios de
// verdad, no leyendo docs** — `whisper-ctranslate2` imprime en stdout «Detected language …» y
// «Transcription results written to '<dir>' directory», así que leer stdout devolvería ESE
// texto como si fuera el dictado.
type dondeSale int

const (
	// enStdout — el transcripto sale por stdout (whisper.cpp con `-nt`).
	enStdout dondeSale = iota
	// enArchivoTxt — el transcripto queda en `<outDir>/<basename>.txt`.
	enArchivoTxt
)

// motor is one supported engine: how to find it, how to call it, and where its output lands.
type motor struct {
	// bin is the executable name looked up on $PATH.
	bin string
	// wav is true when the engine only eats WAV 16 kHz mono (whisper.cpp does), which
	// forces a transcode step. Engines that read containers directly set it false.
	wav bool
	// salida is where the transcript comes from.
	salida dondeSale
	// args builds the argv (sans binary). outDir is where file-based engines must write.
	args func(path, outDir string) []string
	// listo reports whether the engine is actually usable, and why not when it is not.
	// nil = usable con solo estar en $PATH.
	listo func(a *Adapter) string
}

// motores are the engines we know, in preference order: whisper.cpp first (binario chico, el
// que empaquetaría más limpio si T7 cierra por ahí), whisper-ctranslate2 después (el CLI real
// del `faster-whisper` que midió §1.8 — la librería NO expone ejecutable propio, así que el
// binario a buscar es este).
var motores = []motor{
	{
		bin:    "whisper-cli",
		wav:    true,
		salida: enStdout,
		args: func(path, _ string) []string {
			// `-nt` = sin timestamps: queremos el texto pelado, no `[00:00.000 --> …]`.
			return []string{"-m", os.Getenv(EnvModeloWhisperCpp), "-l", "es", "-nt", "-f", path}
		},
		listo: func(a *Adapter) string {
			if os.Getenv(EnvModeloWhisperCpp) == "" {
				return "whisper-cli está instalado pero falta el modelo: fijá " + EnvModeloWhisperCpp + " a un archivo ggml-" + modelo + ".bin"
			}
			return ""
		},
	},
	{
		bin:    "whisper-ctranslate2",
		wav:    false,
		salida: enArchivoTxt,
		args: func(path, outDir string) []string {
			return []string{
				"--model", modelo,
				"--language", "es",
				// `--device cpu --compute_type int8` es la config MEDIDA en §1.8 (2.2 s
				// para 47 s de voz). Y no es solo performance: sin fijar el device, el CLI
				// intenta CUDA y **explota** con `Library libcublas.so.12 is not found`
				// en cualquier máquina sin toolkit de NVIDIA — probado en vivo.
				"--device", "cpu",
				"--compute_type", "int8",
				"--output_format", "txt",
				"--output_dir", outDir,
				path,
			}
		},
	},
}

// mimesSoportados maps the mime the recorder produced to a file extension. WebKitGTK 2.52.3
// only records `audio/mp4`, but the adapter accepts the usual suspects so a different engine
// (or a future WebKit) does not need a code change here.
var mimesSoportados = map[string]string{
	"audio/mp4":   ".mp4",
	"audio/mpeg":  ".mp3",
	"audio/wav":   ".wav",
	"audio/x-wav": ".wav",
	"audio/webm":  ".webm",
	"audio/ogg":   ".ogg",
}

// candidatosPATH are the dirs to look in beyond the inherited $PATH.
//
// **Root cause (bug real, mismo que ya arregló `selfupdate.pathAumentado`):** un proceso
// lanzado desde el launcher gráfico —el `.desktop` del `.deb`— **no hereda `~/.profile` ni
// `~/.bashrc`** como sí hace una terminal interactiva. Sin esto, el motor STT que el operador
// instaló y que CUALQUIER terminal ve queda invisible para la app instalada, y el botón de
// dictado se quedaría gris diciendo «falta el motor» con el motor ahí puesto. El fix va en el
// código, no pidiéndole al operador que reinicie sesión.
func candidatosPATH() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		// Convención XDG de binarios de usuario: es donde aterrizan `pipx install` y los
		// symlinks de un venv propio.
		filepath.Join(home, ".local", "bin"),
		// Venv aislado recomendado para el motor (el patrón «sin sudo» del spike).
		filepath.Join(home, ".local", "share", "arnesia-stt", "bin"),
		"/usr/local/bin",
	}
}

// lookPathAumentado resolves bin on the inherited $PATH first, then on candidatosPATH().
// Devuelve la ruta ABSOLUTA: correr por nombre pelado volvería a depender del $PATH del
// proceso, que es justo lo que falla en el launcher gráfico. En Windows no hay bit de
// ejecución (NTFS reporta 0666) y los binarios llevan extensión — se prueba PATHEXT
// mínimo en vez del chequeo de modo.
func lookPathAumentado(bin string) (string, error) {
	if p, err := exec.LookPath(bin); err == nil {
		return p, nil
	}
	sufijos := []string{""}
	if runtime.GOOS == "windows" {
		sufijos = []string{".exe", ".cmd", ".bat", ""}
	}
	for _, dir := range candidatosPATH() {
		for _, suf := range sufijos {
			cand := filepath.Join(dir, bin+suf)
			st, err := os.Stat(cand)
			if err != nil || st.IsDir() {
				continue
			}
			if runtime.GOOS != "windows" && st.Mode()&0o111 == 0 {
				continue
			}
			return cand, nil
		}
	}
	return "", exec.ErrNotFound
}

// elegido is the engine that won detection, plus the absolute path it resolved to.
type elegido struct {
	motor
	ruta string
}

// Adapter transcribes audio with a locally installed engine.
type Adapter struct {
	// look resolves a binary to its absolute path. Injectable so the tests do not depend on
	// what the machine running them happens to have installed.
	look func(string) (string, error)
	// run executes the engine and returns its stdout. Injectable for the same reason.
	run func(ctx context.Context, bin string, args ...string) ([]byte, error)
	// tmp is the directory temporaries are created in; empty means os.TempDir().
	tmp string
}

var _ ports.TranscriptionPort = (*Adapter)(nil)

// New returns an Adapter wired to the augmented $PATH and the real process runner.
func New() *Adapter {
	return &Adapter{look: lookPathAumentado, run: correr}
}

// correr executes bin and returns its stdout, folding stderr into the error so a failing
// engine explains itself instead of returning a bare exit code.
func correr(ctx context.Context, bin string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, bin, args...) //nolint:gosec // G204: bin comes from our own `motores` allow-list resolved via LookPath — never from the request.
	var errBuf strings.Builder
	cmd.Stderr = &errBuf
	out, err := cmd.Output()
	if err != nil {
		if msg := strings.TrimSpace(errBuf.String()); msg != "" {
			return nil, fmt.Errorf("%s: %w: %s", bin, err, msg)
		}
		return nil, fmt.Errorf("%s: %w", bin, err)
	}
	return out, nil
}

// Detectar returns the first known engine found on $PATH **and usable**, or ok=false when
// there is none. `motivo` explains the last near-miss (instalado pero incompleto), que es más
// útil que un «no hay nada» cuando el operador ya instaló algo.
func (a *Adapter) Detectar() (e elegido, motivo string, ok bool) {
	for _, cand := range motores {
		ruta, err := a.look(cand.bin)
		if err != nil {
			continue
		}
		if cand.listo != nil {
			if falta := cand.listo(a); falta != "" {
				motivo = falta
				continue
			}
		}
		// whisper.cpp only eats WAV: sin transcodificador el mp4 que graba el WebView es
		// inutilizable, así que el motor NO está usable — no es un detalle posterior.
		if cand.wav {
			if _, err := a.look("ffmpeg"); err != nil {
				motivo = cand.bin + " necesita WAV y falta ffmpeg para convertir la grabación"
				continue
			}
		}
		return elegido{motor: cand, ruta: ruta}, "", true
	}
	return elegido{}, motivo, false
}

// nombres lists the engines a operator could install, for the "qué instalar" hint. Solo
// binarios que EXISTEN de verdad: `faster-whisper` (la librería) no expone ejecutable, así
// que ofrecerlo mandaría al operador a instalar algo que no serviría.
func nombres() []string {
	out := make([]string, 0, len(motores))
	for _, m := range motores {
		out = append(out, m.bin)
	}
	return out
}

// Disponible reports whether transcription can run right now, and why not when it cannot.
func (a *Adapter) Disponible(_ context.Context) ports.Disponibilidad {
	e, motivo, ok := a.Detectar()
	if ok {
		return ports.Disponibilidad{Disponible: true, Motor: e.bin}
	}
	if motivo != "" {
		// Hay un motor instalado pero incompleto: decir QUÉ le falta es más útil que
		// «no hay motor», que mandaría a reinstalar lo que ya está.
		return ports.Disponibilidad{Motivo: motivo}
	}
	return ports.Disponibilidad{
		Motivo:   "no hay motor de transcripción instalado",
		Instalar: nombres(),
	}
}

// ErrSinMotor is returned by Transcribir when no engine is installed.
var ErrSinMotor = errors.New("stt: no hay motor de transcripción instalado")

// ErrMimeNoSoportado is returned when the recorder's mime is one we cannot hand to an engine.
var ErrMimeNoSoportado = errors.New("stt: formato de audio no soportado")

// Transcribir writes audio to a short-lived temporary, runs the engine over it, and returns
// the transcript.
//
// **El temporal es la única excepción a "el audio no se persiste"** (V-D5): un motor de
// línea de comandos necesita un archivo para leer. Se crea, se usa y se borra en el mismo
// `defer` — nunca queda un archivo persistente, y nunca se escribe nada en el árbol de un
// arnés.
func (a *Adapter) Transcribir(ctx context.Context, audio []byte, mime string) (string, error) {
	e, motivo, ok := a.Detectar()
	if !ok {
		if motivo != "" {
			return "", fmt.Errorf("%w: %s", ErrSinMotor, motivo)
		}
		return "", ErrSinMotor
	}
	ext, ok := mimesSoportados[normalizarMime(mime)]
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrMimeNoSoportado, mime)
	}

	path, limpiar, err := a.escribirTemporal(audio, ext)
	if err != nil {
		return "", err
	}
	defer limpiar()

	if e.wav && ext != ".wav" {
		wav, limpiarWav, err := a.transcodificar(ctx, path)
		if err != nil {
			return "", err
		}
		defer limpiarWav()
		path = wav
	}

	// outDir es propio y efímero: el motor de archivo escribe ahí y se borra entero. Va
	// aparte del temporal del audio para no tener que adivinar qué archivos dejó el motor.
	outDir, err := os.MkdirTemp(a.tmp, "arnesia-dictado-out-*")
	if err != nil {
		return "", fmt.Errorf("stt: dir de salida: %w", err)
	}
	defer func() { _ = os.RemoveAll(outDir) }()

	// Se corre por RUTA ABSOLUTA (e.ruta), no por nombre: el nombre pelado volvería a
	// depender del $PATH del proceso, que es exactamente lo que falta en el launcher gráfico.
	out, err := a.run(ctx, e.ruta, e.args(path, outDir)...)
	if err != nil {
		return "", err
	}
	if e.salida == enStdout {
		return strings.TrimSpace(string(out)), nil
	}
	return leerTranscripto(outDir, path)
}

// leerTranscripto reads the `.txt` a file-based engine left in outDir.
//
// **NO se lee stdout para estos motores**: `whisper-ctranslate2` imprime ahí «Detected
// language 'Spanish'…» y «Transcription results written to '<dir>' directory» — devolver eso
// sería poblarle el composer al operador con la charla del CLI en vez de con su dictado.
// (Probado en vivo; era un bug real de la primera versión de este adaptador.)
func leerTranscripto(outDir, audioPath string) (string, error) {
	base := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
	b, err := os.ReadFile(filepath.Join(outDir, base+".txt")) //nolint:gosec // G304: outDir es un temporal nuestro y base viene de un nombre que nosotros generamos.
	if err == nil {
		return strings.TrimSpace(string(b)), nil
	}
	// Fallback por si el motor nombra distinto: el único `.txt` del dir efímero es el
	// transcripto (nadie más escribe ahí).
	entradas, derr := os.ReadDir(outDir)
	if derr != nil {
		return "", fmt.Errorf("stt: leer transcripto: %w", err)
	}
	for _, e := range entradas {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		b, rerr := os.ReadFile(filepath.Join(outDir, e.Name())) //nolint:gosec // G304: idem, dir efímero propio.
		if rerr == nil {
			return strings.TrimSpace(string(b)), nil
		}
	}
	return "", fmt.Errorf("stt: el motor no dejó transcripto en %s: %w", outDir, err)
}

// normalizarMime drops the parameters browsers append ("audio/mp4;codecs=mp4a.40.2") and
// lowercases, so the lookup does not miss on cosmetics.
func normalizarMime(mime string) string {
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = mime[:i]
	}
	return strings.ToLower(strings.TrimSpace(mime))
}

// escribirTemporal writes audio to a temp file and returns its path plus the cleanup that
// removes it. The cleanup is always safe to call.
func (a *Adapter) escribirTemporal(audio []byte, ext string) (string, func(), error) {
	f, err := os.CreateTemp(a.tmp, "arnesia-dictado-*"+ext)
	if err != nil {
		return "", func() {}, fmt.Errorf("stt: temporal: %w", err)
	}
	path := f.Name()
	limpiar := func() { _ = os.Remove(path) }
	if _, err := f.Write(audio); err != nil {
		_ = f.Close()
		limpiar()
		return "", func() {}, fmt.Errorf("stt: escribir temporal: %w", err)
	}
	if err := f.Close(); err != nil {
		limpiar()
		return "", func() {}, fmt.Errorf("stt: cerrar temporal: %w", err)
	}
	return path, limpiar, nil
}

// transcodificar converts src to WAV 16 kHz mono with ffmpeg, for engines that only eat that.
func (a *Adapter) transcodificar(ctx context.Context, src string) (string, func(), error) {
	dst := strings.TrimSuffix(src, filepath.Ext(src)) + ".wav"
	limpiar := func() { _ = os.Remove(dst) }
	// Por ruta absoluta, igual que el motor: en el launcher gráfico `ffmpeg` a secas
	// tampoco está en el PATH heredado.
	ffmpeg, err := a.look("ffmpeg")
	if err != nil {
		limpiar()
		return "", func() {}, fmt.Errorf("stt: falta ffmpeg para convertir la grabación: %w", err)
	}
	if _, err := a.run(ctx, ffmpeg, "-nostdin", "-y", "-i", src, "-ar", "16000", "-ac", "1", "-f", "wav", dst); err != nil {
		limpiar()
		return "", func() {}, fmt.Errorf("stt: transcodificar a wav: %w", err)
	}
	return dst, limpiar, nil
}
