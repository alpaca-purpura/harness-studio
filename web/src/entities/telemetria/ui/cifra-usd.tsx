import { cn } from "@/shared/lib/cn"
import { usd } from "../model/selectors"

// CifraUsd — **el ÚNICO lugar donde se formatea dinero** en toda la plataforma (RF-281).
//
// Vive en `entities` y no en `shared/ui` porque el formato de dinero es dominio: lee `usd()` de
// `entities/telemetria/model`, y `shared/ui/**` tiene prohibido importar `entities/*/model`
// (dependency-cruiser `ui-not-domain`, severidad `error`).
//
// ⚠️ El símbolo se llama `CifraUsd` y no `CifraUSD` como fijaban `design.md` §1.3 y **D22**:
// `biome lint/style/useNamingConvention` con `strictCase: true` (severidad `error`, dentro de
// `verify` y de CI) rechaza dos mayúsculas consecutivas en PascalCase. Se renombra el símbolo,
// NO se baja el gate — misma lógica con la que D18 rechazó apagar `fsd/no-cross-imports`.
// El archivo sigue siendo `cifra-usd.tsx` y el puntero de CAP-139 se corrige.
//
// Las tres reglas que las stories asertan:
//  1. `USD` va en su propio `<span>` atenuado — así el Portafolio puede omitirlo (está en el
//     encabezado de la columna) sin duplicar el formateador.
//  2. **`0,004` no se redondea a `0,00`.** Un monto que existe y se muestra como cero es la
//     mentira barata que este paquete existe para no decir; `usd()` sube decimales hasta que
//     el número deja de renderizarse en cero.
//  3. Sin monto ⇒ `sin dato`. **Nunca `0,00`, nunca `—` a secas** (BR-M2): un guion no dice si
//     falta el dato, si no aplica o si es cero.

export interface CifraUsdProps {
  /** Micros de dólar. `null`/`undefined` = no hay monto, y se DICE. */
  micros?: number | null | undefined
  /** El Portafolio lo omite: el `USD` vive en el encabezado de la columna (design §3.5). */
  sinPrefijo?: boolean | undefined
  /** Copy de ausencia. `sin dato` por default; el canvas usa motivos más específicos. */
  ausente?: string | undefined
  className?: string | undefined
}

export function CifraUsd({ micros, sinPrefijo, ausente = "sin dato", className }: CifraUsdProps) {
  const monto = usd(micros)
  if (monto === null) {
    return <span className={cn("num", "num-sindato", className)}>{ausente}</span>
  }
  return (
    <span className={cn("num", className)}>
      {!sinPrefijo && <span className="num-usd">USD</span>}
      <span className="num-monto">{monto}</span>
    </span>
  )
}
