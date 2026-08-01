package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// adoptar.go es «Adoptar como canónico» (DD-2/E-a, deudas de dogfood del paquete
// 2026-07-30-definicion-de-arnes): el camino a canónico para un plugin FORJADO localmente
// que todavía no existe en el marketplace. Rompe el círculo del alta (RN-IDENT-4 exige
// vivir en checkouts · el wizard no escanea ~/.arnesia · B2 exige canónico previo) SIN
// tocar los guards de Publicar: adoptar copia el dir al checkout del marketplace-home y
// sella la entrada con canónico; después B2 publica por el camino de siempre (el publisher
// ya agrega la fila nueva al marketplace.json).

// Errores de Adoptar. El transporte los mapea por errors.Is.
var (
	// ErrAdoptarSinMarketplace: ni el sello declara home ni el request nombra marketplace.
	ErrAdoptarSinMarketplace = errors.New("adoptar: el sello no declara marketplace-home y el request no nombra uno")
	// ErrAdoptarNoPropio: el marketplace destino no es de clase propio (BR-1 lado write).
	ErrAdoptarNoPropio = errors.New("adoptar: solo se adopta hacia un marketplace PROPIO")
	// ErrAdoptarYaCanonico: la identidad ya tiene canónico registrado — nada que adoptar.
	ErrAdoptarYaCanonico = errors.New("adoptar: la identidad ya tiene canónico registrado")
	// ErrAdoptarHomeDiscrepante: el sello declara un home distinto al marketplace pedido.
	ErrAdoptarHomeDiscrepante = errors.New("adoptar: el sello declara un marketplace-home distinto al pedido")
)

// ResultadoAdoptar es lo que la UI pinta tras adoptar. Deriva viaja SIEMPRE (BR-17).
type ResultadoAdoptar struct {
	Entrada       domain.EntradaPortafolio `json:"entrada"`
	Clave         string                   `json:"clave"`
	Destino       string                   `json:"destino"`
	Marketplace   string                   `json:"marketplace"`
	Version       string                   `json:"version,omitempty"`
	Deriva        domain.EstadoDeriva      `json:"deriva"`
	DerivaDetalle string                   `json:"deriva_detalle,omitempty"`
	Avisos        []string                 `json:"avisos,omitempty"`
}

