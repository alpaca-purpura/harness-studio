package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// traer.go es la ATOMICIDAD de `↧ Traer canónico` (BR-15) en UN solo lugar: el usecase es dueño
// del ciclo de vida del temporal, así los dos materializadores quedan tontos y la limpieza tiene
// una sola implementación (design.md §13.5).

// Errores de EJECUCIÓN de Traer. El handler los mapea por errors.Is (§13.8).
var (
	// ErrTraerDestinoPoblado: BR-14 — el destino ya tiene contenido. Se aborta SIN TOCAR NADA:
	// pisar un canónico destruye trabajo, y es la única copia editable (ley anti-drift).
	ErrTraerDestinoPoblado = errors.New("traer: el destino ya tiene contenido: no se pisa el canónico")
	// ErrTraerSHANoCoincide: BR-16 — lo que vino no es lo que el catálogo declara.
	ErrTraerSHANoCoincide = errors.New("traer: el contenido traído no coincide con el sha declarado")
	// ErrTraerSinMaterializador: el camino que el plan pide no está cableado (sin `git`/`gh`) →
	// 503 honesto en vez de un nil-pointer panic.
	ErrTraerSinMaterializador = errors.New("traer: el camino que el catálogo exige no está disponible en esta instalación")
	// ErrTraerLocal: fallo local (disco, permisos, rename, Upsert) → 500 con el motivo real.
	// Se RE-EXPORTA del dominio: lo producen los dos lados (el adapter y este usecase).
	ErrTraerLocal = domain.ErrTraerLocal
	// ErrTraerRemotoNoTiene / ErrTraerSinAuth se re-exportan igual, para que el transporte los
	// clasifique sin importar el adapter (C23 de design.md).
	ErrTraerRemotoNoTiene = domain.ErrTraerRemotoNoTiene
	ErrTraerSinAuth       = domain.ErrTraerSinAuth
)

// SetTraer cablea los 3 puertos de `Traer` (AG-D17). `externo` puede ser nil: el camino B degrada
// a 503 honesto. `deriva` es la MISMA instancia que usa el Portafolio — crear una segunda
// duplicaría la resolución de RutaReferencia y podrían divergir (§12.2 riesgo 15).
// `raizArnesia` es "" en producción (⇒ ~/.arnesia) e inyectable en tests.
// Al cablear se BARREN los temporales huérfanos de un SIGKILL previo (E-86b).
func (s *MarketplaceService) SetTraer(local, externo ports.Materializador, deriva ports.DerivaEvaluator, raizArnesia string) {
	s.matLocal, s.matExterno, s.deriva, s.raizArnesia = local, externo, deriva, raizArnesia
	s.barrerTemporalesHuerfanos()
}

// raiz resuelve la raíz de ArnesIA ("" ⇒ ~/.arnesia).
func (s *MarketplaceService) raiz() (string, error) {
	if s.raizArnesia != "" {
		return s.raizArnesia, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("%w: resolver home: %w", ErrTraerLocal, err)
	}
	return filepath.Join(home, ".arnesia"), nil
}

// barrerTemporalesHuerfanos borra los `traer-*` de más de una hora que dejó un SIGKILL previo
// (E-86b). Se hace AL ARRANCAR y no en caliente a propósito: al boot no puede haber un Traer en
// vuelo de un proceso que ya murió, mientras que barrer en caliente correría el riesgo de matar el
// temporal de otra instancia del daemon.
func (s *MarketplaceService) barrerTemporalesHuerfanos() {
	raiz, err := s.raiz()
	if err != nil {
		return
	}
	dir := domain.RaizTemporales(raiz)
	entradas, rerr := os.ReadDir(dir)
	if rerr != nil {
		return
	}
	limite := s.ahora().Add(-time.Hour)
	for _, e := range entradas {
		if !e.IsDir() || len(e.Name()) < 6 || e.Name()[:6] != "traer-" {
			continue
		}
		info, ierr := e.Info()
		if ierr != nil || info.ModTime().After(limite) {
			continue // reciente: podría ser de otra instancia VIVA. No se toca.
		}
		_ = os.RemoveAll(filepath.Join(dir, e.Name()))
	}
}

