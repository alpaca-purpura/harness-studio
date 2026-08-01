package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// marketplace.go es la POLÍTICA del estante (design.md §3.5/§5.1): qué se lee primero, cuándo se
// refresca, cómo degrada. El MECANISMO vive en los adapters; acá no hay un solo `os.` ni un solo
// `http.` — todo pasa por puertos.

// MarketplaceService orquesta el plano Marketplaces y el catálogo a través de puertos —
// cero I/O acá. `remoto` puede ser nil (sin `gh` ni PAT): el servicio degrada honesto en vez
// de fallar, y la fila dice `sin-acceso` con motivo.
type MarketplaceService struct {
	store      ports.MarketplaceStore
	detector   ports.MarketplaceDetector
	local      ports.CatalogoReader
	remoto     ports.CatalogoReader
	sync       ports.CatalogoSync // nil ⇒ Refrescar no sincroniza el checkout (solo relee).
	validador  ports.CatalogoValidador
	cache      ports.CatalogoCache
	portafolio ports.PortafolioStore
	ahora      func() time.Time // inyectable: los tests no dependen del reloj.

	// ── `↧ Traer canónico` (AG-D17, design.md §13) ──
	matLocal    ports.Materializador  // camino A: copia del checkout de CC. Sin red.
	matExterno  ports.Materializador  // camino B: fetch shallow pineado por sha. nil ⇒ 503 honesto.
	deriva      ports.DerivaEvaluator // firma FIJA: se reusa la MISMA instancia que el Portafolio.
	raizArnesia string                // "" ⇒ ~/.arnesia — inyectable para tests.
	traerMu     sync.Mutex            // serializa el chequeo destino-poblado + el rename (E-102).

	// ── `▲ Publicar` (B2, paquete 2026-07-30-volverlo-de-arnesia-y-publicar) ──
	pub         ports.PublishPort     // el publisher git. nil ⇒ 503 honesto (SetPublicar).
	cargarArnes ports.ArnesLoader     // carga el canónico para leer su versión (plugin.json SoT).
	conf        ports.ConformancePort // el gate: RunGraph verde ANTES de tocar el remoto.
	publicarMu  sync.Mutex            // serializa dos Publicar del mismo daemon (el remoto arbitra el resto).
}

// NewMarketplaceService cablea el usecase a sus 7 puertos de lectura. Los 3 puertos de `Traer`
// se cablean aparte con SetTraer (el composition root los tiene después), así el servicio
// arranca útil aunque `Traer` no esté disponible — y `Traer` da un error honesto en vez de un
// nil-pointer panic.
func NewMarketplaceService(
	store ports.MarketplaceStore,
	detector ports.MarketplaceDetector,
	local, remoto ports.CatalogoReader,
	validador ports.CatalogoValidador,
	cache ports.CatalogoCache,
	portafolio ports.PortafolioStore,
) *MarketplaceService {
	return &MarketplaceService{
		store: store, detector: detector, local: local, remoto: remoto,
		validador: validador, cache: cache, portafolio: portafolio,
		ahora: func() time.Time { return time.Now().UTC() },
	}
}

// SetAhora inyecta el reloj (tests).
func (s *MarketplaceService) SetAhora(f func() time.Time) { s.ahora = f }

// SetSync cablea el sincronizador del checkout (DD-2/E-bis). Aparte del constructor por la
// misma razón que SetTraer: el servicio arranca útil sin él y Refrescar degrada a solo-releer.
func (s *MarketplaceService) SetSync(sync ports.CatalogoSync) { s.sync = sync }

