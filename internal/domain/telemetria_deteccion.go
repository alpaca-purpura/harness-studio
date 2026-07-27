package domain

import (
	"fmt"
	"time"
)

// telemetria_deteccion.go es el CONTRATO de los detectores de puntos de mejora
// (arquitectura-modulo.md §2.4). Las seis implementaciones del MVP y `DetectoresMVP()`
// llegan en su propio ticket; acá vive lo que el resto del módulo necesita para hablar de
// ellos —incluidas las vistas de lectura, que devuelven `PuntoDeMejora`.
//
// La regla que gobierna el archivo entero: **`Aplica()` se consulta SIEMPRE antes que
// `Evaluar()`**, y **`Motivo` es obligatorio cuando `Aplica` es false**. Un detector
// apagado sin razón es un gap escondido, y un detector que devuelve 0 porque no pudo
// correr es una mentira (boundary no-aplica-no-es-cero).

// DetectorID identifica un detector de forma estable. El id viaja al wire y a la UI: es lo
// que hace que una marca de fuga NOMBRE al detector en vez de ser un ⚠ genérico.
type DetectorID string

const (
	DetB4 DetectorID = "b4-gasto-por-arnes-empresa-puesto"
	DetP1 DetectorID = "p1-caja-que-consume-y-se-rechaza"
	DetB2 DetectorID = "b2-costo-de-la-rotacion"
	DetB6 DetectorID = "b6-sesion-abandonada"
	DetB3 DetectorID = "b3-cambio-de-modelo-invalida-cache"
	DetB1 DetectorID = "b1-rewarm-por-ttl"
)

// ContextoDeteccion es lo que un detector necesita para decidir si PUEDE correr. Se arma
// del dato real de la ventana, jamás se asume ni se configura: si se declarara, un arnés
// podría decir que está mejor medido de lo que está.
type ContextoDeteccion struct {
	Runtime   string
	Escenario Escenario
	// TieneCosto: llegó ALGÚN evento con dinero en la ventana. En s2 esto depende de si el
	// arnés lleva el bloque `env` (ANEXO H9), no del escenario — por eso es un hecho de la
	// ventana y no un derivado del modo.
	TieneCosto bool
	// TieneSplitTTL: llegó el `result` del stream-json, que es donde vive el split
	// ephemeral_5m/1h. Es la regla de B1 — así, el día que el split llegue por OTel, el
	// detector se enciende SOLO, sin tocar código.
	TieneSplitTTL       bool
	CatalogoDisponible  bool
	TieneSenalProceso   bool // llegaron eventos de hook/daemon (P1 los exige)
	TieneGateHumano     bool // hubo eventos de gate del daemon (P1 completo vs. parcial)
	TieneEventoRotacion bool // el daemon registró rotaciones (B2)
	// ModelosDistintos cuenta los modelos no vacíos de la ventana (B3 exige ≥2).
	ModelosDistintos int
}

// Aplicabilidad es la respuesta honesta de un detector a «¿podés correr acá?».
type Aplicabilidad struct {
	Aplica bool   `json:"aplica"`
	Motivo string `json:"motivo,omitempty"`
	// Parcial marca que el detector corre pero no ve todo (el caso de P1 en
	// s2-instrumentado: ve reintentos y fracasos de herramienta, no rechazos de gate).
	// Viaja como matiz explícito con su motivo, jamás como un ✅ liso.
	Parcial bool `json:"parcial,omitempty"`
}

