package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// telemetria_service.go convierte filas en la frase del producto —*«este arnés, en este
// puesto, quema $X»*— con cada número cargando **cómo** se atribuyó.
//
// Tres reglas gobiernan el archivo entero:
//
//  1. **La atribución se resuelve en orden y el primero que acierta la fija** (§6.3):
//     exacta → por-hash → por-proceso → sin-dato.
//  2. **Un evento `sin-dato` se guarda y NO se suma al total** (A15). Aparece en la cobertura
//     y en el drill-down. Sumarlo sería atribuir por adivinanza.
//  3. **El escenario se DERIVA de la señal**, jamás lo declara el emisor: un arnés no puede
//     mentir sobre su propio nivel de instrumentación.

// TelemetriaService es el caso de uso del módulo. El HTTP y el CLI usan ESTE, no cada uno el
// suyo (A17): dos implementaciones de la misma consulta divergen el día que alguien arregla
// una sola.
type TelemetriaService struct {
	store      ports.TelemetriaStore
	catalogo   ports.CatalogoPrecios
	atribucion ports.AtribucionRegistry
	forward    ports.ForwardOTLP
	detectores []domain.Detector
	perfil     domain.PerfilRuntime
	reloj      func() time.Time

	// rolResolver da el `rol` del arnés indexado — de ahí sale el `puesto` de la fila del
	// Portafolio (D20). Devuelve "" cuando el arnés no declara rol o no está indexado, y
	// entonces el puesto viaja `null`: «puesto sin declarar», que es el caso NORMAL hoy.
	rolResolver func(ctx context.Context, arnesID string) string
	// instalaciones lista (arnesID, instalacionID, clave, nombre, empresas) del Portafolio.
	instalaciones func(ctx context.Context) []FilaInstalacion

	// Retención (T14). El TTL en días es un valor PROPUESTO, no firmado (J-6): viaja por
	// flag y se muestra rotulado como tal.
	purga          purgador
	recomputa      recomputador
	retencionDias  int
	rollupMeses    int
	descubrimiento string
}

// FilaInstalacion es lo mínimo que el servicio necesita del Portafolio para armar la tabla.
// Se inyecta como función y no como puerto para no atar `telemetria` al modelo del
// Portafolio: si mañana cambia, cambia el adaptador del composition root, no este archivo.
type FilaInstalacion struct {
	ArnesID       string
	InstalacionID string
	Clave         string
	Nombre        string
	Empresas      []string
}

// NewTelemetriaService cablea el servicio. `catalogo`, `atribucion`, `forward`, `rolResolver`
// e `instalaciones` pueden ser nil: el servicio degrada honesto (sin costo calculado, sin
// atribución por hash, sin forward, con `puesto: null`) en vez de fallar.
func NewTelemetriaService(
	store ports.TelemetriaStore,
	catalogo ports.CatalogoPrecios,
	atribucion ports.AtribucionRegistry,
	forward ports.ForwardOTLP,
	detectores []domain.Detector,
	perfil domain.PerfilRuntime,
	reloj func() time.Time,
) *TelemetriaService {
	if reloj == nil {
		reloj = time.Now
	}
	return &TelemetriaService{
		store: store, catalogo: catalogo, atribucion: atribucion, forward: forward,
		detectores: detectores, perfil: perfil, reloj: reloj,
	}
}

// SetPortafolio inyecta las dos funciones que resuelven el Portafolio. Se llama desde el
// composition root, después de construir el servicio del Portafolio.
func (s *TelemetriaService) SetPortafolio(
	rol func(ctx context.Context, arnesID string) string,
	instalaciones func(ctx context.Context) []FilaInstalacion,
) {
	s.rolResolver = rol
	s.instalaciones = instalaciones
}

// SetDescubrimiento registra dónde quedó publicada la ficha del daemon (o el motivo de no
// haberla publicado), para que la salud lo diga. Una ficha ausente que nadie menciona
// convertiría «el hook nunca reporta» en un misterio.
func (s *TelemetriaService) SetDescubrimiento(ruta string) { s.descubrimiento = ruta }

