// Package traer materializa un `domain.PlanTraer` en un directorio de staging: camino A (copia
// del checkout que Claude Code ya mantiene, sin red) y camino B (fetch shallow pineado por `sha`).
//
// El paquete NO mueve al destino, NO registra nada y NO limpia el temporal: eso es del usecase
// (BR-15) — así la atomicidad vive en UN lugar y los dos adapters quedan tontos.
package traer

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// excluidos es el set de exclusión de la copia y es EXACTAMENTE el de
// `portafolio.HashFormaPlugin` (design.md §12.2 riesgo 14 · E-94):
//
//   - excluir MENOS (llevarse `.git/`) haría que el canónico heredara el remote del MARKETPLACE
//     — peligroso: un futuro `Publicar` podría empujar al lugar equivocado;
//   - excluir MÁS alteraría contenido en silencio y el hash del canónico recién traído NO
//     coincidiría con el de la referencia ⇒ BR-17 reportaría `en-deriva` sobre algo sano.
//
// Es la garantía ESTRUCTURAL de que un traído sano dé `al-hilo`. Cambiar este set sin cambiar
// `HashFormaPlugin` rompe E-94, que lo asserta cruzado.
var excluidos = map[string]bool{
	".in_use":      true,
	".orphaned_at": true,
}

// dirExcluido es el dir que se salta entero (mismo criterio que HashFormaPlugin: SkipDir).
const dirExcluido = ".git"

// CopiarArbol copia origen → destino aplicando el set de exclusión, las reglas de symlink y los
// permisos de §13.6 (pasos 2-4). Devuelve los avisos VISIBLES de lo que NO se copió (BR-8):
// un symlink que escapa del árbol, un symlink roto, un nodo que no es archivo ni dir.
// Es la ÚNICA función de copia: los dos caminos la comparten.
func CopiarArbol(origen, destino string) ([]string, error) {
	raiz, err := filepath.EvalSymlinks(origen)
	if err != nil {
		return nil, fmt.Errorf("traer: origen %s: %w", origen, err)
	}
	info, err := os.Stat(raiz)
	if err != nil {
		return nil, fmt.Errorf("traer: origen %s: %w", raiz, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("traer: origen %s no es un directorio", raiz)
	}
	if err := os.MkdirAll(destino, 0o750); err != nil {
		return nil, fmt.Errorf("traer: crear staging %s: %w", destino, err)
	}

	var avisos []string
	werr := filepath.WalkDir(raiz, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(raiz, p)
		if rerr != nil {
			return rerr
		}
		if rel == "." {
			return nil
		}
		nombre := d.Name()
		if d.IsDir() && nombre == dirExcluido {
			return filepath.SkipDir
		}
		if excluidos[nombre] {
			return nil
		}
		target := filepath.Join(destino, rel)

		switch {
		case d.Type()&os.ModeSymlink != 0:
			aviso, cerr := copiarSymlink(raiz, p, rel, destino)
			if cerr != nil {
				return cerr
			}
			if aviso != "" {
				avisos = append(avisos, aviso)
			}
			return nil
		case d.IsDir():
			return os.MkdirAll(target, 0o750)
		case d.Type().IsRegular():
			return copiarArchivo(p, target, d)
		default:
			// FIFO, socket, device: no forman parte de un árbol de arnés y copiarlos sería
			// materializar algo que no es contenido. Se anota, no se copia.
			avisos = append(avisos, fmt.Sprintf("no se copió %s: no es un archivo regular ni un directorio (%s)", rel, d.Type()))
			return nil
		}
	})
	if werr != nil {
		return avisos, fmt.Errorf("traer: copiar %s: %w", raiz, werr)
	}
	return avisos, nil
}