// PuntoDeMejora es la unidad del entregable: un número, su contrafactual, su umbral
// citado, su sesgo declarado EN CONTRA y UN fix (regla A4). Sin las cinco cosas no se
// muestra — una tarjeta sin contrafactual es un reproche, no una recomendación.
type PuntoDeMejora struct {
	// ID identifica al punto de forma estable entre dos consultas del mismo dato. Lo necesita
	// la superficie para poder descartar UNO y no «el tercero de la lista», que cambia de
	// lugar en cuanto el orden por ahorro cambia. Lo arma `Redactar()`.
	ID           string     `json:"id"`
	Detector     DetectorID `json:"detector"`
	ScoreVersion int        `json:"score_version"` // A7: cambiar la fórmula obliga a bumpearlo
	Titulo       string     `json:"titulo"`
	// Lede viaja SIEMPRE (sin `omitempty`): la tarjeta lo pinta sin condición, y un campo
	// ausente donde el FE espera texto es el mismo tipo de hueco que V-5 destapó.
	Lede        string `json:"lede"`
	CajaID      string `json:"caja_id,omitempty"`
	GastoMicros int64  `json:"gasto_micros"`
	// ParteDelTotal va en 0..1.
	ParteDelTotal float64 `json:"parte_del_total"`
	// ContrafactualMicros es qué habría costado el mundo alternativo, con las MISMAS
	// corridas (A2). No es «gastaste X»: es «con el cambio habrías gastado Y».
	ContrafactualMicros int64 `json:"contrafactual_micros"`
	DiferenciaMicros    int64 `json:"diferencia_micros"`
	// Contrafactual es ESA MISMA COSA EN PROSA, con su unidad declarada (RF-249) — la que la
	// tarjeta pinta. La arma `Redactar()` en `telemetria_prosa.go`, **nunca el FE** (D25 ·
	// design.md §1.3). Vacío ⇒ el punto no se publica como tarjeta (regla A4).
	Contrafactual string `json:"contrafactual"`
	// BaseContrafactual es la cláusula «con el arreglo» que abre esa frase, y la pone el
	// detector porque es lo único de la frase que él sabe. No viaja al wire: es insumo de
	// redacción, y duplicarlo afuera invitaría a rearmar la frase en otro lado.
	BaseContrafactual string `json:"-"`
	// Umbral es la desigualdad algebraica CITADA (A1), no un número mágico. `omitempty`
	// porque **no todo detector se decide por un umbral**: P1 se decide por un patrón, y
	// pintarle una desigualdad inventada sería fabricarle el rigor que no tiene (RF-250).
	Umbral string `json:"umbral,omitempty"`
	// Patron reemplaza al umbral en ese caso. Los dos vacíos a la vez sería un punto que no
	// dice por qué lo creemos.
	Patron string `json:"patron,omitempty"`
	// Calculo es la fórmula RESUELTA con los números de esta ventana — lo que despliega «ver
	// el cálculo» (H-4). Es lo que hace auditable a mano el número de arriba.
	Calculo string `json:"calculo,omitempty"`
	// Sesgo se declara y va EN CONTRA de la recomendación (A3): si el sesgo favoreciera
	// al fix que proponemos, la recomendación se estaría auto-justificando.
	Sesgo string `json:"sesgo"`
	// DireccionSesgo es "subestima" | "sobreestima". Nunca vacío ni neutro (RF-252).
	DireccionSesgo string    `json:"direccion_sesgo"`
	Fix            string    `json:"fix"` // UNO, concreto
	FixCodigo      string    `json:"fix_codigo,omitempty"`
	Confianza      Confianza `json:"confianza"`
	// ConfianzaDetalle es el desglose cuando la confianza NO es exacta. Hoy viaja vacío a
	// propósito: el punto conserva la PEOR confianza de la ventana, no el reparto entre
	// exactas y por-huella, y redactar «11 de 14 exactas, 3 por huella» sin ese reparto sería
	// inventarlo. Vacío ⇒ la tarjeta cae a su fórmula honesta con las corridas atribuidas.
	ConfianzaDetalle string `json:"confianza_detalle,omitempty"`
	CorridasUsadas   int    `json:"corridas_usadas"`
	CorridasTotales  int    `json:"corridas_totales"`
	Grave            bool   `json:"grave,omitempty"`
	// SoloS1 marca al detector que solo puede correr con la telemetría de ArnesIA (hoy B1).
	// Sin el chip, en una instalación mayormente S2 la tarjeta insignia desaparecería sin
	// explicación (J-2).
	SoloS1 bool `json:"solo_s1,omitempty"`
}

// Ventana es el corte de datos sobre el que corre un detector: los turnos unidos
// (dinero×proceso) de un arnés en un rango, YA costeados. Un detector no consulta el
// almacén: recibe su ventana y calcula.
type Ventana struct {
	Desde       time.Time
	Hasta       time.Time
	ArnesID     string
	Turnos      []TurnoUnido
	TotalMicros int64
	Contexto    ContextoDeteccion
	// Precios es el catálogo aplicable a los modelos de ESTA ventana, por nombre canónico.
	//
	// Existe por D26.1: B1 y B3 razonan sobre buckets de cache, y antes **sumaban conteos de
	// tokens en campos `micros`** (defecto M2). Un detector no puede consultar el catálogo
	// —no conoce puertos—, así que el servicio le pasa los precios que su ventana necesita y
	// el detector cotiza con `CalcularCosto`, que ya sabe cobrar los dos tramos del cache por
	// separado y **no cobra a cero un bucket sin tarifa**.
	//
	// Un modelo ausente del mapa NO es gratis: es un modelo que no se pudo cotizar, y lo que
	// no se cotiza queda afuera del número y dentro del sesgo declarado.
	Precios map[string]PrecioModelo
}

// PrecioDe devuelve el precio del modelo de un turno. El segundo valor en false significa
// «no se pudo cotizar», jamás «costó cero».
func (v Ventana) PrecioDe(modelo string) (PrecioModelo, bool) {
	if modelo == "" || v.Precios == nil {
		return PrecioModelo{}, false
	}
	p, ok := v.Precios[modelo]
	return p, ok
}

// Detector es el contrato común de los seis del MVP (D16.1).
type Detector interface {
	ID() DetectorID
	Nombre() string
	Aplica(c ContextoDeteccion) Aplicabilidad
	Evaluar(v Ventana) []PuntoDeMejora
}