// Errores centinela: el handler mapea el status por errors.Is, jamás parseando strings (mismo
// patrón que ErrObservarClaveNoEncontrada). ErrNoEsMarketplace/ErrSinViaDeLectura se RE-EXPORTAN
// del dominio: los produce el adapter y `usecase` no puede importarlo (C23 de design.md).
var (
	// ErrMarketplaceNoConocido: el nombre no está ni detectado ni declarado → 404.
	ErrMarketplaceNoConocido = errors.New("marketplace: nombre no conocido")
	// ErrMarketplaceYaRegistrado: BR-7 — no se duplica ni se pisa → 409.
	ErrMarketplaceYaRegistrado = errors.New("marketplace: ya registrado (no se duplica ni se pisa)")
	// ErrURLNoCanonicalizable: la url no parsea a host/owner/repo → 400.
	ErrURLNoCanonicalizable = errors.New("marketplace: url no resuelve a host/owner/repo")
	// ErrNoEsMarketplace: el repo existe pero no tiene `.claude-plugin/marketplace.json`
	// legible → 400 (E-17: mensaje explícito, no un 500 genérico).
	ErrNoEsMarketplace = domain.ErrNoEsMarketplace
	// ErrSinViaDeLectura: no hay checkout local ni vía remota (gh ausente/no autenticado y sin
	// PAT) → 503. Distinto de ErrNoEsMarketplace: no es que la url esté mal, es que NO PUEDO MIRAR.
	ErrSinViaDeLectura = domain.ErrSinViaDeLectura
	// ErrClaseInvalida: clase distinta de propio|referencia → 400.
	ErrClaseInvalida = errors.New("marketplace: clase inválida (propio|referencia)")
	// ErrEntradaNoEnCatalogo: la fila pedida no está en el catálogo del marketplace → 400.
	ErrEntradaNoEnCatalogo = errors.New("marketplace: la entrada no está en el catálogo de este marketplace")
)

// CorruptaMarketplace expone solo el motivo de una fila corrupta del registro — el blob crudo no
// viaja al wire.
type CorruptaMarketplace struct {
	Motivo string `json:"motivo"`
}

// ListadoMarketplaces es el wire del plano. Corruptas y AvisoDetector son VISIBLES: un registro
// dañado o una metadata de CC ilegible nunca se oculta ni impide listar el resto (BR-11).
type ListadoMarketplaces struct {
	Marketplaces      []domain.MarketplaceConocido `json:"marketplaces"`
	Corruptas         []CorruptaMarketplace        `json:"corruptas,omitempty"`
	AvisoDetector     string                       `json:"aviso_detector,omitempty"`
	SinOrigenResuelto int                          `json:"sin_origen_resuelto"`
}

// Listar arma el plano Marketplaces (S2): collect-all detector+store → MergeMarketplaces →
// hidrata Lectura de cada fila desde el CACHÉ (sin tocar red ni parsear catálogos: el caché
// existe para que este GET sea O(1) lecturas chicas por fila, AG-D16). Cuenta también las
// entradas del Portafolio sin origen resuelto (el contador cruzado de AG-D8 decisión 7).
// Un detector que falla NO aborta: la lista sale con las declaradas + el aviso visible (E-68).
func (s *MarketplaceService) Listar(_ context.Context) (ListadoMarketplaces, error) {
	var out ListadoMarketplaces

	detectados, derr := s.detector.Detectados()
	if derr != nil {
		out.AvisoDetector = derr.Error()
	}
	declarados, corruptas := s.store.Listar()
	for _, c := range corruptas {
		out.Corruptas = append(out.Corruptas, CorruptaMarketplace{Motivo: c.Motivo})
	}

	filas := domain.MergeMarketplaces(detectados, declarados)
	for i := range filas {
		filas[i].Lectura = s.lecturaCacheada(filas[i].Nombre)
	}
	// `marketplaces` NUNCA es null: «no conozco ningún marketplace» es una afirmación verdadera
	// y verificable (a diferencia de un catálogo no leído, que viaja null).
	if filas == nil {
		filas = []domain.MarketplaceConocido{}
	}
	out.Marketplaces = filas
	out.SinOrigenResuelto = s.contarSinOrigenResuelto()
	return out, nil
}