// Adoptar toma el árbol de `dir` (el plugin forjado, material local), lo copia al checkout
// del marketplace-home bajo <raiz>/checkouts/<marketplace>/<id> y registra la entrada con
// canónico. `marketplace` es opcional: el home del sello (arnes.Marketplace) manda; si el
// sello no declara, el request lo nombra; si declaran distinto ⇒ 400 discrepante.
//
// Reusa la cola de Traer a propósito (staging bajo <raiz>/tmp + rename + BR-13 con
// symlinks resueltos + registro + deriva): mismos guardarraíles, cero duplicación de
// atomicidad. El origen es CaminoLocal con el dir del operador como OrigenLocal.
func (s *MarketplaceService) Adoptar(ctx context.Context, dir, marketplace string) (ResultadoAdoptar, error) {
	var vacio ResultadoAdoptar
	if s.cargarArnes == nil || s.matLocal == nil {
		return vacio, ErrPublicarNoDisponible
	}
	raiz, err := s.raiz()
	if err != nil {
		return vacio, err
	}

	// 1 · el dir del operador: absoluto y existente. Sin sello NO se rechaza (material
	// crudo es exactamente el caso de la forja) — pero sin id no hay identidad.
	dir = filepath.Clean(dir)
	if !filepath.IsAbs(dir) {
		return vacio, fmt.Errorf("%w: la ruta debe ser absoluta: %s", ErrTraerLocal, dir)
	}
	if fi, serr := os.Stat(dir); serr != nil || !fi.IsDir() {
		return vacio, fmt.Errorf("%w: el origen %s no existe o no es un directorio", ErrTraerLocal, dir)
	}
	g, lerr := s.cargarArnes.Load(dir)
	if lerr != nil {
		return vacio, fmt.Errorf("adoptar: cargar %s: %w", dir, lerr)
	}
	var id, version, homeSello string
	if g.Arnes != nil {
		id, version = g.Arnes.ID, g.Arnes.Version
		if g.Arnes.Marketplace != "" {
			if canon, ok := domain.CanonicalizarRepo(g.Arnes.Marketplace); ok {
				homeSello = canon
			}
		}
	}
	if id == "" {
		return vacio, fmt.Errorf("adoptar: %s no declara id (plugin.json name / arnes.l0.json id) — sin id no hay identidad que adoptar", dir)
	}

	// 2 · el marketplace destino: el home del sello manda; el request desempata o suple.
	mkt, home, err := s.marketplaceDestino(homeSello, marketplace)
	if err != nil {
		return vacio, err
	}
	identidad := domain.IdentidadArnes{Home: home, ID: id}

	// 3 · idempotencia de identidad: un canónico registrado NO se pisa (ley anti-drift:
	// una sola copia editable).
	if actual, ok := s.entradaPorClave(identidad.Clave()); ok && actual.Canonico != nil && actual.Canonico.Path != "" {
		return vacio, fmt.Errorf("%w: %s → %s", ErrAdoptarYaCanonico, identidad.Clave(), actual.Canonico.Path)
	}

	destino, dok := domain.RutaCanonico(raiz, mkt.Nombre, id)
	if !dok {
		return vacio, fmt.Errorf("%w: %s / %s", domain.ErrTraerDestinoEscapa, mkt.Nombre, id)
	}

	s.traerMu.Lock()
	defer s.traerMu.Unlock()

	// 4 · BR-14: destino poblado ⇒ 409 sin tocar nada.
	if poblado, perr := destinoPoblado(destino); perr != nil {
		return vacio, fmt.Errorf("%w: revisar el destino %s: %w", ErrTraerLocal, destino, perr)
	} else if poblado {
		return vacio, fmt.Errorf("%w: %s", ErrTraerDestinoPoblado, destino)
	}

	// 5 · staging bajo <raiz>/tmp (nunca os.TempDir: EXDEV) + copia local del árbol.
	dirTmp := domain.RaizTemporales(raiz)
	if merr := os.MkdirAll(dirTmp, 0o750); merr != nil {
		return vacio, fmt.Errorf("%w: crear %s: %w", ErrTraerLocal, dirTmp, merr)
	}
	tmp, terr := os.MkdirTemp(dirTmp, "traer-")
	if terr != nil {
		return vacio, fmt.Errorf("%w: temporal: %w", ErrTraerLocal, terr)
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	staging := filepath.Join(tmp, "staging")
	plan := domain.PlanTraer{Camino: domain.CaminoLocal, OrigenLocal: dir, Destino: destino, Identidad: identidad}
	_, avisos, merr := s.matLocal.Materializar(ctx, plan, staging)
	if merr != nil {
		return vacio, merr
	}

	// 6 · rename + BR-13 re-verificado con symlinks resueltos (misma defensa que Traer).
	if merr := os.MkdirAll(filepath.Dir(destino), 0o750); merr != nil {
		return vacio, fmt.Errorf("%w: crear el dir padre de %s: %w", ErrTraerLocal, destino, merr)
	}
	if fi, serr := os.Lstat(destino); serr == nil && fi.IsDir() {
		if rmerr := os.Remove(destino); rmerr != nil {
			return vacio, fmt.Errorf("%w: quitar el destino vacío %s: %w", ErrTraerLocal, destino, rmerr)
		}
	}
	if rerr := os.Rename(staging, destino); rerr != nil {
		return vacio, fmt.Errorf("%w: mover el staging a %s: %w", ErrTraerLocal, destino, rerr)
	}
	if !dentroDeCheckoutsResuelto(raiz, destino) {
		_ = os.RemoveAll(destino)
		return vacio, fmt.Errorf("%w: %s (verificado tras crear, con symlinks resueltos)", domain.ErrTraerDestinoEscapa, destino)
	}

	// 7 · registro: entrada existente (instalaciones del proyecto) GANA — solo se le sella
	// el canónico; sin entrada previa se crea una nueva.
	entrada, existia := s.entradaPorClave(identidad.Clave())
	if !existia {
		entrada = domain.EntradaPortafolio{
			Identidad:   identidad,
			Nombre:      primerNoVacioAdoptar(g.Arnes, id),
			Descripcion: descripcionDe(g.Arnes),
			Registries:  []string{home},
			Agregado:    s.ahora().Format(time.RFC3339),
		}
	}
	entrada.Canonico = &domain.Canonico{Path: destino, Version: version}
	if uerr := s.portafolio.Upsert(entrada); uerr != nil {
		// Mismo único-parcial-visible que Traer paso 9: el dir queda, el próximo intento
		// da 409 destino-poblado, jamás silencio.
		return vacio, fmt.Errorf("%w: el canónico quedó en %s pero no se pudo registrar: %w", ErrTraerLocal, destino, uerr)
	}

	res := ResultadoAdoptar{
		Entrada: entrada, Clave: identidad.Clave(), Destino: destino,
		Marketplace: mkt.Nombre, Version: version, Avisos: avisos,
	}
	if !domain.EsSemver(version) {
		res.Avisos = append(res.Avisos, fmt.Sprintf("la versión %q del sello no es semver — Publicar la va a exigir", version))
	}

	// 8 · BR-17: la deriva se evalúa y viaja tal cual salga.
	res.Deriva, res.DerivaDetalle = domain.DerivaNoEvaluable, "sin evaluador de deriva cableado"
	if s.deriva != nil {
		res.Deriva, res.DerivaDetalle = s.deriva.Evaluar(destino, home, id, version)
	}
	return res, nil
}

// marketplaceDestino resuelve el marketplace destino de una adopción: el home del sello
// manda; el request desempata o suple; discrepancia ⇒ error. El destino DEBE ser propio.
func (s *MarketplaceService) marketplaceDestino(homeSello, pedido string) (domain.MarketplaceConocido, string, error) {
	var vacio domain.MarketplaceConocido
	if homeSello == "" && pedido == "" {
		return vacio, "", ErrAdoptarSinMarketplace
	}
	if pedido != "" {
		mkt, err := s.buscar(pedido)
		if err != nil {
			return vacio, "", err
		}
		home, ok := domain.CanonicalizarRepo(mkt.Repo)
		if !ok {
			return vacio, "", fmt.Errorf("%w: el marketplace %q no declara un repo canonicalizable", domain.ErrPublicarSinHome, mkt.Nombre)
		}
		if homeSello != "" && homeSello != home {
			return vacio, "", fmt.Errorf("%w: sello %q ≠ %q (%s)", ErrAdoptarHomeDiscrepante, homeSello, home, mkt.Nombre)
		}
		if domain.ClaseSegura(mkt.Clase) != domain.ClasePropio {
			return vacio, "", fmt.Errorf("%w: %q es de clase %q", ErrAdoptarNoPropio, mkt.Nombre, domain.ClaseSegura(mkt.Clase))
		}
		return mkt, home, nil
	}
	mkt, hallado := s.marketplaceDeHome(homeSello)
	if !hallado {
		return vacio, "", fmt.Errorf("%w: ningún marketplace conocido apunta a %q — registralo primero", domain.ErrPublicarNoPropio, homeSello)
	}
	if domain.ClaseSegura(mkt.Clase) != domain.ClasePropio {
		return vacio, "", fmt.Errorf("%w: %q es de clase %q", ErrAdoptarNoPropio, mkt.Nombre, domain.ClaseSegura(mkt.Clase))
	}
	return mkt, homeSello, nil
}

// primerNoVacioAdoptar resuelve el nombre humano de la entrada: nombre del sello → id.
func primerNoVacioAdoptar(a *domain.Arnes, id string) string {
	if a != nil && a.Nombre != "" {
		return a.Nombre
	}
	return id
}

func descripcionDe(a *domain.Arnes) string {
	if a != nil {
		return a.Descripcion
	}
	return ""
}
