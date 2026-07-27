package domain

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// telemetria_prosa.go es **el único lugar donde la prosa de un punto de mejora se arma** (D25).
//
// La regla viene firmada de `design.md` §1.3: *«el join, los detectores y el contrafactual se
// resuelven en el dominio Go y viajan resueltos en el wire»*. La implementación del Tramo B se la
// salteó —el FE tipaba `contrafactual` como prosa y el Go mandaba `ContrafactualMicros`— y eso
// dejó a la tarjeta insignia sin fuente: era el hueco V-5 de `PARIDAD.md`.
//
// Por qué acá y no en el FE: la frase necesita **el monto alternativo, la diferencia, su unidad y
// el denominador de corridas** — los cuatro salen del mismo cálculo que produjo el punto. Armarla
// en `entities` sería calcular en la UI, que es exactamente lo que §1.3 prohíbe; y sería armarla
// en un lugar donde el sesgo y el umbral que la sostienen ya no están.
//
// **El dinero también se formatea acá, y una sola vez** (RF-281). `CifraUsd` sigue siendo el único
// COMPONENTE que pinta un monto suelto; `MontoMicros` es la única FUNCIÓN que convierte micros a
// texto para la prosa. Las dos obedecen las mismas reglas —coma decimal, dos decimales que suben
// hasta que el número deja de renderizarse en cero, agrupación con espacio fino, «−» tipográfico—
// y `montos.golden.json` las ata: es la misma tabla asertada de los dos lados.

// espacioFino es U+202F (narrow no-break space), el mismo separador de miles que usa
// `agrupar()` en `web/src/entities/telemetria/model/selectors.ts`. Un espacio normal ahí
// permitiría que el número se parta en dos líneas.
const espacioFino = " "

// signoMenos es U+2212, no el guion ASCII: es el mismo que emite `usd()` en el FE.
const signoMenos = "−"

// MontoMicros formatea micros de dólar SIN prefijo de moneda — el espejo exacto de `usd()` del
// FE. Un monto que existe **nunca** sale «0,00»: los decimales suben hasta la precisión real del
// dato (6, que es lo que un micro puede decir) antes que mentir un cero.
func MontoMicros(micros int64) string {
	dolares := float64(micros) / 1_000_000
	decimales := 2
	for decimales < 6 && dolares != 0 && renderizaCero(dolares, decimales) {
		decimales++
	}
	fijo := strconv.FormatFloat(math.Abs(dolares), 'f', decimales, 64)
	entero, frac, _ := strings.Cut(fijo, ".")
	signo := ""
	if dolares < 0 {
		signo = signoMenos
	}
	return signo + agruparMiles(entero) + "," + frac
}

// UsdMicros es `MontoMicros` con el prefijo de moneda, que es como el monto entra en una frase.
func UsdMicros(micros int64) string { return "USD " + MontoMicros(micros) }

func renderizaCero(dolares float64, decimales int) bool {
	v, err := strconv.ParseFloat(strconv.FormatFloat(dolares, 'f', decimales, 64), 64)
	return err == nil && v == 0
}

func agruparMiles(entero string) string {
	if len(entero) <= 3 {
		return entero
	}
	var b strings.Builder
	pre := len(entero) % 3
	if pre > 0 {
		b.WriteString(entero[:pre])
	}
	for i := pre; i < len(entero); i += 3 {
		if b.Len() > 0 {
			b.WriteString(espacioFino)
		}
		b.WriteString(entero[i : i+3])
	}
	return b.String()
}

// Redactar completa los campos de PROSA del punto a partir de sus piezas numéricas. Es
// idempotente y no pisa nada que el detector ya haya escrito: un detector puede redactar su
// propio contrafactual si el genérico no le sirve.
//
// 🔴 **Devuelve el `Contrafactual` VACÍO cuando el ahorro no es positivo**, y eso tiene
// consecuencia: `Mejoras()` no publica ese punto como tarjeta, lo declara como detector
// «sin fix propuesto» con su motivo. Es la regla A4 aplicada donde se puede aplicar — *«una
// tarjeta sin contrafactual es un reproche, no una recomendación»* —, y un contrafactual que
// dice «con el arreglo habrías gastado MÁS» no es una recomendación de nada.
func (p PuntoDeMejora) Redactar() PuntoDeMejora {
	if p.ID == "" {
		p.ID = string(p.Detector)
		if p.CajaID != "" {
			p.ID += ":" + p.CajaID
		}
	}
	if p.Contrafactual == "" {
		p.Contrafactual = p.contrafactualEnProsa()
	}
	return p
}

// contrafactualEnProsa arma la frase que RF-249 pide: el mundo alternativo con las MISMAS
// corridas, la diferencia, y **la unidad de esa diferencia sin ambigüedad**. La iteración 1 del
// mockup decía «0,53 por corrida» sobre una caja de 14 corridas y no cerraba: por eso la frase
// nombra las dos unidades (en la ventana · por corrida, sobre N) y no elige una.
func (p PuntoDeMejora) contrafactualEnProsa() string {
	if p.BaseContrafactual == "" || p.DiferenciaMicros <= 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(p.BaseContrafactual)
	switch {
	case p.CorridasUsadas > 1:
		fmt.Fprintf(&b, ", las mismas %d corridas costaban %s",
			p.CorridasUsadas, UsdMicros(p.ContrafactualMicros))
	case p.CorridasUsadas == 1:
		fmt.Fprintf(&b, ", la misma corrida costaba %s", UsdMicros(p.ContrafactualMicros))
	default:
		fmt.Fprintf(&b, ", el mismo trabajo costaba %s", UsdMicros(p.ContrafactualMicros))
	}
	fmt.Fprintf(&b, " → diferencia %s en la ventana", UsdMicros(p.DiferenciaMicros))
	if p.CorridasUsadas > 0 {
		fmt.Fprintf(&b, " (%s por corrida, sobre %d)",
			UsdMicros(p.DiferenciaMicros/int64(p.CorridasUsadas)), p.CorridasUsadas)
	}
	b.WriteString(".")
	return b.String()
}

// MotivoSinContrafactual es lo que se dice de un detector que ENCONTRÓ algo cotizable y no puede
// proponer un arreglo. No se esconde y no se disfraza de «no encontró nada»: son cosas distintas
// y la de arriba es la cara del boundary `no-aplica-no-es-cero`.
const MotivoSinContrafactual = "encontró hallazgos pero ninguno tiene un mundo alternativo más " +
	"barato que el actual: sin ahorro positivo no hay recomendación que hacer, así que se declara " +
	"en vez de mostrar una tarjeta que reprocha sin proponer"
