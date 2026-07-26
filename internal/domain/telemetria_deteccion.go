package domain

import "time"

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
	Detector     DetectorID `json:"detector"`
	ScoreVersion int        `json:"score_version"` // A7: cambiar la fórmula obliga a bumpearlo
	Titulo       string     `json:"titulo"`
	Lede         string     `json:"lede,omitempty"`
	CajaID       string     `json:"caja_id,omitempty"`
	GastoMicros  int64      `json:"gasto_micros"`
	// ParteDelTotal va en 0..1.
	ParteDelTotal float64 `json:"parte_del_total"`
	// ContrafactualMicros es qué habría costado el mundo alternativo, con las MISMAS
	// corridas (A2). No es «gastaste X»: es «con el cambio habrías gastado Y».
	ContrafactualMicros int64 `json:"contrafactual_micros"`
	DiferenciaMicros    int64 `json:"diferencia_micros"`
	// Umbral es la desigualdad algebraica CITADA (A1), no un número mágico.
	Umbral string `json:"umbral"`
	// Sesgo se declara y va EN CONTRA de la recomendación (A3): si el sesgo favoreciera
	// al fix que proponemos, la recomendación se estaría auto-justificando.
	Sesgo string `json:"sesgo"`
	// DireccionSesgo es "subestima" | "sobreestima". Nunca vacío ni neutro (RF-252).
	DireccionSesgo  string    `json:"direccion_sesgo"`
	Fix             string    `json:"fix"` // UNO, concreto
	Confianza       Confianza `json:"confianza"`
	CorridasUsadas  int       `json:"corridas_usadas"`
	CorridasTotales int       `json:"corridas_totales"`
	Grave           bool      `json:"grave,omitempty"`
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
const ScoreVersionMVP = 1

// DetectoresMVP devuelve los seis en orden estable. Es la lista COMPLETA: los otros siete de
// la familia B se reportan como «no medidos todavía» por el servicio, **no por un detector
// vacío que devuelve cero**.
func DetectoresMVP() []Detector {
	return []Detector{detB4{}, detP1{}, detB2{}, detB6{}, detB3{}, detB1{}}
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
			Umbral:              "parte del total > 1/3 (0,33)",
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
		return Aplicabilidad{Aplica: true, Parcial: true,
			Motivo: "cobertura parcial: se ven reintentos y fracasos de herramienta, no rechazos de gate " +
				"(el veredicto del gate es un evento del daemon y esta corrida no lo tuvo)"}
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
			Umbral:              "≥ 1 corrida rechazada o reintentada con costo atribuido",
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
		Umbral:              "≥ 1 rotación registrada por el daemon en la ventana",
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
		Umbral:              "sesión con costo > 0 y sin ningún evento de proceso que la cierre",
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
	if c.ModelosDistintos < 2 {
		return Aplicabilidad{Motivo: "un solo modelo en la ventana: no hubo cambio que invalidara el cache"}
	}
	return Aplicabilidad{Aplica: true}
}

func (detB3) Evaluar(v Ventana) []PuntoDeMejora {
	// Cada cambio de modelo obliga a re-escribir el cache: el bucket de escritura del primer
	// turno con el modelo nuevo es lo que se pagó de más.
	var micros int64
	var cambios int
	anterior := ""
	for _, t := range v.Turnos {
		if t.Modelo == "" || t.Atribucion == ConfianzaSinDato {
			continue
		}
		if anterior != "" && t.Modelo != anterior {
			cambios++
			if t.Tokens.CacheEscritura5m != nil {
				micros += *t.Tokens.CacheEscritura5m
			}
			if t.Tokens.CacheEscritura1h != nil {
				micros += *t.Tokens.CacheEscritura1h
			}
		}
		anterior = t.Modelo
	}
	if cambios == 0 {
		return nil
	}
	return []PuntoDeMejora{{
		Detector: DetB3, ScoreVersion: ScoreVersionMVP,
		Titulo:      "cambiar de modelo a mitad tira el cache",
		Lede:        "cada cambio de modelo obliga a re-escribir el contexto que ya estaba cacheado",
		GastoMicros: micros, ParteDelTotal: parteDe(micros, v.TotalMicros),
		ContrafactualMicros: 0,
		DiferenciaMicros:    micros,
		Umbral:              "≥ 2 modelos distintos en la ventana, con re-escritura de cache tras el cambio",
		Sesgo: "se cuentan tokens de re-escritura, no su costo exacto por modelo; si el modelo nuevo " +
			"es más barato, el desperdicio real es MENOR",
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
	return Aplicabilidad{Aplica: true}
}

func (detB1) Evaluar(v Ventana) []PuntoDeMejora {
	// El break-even es ÁLGEBRA, no medición: escribir a 1 h cuesta 2× la entrada y a 5 min
	// cuesta 1,25×; leer de cache cuesta 0,1×. Conviene el TTL largo cuando la fracción de
	// re-warms evitados supera (2 − 1,25) / (2 − 0,1) = 39,47 %.
	var escrito5m, escrito1h int64
	var usados int
	for _, t := range v.Turnos {
		if t.Atribucion == ConfianzaSinDato {
			continue
		}
		if t.Tokens.CacheEscritura5m != nil {
			escrito5m += *t.Tokens.CacheEscritura5m
		}
		if t.Tokens.CacheEscritura1h != nil {
			escrito1h += *t.Tokens.CacheEscritura1h
		}
		usados++
	}
	if escrito5m == 0 {
		return nil // sin escrituras de 5 min no hay re-warm que evitar.
	}
	// Lo que se paga hoy por esas escrituras cortas, y lo que costarían a TTL largo.
	const tarifa5m, tarifa1h = 1.25, 2.0
	hoy := int64(float64(escrito5m) * tarifa5m)
	conTTLLargo := int64(float64(escrito5m) * tarifa1h)
	return []PuntoDeMejora{{
		Detector: DetB1, ScoreVersion: ScoreVersionMVP,
		Titulo:      "el cache se re-escribe porque vence antes de reusarse",
		Lede:        "se está pagando la escritura corta del cache más veces de las necesarias",
		GastoMicros: hoy, ParteDelTotal: parteDe(hoy, v.TotalMicros),
		ContrafactualMicros: conTTLLargo,
		DiferenciaMicros:    hoy - conTTLLargo,
		// El umbral CITADO ENTERO, que es lo que pide RF-250: es álgebra, no un número
		// que salió de esta corrida.
		Umbral: "(2 − 1,25) / (2 − 0,1) = 39,47 % — conviene el vencimiento largo cuando evita " +
			"más de ese porcentaje de re-escrituras",
		Sesgo: "se cuentan tokens escritos, no cuántos se habrían reusado de verdad; si se reusaran " +
			"menos que el umbral, el cambio costaría MÁS y esta recomendación estaría de más",
		DireccionSesgo: "sobreestima",
		Fix:            "subir el vencimiento del cache a una hora en este arnés",
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