// ResultadoTraer es lo que la UI pinta tras traer. `Deriva` viaja SIEMPRE, incluso
// `deriva-no-evaluable` con motivo: BR-17 dice que se muestra lo que salga, no lo que conviene.
type ResultadoTraer struct {
	Entrada       domain.EntradaPortafolio `json:"entrada"`
	Clave         string                   `json:"clave"`
	Destino       string                   `json:"destino"`
	Camino        domain.CaminoTraer       `json:"camino"`
	SHAEfectivo   string                   `json:"sha_efectivo,omitempty"`
	Deriva        domain.EstadoDeriva      `json:"deriva"`
	DerivaDetalle string                   `json:"deriva_detalle,omitempty"`
	Avisos        []string                 `json:"avisos,omitempty"`
}

// Traer materializa la fila `entrada` del catálogo de `nombre` como canónico editable.
// Secuencia EXACTA de design.md §13.5 — el orden ES el contrato de atomicidad (BR-15):
//
//	1-2. planificar (CERO I/O)                        → 400 si clase referencia / source / destino
//	3.   destino poblado                              → 409, NADA se tocó
//	4.   tmp bajo <raiz>/tmp + defer RemoveAll        ← la ÚNICA limpieza, cubre las 6 ramas
//	5.   materializar en staging                      → 502/503/500, el defer limpia
//	6.   BR-16: sha efectivo == sha declarado         → 502, el defer limpia
//	7.   rename(staging, destino)                     ← el ÚNICO efecto observable
//	8.   re-verificar BR-13 con symlinks resueltos    → 400 + borrar el destino
//	9.   registrar el canónico                        → 500 (el dir QUEDA: es el único parcial)
//	10.  BR-17: evaluar deriva y devolverla tal cual salga
func (s *MarketplaceService) Traer(ctx context.Context, nombre, entrada string) (ResultadoTraer, error) {
	raiz, err := s.raiz()
	if err != nil {
		return ResultadoTraer{}, err
	}

	// 1 · la fila mergeada del marketplace + la fila del catálogo.
	mkt, err := s.buscar(nombre)
	if err != nil {
		return ResultadoTraer{}, err
	}
	fila, err := s.filaDeCatalogo(ctx, mkt, entrada)
	if err != nil {
		return ResultadoTraer{}, err
	}

	// 2 · el PLAN, puro y sin I/O. Acá se rechaza la clase `referencia` (BR-1 en el dominio: un
	// POST puede llegar sin pasar por la UI) y el source no materializable.
	installExiste := mkt.InstallLocation != "" && esDir(mkt.InstallLocation)
	plan, err := domain.PlanificarTraer(mkt, fila, installExiste, raiz)
	if err != nil {
		return ResultadoTraer{}, err
	}

	mat := s.materializadorDe(plan.Camino)
	if mat == nil {
		return ResultadoTraer{}, fmt.Errorf("%w (camino %q)", ErrTraerSinMaterializador, plan.Camino)
	}

	// El candado serializa destino-poblado + rename: es lo que hace que dos Traer concurrentes del
	// MISMO destino den exactamente 1 éxito y 1 conflicto (E-102), sin `singleflight`.
	s.traerMu.Lock()
	defer s.traerMu.Unlock()

	// 3 · BR-14 · destino poblado ⇒ se aborta SIN TOCAR NADA. Se chequea con ReadDir, no con
	// Stat: un dir VACÍO no es «poblado» y se puede usar (E-79b); un ARCHIVO sí lo es (E-79c).
	if poblado, perr := destinoPoblado(plan.Destino); perr != nil {
		return ResultadoTraer{}, fmt.Errorf("%w: revisar el destino %s: %w", ErrTraerLocal, plan.Destino, perr)
	} else if poblado {
		return ResultadoTraer{}, fmt.Errorf("%w: %s", ErrTraerDestinoPoblado, plan.Destino)
	}

	// 4 · el temporal vive bajo <raiz>/tmp, NO en os.TempDir(): `os.Rename` entre filesystems
	// falla con EXDEV y /tmp es tmpfs en muchas instalaciones (§12.2 riesgo 11, E-101).
	dirTmp := domain.RaizTemporales(raiz)
	if merr := os.MkdirAll(dirTmp, 0o750); merr != nil {
		return ResultadoTraer{}, fmt.Errorf("%w: crear %s: %w", ErrTraerLocal, dirTmp, merr)
	}
	tmp, terr := os.MkdirTemp(dirTmp, "traer-")
	if terr != nil {
		return ResultadoTraer{}, fmt.Errorf("%w: temporal: %w", ErrTraerLocal, terr)
	}
	defer func() { _ = os.RemoveAll(tmp) }() // la ÚNICA limpieza: cubre las 6 ramas de falla.
	staging := filepath.Join(tmp, "staging")

	// 5 · materializar. Cualquier falla acá deja CERO residuo (el defer) y CERO registro.
	shaEfectivo, avisos, merr := mat.Materializar(ctx, plan, staging)
	avisos = append(append([]string{}, plan.Avisos...), avisos...)
	if merr != nil {
		return ResultadoTraer{}, merr
	}

	// 6 · BR-16 · el pin es el `sha` del catálogo y es la AUTORIDAD.
	if plan.SHAEsperado != "" && shaEfectivo != "" && shaEfectivo != plan.SHAEsperado {
		return ResultadoTraer{}, fmt.Errorf("%w: el catálogo declara %s y se materializó %s",
			ErrTraerSHANoCoincide, plan.SHAEsperado, shaEfectivo)
	}

	// 7 · el ÚNICO efecto observable.
	if merr := os.MkdirAll(filepath.Dir(plan.Destino), 0o750); merr != nil {
		return ResultadoTraer{}, fmt.Errorf("%w: crear el dir padre de %s: %w", ErrTraerLocal, plan.Destino, merr)
	}
	// ⚠ `os.Rename` de Go devuelve EEXIST si el destino existe y es un dir — INCLUSO VACÍO, a
	// diferencia de `rename(2)`, que permite reemplazar un dir vacío. El paso 3 ya verificó que
	// está vacío, así que quitarlo acá es seguro y es lo que hace pasar E-79b.
	if fi, serr := os.Lstat(plan.Destino); serr == nil && fi.IsDir() {
		if rmerr := os.Remove(plan.Destino); rmerr != nil {
			return ResultadoTraer{}, fmt.Errorf("%w: quitar el destino vacío %s: %w", ErrTraerLocal, plan.Destino, rmerr)
		}
	}
	if rerr := os.Rename(staging, plan.Destino); rerr != nil {
		return ResultadoTraer{}, fmt.Errorf("%w: mover el staging a %s: %w", ErrTraerLocal, plan.Destino, rerr)
	}

	// 8 · BR-13 re-verificado CON symlinks resueltos: un atacante con acceso al FS pudo haber
	// puesto un symlink en `<raiz>/checkouts/<algo>` apuntando afuera. Defensa en profundidad
	// real, no ceremonia (§13.9 capa 2).
	if !dentroDeCheckoutsResuelto(raiz, plan.Destino) {
		_ = os.RemoveAll(plan.Destino)
		return ResultadoTraer{}, fmt.Errorf("%w: %s (verificado tras crear, con symlinks resueltos)",
			domain.ErrTraerDestinoEscapa, plan.Destino)
	}

	// 9 · REGISTRO del canónico. Identidad con Home = repo CANONICALIZADO (C17).
	nueva := domain.EntradaPortafolio{
		Identidad:   plan.Identidad,
		Nombre:      fila.Nombre,
		Descripcion: fila.Descripcion,
		Registries:  []string{plan.Identidad.Home},
		Canonico:    &domain.Canonico{Path: plan.Destino, Version: fila.Version},
		Agregado:    s.ahora().Format(time.RFC3339),
	}
	if uerr := s.portafolio.Upsert(nueva); uerr != nil {
		// ⚠ El ÚNICO estado parcial posible, y es VISIBLE: el dir YA existe y NO se borra —
		// destruir contenido que costó red por un fallo de escritura del registro es peor. El
		// próximo Traer da 409 destino-poblado, no un silencio (E-100).
		return ResultadoTraer{}, fmt.Errorf("%w: el canónico quedó en %s pero no se pudo registrar: %w",
			ErrTraerLocal, plan.Destino, uerr)
	}

	res := ResultadoTraer{
		Destino:     plan.Destino,
		Camino:      plan.Camino,
		SHAEfectivo: shaEfectivo,
		Avisos:      avisos,
	}

	// 10 · BR-17 · la deriva se evalúa DE INMEDIATO y se devuelve TAL CUAL salga (incluso
	// `no-evaluable` con motivo): un veredicto incómodo no es un fallo de la operación.
	res.Deriva, res.DerivaDetalle = domain.DerivaNoEvaluable, "sin evaluador de deriva cableado"
	if s.deriva != nil {
		res.Deriva, res.DerivaDetalle = s.deriva.Evaluar(plan.Destino, plan.Identidad.Home, plan.Identidad.ID, fila.Version)
	}
	// El veredicto se PERSISTE en la entrada (el Portafolio nunca finge al-hilo).
	nueva.Instalaciones = nil
	if actual, ok := s.entradaPorClave(plan.Identidad.Clave()); ok {
		nueva = actual
	}
	res.Entrada = nueva
	res.Clave = nueva.Identidad.Clave()
	return res, nil
}

