package domain

import "time"

// telemetria_vistas.go son las formas de LECTURA del módulo: lo que el HTTP y el CLI
// devuelven (arquitectura-modulo.md §3.2 y §8.2).
//
// Regla de contrato transversal, que un implementador «prolijo» rompería:
// **`null` y `0` significan cosas distintas en toda esta superficie.** Convertir un `null`
// en `0` es el pass fabricado que la doctrina prohíbe. Por eso los campos de dinero y de
// tokens son punteros y por eso las ausencias viajan explícitas, nunca omitidas.
// Enforced por TestNoAplicaNoEsCeroEnElWire.

// VersionCatalogoPrecios identifica con qué precios se costeó.
//
// ⚠️ Se llama `VersionCatalogoPrecios` y no `VersionCatalogo` como decía
// `arquitectura-modulo.md` §8.2: ese nombre YA está tomado en `internal/domain/marketplace.go`
// por una fila de `catalogo.json#versiones[]` del marketplace, que es otra cosa. Colisión de
// vocabulario detectada al construir (mockups/INDEX.md regla 4 aplicada al Go). `Refrescado` nil = NUNCA se
// refrescó, y se muestra como tal (§8.3) — no como «hace 0 h».
type VersionCatalogoPrecios struct {
	Version    string     `json:"version"` // "2026-07-20" — fecha del rev de LiteLLM
	Rev        string     `json:"rev"`     // sha del commit de origen
	SHA256     string     `json:"sha256"`  // del archivo embebido
	Modelos    int        `json:"modelos"`
	Refrescado *time.Time `json:"refrescado,omitempty"`
}

// Cobertura es la barra de §1.2 del mockup. `Esperados` sale de la conciliación (A9): los
// turnos que SABEMOS que ocurrieron. `NoLlegaron` viaja EXPLÍCITO, no se calcula por resta
// en el FE — obligarlo a restar es obligarlo a inventar cuando algo no cuadra.
//
// En S2 no hay denominador independiente: `Esperados` viaja nil y la UI dice «cobertura
// desconocida fuera de ArnesIA», que es la verdad.
type Cobertura struct {
	Esperados  *int `json:"esperados"`
	Exacta     int  `json:"exacta"`
	PorHash    int  `json:"por_hash"`
	PorProceso int  `json:"por_proceso"`
	SinDato    int  `json:"sin_dato"`
	NoLlegaron *int `json:"no_llegaron"` // esperados − medidos: el agujero WSL2/devcontainer
}

// ResumenTelemetria alimenta la franja de la barra del Mapa y `arnesia telemetria resumen`.
type ResumenTelemetria struct {
	Desde time.Time `json:"desde"`
	Hasta time.Time `json:"hasta"`
	// Estimado es SIEMPRE true mientras la fuente sea `cost_usd_micros`: está documentado
	// como "Estimated cost", no facturación (ANEXO H4). La UI lo dice en la superficie.
	Estimado             bool   `json:"estimado"`
	CostoReportadoMicros *int64 `json:"costo_reportado_micros"` // null = ningún turno lo trajo
	CostoCalculadoMicros *int64 `json:"costo_calculado_micros"` // null = sin catálogo aplicable
	// CostoCompleto dice si el costo CALCULADO cotizó todos los buckets que tenían tokens.
	// `false` significa que la cifra es una **cota inferior declarada**, no un total.
	//
	// Viaja hasta acá a propósito: un flag que existe y nadie ve no sirve de nada — y el
	// caso que lo dispara (escritura de cache sin tier declarado, que es lo único que el
	// canal OTLP puede dar) produce una divergencia del 33 % contra lo reportado.
	CostoCompleto *bool    `json:"costo_completo"`
	SinTarifa     []string `json:"sin_tarifa,omitempty"`
	// DivergenciaPct compara los dos costos, en porcentaje del reportado. `nil` cuando falta
	// alguno: sin dos números no hay divergencia que calcular, y un 0 diría «coinciden».
	//
	// DivergenciaSospechosa se enciende al pasar el umbral. **El oráculo de doble costo es lo
	// que caza los errores de costeo** —catálogo viejo, bucket perdido, tramo equivocado— que
	// ningún test de tabla ve venir, porque un test de tabla solo conoce los casos que su
	// autor imaginó. Esto es lo que hace que el sistema lo MIRE.
	DivergenciaPct        *float64 `json:"divergencia_pct"`
	DivergenciaSospechosa bool     `json:"divergencia_sospechosa"`
	Corridas              int      `json:"corridas"`
	Sesiones              int      `json:"sesiones"`
	Turnos                int      `json:"turnos"`
	// Cajas es cuántas cajas distintas tuvieron actividad atribuida en la ventana.
	//
	// 🔴 Lo destapó el candado de contrato al extenderse a este struct (D26.4): **el FE lo
	// tipaba y el Go nunca lo mandaba**, así que el denominador de la franja —«… · 4 cajas»,
	// que es parte del copy firmado— se pintaba desde un campo que nadie producía. Misma
	// clase exacta que V-5.
	Cajas     int                    `json:"cajas"`
	Escenario Escenario              `json:"escenario"`
	Confianza Confianza              `json:"confianza"`
	Cobertura Cobertura              `json:"cobertura"`
	Catalogo  VersionCatalogoPrecios `json:"catalogo"`
	// UltimaCorrida es la última medición del arnés **ignorando la ventana** (D26.4, estado
	// 1b). Sin este campo, «0 corridas en los últimos 7 días» y «este arnés nunca corrió» se
	// dicen igual — y sobre un arnés con dos años de historial la segunda es falsa.
	//
	// `nil` = nunca hubo una. Ese sí es el estado 1.
	UltimaCorrida *time.Time `json:"ultima_corrida"`
	// Runtime es con qué corre este arnés, tomado del dato REAL de la ventana (no de una
	// config). Vacío = no llegó ningún evento que lo diga.
	//
	// RuntimeSoportado dice si sabemos medirlo. `false` con `Runtime` no vacío es el estado
	// «todavía no medimos ese runtime», que se DICE en vez de mostrar un cero.
	Runtime          string `json:"runtime,omitempty"`
	RuntimeSoportado bool   `json:"runtime_soportado"`
}