// copiarSymlink aplica §13.6 paso 3: un symlink cuyo destino cae DENTRO del árbol se copia COMO
// symlink (se preserva la estructura); uno que ESCAPA no se sigue ni se copia (copiar su
// contenido metería en el canónico un archivo de otro árbol — mismo criterio que el walker del
// Portafolio, C-P-12); uno ROTO tampoco. Los dos últimos dejan aviso.
func copiarSymlink(raiz, p, rel, destino string) (string, error) {
	resuelto, everr := filepath.EvalSymlinks(p)
	if everr != nil {
		// Un symlink ROTO no es un error de la copia: es un dato del árbol origen que se ANOTA
		// (BR-8) y no se copia. Devolver el error abortaría toda la materialización por un
		// symlink colgante ajeno.
		crudo, _ := os.Readlink(p)
		return fmt.Sprintf("no se copió el symlink %s: está roto (apunta a %q)", rel, crudo), nil //nolint:nilerr // un symlink roto es un DATO del árbol origen que se ANOTA (BR-8), no un error de la copia: abortar toda la materialización por un colgante ajeno sería peor.
	}
	if !dentroDe(raiz, resuelto) {
		return fmt.Sprintf("no se copió el symlink %s: escapa del árbol del arnés (apunta a %q)", rel, resuelto), nil
	}

	// Se re-apunta al equivalente DENTRO del staging, en forma relativa: así el árbol copiado es
	// autocontenido y no depende de la ruta del origen.
	relDestino, relerr := filepath.Rel(raiz, resuelto)
	if relerr != nil {
		// Ídem: se ANOTA y no se copia, en vez de abortar la materialización entera.
		return fmt.Sprintf("no se copió el symlink %s: no se pudo relativizar su destino (%v)", rel, relerr), nil
	}
	target := filepath.Join(destino, rel)
	if merr := os.MkdirAll(filepath.Dir(target), 0o750); merr != nil {
		return "", fmt.Errorf("traer: crear dir de %s: %w", rel, merr)
	}
	nuevoTarget, rerr2 := filepath.Rel(filepath.Dir(target), filepath.Join(destino, relDestino))
	if rerr2 != nil {
		return "", fmt.Errorf("traer: relativizar symlink %s: %w", rel, rerr2)
	}
	if serr := os.Symlink(nuevoTarget, target); serr != nil {
		return "", fmt.Errorf("traer: symlink %s: %w", rel, serr)
	}
	return "", nil
}

// copiarArchivo copia un archivo regular con los permisos de §13.6 paso 4: dirs 0o750, archivos
// 0o640 MÁS el bit de ejecución del origen si lo tenía (un hook o script del arnés tiene que
// seguir siendo ejecutable). setuid/setgid/sticky se DESCARTAN: nunca los necesita un árbol de
// arnés y son superficie de riesgo.
func copiarArchivo(origen, destino string, d fs.DirEntry) error {
	info, err := d.Info()
	if err != nil {
		return err
	}
	if merr := os.MkdirAll(filepath.Dir(destino), 0o750); merr != nil {
		return merr
	}
	perm := os.FileMode(0o640) | (info.Mode().Perm() & 0o111)

	src, err := os.Open(origen) //nolint:gosec // G304: ruta bajo el árbol que el plan eligió copiar.
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	dst, err := os.OpenFile(destino, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm) //nolint:gosec // G304: ruta bajo el staging del usecase.
	if err != nil {
		return err
	}
	if _, cerr := io.Copy(dst, src); cerr != nil {
		_ = dst.Close()
		return cerr
	}
	if cerr := dst.Close(); cerr != nil {
		return cerr
	}
	// O_CREATE respeta el umask: se fuerza el modo para que el resultado sea determinista.
	return os.Chmod(destino, perm)
}

// dentroDe reporta si child es igual a, o está contenido en, parent.
func dentroDe(parent, child string) bool {
	rel, err := filepath.Rel(filepath.Clean(parent), filepath.Clean(child))
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// TamanoDeArbol suma los bytes de los archivos regulares de dir (guarda de techo del camino B,
// §13.7). No sigue symlinks.
func TamanoDeArbol(dir string) (int64, error) {
	var total int64
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		info, ierr := d.Info()
		if ierr != nil {
			return ierr
		}
		total += info.Size()
		return nil
	})
	return total, err
}