// materializadorDe elige el adapter del camino que el plan pide.
func (s *MarketplaceService) materializadorDe(camino domain.CaminoTraer) ports.Materializador {
	switch camino {
	case domain.CaminoLocal:
		return s.matLocal
	case domain.CaminoExterno:
		return s.matExterno
	default:
		return nil
	}
}

// filaDeCatalogo busca la fila `entrada` en el catálogo de mkt. Si no hay caché, LEE primero
// (E-104): una lectura que falla devuelve el error de lectura, no un «entrada no encontrada»
// engañoso.
func (s *MarketplaceService) filaDeCatalogo(ctx context.Context, mkt domain.MarketplaceConocido, entrada string) (domain.EntradaCatalogo, error) {
	cat, motivo, ok := s.cache.Leer(mkt.Nombre)
	if !ok {
		fresco, lerr := s.leerFresco(ctx, mkt)
		if lerr != nil {
			return domain.EntradaCatalogo{}, lerr
		}
		cat = fresco
		_ = motivo
	}
	for _, e := range cat.Entradas {
		if e.Nombre == entrada {
			return e, nil
		}
	}
	return domain.EntradaCatalogo{}, fmt.Errorf("%w: %q en %q", ErrEntradaNoEnCatalogo, entrada, mkt.Nombre)
}