// lecturaCacheada hidrata el EstadoLectura de una fila desde el caché, SIN leer el catálogo.
// Ausente ⇒ `no-leido` (nunca «0 entradas»); corrupto ⇒ `no-leido` CON el motivo visible (E-48).
func (s *MarketplaceService) lecturaCacheada(nombre string) domain.EstadoLectura {
	cat, motivo, ok := s.cache.Leer(nombre)
	if ok {
		return cat.Lectura
	}
	return domain.EstadoLectura{Tipo: domain.LecturaNoLeida, Motivo: motivo}
}

// contarSinOrigenResuelto es el contador cruzado de AG-D8 decisión 7: identidades provisionales
// (Home=="") que el operador todavía NO confirmó dejar sin origen.
func (s *MarketplaceService) contarSinOrigenResuelto() int {
	entradas, _ := s.portafolio.Listar()
	n := 0
	for _, e := range entradas {
		if e.Identidad.Provisional() && e.OrigenSinResolverDesde == "" {
			n++
		}
	}
	return n
}

// buscar devuelve la fila mergeada de `nombre`. No existe ⇒ ErrMarketplaceNoConocido (404).
func (s *MarketplaceService) buscar(nombre string) (domain.MarketplaceConocido, error) {
	detectados, _ := s.detector.Detectados()
	declarados, _ := s.store.Listar()
	for _, m := range domain.MergeMarketplaces(detectados, declarados) {
		if m.Nombre == nombre {
			// La Lectura se hidrata del caché igual que en Listar: una fila del wire NUNCA viaja con
			// `tipo: ""` — `no-leido` es el cero del tipo A PROPÓSITO («no leído aún»), y un enum
			// vacío obligaría al FE a inventar el estado.
			m.Lectura = s.lecturaCacheada(m.Nombre)
			return m, nil
		}
	}
	return domain.MarketplaceConocido{}, fmt.Errorf("%w: %q", ErrMarketplaceNoConocido, nombre)
}

// Catalogo devuelve el catálogo de `nombre` con la situación ya calculada por fila.
// refrescar=false → caché primero; si no hay caché (o está corrupto), hace UNA lectura y la
// persiste. refrescar=true → lectura fresca obligatoria (AG-D8 decisión 3).
// Si la lectura falla y HAY caché: devuelve el caché + Lectura degradada con motivo + el
// `Cuando` viejo (BR-3+BR-4). Si falla y NO hay caché: Entradas=nil (null en el wire) +
// Lectura con motivo. JAMÁS un `[]` fabricado, JAMÁS un 500.
func (s *MarketplaceService) Catalogo(ctx context.Context, nombre string, refrescar bool) (domain.Catalogo, error) {
	m, err := s.buscar(nombre)
	if err != nil {
		return domain.Catalogo{}, err
	}

	var motivoCache string
	if !refrescar {
		if cat, motivo, ok := s.cache.Leer(nombre); ok {
			return s.conSituacion(m, cat), nil // camino normal: cero I/O grande.
		} else if motivo != "" {
			motivoCache = motivo // se arrastra al Lectura.Motivo del intento que sigue.
		}
	}

	// DD-2/E-bis · «↻ Refrescar» = fetch/pull REAL, no solo releer el clone stale: en
	// refresco EXPLÍCITO y SOLO para un marketplace PROPIO se sincroniza el checkout
	// antes de leer (referencia jamás: sobre lo ajeno solo se lee — invariante 3 del
	// boundary). Best-effort: el pull fallido se arrastra a Lectura.Motivo — TAMBIÉN en
	// lectura exitosa (el clone quedó stale y eso se dice) — y la lectura sigue.
	var motivoSync string
	if refrescar && s.sync != nil && m.InstallLocation != "" && domain.ClaseSegura(m.Clase) == domain.ClasePropio {
		if _, serr := s.sync.Sincronizar(ctx, m); serr != nil {
			motivoSync = serr.Error()
		}
	}

	cat, lerr := s.leerFresco(ctx, m)
	if lerr == nil {
		cat.Lectura.Tipo = domain.LecturaLeida
		cat.Lectura.Cuando = s.ahora().Format(time.RFC3339)
		cat.Lectura.Entradas = len(cat.Entradas)
		cat.Lectura.Motivo = juntarMotivos(cat.Lectura.Motivo, motivoSync)
		if gerr := s.cache.Guardar(nombre, cat); gerr != nil {
			// La lectura NO se pierde por no poder guardarla (E-51).
			cat.Lectura.Motivo = strings.TrimSpace(cat.Lectura.Motivo + fmt.Sprintf(" (no se pudo cachear: %v)", gerr))
		}
		return s.conSituacion(m, cat), nil
	}
	motivoCache = juntarMotivos(motivoCache, motivoSync)

	// Falló la lectura fresca. Con caché: se devuelve el caché + la degradación VISIBLE + el
	// `Cuando` VIEJO — la UI dice «leído hace 2 días · ahora sin acceso», nunca «al día» (E-29/E-63).
	if viejo, _, hay := s.cache.Leer(nombre); hay {
		viejo.Lectura = domain.EstadoLectura{
			Tipo:     tipoDeErrorLectura(lerr),
			Cuando:   viejo.Lectura.Cuando,
			Entradas: len(viejo.Entradas),
			Motivo:   juntarMotivos(motivoCache, lerr.Error()),
			Fuente:   viejo.Lectura.Fuente,
		}
		return s.conSituacion(m, viejo), nil
	}

	// Sin caché: `entradas` viaja NULL con su motivo. Nunca un `[]` fabricado (BR-4).
	degradado := domain.Catalogo{
		Marketplace: m.Nombre, Clase: m.Clase, Repo: m.Repo,
		Lectura: domain.EstadoLectura{Tipo: tipoDeErrorLectura(lerr), Motivo: juntarMotivos(motivoCache, lerr.Error())},
	}
	return degradado, nil
}

