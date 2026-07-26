// Package logfile writes the daemon's log to a rotating file under ~/.arnesia/logs/.
//
// **Por qué existe** (paquete 2026-07-25-spike-voz-dictado, RF-230): hasta acá el daemon
// logueaba SOLO a stderr, y la app instalada arranca desde el `.desktop` del `.deb` — que no
// tiene terminal donde mirar ese stderr. En la práctica eso significó que un bug del dictado
// viviera una versión entera sin dejar rastro: lo único que quedaba del incidente era la
// frase que el operador leyó en la pantalla.
//
// El archivo es el destino DE LA CASA: local, del usuario, sin red. La telemetría de producto
// (OTel, HS-27) es otro paquete, otra decisión y otro destino — esto no la adelanta ni la
// reemplaza.
package logfile

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// maxDefault is the size at which the log rotates. 8 MB de JSON son decenas de miles de
// líneas: alcanza de sobra para reconstruir un incidente sin volverse un archivo que nadie
// puede abrir.
const maxDefault = 8 << 20

// Writer is an io.Writer that rotates when the file grows past max.
//
// Guarda UNA generación vieja (`arnesia.log.1`) y nada más. Dos rotaciones son suficientes
// para no perder el contexto anterior a un crash, y un esquema de N archivos con fechas
// pediría una política de retención que nadie va a operar.
type Writer struct {
	mu   sync.Mutex
	path string
	max  int64
	f    *os.File
	n    int64
}

// Ruta returns the default log path (~/.arnesia/logs/arnesia.log), mirroring the convention
// of the daemon's other file-backed adapters (sessions.json, index.db).
func Ruta() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("logfile: resolver home: %w", err)
	}
	return filepath.Join(home, ".arnesia", "logs", "arnesia.log"), nil
}

// Abrir opens (creating the parent dir) the log at path. An empty path uses Ruta(); a max of
// zero uses the default size.
//
// Abre en modo append: dos daemons a la vez es un accidente, pero si pasa, el segundo NO le
// trunca el log al primero.
func Abrir(path string, max int64) (*Writer, error) {
	if path == "" {
		p, err := Ruta()
		if err != nil {
			return nil, err
		}
		path = p
	}
	if max <= 0 {
		max = maxDefault
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("logfile: crear dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644) //nolint:gosec // G304: la ruta la fija el daemon (flag o default), nunca un request.
	if err != nil {
		return nil, fmt.Errorf("logfile: abrir %s: %w", path, err)
	}
	var n int64
	if st, serr := f.Stat(); serr == nil {
		n = st.Size()
	}
	return &Writer{path: path, max: max, f: f, n: n}, nil
}

// Write appends p, rotating first when it would push the file past max.
//
// Un error de rotación NO pierde la línea: se sigue escribiendo en el archivo actual. Perder
// el log por no poder rotarlo sería exactamente el silencio que este paquete vino a sacar.
func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return 0, os.ErrClosed
	}
	if w.n+int64(len(p)) > w.max {
		_ = w.rotar()
	}
	n, err := w.f.Write(p)
	w.n += int64(n)
	return n, err
}

// Close closes the underlying file.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return nil
	}
	err := w.f.Close()
	w.f = nil
	return err
}

// Path returns the file being written, so the daemon can name it at boot: un log que nadie
// sabe dónde está no se lee.
func (w *Writer) Path() string { return w.path }

// rotar closes the current file, moves it to .1 and opens a fresh one. Se llama con el mutex
// tomado.
func (w *Writer) rotar() error {
	if err := w.f.Close(); err != nil {
		return err
	}
	w.f = nil
	if err := os.Rename(w.path, w.path+".1"); err != nil && !os.IsNotExist(err) {
		// Reabrir igual: sin archivo abierto, el daemon se queda mudo.
		f, oerr := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644) //nolint:gosec // G304: idem, ruta del daemon.
		if oerr != nil {
			return oerr
		}
		w.f = f
		return err
	}
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644) //nolint:gosec // G304: idem, ruta del daemon.
	if err != nil {
		return err
	}
	w.f = f
	w.n = 0
	return nil
}