// ── Los SEIS detectores del MVP (D16.1) ──────────────────────────────────────────────────
//
// Cada uno cumple la regla A4: número · contrafactual · umbral citado · sesgo declarado EN
// CONTRA · UN fix. Sin las cinco cosas el punto no se muestra — una tarjeta sin contrafactual
// es un reproche, no una recomendación.
//
// `ScoreVersionMVP` versiona la FÓRMULA. Cambiar cómo se calcula un punto obliga a bumpearlo
// (A7/RF-254): sin eso, dos cifras calculadas con reglas distintas se leerían como
// comparables.
//
// **v2 (D26.1, 2026-07-27)** — dos correcciones que cambian los números de B1 y B3:
//
//  1. Los buckets de cache se **cotizan con el catálogo**. Antes se sumaban conteos de tokens
//     en campos `micros` (defecto M2): la cifra no era dinero, y se mostraba como si lo fuera.
//  2. El contrafactual de B1 pasa a modelar **las re-escrituras que el vencimiento largo
//     evita** (`W` escrituras cortas contra UNA larga). Antes comparaba las mismas escrituras
//     a dos tarifas y la larga es más cara, así que el «ahorro» salía negativo siempre.
const ScoreVersionMVP = 2

// DetectoresMVP devuelve los seis en orden estable. Es la lista COMPLETA: los otros siete de
// la familia B se reportan como «no medidos todavía» por el servicio, **no por un detector
// vacío que devuelve cero**.
func DetectoresMVP() []Detector {
	crudos := []Detector{detB4{}, detP1{}, detB2{}, detB6{}, detB3{}, detB1{}}
	out := make([]Detector, len(crudos))
	for i, d := range crudos {
		out[i] = redactado{d}
	}
	return out
}

// redactado envuelve a cada detector para que **ningún punto salga de acá sin su prosa** (D25).
// Envolver la lista y no llamar a `Redactar()` en el servicio es deliberado: el servicio no es el
// único consumidor —los tests de dominio evalúan detectores directo—, y un punto a medio redactar
// que se cuela por un camino que nadie miró es cómo nació el hueco V-5.
type redactado struct{ Detector }

func (r redactado) Evaluar(v Ventana) []PuntoDeMejora {
	puntos := r.Detector.Evaluar(v)
	for i := range puntos {
		puntos[i] = puntos[i].Redactar()
	}
	return puntos
}

// ── B4 · gasto por arnés × empresa × puesto ──────────────────────────────────────────────

type detB4 struct{}

func (detB4) ID() DetectorID { return DetB4 }
func (detB4) Nombre() string { return "gasto por arnés × empresa × puesto" }

func (detB4) Aplica(c ContextoDeteccion) Aplicabilidad {
	if !c.TieneCosto {
		return Aplicabilidad{Motivo: "este arnés no reporta costo: no lleva el bloque de telemetría, " +
			"así que se ve el proceso pero no el dinero"}
	}
	return Aplicabilidad{Aplica: true}
}

// Evaluar señala el gasto concentrado. **Objeción registrada (plan §9.2): B4 es el único de
// los seis cuya recomendación es «dónde mirar» y no un ajuste concreto.** Entra porque D16.1
// lo firmó; si en uso resulta que no pasa su propio filtro, el lugar honesto es la franja
// (el total con su cobertura), no una tarjeta.
func (detB4) Evaluar(v Ventana) []PuntoDeMejora {
	if v.TotalMicros <= 0 {
		return nil
	}
	porCaja := map[string]int64{}
	corridas := map[string]int{}
	for _, t := range v.Turnos {
		if t.Atribucion == ConfianzaSinDato || t.CostoReportadoMicros == nil {
			continue // A15: lo no atribuido no entra a ningún desglose.
		}
		porCaja[t.CajaID] += *t.CostoReportadoMicros
		corridas[t.CajaID]++
	}
	var out []PuntoDeMejora
	for caja, micros := range porCaja {
		parte := float64(micros) / float64(v.TotalMicros)
		// Umbral CITADO, no mágico: una caja concentra cuando se lleva más de un tercio.
		if parte < 0.33 {
			continue
		}
		out = append(out, PuntoDeMejora{
			Detector: DetB4, ScoreVersion: ScoreVersionMVP,
			Titulo: "una sola caja se lleva la mayor parte del gasto",
			Lede:   "el gasto de esta ventana está concentrado; conviene mirar esa caja antes que el resto",
			CajaID: caja, GastoMicros: micros, ParteDelTotal: parte,
			// El contrafactual de B4 es el mundo donde esa caja gasta lo que gasta la
			// mediana de las demás — no «cero», que sería fantasía.
			ContrafactualMicros: v.TotalMicros - micros + medianaSalvo(porCaja, caja),
			DiferenciaMicros:    micros - medianaSalvo(porCaja, caja),
			BaseContrafactual:   "Si esta caja gastara lo que gasta una caja típica de este arnés",
			Umbral:              "parte del total > 1/3 (0,33)",
			Calculo: fmt.Sprintf("%s de %s = %.1f %% del total · mediana de las otras cajas: %s",
				UsdMicros(micros), UsdMicros(v.TotalMicros), parte*100,
				UsdMicros(medianaSalvo(porCaja, caja))),
			Sesgo: "solo entran los turnos atribuidos: si parte del gasto no se pudo atribuir, " +
				"la concentración medida es MENOR que la real",
			DireccionSesgo: "subestima",
			Fix:            "revisar el contenido de esta caja antes que el de las demás",
			Confianza:      peorDeLaVentana(v),
			CorridasUsadas: corridas[caja], CorridasTotales: len(v.Turnos),
		})
	}
	return out
}

