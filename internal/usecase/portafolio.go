package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// Candidato es un resultado de Escanear: NO PERSISTIDO — el usuario elige cuáles agregar
// (spec §7.1). Trae ya resuelta la identidad/origen/deriva de un hallazgo del walker.
// Lleva tags json: es el tipo que cruza al wire (HTTP/CLI) tal cual, sin DTO aparte —
// mismo patrón que usecase.SelfUpdateReport.
type Candidato struct {
	// Clave es Identidad.Clave() ya resuelto: lo que el cliente devuelve en `elegidos[]`
	// de AgregarProyecto (evita que HTTP/CLI reimplementen el slug).
	Clave       string                `json:"clave"`
	Identidad   domain.IdentidadArnes `json:"identidad"`
	Nombre      string                `json:"nombre,omitempty"`
	Descripcion string                `json:"descripcion,omitempty"`
	Empresas    []string              `json:"empresas,omitempty"`
	Instalacion domain.Instalacion    `json:"instalacion"`
	// EsCanonico marca un hallazgo cuyo path ES un checkout conocido del store
	// (RN-IDENT-4): se ofrece como candidato a CANÓNICO, jamás como instalación.
	EsCanonico bool `json:"es_canonico,omitempty"`
}

// PortafolioService orquesta el ciclo escanear→elegir→persistir del Portafolio (Slice 0)
// a través de puertos — cero lógica de I/O acá, eso vive en los adapters inyectados.
type PortafolioService struct {
	store  ports.PortafolioStore
	scan   ports.PortafolioScanner
	cargar ports.ArnesLoader
	deriva ports.DerivaEvaluator
	// indice es el 5° puerto (Slice 1, S1-D1): publica presencias del Portafolio al
	// índice del Mapa vía ObservarEnMapa. nil en el subcomando CLI (no lo necesita) — un
	// nil explícito, no un adapter vacío, así ObservarEnMapa puede dar el error honesto
	// «requiere el daemon» en vez de un nil-pointer panic.
	indice ports.IndexPort
}

// NewPortafolioService cablea el usecase a sus 5 puertos (cmd es quien inyecta los
// adapters concretos — composition root, igual que el resto de services). indice puede
// ser nil (subcomando CLI: la observación en Mapa no aplica sin daemon).
func NewPortafolioService(store ports.PortafolioStore, scan ports.PortafolioScanner, cargar ports.ArnesLoader, deriva ports.DerivaEvaluator, indice ports.IndexPort) *PortafolioService {
	return &PortafolioService{store: store, scan: scan, cargar: cargar, deriva: deriva, indice: indice}
}

// Los errores centinela de ObservarEnMapa (S1-D1) — distinguibles por errors.Is desde el
// handler HTTP para mapear el status code correcto (404/400/500), mismo patrón que
// usecase.ErrBusy/ErrNadaQueInterrumpir en sessions.go.
var (
	// ErrObservarClaveNoEncontrada: solo se observa lo YA PERSISTIDO en el Portafolio —
	// jamás un candidato de un escaneo (S1-D1).
	ErrObservarClaveNoEncontrada = errors.New("portafolio: clave no encontrada en el Portafolio")
	// ErrObservarInstallPathAjeno: install_path no es una presencia real de la entrada
	// (ni instalaciones[].install_path ni canonico.path) — el endpoint jamás carga un
	// directorio arbitrario, la autoridad es el store (S1-D1).
	ErrObservarInstallPathAjeno = errors.New("portafolio: install_path ajeno a la entrada")
	// ErrObservarSinIndice: PortafolioService no tiene el 5° puerto cableado (subcomando
	// CLI) — la observación en Mapa requiere el daemon.
	ErrObservarSinIndice = errors.New("portafolio: observar en Mapa requiere el daemon")
)