// MarcaDeFuga NOMBRA el detector, nunca es un ⚠ genérico: un signo de admiración sin
// nombre obliga al usuario a adivinar qué le están señalando.
type MarcaDeFuga struct {
	Detector DetectorID `json:"detector"`
	Nombre   string     `json:"nombre"`
	Grave    bool       `json:"grave"`
}

// GastoCaja es una fila del desglose por caja. Las cajas SIN dato viajan igual, con
// `Atribuible:false` + `Motivo`: omitirlas obligaría al FE a inventar por qué faltan.
type GastoCaja struct {
	CajaID      string    `json:"caja_id"`
	Nombre      string    `json:"nombre"`
	Atribuible  bool      `json:"atribuible"`       // false ⇒ el FE pinta «sin dato atribuible»
	Motivo      string    `json:"motivo,omitempty"` // OBLIGATORIO si Atribuible=false
	CostoMicros *int64    `json:"costo_micros"`     // null, no 0, cuando no es atribuible
	Parte       *float64  `json:"parte,omitempty"`
	Confianza   Confianza `json:"confianza"`
	Corridas    int       `json:"corridas"`
	// SinAtribucion distingue «no se pudo asignar a esta caja» de «se asignó pero no hubo
	// dinero». Son dos ausencias distintas y la UI las dice distinto.
	SinAtribucion bool          `json:"sin_atribucion,omitempty"`
	Marcas        []MarcaDeFuga `json:"marcas,omitempty"`
}

// TurnoUnido es el resultado del join dinero×proceso por `(SesionID, TurnoID)` — igualdad
// de dos campos, sin heurística de tiempo ni de orden (ANEXO H1).
//
// `TieneDinero`/`TieneProceso` viajan explícitos porque un turno puede tener solo una de
// las dos mitades y eso es un dato, no un error.
type TurnoUnido struct {
	SesionID      string     `json:"sesion_id"`
	TurnoID       string     `json:"turno_id"`
	ArnesID       string     `json:"arnes_id,omitempty"`
	InstalacionID string     `json:"instalacion_id,omitempty"`
	CajaID        string     `json:"caja_id,omitempty"`
	CorridaID     string     `json:"corrida_id,omitempty"`
	TS            time.Time  `json:"ts"`
	Modelo        string     `json:"modelo,omitempty"`
	Tokens        Tokens     `json:"tokens"`
	Aritmetica    Aritmetica `json:"aritmetica,omitempty"`

	CostoReportadoMicros *int64 `json:"costo_reportado_micros"`
	CostoCalculadoMicros *int64 `json:"costo_calculado_micros"`
	DuracionMs           *int64 `json:"duracion_ms"`

	Atribucion   Confianza `json:"atribucion"`
	Escenario    Escenario `json:"escenario"`
	TieneDinero  bool      `json:"tiene_dinero"`
	TieneProceso bool      `json:"tiene_proceso"`
	// Resultados son los desenlaces de proceso del turno (gate, corrida, herramienta).
	Resultados []Resultado `json:"resultados,omitempty"`
	Gates      []string    `json:"gates,omitempty"`
	Rotaciones int         `json:"rotaciones"`
	// Herramientas es el conteo por herramienta usada en el turno (ANEXO H8).
	Herramientas map[string]int `json:"herramientas,omitempty"`
}