// ── P1 · caja que consume y se rechaza ───────────────────────────────────────────────────

type detP1 struct{}

func (detP1) ID() DetectorID { return DetP1 }
func (detP1) Nombre() string { return "caja que consume y se rechaza" }

func (detP1) Aplica(c ContextoDeteccion) Aplicabilidad {
	if !c.TieneSenalProceso {
		return Aplicabilidad{Motivo: "este arnés no porta el hook de proceso — no hay veredicto de gate que leer"}
	}
	// ⚠️ El matiz de H8: fuera de ArnesIA llega parte de la señal de proceso por telemetría
	// (decisión de permiso, éxito y duración por herramienta) pero NO el veredicto del gate
	// humano ni la corrida de caja, que son eventos del daemon. Corre PARCIAL, y eso viaja
	// como matiz explícito — jamás como un visto bueno liso.
	if !c.TieneGateHumano {
		return Aplicabilidad{
			Aplica: true, Parcial: true,
			Motivo: "cobertura parcial: se ven reintentos y fracasos de herramienta, no rechazos de gate " +
				"(el veredicto del gate es un evento del daemon y esta corrida no lo tuvo)",
		}
	}
	return Aplicabilidad{Aplica: true}
}

func (detP1) Evaluar(v Ventana) []PuntoDeMejora {
	desperdicio := map[string]int64{}
	rechazos := map[string]int{}
	intentos := map[string]int{}
	for _, t := range v.Turnos {
		if t.Atribucion == ConfianzaSinDato {
			continue
		}
		intentos[t.CajaID]++
		fallo := false
		for _, r := range t.Resultados {
			if r == ResultadoRechazado || r == ResultadoReintento {
				fallo = true
			}
		}
		if !fallo {
			continue
		}
		rechazos[t.CajaID]++
		if t.CostoReportadoMicros != nil {
			desperdicio[t.CajaID] += *t.CostoReportadoMicros
		}
	}
	var out []PuntoDeMejora
	for caja, micros := range desperdicio {
		if rechazos[caja] == 0 {
			continue
		}
		out = append(out, PuntoDeMejora{
			Detector: DetP1, ScoreVersion: ScoreVersionMVP,
			Titulo: "una caja consume y después se rechaza",
			Lede:   "lo que se gastó en las corridas rechazadas de esta caja no produjo nada",
			CajaID: caja, GastoMicros: micros,
			ParteDelTotal: parteDe(micros, v.TotalMicros),
			// Contrafactual: el mundo donde esas corridas no se hubieran hecho. Es el ÚNICO
			// caso de los seis donde el contrafactual es cero y no una alternativa, porque
			// una corrida rechazada literalmente no entregó nada.
			ContrafactualMicros: 0,
			DiferenciaMicros:    micros,
			BaseContrafactual:   "Si el gate pasara a la primera",
			// RF-250 — **P1 no lleva `Umbral`, lleva `Patron`.** No se decide por una
			// desigualdad: se decide porque hubo rechazos. Pintarle una desigualdad sería
			// fabricarle un rigor algebraico que su regla no tiene.
			Patron: fmt.Sprintf("%d de %d corridas atribuidas de esta caja terminaron en rechazo "+
				"o reintento", rechazos[caja], intentos[caja]),
			Calculo: fmt.Sprintf("costo de las corridas rechazadas: %s · corridas rechazadas: %d de %d",
				UsdMicros(micros), rechazos[caja], intentos[caja]),
			Sesgo: "solo cuenta el costo de los turnos rechazados que se pudieron atribuir a esta caja; " +
				"lo no atribuido queda afuera, así que el desperdicio real es MAYOR",
			DireccionSesgo: "subestima",
			Fix:            "endurecer el contrato de esta caja antes de correrla, no después",
			Confianza:      peorDeLaVentana(v),
			CorridasUsadas: rechazos[caja], CorridasTotales: intentos[caja],
			Grave: rechazos[caja] > 2,
		})
	}
	return out
}

// ── B2 · costo de la rotación de contexto ────────────────────────────────────────────────

type detB2 struct{}

func (detB2) ID() DetectorID { return DetB2 }
func (detB2) Nombre() string { return "costo de la rotación de contexto" }