// Escanear recorre root y devuelve TODOS los candidatos crudos (collect-all, spec §7.1):
// walk → por hallazgo, cargar (loader con fallback plugin.json) → ResolverIdentidad →
// ResolverOrigen → clasificar checkout-vs-instalación (RN-IDENT-4) → EvaluarDeriva. NO
// persiste nada — root pasa por la misma política de path protegido que el registro de
// arneses (abs·existe·no-protegido).
func (s *PortafolioService) Escanear(ctx context.Context, root string) ([]Candidato, error) {
	if err := validarRootPortafolio(root); err != nil {
		return nil, err
	}
	hallazgos, err := s.scan.Escanear(ctx, root)
	if err != nil {
		return nil, err
	}
	checkouts := s.store.Checkouts()

	out := make([]Candidato, 0, len(hallazgos))
	for _, h := range hallazgos {
		out = append(out, s.candidatoDe(root, h, checkouts))
	}
	return out, nil
}

// candidatoDe resuelve UN hallazgo crudo a un Candidato: la pieza central del pipeline
// (§2.7 del plan).
func (s *PortafolioService) candidatoDe(root string, h domain.HallazgoInstalacion, checkouts []string) Candidato {
	esCanonico := h.Dir != "" && contieneEn(checkouts, h.Dir)

	var a *domain.Arnes
	var avisoCarga string
	if h.Dir != "" {
		if g, lerr := s.cargar.Load(h.Dir); lerr != nil {
			avisoCarga = lerr.Error()
		} else {
			a = g.Arnes
		}
	}

	identidad, avisoID := domain.ResolverIdentidad(a, h.IDConocido, scopeLocalDe(root, h.Dir), scopeRemotoDe(h.Eslabones))
	origen := domain.ResolverOrigen(h.Eslabones)

	inst := domain.Instalacion{
		ProyectoPath: root,
		InstallPath:  h.Dir,
		Tipo:         h.Tipo,
		Origen:       origen,
		Aviso:        primerNoVacio(h.Aviso, avisoCarga, avisoID),
	}

	switch {
	case esCanonico:
		inst.Deriva = domain.DerivaNoEvaluable
		inst.DerivaDetalle = "es el canónico conocido, no una instalación (RN-IDENT-4)"
	case h.Dir == "":
		inst.Deriva = domain.DerivaNoEvaluable
		inst.DerivaDetalle = "sin dir físico resoluble"
	default:
		// homeParaDeriva: Home (arnes.l0.marketplace) manda; sin arnes.l0.json (el caso
		// MÁS COMÚN real — un plugin CC normal sin manifiesto propio, C-N-14) la identidad
		// queda honestamente provisional (RN-IDENT-2) pero origen.Registry YA canonicalizó
		// un repo vía cc-plugins/lock — usarlo acá es la diferencia entre evaluar deriva de
		// verdad o `deriva-no-evaluable` en la mayoría de los casos reales (BR-4: hash
		// siempre que HAYA una referencia accesible, nunca negarla por falta de manifiesto).
		homeParaDeriva := identidad.Home
		if homeParaDeriva == "" {
			// origen.Registry es el valor CRUDO del eslabón ganador (p.ej. "owner/repo"
			// corto) — RutaReferencia compara contra la forma canónica, así que se
			// canonicaliza acá antes de usarlo como fallback.
			if canon, ok := domain.CanonicalizarRepo(origen.Registry); ok {
				homeParaDeriva = canon
			}
		}
		inst.Deriva, inst.DerivaDetalle = s.deriva.Evaluar(h.Dir, homeParaDeriva, identidad.ID, origen.Version)
	}

	c := Candidato{Clave: identidad.Clave(), Identidad: identidad, Instalacion: inst, EsCanonico: esCanonico}
	if a != nil {
		c.Nombre = a.Nombre
		c.Descripcion = a.Descripcion
		c.Empresas = a.Empresas
	}
	return c
}

// scopeLocalDe es el install-path relativo al proyecto (RN-IDENT-2, forma local de scope).
func scopeLocalDe(root, dir string) string {
	if dir == "" {
		return ""
	}
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return ""
	}
	return filepath.ToSlash(rel)
}

// scopeRemotoDe extrae el remote canonicalizado del PROYECTO (eslabón git-proyecto) para
// la forma remota de scope (RN-IDENT-2) — nunca el home/registry del arnés (RN-GIT-1).
func scopeRemotoDe(eslabones []domain.EslabonOrigen) string {
	for _, e := range eslabones {
		if e.Fuente == "git-proyecto" && e.Campo == "proyecto-remote" {
			if canon, ok := domain.CanonicalizarRepo(e.Valor); ok {
				return canon
			}
		}
	}
	return ""
}

