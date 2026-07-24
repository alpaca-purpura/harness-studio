package usecase

import (
	"context"
	"encoding/json"
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
	// ErrIdentificarYaSellado: el dir ya tiene arnes.l0.json — Identificar NO pisa un sello
	// existente (S1-D28, guarda 2). Editar un sello ya escrito es otra operación (S2+).
	ErrIdentificarYaSellado = errors.New("portafolio: identificar: el directorio ya tiene arnes.l0.json (no se pisa)")
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

	identidad, avisoID := domain.ResolverIdentidad(a, h.IDConocido, scopeLocalDe(root, h.Dir), scopeRemotoDe(h.Eslabones), canonicalPathPortafolio(h.Dir))
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
		if !quiere[c.Clave] {
			continue
		}
		e := entradaDeCandidato(c)
		if uerr := s.store.Upsert(e); uerr != nil {
			return nil, uerr
		}
		out = append(out, e)
	}
	return out, nil
}

// entradaDeCandidato construye la EntradaPortafolio persistible de un candidato: facet
// Registries canonicalizado o CRUDO visible (S1-D3, cierra GAP-3), y el path clasificado
// como canónico (RN-IDENT-4) o como instalación. Compartido por AgregarProyecto e
// Identificar (el re-key tras sellar) — una sola tubería de identidad, sin duplicar.
func entradaDeCandidato(c Candidato) domain.EntradaPortafolio {
	e := domain.EntradaPortafolio{
		Identidad:   c.Identidad,
		Nombre:      c.Nombre,
		Descripcion: c.Descripcion,
		Empresas:    c.Empresas,
		Agregado:    time.Now().UTC().Format(time.RFC3339),
	}
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
	return e
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
// cargable jamás produce un grafo inventado. OK → indice.Upsert(ctx, clave, g) y devuelve
// `clave`: el índice se indexa por la identidad CALIFICADA, no por `g.Arnes.ID` a secas —
// cierra GAP-2/S0-D6 (deuda BACKLOG «re-key (home,id,scope)», 2026-07-23).
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
		// Modo degradado (S1-D27, honra nomenclatura-arnes.md §2): sin sello no hay id, así
		// que se SINTETIZA un arnés mínimo cuya id es la huella de la ruta física (S1-D29) —
		// llave sintética estable que unifica la del Portafolio y la del índice del Mapa, y
		// no colisiona con otra presencia. Los nodos que el loader SÍ reconoció ya están en
		// g.Nodes: el grafo es fino, jamás inventado. El FE lo pinta con la marca roja
		// `manifiesto-ausente` (g.Degradado, que el loader ya prendió).
		nombre := entrada.Nombre
		if nombre == "" {
			nombre = filepath.Base(installPath)
		}
		g.Arnes = &domain.Arnes{ID: domain.HuellaPath(canonicalPathPortafolio(installPath)), Nombre: nombre}
		g.Degradado = true
	}

	// Se indexa bajo `clave` (calificada home,id,scope) — jamás bajo `g.Arnes.ID` a secas,
	// que dos arneses de homes distintos pueden compartir (deuda BACKLOG «re-key», cerrada
	// 2026-07-23: antes el índice colisionaba en silencio y solo un aviso de la UI lo
	// mitigaba; ahora cada clave tiene su propio slot, la colisión es estructuralmente
	// imposible). El FE usa este `id` de vuelta como el harnessID de TODO fetch del Mapa.
	if uerr := s.indice.Upsert(ctx, clave, g); uerr != nil {
		return "", fmt.Errorf("portafolio: observar en Mapa: indexar: %w", uerr)
	}
	return clave, nil
}