func (detB2) Aplica(c ContextoDeteccion) Aplicabilidad {
	if !c.TieneEventoRotacion {
		return Aplicabilidad{Motivo: "la rotación es una decisión de ArnesIA: corriendo fuera, " +
			"no hay rotaciones que medir"}
	}
	if !c.TieneCosto {
		return Aplicabilidad{Motivo: "hubo rotaciones pero no hay señal de dinero para cotizarlas"}
	}
	return Aplicabilidad{Aplica: true}
}

func (detB2) Evaluar(v Ventana) []PuntoDeMejora {
	var micros int64
	var rotaciones, usados int
	for _, t := range v.Turnos {
		if t.Rotaciones == 0 || t.Atribucion == ConfianzaSinDato {
			continue
		}
		rotaciones += t.Rotaciones
		usados++
		if t.CostoReportadoMicros != nil {
			micros += *t.CostoReportadoMicros
		}
	}
	if rotaciones == 0 {
		return nil
	}
	return []PuntoDeMejora{{
		Detector: DetB2, ScoreVersion: ScoreVersionMVP,
		Titulo:      "rotar el contexto cuesta un turno completo cada vez",
		Lede:        "cada rotación re-carga el contexto: los turnos que rotaron pagaron esa recarga",
		GastoMicros: micros, ParteDelTotal: parteDe(micros, v.TotalMicros),
		// El contrafactual es el turno SIN la recarga: se estima como el costo del turno
		// menos la parte de entrada+cache que la rotación volvió a pagar.
		ContrafactualMicros: micros / 2,
		DiferenciaMicros:    micros - micros/2,
		BaseContrafactual:   "Sin la recarga de contexto que impone cada rotación",
		Umbral:              "≥ 1 rotación registrada por el daemon en la ventana",
		Calculo: fmt.Sprintf("%d rotaciones en %d turnos con costo atribuido · %s gastados, "+
			"la mitad imputada a la recarga", rotaciones, usados, UsdMicros(micros)),
		Sesgo: "el contrafactual asume que la mitad del turno rotado fue recarga; si fue menos, " +
			"el ahorro estimado es MAYOR que el real",
		DireccionSesgo: "sobreestima",
		Fix:            "subir el umbral de rotación para que corte más tarde",
		Confianza:      peorDeLaVentana(v),
		CorridasUsadas: usados, CorridasTotales: len(v.Turnos),
	}}
}

// ── B6 · sesión abandonada ───────────────────────────────────────────────────────────────

type detB6 struct{}

func (detB6) ID() DetectorID { return DetB6 }
func (detB6) Nombre() string { return "sesión abandonada" }

func (detB6) Aplica(c ContextoDeteccion) Aplicabilidad {
	if !c.TieneCosto {
		return Aplicabilidad{Motivo: "sin señal de dinero no se puede decir que una sesión se desperdició"}
	}
	return Aplicabilidad{Aplica: true}
}

func (detB6) Evaluar(v Ventana) []PuntoDeMejora {
	// Una sesión abandonada es la que gastó y NUNCA cerró un turno: hubo dinero y no hubo
	// señal de proceso que diga que terminó.
	porSesion := map[string]int64{}
	cerro := map[string]bool{}
	for _, t := range v.Turnos {
		if t.Atribucion == ConfianzaSinDato {
			continue
		}
		if t.CostoReportadoMicros != nil {
			porSesion[t.SesionID] += *t.CostoReportadoMicros
		}
		if t.TieneProceso {
			cerro[t.SesionID] = true
		}
	}
	var micros int64
	var abandonadas int
	for sesion, m := range porSesion {
		if cerro[sesion] || m == 0 {
			continue
		}
		micros += m
		abandonadas++
	}
	if abandonadas == 0 {
		return nil
	}
	return []PuntoDeMejora{{
		Detector: DetB6, ScoreVersion: ScoreVersionMVP,
		Titulo:      "hay sesiones que gastaron y nunca cerraron",
		Lede:        "esas sesiones consumieron y no dejaron señal de haber terminado un turno",
		GastoMicros: micros, ParteDelTotal: parteDe(micros, v.TotalMicros),
		ContrafactualMicros: 0,
		DiferenciaMicros:    micros,
		BaseContrafactual:   "Si esas sesiones se hubieran cerrado en vez de quedar colgadas",
		Umbral:              "sesión con costo > 0 y sin ningún evento de proceso que la cierre",
		Calculo: fmt.Sprintf("%d de %d sesiones de la ventana gastaron y no cerraron ningún turno · %s",
			abandonadas, len(porSesion), UsdMicros(micros)),
		Sesgo: "una sesión puede haber cerrado sin que el hook lo reportara; en ese caso esto " +
			"cuenta de más y el desperdicio real es MENOR",
		DireccionSesgo: "sobreestima",
		Fix:            "cerrar la sesión desde el Dock en vez de dejarla colgada",
		Confianza:      peorDeLaVentana(v),
		CorridasUsadas: abandonadas, CorridasTotales: len(porSesion),
	}}
}