// primerNoVacio devuelve el primer string no vacío de vals.
func primerNoVacio(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// contieneEn reporta si v está en xs.
func contieneEn(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// AgregarProyecto re-escanea root y persiste SOLO los candidatos cuya Identidad.Clave()
// está en elegidos (TOCTOU-safe: se re-escanea, no se confía en un snapshot viejo).
// Idempotente por identidad+installPath — re-agregar = re-escanear sin duplicar (C-P-8).
func (s *PortafolioService) AgregarProyecto(ctx context.Context, root string, elegidos []string) ([]domain.EntradaPortafolio, error) {
	candidatos, err := s.Escanear(ctx, root)
	if err != nil {
		return nil, err
	}
	quiere := make(map[string]bool, len(elegidos))
	for _, c := range elegidos {
		quiere[c] = true
	}

	var out []domain.EntradaPortafolio
	for _, c := range candidatos {
		clave := c.Identidad.Clave()
		if !quiere[clave] {
			continue
		}
		e := domain.EntradaPortafolio{
			Identidad:   c.Identidad,
			Nombre:      c.Nombre,
			Descripcion: c.Descripcion,
			Empresas:    c.Empresas,
			Agregado:    time.Now().UTC().Format(time.RFC3339),
		}
		// S1-D3 (cierra GAP-3): el registry de origen resuelto se puebla como facet —
		// canonicalizado si RutaReferencia lo reconoce, crudo VISIBLE si no (el dato no
		// se descarta por no parsear).
		if r := c.Instalacion.Origen.Registry; r != "" {
			canon, ok := domain.CanonicalizarRepo(r)
			if !ok {
				canon = r
			}
			e.Registries = []string{canon}
		}
		if c.EsCanonico {
			e.Canonico = &domain.Canonico{Path: c.Instalacion.InstallPath, Version: c.Instalacion.Origen.Version}
		} else {
			e.Instalaciones = []domain.Instalacion{c.Instalacion}
		}
		if uerr := s.store.Upsert(e); uerr != nil {
			return nil, uerr
		}
		out = append(out, e)
	}
	return out, nil
}

// ObservarEnMapa publica una presencia YA PERSISTIDA del Portafolio al índice del Mapa —
// read-only, espejo de la ley anti-drift (S1-D1, cierra GAP-1): jamás registra cwd, jamás
// toca `arneses.json`/`ArnesRegistry` (esa es otra frontera — confinamiento de sesión).
// Orden de validación (decisiones.md S1-D1): (1) clave debe existir en el store — solo se
// observa lo persistido, nunca un candidato del escaneo; (2) installPath debe ser una
// presencia REAL de esa entrada (∈ instalaciones[].InstallPath o == canonico.Path,
// comparación canónica EvalSymlinks+Clean — así un checkout clasificado CANÓNICO por
// RN-IDENT-4/C-N-12 también es observable); (3) el 5° puerto debe estar cableado (nil en
// el subcomando CLI); (4) el loader debe resolver un *domain.Arnes real — un dir no
// cargable jamás produce un grafo inventado. OK → indice.Upsert(ctx, g) y devuelve
// g.Arnes.ID: el bare id EFECTIVO que quedó indexado (S1-D2 — el FE lo usa para apuntar
// el Mapa incluso cuando colisiona con otra clave; el re-key del índice sigue siendo
// deuda de GAP-2, fuera de este slice).
func (s *PortafolioService) ObservarEnMapa(ctx context.Context, clave, installPath string) (string, error) {
	sanas, _ := s.store.Listar()
	var entrada domain.EntradaPortafolio
	var encontrada bool
	for _, e := range sanas {
		if e.Identidad.Clave() == clave {
			entrada = e
			encontrada = true
			break
		}
	}
	if !encontrada {
		return "", fmt.Errorf("%w: %q", ErrObservarClaveNoEncontrada, clave)
	}

	if !instalPathPerteneceA(entrada, installPath) {
		return "", fmt.Errorf("%w: %q", ErrObservarInstallPathAjeno, installPath)
	}

	if s.indice == nil {
		return "", ErrObservarSinIndice
	}

	g, lerr := s.cargar.Load(installPath)
	if lerr != nil {
		return "", fmt.Errorf("portafolio: observar en Mapa: cargar %q: %w", installPath, lerr)
	}
	if g.Arnes == nil {
		return "", fmt.Errorf("portafolio: observar en Mapa: %q no resolvió un arnés cargable", installPath)
	}

	if uerr := s.indice.Upsert(ctx, g); uerr != nil {
		return "", fmt.Errorf("portafolio: observar en Mapa: indexar: %w", uerr)
	}
	return g.Arnes.ID, nil
}

// instalPathPerteneceA reporta si installPath es una presencia REAL de entrada: alguna de
// sus instalaciones, o su canónico (S1-D1) — comparación canónica (EvalSymlinks+Clean)
// para que un alias/symlink del mismo dir no cuente como ajeno.
func instalPathPerteneceA(entrada domain.EntradaPortafolio, installPath string) bool {
	target := canonicalPathPortafolio(installPath)
	if target == "" {
		return false
	}
	if entrada.Canonico != nil && canonicalPathPortafolio(entrada.Canonico.Path) == target {
		return true
	}
	for _, inst := range entrada.Instalaciones {
		if canonicalPathPortafolio(inst.InstallPath) == target {
			return true
		}
	}
	return false
}

// canonicalPathPortafolio normaliza p para comparación estable: Clean siempre, más
// EvalSymlinks best-effort — un path que no resuelve (p.ej. ya no existe en disco) sigue
// comparándose por su forma Clean, jamás aborta la comparación silenciosamente.
func canonicalPathPortafolio(p string) string {
	if p == "" {
		return ""
	}
	clean := filepath.Clean(p)
	if resolved, err := filepath.EvalSymlinks(clean); err == nil {
		return resolved
	}
	return clean
}

// Listar pasa a través del store (BR-11: entradas sanas + corruptas visibles aparte). ctx
// no se usa hoy (I/O local síncrono) — se acepta por simetría con el resto de services.
func (s *PortafolioService) Listar(_ context.Context) ([]domain.EntradaPortafolio, []domain.EntradaCorrupta, error) {
	sanas, corruptas := s.store.Listar()
	return sanas, corruptas, nil
}

// Desvincular quita clave del registro. NO desinstala, no borra ningún clon (C-UNL-3).
func (s *PortafolioService) Desvincular(_ context.Context, clave string) (bool, error) {
	return s.store.Desvincular(clave)
}

// validarRootPortafolio reusa la política mínima de arnes_registry.validate (abs · existe
// · no-protegido) — duplicada acá porque son stores independientes (§2.7 del plan): el
// scan del Portafolio NUNCA registra el proyecto en arneses.json.
func validarRootPortafolio(root string) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return errors.New("portafolio: root vacío")
	}
	if !filepath.IsAbs(root) {
		return fmt.Errorf("portafolio: root %q debe ser absoluto", root)
	}
	clean := filepath.Clean(root)
	if resolved, err := filepath.EvalSymlinks(clean); err == nil {
		clean = resolved
	}
	info, err := os.Stat(clean)
	if err != nil {
		return fmt.Errorf("portafolio: root %q: %w", clean, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("portafolio: root %q no es un directorio", clean)
	}

	home, herr := os.UserHomeDir()
	if herr != nil {
		return nil //nolint:nilerr // sin home resoluble no podemos chequear protección — no bloqueamos por eso.
	}
	if clean == string(filepath.Separator) || clean == home {
		return fmt.Errorf("portafolio: root %q es una raíz protegida", clean)
	}
	protegidos := []string{
		filepath.Join(home, ".claude"), filepath.Join(home, ".arnesia"),
		filepath.Join(home, ".ssh"), filepath.Join(home, ".gnupg"), filepath.Join(home, ".config"),
	}
	for _, p := range protegidos {
		if clean == p || dentroDePath(p, clean) || dentroDePath(clean, p) {
			return fmt.Errorf("portafolio: root %q se superpone con una ubicación protegida %q", clean, p)
		}
	}
	return nil
}

// dentroDePath reporta si child está contenido dentro de parent (parent es ancestro).
func dentroDePath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "."
}