// Identificar escribe el sello `arnes.l0.json` (el manifiesto de la fábrica, S1-D28) en la
// instalación installPath de la entrada clave, y re-keya la entrada con su identidad ya
// sellada. V1 = scaffold mínimo (id·nombre·empresas·version), SIN clon: se sella IN-SITU
// (S1-D27 decisión 2 — git del proyecto es la red de seguridad, no una 2ª copia). Guardas:
// (1) la entrada existe y installPath le pertenece (misma autoridad que Observar — el store,
// jamás un dir arbitrario); (2) path no protegido (validarRootPortafolio, así un sello jamás
// se escribe en ~/.claude/~/.ssh/etc.); (3) NO pisa un sello existente. Sellar un dir sin
// identidad previa NO viola la ley anti-drift: CREA la identidad, no edita una copia
// downstream de un canónico. Devuelve la entrada re-keyed.
func (s *PortafolioService) Identificar(ctx context.Context, clave, installPath, id, nombre string) (domain.EntradaPortafolio, error) {
	entrada, encontrada := s.buscarPorClave(clave)
	if !encontrada {
		return domain.EntradaPortafolio{}, fmt.Errorf("%w: %q", ErrObservarClaveNoEncontrada, clave)
	}
	if !instalPathPerteneceA(entrada, installPath) {
		return domain.EntradaPortafolio{}, fmt.Errorf("%w: %q", ErrObservarInstallPathAjeno, installPath)
	}
	if err := validarRootPortafolio(installPath); err != nil {
		return domain.EntradaPortafolio{}, fmt.Errorf("portafolio: identificar: %w", err)
	}
	ruta := filepath.Join(installPath, "arnes.l0.json")
	if _, serr := os.Stat(ruta); serr == nil {
		return domain.EntradaPortafolio{}, fmt.Errorf("%w: %s", ErrIdentificarYaSellado, ruta)
	}

	sello := selloDe(installPath, id, nombre, entrada.Empresas)
	b, merr := json.MarshalIndent(sello, "", "  ")
	if merr != nil {
		return domain.EntradaPortafolio{}, fmt.Errorf("portafolio: identificar: serializar sello: %w", merr)
	}
	if werr := os.WriteFile(ruta, append(b, '\n'), 0o644); werr != nil { //nolint:gosec // G306: el sello es doc pública versionable, no secreto.
		return domain.EntradaPortafolio{}, fmt.Errorf("portafolio: identificar: escribir %s: %w", ruta, werr)
	}

	// Re-key: re-escanear el proyecto reconstruye la identidad YA sellada por la misma tubería
	// que AgregarProyecto (cero lógica de identidad duplicada) y migra la entrada; la clave
	// vieja (anónima/provisional) se desvincula solo si cambió.
	proyectoPath := proyectoPathDe(entrada, installPath)
	if proyectoPath == "" {
		proyectoPath = installPath
	}
	cands, serr := s.Escanear(ctx, proyectoPath)
	if serr != nil {
		return domain.EntradaPortafolio{}, fmt.Errorf("portafolio: identificar: re-escanear: %w", serr)
	}
	target := canonicalPathPortafolio(installPath)
	for _, c := range cands {
		if canonicalPathPortafolio(c.Instalacion.InstallPath) != target {
			continue
		}
		nueva := entradaDeCandidato(c)
		if uerr := s.store.Upsert(nueva); uerr != nil {
			return domain.EntradaPortafolio{}, uerr
		}
		if nueva.Identidad.Clave() != clave {
			if _, derr := s.store.Desvincular(clave); derr != nil {
				return domain.EntradaPortafolio{}, derr
			}
		}
		return nueva, nil
	}
	return domain.EntradaPortafolio{}, fmt.Errorf("portafolio: identificar: el sello se escribió en %s pero el re-escaneo no reencontró la instalación", ruta)
}

// buscarPorClave devuelve la entrada sana cuya clave coincide (solo lo persistido).
func (s *PortafolioService) buscarPorClave(clave string) (domain.EntradaPortafolio, bool) {
	sanas, _ := s.store.Listar()
	for _, e := range sanas {
		if e.Identidad.Clave() == clave {
			return e, true
		}
	}
	return domain.EntradaPortafolio{}, false
}

// selloDe arma el manifiesto mínimo de Identificar V1 (S1-D28): id = el dado o el basename
// del install-path sluggeado; nombre = el dado o el id; empresas heredadas de la entrada;
// version semilla "0.1.0". Sin marketplace (identidad provisional honesta hasta que se
// declare un home) ni fases/spine (opcionales, §2 de nomenclatura-arnes). reporta_a = null.
func selloDe(installPath, id, nombre string, empresas []string) domain.Arnes {
	id = domain.Slug(primerNoVacio(id, filepath.Base(installPath)))
	if nombre == "" {
		nombre = id
	}
	return domain.Arnes{ID: id, Nombre: nombre, Empresas: empresas, Version: "0.1.0", ReportaA: nil}
}

// proyectoPathDe devuelve el ProyectoPath de la instalación de entrada que coincide con
// installPath (para re-escanear el proyecto correcto tras sellar); "" si installPath es el
// canónico u otra forma sin ProyectoPath — el caller cae a installPath como root.
func proyectoPathDe(entrada domain.EntradaPortafolio, installPath string) string {
	target := canonicalPathPortafolio(installPath)
	for _, inst := range entrada.Instalaciones {
		if canonicalPathPortafolio(inst.InstallPath) == target {
			return inst.ProyectoPath
		}
	}
	return ""
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

// ReevaluarDeriva re-corre el evaluador de deriva de la instalación cuyo InstallPath
// coincide con installPath (RF-193, mejorar-arnes-conversando T1): tras una edición por
// chat, el Portafolio nunca finge `al-hilo`. Devuelve si algo cambió (y persiste solo en
// ese caso). Path fuera del Portafolio o canónico → (false, nil): nada que re-evaluar,
// no es un error. Misma resolución de home/version que candidatoDe (BR-4).
func (s *PortafolioService) ReevaluarDeriva(_ context.Context, installPath string) (bool, error) {
	canon := canonicalPathPortafolio(installPath)
	if canon == "" || s.deriva == nil {
		return false, nil
	}
	entradas, _ := s.store.Listar()
	for _, e := range entradas {
		for i := range e.Instalaciones {
			inst := &e.Instalaciones[i]
			if canonicalPathPortafolio(inst.InstallPath) != canon {
				continue
			}
			homeParaDeriva := e.Identidad.Home
			if homeParaDeriva == "" {
				if c, ok := domain.CanonicalizarRepo(inst.Origen.Registry); ok {
					homeParaDeriva = c
				}
			}
			estado, detalle := s.deriva.Evaluar(inst.InstallPath, homeParaDeriva, e.Identidad.ID, inst.Origen.Version)
			if estado == inst.Deriva && detalle == inst.DerivaDetalle {
				return false, nil
			}
			inst.Deriva, inst.DerivaDetalle = estado, detalle
			if err := s.store.Upsert(e); err != nil {
				return false, fmt.Errorf("portafolio: re-evaluar deriva de %s: %w", canon, err)
			}
			return true, nil
		}
	}
	return false, nil
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