// ── B3 · cambio de modelo que invalida el cache ──────────────────────────────────────────

type detB3 struct{}

func (detB3) ID() DetectorID { return DetB3 }
func (detB3) Nombre() string { return "cambio de modelo invalida el cache" }

func (detB3) Aplica(c ContextoDeteccion) Aplicabilidad {
	if !c.TieneCosto {
		return Aplicabilidad{Motivo: "sin señal de dinero no se puede cotizar el cache perdido"}
	}
	// D26.1 — sin catálogo el número saldría en TOKENS y se mostraría como dinero (defecto
	// M2). Antes que una cifra en la unidad equivocada, se declara que no se puede cotizar.
	if !c.CatalogoDisponible {
		return Aplicabilidad{Motivo: "sin catálogo de precios no se puede poner un monto a la " +
			"re-escritura de cache: el número saldría en tokens, no en dinero"}
	}
	if c.ModelosDistintos < 2 {
		return Aplicabilidad{Motivo: "un solo modelo en la ventana: no hubo cambio que invalidara el cache"}
	}
	return Aplicabilidad{Aplica: true}
}

func (detB3) Evaluar(v Ventana) []PuntoDeMejora {
	// Cada cambio de modelo obliga a re-escribir el cache: el bucket de escritura del primer
	// turno con el modelo nuevo es lo que se pagó de más.
	//
	// D26.1 — ese bucket se **COTIZA**, no se cuenta. Antes se sumaban tokens en un campo
	// `micros` y la tarjeta mostraba el conteo como si fuera dinero.
	var micros int64
	var cambios, sinCotizar int
	anterior := ""
	for _, t := range v.Turnos {
		if t.Modelo == "" || t.Atribucion == ConfianzaSinDato {
			continue
		}
		if anterior != "" && t.Modelo != anterior {
			cambios++
			m, ok := costoDelCacheEscrito(v, t)
			if !ok {
				// El modelo no está en el catálogo. **No entra como cero**: queda afuera del
				// número y adentro del sesgo.
				sinCotizar++
			} else {
				micros += m
			}
		}
		anterior = t.Modelo
	}
	if cambios == 0 || micros <= 0 {
		return nil
	}
	return []PuntoDeMejora{{
		Detector: DetB3, ScoreVersion: ScoreVersionMVP,
		Titulo:      "cambiar de modelo a mitad tira el cache",
		Lede:        "cada cambio de modelo obliga a re-escribir el contexto que ya estaba cacheado",
		GastoMicros: micros, ParteDelTotal: parteDe(micros, v.TotalMicros),
		ContrafactualMicros: 0,
		DiferenciaMicros:    micros,
		BaseContrafactual:   "Fijando un modelo por frente de trabajo",
		Umbral:              "≥ 2 modelos distintos en la ventana, con re-escritura de cache tras el cambio",
		Calculo: fmt.Sprintf("%d cambios de modelo en %d turnos · re-escritura de cache tras el "+
			"cambio, cotizada con el catálogo: %s%s", cambios, len(v.Turnos), UsdMicros(micros),
			sufijoSinCotizar(sinCotizar)),
		// A3 — el sesgo va EN CONTRA de la recomendación. Se cuenta como desperdicio TODA la
		// re-escritura posterior al cambio; parte de ella se habría pagado igual, así que el
		// desperdicio real es MENOR. Lo no cotizado tira para el otro lado y se dice, pero la
		// dirección que se declara es la que nos incomoda.
		Sesgo: "se imputa al cambio de modelo toda la re-escritura de cache posterior; parte se " +
			"habría pagado igual, así que el desperdicio real es MENOR" + fraseSinCotizar(sinCotizar),
		DireccionSesgo: "sobreestima",
		Fix:            "fijar un modelo por frente de trabajo en vez de alternar",
		Confianza:      peorDeLaVentana(v),
		CorridasUsadas: cambios, CorridasTotales: len(v.Turnos),
	}}
}

// ── B1 · re-warm por TTL ─────────────────────────────────────────────────────────────────

type detB1 struct{}

func (detB1) ID() DetectorID { return DetB1 }
func (detB1) Nombre() string { return "re-warm por TTL del cache" }

