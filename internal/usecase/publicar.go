package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// publicar.go es la POLÍTICA de `▲ Publicar` (B2 del paquete
// 2026-07-30-volverlo-de-arnesia-y-publicar, B-D5): las guardas (canónico → home → clase propio
// → semver → gate de conformance) corren TODAS antes de que el publisher toque nada. El
// mecanismo git vive en internal/adapters/publish; acá no hay un solo `os.`/`exec.`.

// ErrPublicarNoDisponible marca que SetPublicar no se cableó (subcomando sin daemon o build
// parcial) → 503 honesto en vez de un nil-pointer panic — mismo criterio que
// ErrTraerSinMaterializador.
var ErrPublicarNoDisponible = errors.New("publicar: no está disponible en esta instalación (sin publisher cableado)")

// ErrorConformancePublicar envuelve ErrPublicarConformanceRojo CON el reporte adjunto: el
// transporte lo desempaqueta por errors.As para responder 409 `{error, conformance}` — el
// operador ve QUÉ checks bloquearon, no solo que «está rojo».
type ErrorConformancePublicar struct {
	Reporte domain.ConformanceReport
}

func (e *ErrorConformancePublicar) Error() string {
	t := e.Reporte.Tally()
	return fmt.Sprintf("%v (fail %d · error %d)", domain.ErrPublicarConformanceRojo,
		t[domain.VeredictoFail], t[domain.VeredictoError])
}

// Unwrap hace que errors.Is(err, domain.ErrPublicarConformanceRojo) funcione a través del tipo.
func (e *ErrorConformancePublicar) Unwrap() error { return domain.ErrPublicarConformanceRojo }

// SetPublicar cablea los 3 puertos de `▲ Publicar` (patrón SetTraer): el publisher git, el
// loader del canónico (la MISMA función que usa el Portafolio — cero lógica de carga duplicada)
// y el motor de conformance del daemon (el MISMO ConformancePort que sirve GET …/conformance).
func (s *MarketplaceService) SetPublicar(pub ports.PublishPort, cargar ports.ArnesLoader, conf ports.ConformancePort) {
	s.pub, s.cargarArnes, s.conf = pub, cargar, conf
}

// ResultadoPublicar es lo que la UI pinta tras publicar. Avisos viaja SIEMPRE que exista
// (tag caído, registro local no actualizado): éxito parcial VISIBLE, jamás silencioso.
type ResultadoPublicar struct {
	Clave       string   `json:"clave"`
	ID          string   `json:"id"`
	Version     string   `json:"version"`
	Marketplace string   `json:"marketplace"`
	Commit      string   `json:"commit"`
	Tag         string   `json:"tag,omitempty"`
	Avisos      []string `json:"avisos,omitempty"`
}