// ParidadCosto compara lo que dijo el runtime con lo que dice nuestro catálogo. Si
// divergen, no se elige uno en silencio: se muestran los dos y la divergencia
// (boundary cifra-viaja-con-su-confianza).
type ParidadCosto struct {
	ReportadoMicros *int64 `json:"reportado_micros"`
	CalculadoMicros *int64 `json:"calculado_micros"`
	// DivergenciaPct es nil cuando falta alguno de los dos: sin dos números no hay
	// divergencia que calcular, y un 0 diría «coinciden», que es otra cosa.
	DivergenciaPct *float64 `json:"divergencia_pct"`
	Completo       bool     `json:"completo"`
	SinTarifa      []string `json:"sin_tarifa,omitempty"`
}

// DetalleCaja es la 4ª tab del inspector: el desglose de UNA caja.
type DetalleCaja struct {
	CajaID     string       `json:"caja_id"`
	Nombre     string       `json:"nombre"`
	Atribuible bool         `json:"atribuible"`
	Motivo     string       `json:"motivo,omitempty"`
	Desde      time.Time    `json:"desde"`
	Hasta      time.Time    `json:"hasta"`
	Tokens     Tokens       `json:"tokens"` // los buckets nil NO serializan: «no aplica» ≠ 0
	Paridad    ParidadCosto `json:"paridad"`
	Confianza  Confianza    `json:"confianza"`
	// Turnos es una PÁGINA, no la lista completa. `TurnosTotales` dice cuántos hay de verdad
	// y `Truncado` lo declara: una degradación que no se declara es un total parcial
	// disfrazado de total (defecto C4 de la auditoría). Los agregados de arriba salen de la
	// ventana ENTERA, no de esta página.
	Turnos        []TurnoUnido           `json:"turnos"`
	TurnosTotales int                    `json:"turnos_totales"`
	Truncado      bool                   `json:"truncado"`
	Detectores    []EstadoDetector       `json:"detectores"`
	Catalogo      VersionCatalogoPrecios `json:"catalogo"`
}

// AgregadoVentana es la suma de una ventana ENTERA, calculada en el almacén.
//
// Existe para que el detalle no sume sobre la página que le devolvieron: con más turnos que el
// límite, sumar la página daría un total parcial presentado como total — y dos pantallas del
// mismo dato mostrando cifras distintas.
type AgregadoVentana struct {
	Turnos               int      `json:"turnos"`
	Tokens               Tokens   `json:"tokens"`
	CostoReportadoMicros *int64   `json:"costo_reportado_micros"`
	CostoCalculadoMicros *int64   `json:"costo_calculado_micros"`
	CostoCompleto        *bool    `json:"costo_completo"`
	SinTarifa            []string `json:"sin_tarifa,omitempty"`
}

// EstadoDetector es un detector con su veredicto de aplicabilidad. `Motivo` es obligatorio
// cuando `Aplica` es false (boundary no-aplica-no-es-cero): un detector apagado sin razón
// es un gap escondido.
type EstadoDetector struct {
	Detector DetectorID `json:"detector"`
	Nombre   string     `json:"nombre"`
	Aplica   bool       `json:"aplica"`
	Motivo   string     `json:"motivo,omitempty"`
	// CoberturaParcial marca el caso de P1 en s2-instrumentado: corre, pero no ve todo.
	// Viaja como matiz explícito, jamás como un ✅ liso.
	CoberturaParcial bool `json:"cobertura_parcial,omitempty"`
	Hallazgos        int  `json:"hallazgos"`
	// SinFix es la contraparte de la regla A4: el detector corrió, encontró algo cotizable y
	// **no puede proponer un arreglo**. No genera tarjeta —una tarjeta sin contrafactual es un
	// reproche— pero tampoco se esconde: el inspector lo lista con su motivo. Sin este campo,
	// «encontró y no sabe qué hacer» se leería igual que «no encontró nada».
	SinFix bool `json:"sin_fix,omitempty"`
}