// Aplica pide `TieneSplitTTL`, que es un hecho de la VENTANA y no del escenario. Esa
// diferencia es deliberada: el día que el split llegue por otro canal, el detector se
// enciende SOLO, sin tocar código.
//
// Y el motivo distingue los dos casos, que se confundían en el diseño previo: fuera de
// ArnesIA con telemetría encendida **sí hay dinero** — lo que falta es el split, no la
// telemetría.
func (detB1) Aplica(c ContextoDeteccion) Aplicabilidad {
	if c.Runtime != "claude-code" {
		return Aplicabilidad{Motivo: "el break-even de TTL depende de la estructura de precios de " +
			"este proveedor; todavía no medimos ese runtime"}
	}
	if !c.TieneSplitTTL {
		return Aplicabilidad{Motivo: "falta el desglose del cache por vencimiento: viaja en el cierre " +
			"de turno del subproceso, que solo se ve cuando ArnesIA es el proceso padre"}
	}
	// D26.1 — misma regla que B3: sin catálogo, el break-even es álgebra sin números.
	if !c.CatalogoDisponible {
		return Aplicabilidad{Motivo: "sin catálogo de precios no se puede cotizar la escritura de " +
			"cache: el break-even es álgebra, pero el ahorro es dinero"}
	}
	return Aplicabilidad{Aplica: true}
}

func (detB1) Evaluar(v Ventana) []PuntoDeMejora {
	// El break-even es ÁLGEBRA, no medición: escribir a 1 h cuesta 2× la entrada y a 5 min
	// cuesta 1,25×; leer de cache cuesta 0,1×. Conviene el TTL largo cuando la fracción de
	// re-warms evitados supera (2 − 1,25) / (2 − 0,1) = 39,47 %.
	//
	// **v2 (D26.1)** — el contrafactual dejó de ser «las mismas escrituras a la otra tarifa»,
	// que era más caro por construcción y daba un ahorro negativo siempre. El ahorro de B1 son
	// las **re-escrituras que el vencimiento largo EVITA**: si el contexto se re-escribió `W`
	// veces a 5 min, con 1 h se habría escrito **una sola vez**. Por eso `W ≥ 2` es la
	// condición de existencia del punto — con una sola escritura no hay nada que evitar.
	//
	// Y el monto **se cotiza con el catálogo**: antes se sumaban tokens en un campo `micros`
	// (defecto M2) y la tarjeta mostraba ese conteo como si fuera dinero.
	porModelo := map[string]int64{} // tokens escritos a 5 min, por modelo canónico
	escrituras := map[string]int{}  // W: cuántas veces se re-escribió, por modelo
	var usados, sinCotizar int
	for _, t := range v.Turnos {
		if t.Atribucion == ConfianzaSinDato {
			continue
		}
		usados++
		tok := valorTok(t.Tokens.CacheEscritura5m)
		if tok <= 0 {
			continue
		}
		porModelo[t.Modelo] += tok
		escrituras[t.Modelo]++
	}

	var hoy, conTTLLargo int64
	var reWarms int
	for modelo, tokens := range porModelo {
		w := escrituras[modelo]
		if w < 2 {
			// Una sola escritura no tiene re-warm que evitar. No es un hallazgo chico: no es
			// un hallazgo.
			continue
		}
		precio, ok := v.PrecioDe(modelo)
		if !ok {
			// Modelo fuera del catálogo. **No entra como cero**: queda afuera del número y
			// dentro del sesgo declarado.
			sinCotizar++
			continue
		}
		arit := aritmeticaDe(v, modelo)
		actual := CalcularCosto(Tokens{CacheEscritura5m: &tokens}, precio, arit)
		unaSola := tokens / int64(w)
		alterno := CalcularCosto(Tokens{CacheEscritura1h: &unaSola}, precio, arit)
		if actual.SinNingunaTarifa || alterno.SinNingunaTarifa {
			sinCotizar++
			continue
		}
		hoy += actual.Micros
		conTTLLargo += alterno.Micros
		reWarms += w - 1
	}
	if hoy <= 0 || hoy-conTTLLargo <= 0 {
		// Sin escrituras repetidas cotizables no hay ahorro que proponer. El motor lo declara
		// como detector sin fix, no como un cero.
		return nil
	}

	return []PuntoDeMejora{{
		Detector: DetB1, ScoreVersion: ScoreVersionMVP,
		Titulo:      "el cache se re-escribe porque vence antes de reusarse",
		Lede:        "se está pagando la escritura corta del cache más veces de las necesarias",
		GastoMicros: hoy, ParteDelTotal: parteDe(hoy, v.TotalMicros),
		ContrafactualMicros: conTTLLargo,
		DiferenciaMicros:    hoy - conTTLLargo,
		BaseContrafactual:   "Con el vencimiento del cache en 1 h",
		// El umbral CITADO ENTERO, que es lo que pide RF-250: es álgebra, no un número
		// que salió de esta corrida.
		Umbral: "(2 − 1,25) / (2 − 0,1) = 39,47 % — conviene el vencimiento largo cuando evita " +
			"más de ese porcentaje de re-escrituras",
		Calculo: fmt.Sprintf("%d re-escrituras evitables · pagado a 5 min: %s · una sola "+
			"escritura a 1 h: %s%s", reWarms, UsdMicros(hoy), UsdMicros(conTTLLargo),
			sufijoSinCotizar(sinCotizar)),
		// J-2 — B1 exige el desglose por vencimiento, que solo llega cuando ArnesIA es el
		// proceso padre. Sin el chip, en una instalación mayormente S2 la tarjeta insignia
		// desaparecería sin decir por qué.
		SoloS1: true,
		// A3 — el sesgo va EN CONTRA: se asume que TODAS las re-escrituras se habrían evitado
		// con el vencimiento largo. Si algunas caían igual fuera de la hora, el ahorro real es
		// MENOR y el cambio puede no convenir.
		Sesgo: "se asume que el vencimiento largo evita todas las re-escrituras medidas; las que " +
			"cayeran igual fuera de la hora no se ahorran, así que el ahorro real es MENOR" +
			fraseSinCotizar(sinCotizar),
		DireccionSesgo: "sobreestima",
		// El ajuste concreto lleva su token de código aparte para que la tarjeta lo pinte en
		// `<code>`: es lo que se copia y se pega, no prosa. `design.md` §7.4 dice «en esta
		// caja»; acá dice «en este arnés» porque **B1 agrega la ventana entera y no tiene
		// caja** — desviación declarada, no un descuido de copy.
		Fix:            "fijar cache_ttl: 1h en este arnés",
		FixCodigo:      "cache_ttl: 1h",
		Confianza:      peorDeLaVentana(v),
		CorridasUsadas: usados, CorridasTotales: len(v.Turnos),
	}}
}