// leerFresco aplica la política de §5.1: LOCAL primero (barato, offline, cero red — el checkout
// que CC ya mantiene), remoto como fallback. Sin ninguna vía ⇒ ErrSinViaDeLectura, con el motivo
// del intento local adentro para no perderlo (E-35).
func (s *MarketplaceService) leerFresco(ctx context.Context, m domain.MarketplaceConocido) (domain.Catalogo, error) {
	var motivoLocal string
	if s.local != nil && m.InstallLocation != "" {
		cat, err := s.local.Leer(ctx, m)
		if err == nil {
			return cat, nil
		}
		motivoLocal = err.Error()
	}
	if s.remoto != nil {
		cat, err := s.remoto.Leer(ctx, m)
		if err == nil {
			return cat, nil
		}
		return domain.Catalogo{}, err
	}
	if motivoLocal != "" {
		return domain.Catalogo{}, fmt.Errorf("%w: %s", ErrSinViaDeLectura, motivoLocal)
	}
	return domain.Catalogo{}, ErrSinViaDeLectura
}

// tipoDeErrorLectura clasifica el error a uno de los 4 TipoLectura. Cualquiera de los dos
// degradados EXIGE motivo — un degradado mudo no informa nada.
func tipoDeErrorLectura(err error) domain.TipoLectura {
	switch {
	case errors.Is(err, ErrSinViaDeLectura), errors.Is(err, os.ErrPermission):
		return domain.LecturaSinAcceso
	case errors.Is(err, ErrNoEsMarketplace):
		return domain.LecturaURLNoResuelve
	default:
		return domain.LecturaSinAcceso
	}
}

// juntarMotivos concatena motivos no vacíos con " · " (el arrastre del caché corrupto + el error
// del intento fresco viajan juntos, ninguno se pierde).
func juntarMotivos(vs ...string) string {
	var out []string
	for _, v := range vs {
		if strings.TrimSpace(v) != "" {
			out = append(out, v)
		}
	}
	return strings.Join(out, " · ")
}