// Publicar publica el canónico de la entrada `clave` en su marketplace-home. Secuencia (RF-B2.3;
// el orden ES el contrato — ninguna guarda corre después de tocar el remoto):
//
//  1. entrada por clave                       → 404 si no existe
//  2. canónico presente                       → 400 ErrPublicarSinCanonico
//  3. home declarado                          → 400 ErrPublicarSinHome
//  4. marketplace del home es ClasePropio     → 400 ErrPublicarNoPropio (BR-1 lado write)
//  5. versión del canónico semver             → 400 ErrPublicarVersionInvalida
//  6. gate conformance (RunGraph) verde       → 409 ErrPublicarConformanceRojo + reporte
//  7. publisher (idempotencia/push/tag en el adapter)
//  8. post-éxito: invalidar caché del catálogo + Canonico.Version al día (fallos = avisos)
func (s *MarketplaceService) Publicar(ctx context.Context, clave string) (ResultadoPublicar, error) {
	var vacio ResultadoPublicar
	if s.pub == nil || s.cargarArnes == nil || s.conf == nil {
		return vacio, ErrPublicarNoDisponible
	}

	s.publicarMu.Lock()
	defer s.publicarMu.Unlock()

	// 1 · la entrada persistida (la autoridad es el store, jamás un dir arbitrario).
	entrada, ok := s.entradaPorClave(clave)
	if !ok {
		return vacio, fmt.Errorf("%w: %q", ErrObservarClaveNoEncontrada, clave)
	}
	// 2 · solo se publica la única copia editable (ley anti-drift).
	if entrada.Canonico == nil || entrada.Canonico.Path == "" {
		return vacio, fmt.Errorf("%w (%s)", domain.ErrPublicarSinCanonico, clave)
	}
	// 3 · sin home no hay estante destino.
	home := entrada.Identidad.Home
	if home == "" {
		return vacio, fmt.Errorf("%w (%s)", domain.ErrPublicarSinHome, clave)
	}
	// 4 · el marketplace del home tiene que ser PROPIO — con el merge detector+store que ya usa
	// Listar (la clase declarada manda; lo desconocido degrada a referencia, el lado seguro).
	mkt, hallado := s.marketplaceDeHome(home)
	if !hallado {
		return vacio, fmt.Errorf("%w: ningún marketplace conocido apunta a %q — registralo primero", domain.ErrPublicarNoPropio, home)
	}
	if domain.ClaseSegura(mkt.Clase) != domain.ClasePropio {
		return vacio, fmt.Errorf("%w: %q es de clase %q", domain.ErrPublicarNoPropio, mkt.Nombre, domain.ClaseSegura(mkt.Clase))
	}

	// 5 · la versión sale del canónico REAL (loader → plugin.json.version, la SoT de
	// conventions/versionado-arnes.md) — jamás de la fila del catálogo ni tecleada.
	g, lerr := s.cargarArnes.Load(entrada.Canonico.Path)
	if lerr != nil {
		return vacio, fmt.Errorf("publicar: cargar el canónico %s: %w", entrada.Canonico.Path, lerr)
	}
	var version string
	if g.Arnes != nil {
		version = g.Arnes.Version
	}
	if !domain.EsSemver(version) {
		return vacio, fmt.Errorf("%w: %q", domain.ErrPublicarVersionInvalida, version)
	}

	// 6 · el gate: el MISMO RunGraph del endpoint de conformance, sobre el grafo recién
	// cargado. Rojo ⇒ el reporte viaja adjunto — se muestra QUÉ bloqueó (innegociable).
	raw, merr := json.Marshal(g)
	if merr != nil {
		return vacio, fmt.Errorf("publicar: serializar el grafo del canónico: %w", merr)
	}
	rep, rerr := s.conf.RunGraph(ctx, raw, entrada.Canonico.Path)
	if rerr != nil {
		return vacio, fmt.Errorf("publicar: correr el gate de conformance: %w", rerr)
	}
	if !rep.OK() {
		return vacio, &ErrorConformancePublicar{Reporte: rep}
	}

	// 7 · el mecanismo. Idempotencia, push-sin-force y tag-tras-push viven en el adapter.
	res, perr := s.pub.Publicar(ctx, domain.SolicitudPublicacion{
		RepoHome:  home,
		ID:        entrada.Identidad.ID,
		Version:   version,
		OrigenDir: entrada.Canonico.Path,
	})
	if perr != nil {
		return vacio, perr
	}

	out := ResultadoPublicar{
		Clave:       clave,
		ID:          entrada.Identidad.ID,
		Version:     version,
		Marketplace: mkt.Nombre,
		Commit:      res.Commit,
		Tag:         res.Tag,
		Avisos:      res.Avisos,
	}

	// 8 · post-éxito. El remoto YA cambió: un fallo local acá no puede «des-publicar», así que
	// degrada a aviso visible, jamás a un error que insinúe que no se publicó.
	if oerr := s.cache.Olvidar(mkt.Nombre); oerr != nil {
		out.Avisos = append(out.Avisos, fmt.Sprintf("publicado, pero no se pudo invalidar el caché del catálogo de %q: %v", mkt.Nombre, oerr))
	}
	entrada.Canonico.Version = version
	if uerr := s.portafolio.Upsert(entrada); uerr != nil {
		out.Avisos = append(out.Avisos, fmt.Sprintf("publicado, pero no se pudo actualizar la versión del canónico en el registro: %v", uerr))
	}
	return out, nil
}

// marketplaceDeHome busca en el merge detector+store la fila cuyo Repo canonicaliza al mismo
// home (RN-IDENT-1: se comparan los DOS lados canonicalizados, el bug clásico es comparar un
// owner/repo corto contra un host/owner/repo).
func (s *MarketplaceService) marketplaceDeHome(home string) (domain.MarketplaceConocido, bool) {
	detectados, _ := s.detector.Detectados()
	declarados, _ := s.store.Listar()
	for _, m := range domain.MergeMarketplaces(detectados, declarados) {
		if m.Repo != "" && mismoRepoCanon(m.Repo, home) {
			return m, true
		}
	}
	return domain.MarketplaceConocido{}, false
}