var _ ports.TelemetriaSink = (*TelemetriaService)(nil)

// Ingerir es la ÚNICA puerta de escritura del módulo: atribuye, costea y persiste. Los cuatro
// emisores pasan por acá.
func (s *TelemetriaService) Ingerir(ctx context.Context, evs []domain.EventoTelemetria) (int, error) {
	if len(evs) == 0 {
		return 0, nil
	}
	listos := make([]domain.EventoTelemetria, 0, len(evs))
	for _, e := range evs {
		s.Atribuir(&e)
		s.costear(&e)
		if e.TSRecibido.IsZero() {
			e.TSRecibido = s.reloj().UTC()
		}
		listos = append(listos, e)
	}
	n, err := s.store.Ingerir(ctx, listos)
	// El forward va DESPUÉS de persistir y con el evento YA PROYECTADO — jamás el cuerpo
	// crudo, que llevaría el email de quien corra el arnés.
	if s.forward != nil && s.forward.Activo() {
		if ferr := s.forward.Enviar(ctx, listos); ferr != nil {
			slog.Warn("telemetria: forward falló — la telemetría local no se ve afectada", "err", ferr)
		}
	}
	return n, err
}

// Atribuir resuelve a qué unidad de trabajo pertenece un evento, en el orden de preferencia
// de §6.3. **El primero que acierta fija la confianza.**
//
// Modifica el evento en el lugar; no devuelve nada, porque no hay caso de fallo: el peor
// desenlace es `sin-dato`, que también es una respuesta.
func (s *TelemetriaService) Atribuir(e *domain.EventoTelemetria) {
	// 1 · exacta — los `arnesia.*` que inyectamos nosotros al spawn (S1). Si ya vinieron,
	//     no hay nada mejor que buscar.
	if e.ArnesID != "" {
		e.Atribucion = domain.ConfianzaExacta
		return
	}
	// 2 · por-hash — `plugin_id_hash` resuelve en la tabla aprendida. Da el ARNÉS, no la
	//     caja: el hash identifica el paquete instalado, no qué parte de él corrió.
	if s.atribucion != nil && e.PluginIDHash != "" {
		if arnes, inst, ok := s.atribucion.PorHash(e.PluginIDHash); ok {
			e.ArnesID, e.InstalacionID = arnes, inst
			e.Atribucion = domain.ConfianzaPorHash
			return
		}
	}
	// 3 · por-proceso — el `cwd` del hook mapea a una instalación conocida del Portafolio.
	//     Da arnés + instalación, tampoco la caja.
	if s.atribucion != nil && e.CWDHuella != "" {
		if arnes, inst, ok := s.atribucion.PorCWD(e.CWDHuella); ok {
			e.ArnesID, e.InstalacionID = arnes, inst
			e.Atribucion = domain.ConfianzaPorProceso
			return
		}
	}
	// 4 · sin-dato — nada resolvió. **Se guarda igual** y no suma a ningún total (A15).
	if e.Atribucion == "" || e.ArnesID == "" {
		e.Atribucion = domain.ConfianzaSinDato
	}
}

// costear cotiza el evento con NUESTRO catálogo, sin pisar lo que reportó el runtime. Los DOS
// costos conviven: si divergen, la prueba de paridad corre en producción y gratis.
//
// Un modelo que el catálogo no conoce deja `CostoCalculadoMicros` en **nil**, no en 0: un 0
// en dinero se lee como «salió gratis».
func (s *TelemetriaService) costear(e *domain.EventoTelemetria) {
	if s.catalogo == nil || e.Modelo == "" {
		return
	}
	canonico := s.catalogo.Canonizar(e.Modelo)
	e.ModeloCanonico = canonico
	precio, ok := s.catalogo.Precio(canonico)
	if !ok {
		return // no se inventa una tarifa; el costo calculado queda nil y la UI lo dice.
	}
	e.Proveedor = precio.Proveedor
	arit := e.Aritmetica
	if arit == "" {
		arit = s.perfil.Aritmetica
	}
	c := domain.CalcularCosto(e.Tokens, precio, arit)
	if c.SinNingunaTarifa {
		return // no se pudo cotizar NADA: null, no 0.
	}
	micros := c.Micros
	e.CostoCalculadoMicros = &micros
	completo := c.Completo
	e.CostoCompleto = &completo
	e.CatalogoVersion = s.catalogo.Version().Version
}