// conSituacion cruza CADA fila del catálogo contra el Portafolio y calcula su situación + acción
// (design.md §6). Se recalcula SIEMPRE, en todos los caminos (caché incluido) y NUNCA se
// persiste: el Portafolio cambia con cada Traer/Identificar/escaneo, y una situación cacheada
// mentiría con el estado viejo — E-106 lo vigila (C24 de design.md).
func (s *MarketplaceService) conSituacion(m domain.MarketplaceConocido, cat domain.Catalogo) domain.Catalogo {
	cat.Marketplace = m.Nombre
	cat.Clase = m.Clase
	if cat.Repo == "" {
		cat.Repo = m.Repo
	}
	if cat.Entradas == nil {
		return cat // sin lectura no hay filas que cruzar: `null` viaja tal cual (BR-4).
	}
	entradas, _ := s.portafolio.Listar()
	for i := range cat.Entradas {
		coincidencias := domain.CruzarConPortafolio(m.Repo, cat.Entradas[i], entradas)
		cat.Entradas[i].Situacion = domain.CalcularSituacion(cat.Entradas[i], coincidencias)
		cat.Entradas[i].Accion = domain.AccionDeSituacion(cat.Entradas[i].Situacion, m.Clase)
	}
	return cat
}

// Validacion es lo que S6 pinta en el ✓: datos LEÍDOS del archivo real, nunca inferidos.
type Validacion struct {
	URLCanonica  string                  `json:"url_canonica"`
	Nombre       string                  `json:"nombre"`
	OwnerNombre  string                  `json:"owner_nombre,omitempty"`
	OwnerEmail   string                  `json:"owner_email,omitempty"`
	OwnerURL     string                  `json:"owner_url,omitempty"`
	Descripcion  string                  `json:"descripcion,omitempty"`
	Entradas     int                     `json:"entradas"`
	Fuente       string                  `json:"fuente"` // local | remoto
	YaRegistrado bool                    `json:"ya_registrado"`
	ClaseActual  domain.ClaseMarketplace `json:"clase_actual,omitempty"` // solo si YaRegistrado.
}

// Validar prueba una url SIN persistir nada (S6 paso 1, BR-5): canonicaliza → si algún
// marketplace conocido ya apunta a ese repo usa su checkout local (cero red) → si no, remoto.
// Devuelve el Catalogo leído (con Entradas reales) + si el nombre ya está registrado.
// **Solo un 200 pinta ✓**: no hay forma de fabricar un ✓ sin un archivo leído (G3).
func (s *MarketplaceService) Validar(ctx context.Context, url string) (Validacion, error) {
	canon, ok := domain.CanonicalizarRepo(url)
	if !ok {
		return Validacion{}, fmt.Errorf("%w: %q", ErrURLNoCanonicalizable, url)
	}

	cat, err := s.leerPorRepo(ctx, canon)
	if err != nil {
		return Validacion{}, err
	}

	v := Validacion{
		URLCanonica: canon,
		Nombre:      cat.Marketplace,
		OwnerNombre: cat.OwnerNombre,
		OwnerEmail:  cat.OwnerEmail,
		OwnerURL:    cat.OwnerURL,
		Descripcion: cat.Descripcion,
		Entradas:    len(cat.Entradas),
		Fuente:      cat.Lectura.Fuente,
	}
	declarados, _ := s.store.Listar()
	for _, d := range declarados {
		if d.Nombre == v.Nombre {
			v.YaRegistrado, v.ClaseActual = true, d.Clase
			break
		}
	}
	return v, nil
}