// ── helpers compartidos ──────────────────────────────────────────────────────────────────

// peorDeLaVentana devuelve la MÍNIMA confianza de los turnos usados. Nunca la máxima ni la
// moda: un punto de mejora calculado sobre datos mezclados vale lo que su peor parte.
func peorDeLaVentana(v Ventana) Confianza {
	peor := Confianza("")
	for _, t := range v.Turnos {
		if peor == "" {
			peor = t.Atribucion
			continue
		}
		peor = PeorConfianza(peor, t.Atribucion)
	}
	if peor == "" {
		return ConfianzaSinDato
	}
	return peor
}

func parteDe(micros, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(micros) / float64(total)
}

// medianaSalvo devuelve la mediana de los valores del mapa excluyendo una clave. Se usa como
// contrafactual de B4: «qué pasaría si esta caja gastara lo que gasta una caja típica».
func medianaSalvo(m map[string]int64, salvo string) int64 {
	var xs []int64
	for k, v := range m {
		if k == salvo {
			continue
		}
		xs = append(xs, v)
	}
	if len(xs) == 0 {
		return 0
	}
	for i := 1; i < len(xs); i++ {
		for j := i; j > 0 && xs[j] < xs[j-1]; j-- {
			xs[j], xs[j-1] = xs[j-1], xs[j]
		}
	}
	return xs[len(xs)/2]
}

// ── helpers de costeo (D26.1) ────────────────────────────────────────────────────────────
//
// B1 y B3 razonan sobre buckets de cache. Antes sumaban **tokens** en campos `micros`; ahora
// cotizan con el catálogo de la ventana. Lo que no se puede cotizar **no vale cero**: queda
// afuera del número y se declara en el sesgo.

func valorTok(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

// costoDelCacheEscrito cotiza los DOS tramos de escritura de cache de un turno. El segundo
// valor en false significa «no se pudo cotizar» — nunca «costó cero».
func costoDelCacheEscrito(v Ventana, t TurnoUnido) (int64, bool) {
	e5, e1 := valorTok(t.Tokens.CacheEscritura5m), valorTok(t.Tokens.CacheEscritura1h)
	if e5 == 0 && e1 == 0 {
		return 0, true // no hubo escritura: es un cero legítimo, no una ausencia.
	}
	precio, ok := v.PrecioDe(t.Modelo)
	if !ok {
		return 0, false
	}
	c := CalcularCosto(Tokens{CacheEscritura5m: &e5, CacheEscritura1h: &e1}, precio, t.Aritmetica)
	if c.SinNingunaTarifa {
		return 0, false
	}
	return c.Micros, true
}

// aritmeticaDe devuelve la aritmética de los turnos de un modelo. Es la del turno, no una
// constante: `cache_lectura ⊂ entrada` depende del emisor, y asumirla duplica el conteo
// (langfuse#12306).
func aritmeticaDe(v Ventana, modelo string) Aritmetica {
	for _, t := range v.Turnos {
		if t.Modelo == modelo && t.Aritmetica != "" {
			return t.Aritmetica
		}
	}
	return AritmeticaDisjunta
}

// sufijoSinCotizar y fraseSinCotizar dicen, en el cálculo y en el sesgo, cuántos modelos
// quedaron afuera por no estar en el catálogo. Callarlo convertiría una cota inferior en un
// total.
func sufijoSinCotizar(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf(" · %d modelo(s) sin tarifa quedaron afuera del monto", n)
}

func fraseSinCotizar(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("; además, %d modelo(s) de la ventana no están en el catálogo y no se "+
		"cotizaron, así que el monto es una cota inferior", n)
}