// Resumen agrega la ventana. Delega en el almacén (que ya aplica A15 y la conciliación) y le
// agrega la versión del catálogo, que es del servicio.
func (s *TelemetriaService) Resumen(ctx context.Context, q ports.ConsultaTelemetria) (domain.ResumenTelemetria, error) {
	r, err := s.store.Resumen(ctx, s.ventana(q))
	if err != nil {
		return r, err
	}
	if s.catalogo != nil {
		r.Catalogo = s.catalogo.Version()
	}
	return r, nil
}

// PorCaja devuelve el gasto por caja, con las marcas de fuga que los detectores encontraron.
func (s *TelemetriaService) PorCaja(ctx context.Context, q ports.ConsultaTelemetria) ([]domain.GastoCaja, error) {
	cajas, err := s.store.PorCaja(ctx, s.ventana(q))
	if err != nil {
		return nil, err
	}
	mejoras, merr := s.Mejoras(ctx, q)
	if merr != nil {
		// El desglose vale sin las marcas; no se cae por eso.
		slog.Warn("telemetria: no se pudieron calcular las marcas de fuga", "err", merr)
		return cajas, nil
	}
	porCaja := map[string][]domain.MarcaDeFuga{}
	for _, p := range mejoras.Puntos {
		if p.CajaID == "" {
			continue
		}
		porCaja[p.CajaID] = append(porCaja[p.CajaID], domain.MarcaDeFuga{
			Detector: p.Detector, Nombre: nombreDetector(s.detectores, p.Detector), Grave: p.Grave,
		})
	}
	for i := range cajas {
		cajas[i].Marcas = porCaja[cajas[i].CajaID]
	}
	return cajas, nil
}

// Turnos es el drill-down: el join dinero×proceso por `(sesion_id, turno_id)`.
func (s *TelemetriaService) Turnos(ctx context.Context, q ports.ConsultaTelemetria) ([]domain.TurnoUnido, error) {
	return s.store.Turnos(ctx, s.ventana(q))
}

