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
}

// NewPortafolioService cablea el usecase a sus 4 puertos (cmd es quien inyecta los
// adapters concretos — composition root, igual que el resto de services).
func NewPortafolioService(store ports.PortafolioStore, scan ports.PortafolioScanner, cargar ports.ArnesLoader, deriva ports.DerivaEvaluator) *PortafolioService {
	return &PortafolioService{store: store, scan: scan, cargar: cargar, deriva: deriva}
}

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
		inst.Deriva, inst.DerivaDetalle = s.deriva.Evaluar(h.Dir, identidad.Home, identidad.ID, origen.Version)
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