// leerPorRepo lee el catálogo de un repo canonicalizado: si algún marketplace CONOCIDO ya apunta
// a ese repo y tiene checkout, se usa el local (cero red); si no, el validador remoto.
func (s *MarketplaceService) leerPorRepo(ctx context.Context, canon string) (domain.Catalogo, error) {
	detectados, _ := s.detector.Detectados()
	declarados, _ := s.store.Listar()
	for _, m := range domain.MergeMarketplaces(detectados, declarados) {
		if m.InstallLocation == "" || !mismoRepoCanon(m.Repo, canon) || s.local == nil {
			continue
		}
		if cat, err := s.local.Leer(ctx, m); err == nil {
			return cat, nil
		}
	}
	if s.validador == nil {
		return domain.Catalogo{}, fmt.Errorf("%w: sin lector remoto cableado", ErrSinViaDeLectura)
	}
	return s.validador.Validar(ctx, canon)
}

// mismoRepoCanon compara canonicalizando los DOS lados (RN-IDENT-1).
func mismoRepoCanon(a, b string) bool {
	ca, oka := domain.CanonicalizarRepo(a)
	cb, okb := domain.CanonicalizarRepo(b)
	if oka && okb {
		return ca == cb
	}
	return a != "" && a == b
}

// Registrar persiste el lado declarado (S6 paso 4). Re-valida (TOCTOU-safe, mismo criterio que
// AgregarProyecto) y devuelve la fila mergeada lista para aterrizar en el catálogo.
// Nombre ya registrado → ErrMarketplaceYaRegistrado (BR-7): no duplica, NO PISA.
// NO clona, NO instala, NO escribe fuera de `~/.arnesia/` (boundary §11.3).
func (s *MarketplaceService) Registrar(ctx context.Context, url string, clase domain.ClaseMarketplace) (domain.MarketplaceConocido, error) {
	if !domain.ClaseValida(clase) {
		return domain.MarketplaceConocido{}, fmt.Errorf("%w: %q", ErrClaseInvalida, string(clase))
	}
	v, err := s.Validar(ctx, url)
	if err != nil {
		return domain.MarketplaceConocido{}, err
	}
	if v.YaRegistrado {
		return domain.MarketplaceConocido{Nombre: v.Nombre}, fmt.Errorf("%w: %q", ErrMarketplaceYaRegistrado, v.Nombre)
	}

	fila := domain.MarketplaceConocido{
		Nombre:     v.Nombre,
		Repo:       v.URLCanonica,
		Clase:      clase,
		Registrado: s.ahora().Format(time.RFC3339),
	}
	if uerr := s.store.Upsert(fila); uerr != nil {
		return domain.MarketplaceConocido{}, uerr
	}
	// Se devuelve la fila YA MERGEADA (con el eslabón de CC si lo hay) para que el FE aterrice
	// en el catálogo correcto sin un GET extra.
	if mergeada, berr := s.buscar(v.Nombre); berr == nil {
		return mergeada, nil
	}
	return fila, nil
}

// Olvidar quita el lado DECLARADO. Si CC igual lo conoce, la fila sigue apareciendo con
// eslabón `cc-known-marketplaces` solo (y clase degradada a `referencia`, fail-safe) — es lo
// honesto: no podemos hacer que Claude Code deje de conocerlo (E-71).
func (s *MarketplaceService) Olvidar(_ context.Context, nombre string) (olvidado, sigueDetectado bool, err error) {
	olvidado, err = s.store.Olvidar(nombre)
	if err != nil {
		return false, false, err
	}
	detectados, _ := s.detector.Detectados()
	for _, d := range detectados {
		if d.Nombre == nombre {
			sigueDetectado = true
			break
		}
	}
	return olvidado, sigueDetectado, nil
}

// CandidatoOrigen es una opción del selector de S7.
type CandidatoOrigen struct {
	Nombre string                  `json:"nombre"`
	Repo   string                  `json:"repo,omitempty"`
	Clase  domain.ClaseMarketplace `json:"clase"`
	// Senal es el texto que la fila muestra debajo del nombre — se ARMA acá, no en el widget,
	// para que sea testeable.
	Senal string `json:"senal,omitempty"`
}