// DetalleCaja arma la 4ª tab del inspector: buckets con sus nulos, los dos costos y su
// divergencia, el join a nivel nodo y el estado de cada detector.
func (s *TelemetriaService) DetalleCaja(ctx context.Context, q ports.ConsultaTelemetria) (domain.DetalleCaja, error) {
	v := s.ventana(q)
	out := domain.DetalleCaja{CajaID: q.CajaID, Nombre: q.CajaID, Desde: v.Desde, Hasta: v.Hasta}
	turnos, err := s.store.Turnos(ctx, v)
	if err != nil {
		return out, err
	}
	out.Turnos = turnos
	out.Confianza = domain.ConfianzaSinDato
	var rep, calc int64
	var hayRep, hayCalc bool
	for _, t := range turnos {
		out.Atribuible = out.Atribuible || t.Atribucion != domain.ConfianzaSinDato
		out.Confianza = domain.PeorConfianza(out.Confianza, t.Atribucion)
		sumarPtr(&out.Tokens.Entrada, t.Tokens.Entrada)
		sumarPtr(&out.Tokens.Salida, t.Tokens.Salida)
		sumarPtr(&out.Tokens.CacheLectura, t.Tokens.CacheLectura)
		sumarPtr(&out.Tokens.CacheEscritura5m, t.Tokens.CacheEscritura5m)
		sumarPtr(&out.Tokens.CacheEscritura1h, t.Tokens.CacheEscritura1h)
		sumarPtr(&out.Tokens.Razonamiento, t.Tokens.Razonamiento)
		if t.CostoReportadoMicros != nil {
			rep += *t.CostoReportadoMicros
			hayRep = true
		}
		if t.CostoCalculadoMicros != nil {
			calc += *t.CostoCalculadoMicros
			hayCalc = true
		}
	}
	if !out.Atribuible {
		out.Motivo = "sin dato atribuible: ningún evento de esta ventana pudo asignarse a esta caja"
	}
	if hayRep {
		out.Paridad.ReportadoMicros = &rep
	}
	if hayCalc {
		out.Paridad.CalculadoMicros = &calc
	}
	// La divergencia solo existe cuando hay DOS números. Con uno solo viaja nil, no 0: un 0
	// diría «coinciden», que es otra afirmación.
	if hayRep && hayCalc && rep != 0 {
		d := (float64(calc) - float64(rep)) / float64(rep) * 100
		out.Paridad.DivergenciaPct = &d
	}
	out.Paridad.Completo = hayRep && hayCalc

	mejoras, merr := s.Mejoras(ctx, q)
	if merr == nil {
		out.Detectores = append(append([]domain.EstadoDetector{}, mejoras.NoAplican...), mejoras.NoMedidos...)
		vistos := map[domain.DetectorID]bool{}
		for _, d := range out.Detectores {
			vistos[d.Detector] = true
		}
		for _, p := range mejoras.Puntos {
			if vistos[p.Detector] {
				continue
			}
			vistos[p.Detector] = true
			out.Detectores = append(out.Detectores, domain.EstadoDetector{
				Detector: p.Detector, Nombre: nombreDetector(s.detectores, p.Detector),
				Aplica: true, Hallazgos: 1,
			})
		}
	}
	if s.catalogo != nil {
		out.Catalogo = s.catalogo.Version()
	}
	return out, nil
}

// Portafolio arma una fila por **(arnés, instalación)** — la unidad que el Portafolio modela
// (D20). El «puesto» sale del `rol` del arnés indexado, resuelto acá porque el servicio tiene
// el índice a mano; `null` cuando el arnés no lo declara, y entonces la UI dice
// «puesto sin declarar».
func (s *TelemetriaService) Portafolio(ctx context.Context, q ports.ConsultaTelemetria) ([]domain.FilaPortafolio, error) {
	var filas []FilaInstalacion
	if s.instalaciones != nil {
		filas = s.instalaciones(ctx)
	}
	out := make([]domain.FilaPortafolio, 0, len(filas))
	for _, f := range filas {
		q2 := s.ventana(q)
		q2.ArnesID = f.ArnesID
		q2.InstalacionID = f.InstalacionID
		r, err := s.store.Resumen(ctx, q2)
		if err != nil {
			return nil, err
		}
		fila := domain.FilaPortafolio{
			ArnesID: f.ArnesID, Clave: f.Clave, Nombre: f.Nombre,
			InstalacionID: f.InstalacionID, Empresas: f.Empresas,
			Corridas: r.Corridas, CostoMicros: r.CostoReportadoMicros,
			Confianza: r.Confianza,
		}
		// D20 · el puesto sale del `rol` del arnés, no de un campo del Portafolio.
		if s.rolResolver != nil {
			if rol := s.rolResolver(ctx, f.ArnesID); rol != "" {
				p := rol
				fila.Puesto = &p
			}
		}
		// `costo_por_corrida: null` para el que nunca corrió. Un 0 diría «corrió gratis».
		if r.CostoReportadoMicros != nil && r.Corridas > 0 {
			cpc := *r.CostoReportadoMicros / int64(r.Corridas)
			fila.CostoPorCorrida = &cpc
		}
		if mej, merr := s.Mejoras(ctx, q2); merr == nil {
			fila.PuntosDeMejora = len(mej.Puntos)
		}
		out = append(out, fila)
	}
	return out, nil
}

// Conciliar registra que un turno OCURRIÓ, lo sepamos medir o no. Es el denominador de la
// cobertura: sin esto, «cuánto medimos» se leería como «cuánto hubo».
func (s *TelemetriaService) Conciliar(ctx context.Context, sesionID, turnoID, arnesID, cajaID string) error {
	return s.store.EsperarTurno(ctx, sesionID, turnoID, arnesID, cajaID)
}