// RespuestaMejoras lleva las TRES listas SIEMPRE (decisión A16). Omitir `no_aplican`
// obligaría al FE a elegir entre no mostrar nada (gap escondido) o mostrar 0 (mentira).
type RespuestaMejoras struct {
	Puntos    []PuntoDeMejora  `json:"puntos"`
	NoAplican []EstadoDetector `json:"no_aplican"`
	NoMedidos []EstadoDetector `json:"no_medidos"`
	Escenario Escenario        `json:"escenario"`
	// Descartados es cuántos puntos se ocultaron porque el operador los descartó (D26.4).
	// Viaja para que la lista pueda decir «hay 2 descartados» en vez de mostrarse más corta
	// sin explicación — que se leería como «no encontramos nada más».
	Descartados int `json:"descartados"`
	Ventana     struct {
		Desde time.Time `json:"desde"`
		Hasta time.Time `json:"hasta"`
	} `json:"ventana"`
}

// FilaPortafolio es una fila de la tabla del Portafolio: arnés × instalación (D20).
//
// `Puesto` es *string y NO string: sale del `rol` del arnés indexado, resuelto server-side,
// y `null` significa «este arnés no declara rol» — la UI dice «puesto sin declarar». Ese es
// el caso NORMAL hoy, no un borde: ningún arnés del dogfood declara `rol`.
type FilaPortafolio struct {
	ArnesID         string    `json:"arnes_id"`
	Clave           string    `json:"clave"`
	Nombre          string    `json:"nombre"`
	InstalacionID   string    `json:"instalacion_id"`
	Puesto          *string   `json:"puesto"`
	Empresas        []string  `json:"empresas,omitempty"`
	Corridas        int       `json:"corridas"`
	CostoMicros     *int64    `json:"costo_micros"`
	CostoPorCorrida *int64    `json:"costo_por_corrida"` // null para el que nunca corrió
	Confianza       Confianza `json:"confianza"`
	Tendencia       string    `json:"tendencia,omitempty"` // "alza" | "estable" | "baja" | ""
	// PuntosDeMejora es cuántas marcas de fuga tiene la fila. Un 0 acá SÍ es un dato:
	// «se midió y no se encontró nada», distinto de `CostoMicros: null` («no se midió»).
	PuntosDeMejora int `json:"puntos_de_mejora"`
}

// SaludTelemetria son los contadores del receptor + el estado del almacén. Se persisten
// para que sobrevivan a un reinicio: «cuántos descarté» es dato de honestidad, no un gauge
// en memoria que se pierde.
type SaludTelemetria struct {
	Recibidos               int        `json:"recibidos"`
	Aceptados               int        `json:"aceptados"`
	DescartadosColaLlena    int        `json:"descartados_cola_llena"`
	RechazadosFormato       int        `json:"rechazados_formato"`
	RechazadosTamano        int        `json:"rechazados_tamano"`
	AtributosFueraDeLista   int        `json:"atributos_fuera_de_lista"`
	TemporalidadNoSoportada int        `json:"temporalidad_no_soportada"`
	UltimaRecepcion         *time.Time `json:"ultima_recepcion"`
	AlmacenDisponible       bool       `json:"almacen_disponible"`
	AlmacenMotivo           string     `json:"almacen_motivo,omitempty"`
	TamanoBytes             int64      `json:"tamano_bytes"`
	AvisoTamano             bool       `json:"aviso_tamano"`
	HistoriaArchivada       string     `json:"historia_archivada,omitempty"`
	// RetencionDias es el TTL vigente, **firmado en 90 días** (D26.3). Sigue viniendo de la
	// config y no de una constante en la UI: firmar el default no es clavarlo.
	RetencionDias        int                    `json:"retencion_dias"`
	RollupMeses          int                    `json:"rollup_meses"`
	Catalogo             VersionCatalogoPrecios `json:"catalogo"`
	Forward              bool                   `json:"forward"`
	ForwardDestino       string                 `json:"forward_destino,omitempty"`
	IngestaTokenEstricto bool                   `json:"ingesta_token_estricto"`
	Descubrimiento       string                 `json:"descubrimiento"` // ruta publicada, o el motivo de no haberla publicado
}