// CandidatosDeOrigen alimenta el selector de S7: los marketplaces conocidos ordenados por
// señales BLANDAS respecto de la entrada `clave` (PENDIENTE-01 §3). Orden:
//  1. el repo ya está en RegistriesDe(entrada) — la señal más fuerte que existe sin manifiesto;
//  2. coincidencia de owner del repo con las `empresas` de la entrada;
//  3. clase `propio` antes que `referencia`;
//  4. resto por nombre.
//
// NADA premarcado: el orden es sugerencia, la elección es del operador (BR-11).
func (s *MarketplaceService) CandidatosDeOrigen(_ context.Context, clave string) ([]CandidatoOrigen, string, error) {
	entradas, _ := s.portafolio.Listar()
	var entrada domain.EntradaPortafolio
	var hallada bool
	for _, e := range entradas {
		if e.Identidad.Clave() == clave {
			entrada, hallada = e, true
			break
		}
	}
	if !hallada {
		return nil, "", fmt.Errorf("%w: %q", ErrObservarClaveNoEncontrada, clave)
	}

	registries := domain.RegistriesDe(entrada)
	detectados, _ := s.detector.Detectados()
	declarados, _ := s.store.Listar()
	filas := domain.MergeMarketplaces(detectados, declarados)
	for i := range filas {
		// La Lectura se hidrata del caché igual que en Listar/buscar: sin esto la señal decía «no
		// leído aún» de un marketplace que SÍ se había leído (bug cazado en el E2E vivo).
		filas[i].Lectura = s.lecturaCacheada(filas[i].Nombre)
	}

	type puntuado struct {
		c     CandidatoOrigen
		rango int
	}
	puntuados := make([]puntuado, 0, len(filas))
	for _, m := range filas {
		c := CandidatoOrigen{Nombre: m.Nombre, Repo: m.Repo, Clase: m.Clase}
		var senales []string
		rango := 4
		switch {
		case m.Repo != "" && contieneCanon(registries, m.Repo):
			rango = 1
			senales = append(senales, "el registry de tu copia ya apunta acá")
		case ownerCoincide(m.Repo, entrada.Empresas):
			rango = 2
			senales = append(senales, "autor coincide")
		case m.Clase == domain.ClasePropio:
			rango = 3
		}
		senales = append(senales, claseLegible(m.Clase))
		if m.Lectura.Tipo == domain.LecturaNoLeida || m.Lectura.Tipo == "" {
			senales = append(senales, "no leído aún")
		}
		c.Senal = strings.Join(senales, " · ")
		puntuados = append(puntuados, puntuado{c: c, rango: rango})
	}
	sort.SliceStable(puntuados, func(i, j int) bool {
		if puntuados[i].rango != puntuados[j].rango {
			return puntuados[i].rango < puntuados[j].rango
		}
		return puntuados[i].c.Nombre < puntuados[j].c.Nombre
	})

	out := make([]CandidatoOrigen, 0, len(puntuados))
	for _, p := range puntuados {
		out = append(out, p.c)
	}
	return out, entrada.Identidad.Home, nil
}

// claseLegible traduce la clase al vocabulario de la UI (un solo lugar, testeable).
func claseLegible(c domain.ClaseMarketplace) string {
	if c == domain.ClasePropio {
		return "propio"
	}
	return "de referencia"
}

// contieneCanon reporta si repo (canonicalizado) está en xs.
func contieneCanon(xs []string, repo string) bool {
	canon, ok := domain.CanonicalizarRepo(repo)
	if !ok {
		canon = repo
	}
	for _, x := range xs {
		if x == canon {
			return true
		}
	}
	return false
}

// ownerCoincide reporta si el owner del repo coincide (case-insensitive) con alguna empresa
// declarada de la entrada — señal BLANDA, nunca una elección automática.
func ownerCoincide(repo string, empresas []string) bool {
	partes := strings.Split(repo, "/")
	if len(partes) < 2 {
		return false
	}
	owner := strings.ToLower(partes[1])
	for _, e := range empresas {
		if strings.ToLower(strings.TrimSpace(e)) == owner {
			return true
		}
	}
	return false
}