// entradaPorClave devuelve la entrada persistida de clave (para reflejar el merge del store).
func (s *MarketplaceService) entradaPorClave(clave string) (domain.EntradaPortafolio, bool) {
	entradas, _ := s.portafolio.Listar()
	for _, e := range entradas {
		if e.Identidad.Clave() == clave {
			return e, true
		}
	}
	return domain.EntradaPortafolio{}, false
}

// destinoPoblado reporta si el destino tiene contenido. ReadDir (no Stat): un dir VACÍO no es
// poblado y se puede usar; un ARCHIVO en la ruta del destino SÍ cuenta como poblado.
func destinoPoblado(destino string) (bool, error) {
	fi, err := os.Lstat(destino)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !fi.IsDir() {
		return true, nil // un archivo (o symlink) ocupando el destino: no se pisa.
	}
	entradas, rerr := os.ReadDir(destino)
	if rerr != nil {
		return false, rerr
	}
	return len(entradas) > 0, nil
}

// esDir reporta si p existe y es un directorio (resolviendo symlinks).
func esDir(p string) bool {
	resuelto, err := filepath.EvalSymlinks(p)
	if err != nil {
		return false
	}
	fi, serr := os.Stat(resuelto)
	return serr == nil && fi.IsDir()
}

// dentroDeCheckoutsResuelto es la variante de domain.DentroDeCheckouts que SÍ resuelve symlinks —
// el chequeo POST-creación de §13.9. Vive en el usecase porque hace I/O (EvalSymlinks), que el
// dominio no puede hacer.
func dentroDeCheckoutsResuelto(raizArnesia, p string) bool {
	raiz := domain.RaizCheckouts(raizArnesia)
	raizResuelta, err := filepath.EvalSymlinks(raiz)
	if err != nil {
		raizResuelta = filepath.Clean(raiz)
	}
	destinoResuelto, derr := filepath.EvalSymlinks(p)
	if derr != nil {
		destinoResuelto = filepath.Clean(p)
	}
	rel, rerr := filepath.Rel(raizResuelta, destinoResuelto)
	if rerr != nil {
		return false
	}
	return rel != ".." && !filepathEmpiezaConDosPuntos(rel)
}

// filepathEmpiezaConDosPuntos reporta si rel arranca con "../".
func filepathEmpiezaConDosPuntos(rel string) bool {
	return len(rel) >= 3 && rel[:3] == ".."+string(filepath.Separator)
}