// Salud reporta el estado del módulo, con el catálogo y el forward que son del servicio.
func (s *TelemetriaService) Salud(ctx context.Context) (domain.SaludTelemetria, error) {
	sal, err := s.store.Salud(ctx)
	if err != nil {
		return sal, err
	}
	if s.catalogo != nil {
		sal.Catalogo = s.catalogo.Version()
	}
	sal.Descubrimiento = s.descubrimiento
	// El TTL que se muestra es el de la CONFIG, no un número escrito en la UI (J-6).
	sal.RetencionDias = s.RetencionDias()
	sal.RollupMeses = s.RollupMeses()
	sal.RetencionPropuesta = true
	if s.forward != nil {
		sal.Forward = s.forward.Activo()
		if sal.Forward {
			sal.ForwardDestino = s.forward.Destino()
		}
	}
	return sal, nil
}

// Forward reenvía eventos al destino externo del OPERADOR, si está encendido. Apagado
// devuelve nil sin abrir un socket: el default es no egresar.
func (s *TelemetriaService) Forward(ctx context.Context, evs []domain.EventoTelemetria) error {
	if s.forward == nil || !s.forward.Activo() {
		return nil
	}
	return s.forward.Enviar(ctx, evs)
}

// ventana normaliza el rango. Sin `Hasta` usa el reloj; sin `Desde` usa 30 días atrás — un
// default acotado y explícito, no «toda la historia», que haría que el primer tablero de una
// base grande tardara lo que no debe.
func (s *TelemetriaService) ventana(q ports.ConsultaTelemetria) ports.ConsultaTelemetria {
	if q.Hasta.IsZero() {
		q.Hasta = s.reloj().UTC()
	}
	if q.Desde.IsZero() {
		q.Desde = q.Hasta.AddDate(0, 0, -30)
	}
	return q
}

func sumarPtr(dst **int64, v *int64) {
	if v == nil {
		return
	}
	if *dst == nil {
		n := *v
		*dst = &n
		return
	}
	**dst += *v
}

func nombreDetector(ds []domain.Detector, id domain.DetectorID) string {
	for _, d := range ds {
		if d.ID() == id {
			return d.Nombre()
		}
	}
	return string(id)
}

// registroAtribucion adapta el Store al puerto `AtribucionRegistry`. Vive acá y no en el
// adaptador porque el mapeo `huella → instalación` necesita el Portafolio, que es del caso de
// uso; el store solo sabe de la tabla `atribucion_hash`.
type registroAtribucion struct {
	buscarHash func(ctx context.Context, hash string) (string, string, bool)
	porCWD     func(huella string) (string, string, bool)
	aprender   func(ctx context.Context, hash, arnesID, instalacionID, como string) error
}

// NuevoRegistroAtribucion arma el registro desde funciones. Se cablea en el composition root.
func NuevoRegistroAtribucion(
	buscarHash func(ctx context.Context, hash string) (string, string, bool),
	porCWD func(huella string) (string, string, bool),
	aprender func(ctx context.Context, hash, arnesID, instalacionID, como string) error,
) ports.AtribucionRegistry {
	return &registroAtribucion{buscarHash: buscarHash, porCWD: porCWD, aprender: aprender}
}

func (r *registroAtribucion) PorHash(hash string) (string, string, bool) {
	if r.buscarHash == nil {
		return "", "", false
	}
	return r.buscarHash(context.Background(), hash)
}

func (r *registroAtribucion) PorCWD(huella string) (string, string, bool) {
	if r.porCWD == nil {
		return "", "", false
	}
	return r.porCWD(huella)
}

func (r *registroAtribucion) Aprender(hash, arnesID, instalacionID, como string) error {
	if r.aprender == nil {
		return fmt.Errorf("telemetria: registro de atribución sin escritura cableada")
	}
	return r.aprender(context.Background(), hash, arnesID, instalacionID, como)
}
