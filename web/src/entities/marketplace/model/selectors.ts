// Selectores puros de Marketplace: sin React, sin transporte — importables desde cualquier capa
// (FSD: entities/*/model). Unit-tested por el proyecto vitest `unit` (environment node, NO
// Storybook/browser) — ver selectors.test.ts.
//
// **SON PRESENTADORES, NO CALCULADORAS** (design.md §2 C1, normativo): la situación de una fila
// del catálogo la calcula el DOMINIO GO (`CalcularSituacion` + `AccionDeSituacion`) y viaja
// resuelta en el wire. Acá solo se la presenta. Cruzar catálogo × Portafolio en el FE exigiría
// importar `EntradaPortafolio` de `entities/portafolio` = cross-import entity↔entity, que
// `steiger fsd/no-cross-imports` rompe (y `pnpm run fsd` está dentro de `verify`).

import type {
  EntradaCatalogo,
  EstadoLectura,
  MarketplaceConocido,
  SituacionCatalogo,
  TipoSituacion,
} from "./types"

// SaludVisual — alias LOCAL de esta entidad con los mismos 3 valores que `SaludPortafolio` a
// propósito: el átomo `DotSaludPortafolio` los consume, y el WIDGET es el que lo importa del
// barrel de `entities/portafolio`. **NO se importa el tipo de la otra entidad** (`steiger
// fsd/no-cross-imports`). Duplicar 3 strings es más barato que abrir un `@x` entre entidades.
export type SaludVisual = "ok" | "atencion" | "sin-senal"

// ── Situación: presentación de las 6 ramas ──────────────────────────────────────────────────

const SITUACION_ROTULO: Record<TipoSituacion, string> = {
  "no-lo-tengo": "no lo tengo",
  "al-hilo": "lo tengo · al hilo",
  "mi-copia-adelantada": "mi copia adelantada",
  "estante-adelantado": "el estante adelantado",
  "instalaciones-en-deriva": "instalaciones en deriva",
  "no-comparable": "no comparable",
}

// rotuloDeTipoSituacion — el rótulo pelado de una rama, sin el detalle. Lo usan las chips del
// panel del filtro: la MISMA palabra que la fila, para que el operador no tenga que traducir.
export function rotuloDeTipoSituacion(t: TipoSituacion): string {
  return SITUACION_ROTULO[t]
}

// etiquetaDeSituacion — el rótulo textual de la fila, con el detalle que el dominio mandó
// (versiones, cuántas instalaciones). Literales calcados del mockup firmado. Nunca teclea una
// cifra: `mia`/`estante`/`cuantas` salen del wire.
export function etiquetaDeSituacion(s: SituacionCatalogo): string {
  const base = SITUACION_ROTULO[s.tipo]
  switch (s.tipo) {
    case "mi-copia-adelantada":
    case "estante-adelantado":
      if (s.mia && s.estante) return `${base} — estante v${s.estante} · tu canónico v${s.mia}`
      return base
    case "instalaciones-en-deriva": {
      const n = s.cuantas ?? 0
      return n === 1 ? "1 instalación en deriva" : `${n} instalaciones en deriva`
    }
    default:
      return base
  }
}

// tonoDeSituacion — el tono de salud con el que se pinta la fila. **`no-comparable` JAMÁS cae en
// `ok`** (spec §1: «no sé» ≠ «sano»); `no-lo-tengo` tampoco — no tenerlo no es estar sano.
export function tonoDeSituacion(s: SituacionCatalogo): SaludVisual {
  switch (s.tipo) {
    case "al-hilo":
      return "ok"
    case "mi-copia-adelantada":
    case "estante-adelantado":
    case "instalaciones-en-deriva":
      return "atencion"
    default:
      return "sin-senal"
  }
}

// ── Lectura: «cuándo se leyó» (BR-3) sin mentir sobre la frescura ───────────────────────────

const LECTURA_ROTULO: Record<Exclude<EstadoLectura["tipo"], "leido">, string> = {
  "no-leido": "no leído aún",
  "sin-acceso": "sin acceso",
  "url-no-resuelve": "url no resuelve",
}

// haceCuanto — delta humanizado DETERMINISTA (sin Intl/locale: el mismo insumo da el mismo
// texto en cualquier máquina, requisito de story=test). `ahora` se inyecta por la misma razón.
function haceCuanto(iso: string, ahora: Date): string {
  const t = Date.parse(iso)
  if (Number.isNaN(t)) return "fecha desconocida"
  const segundos = Math.max(0, Math.floor((ahora.getTime() - t) / 1000))
  if (segundos < 60) return "menos de 1 min"
  const minutos = Math.floor(segundos / 60)
  if (minutos < 60) return `${minutos} min`
  const horas = Math.floor(minutos / 60)
  if (horas < 24) return horas === 1 ? "1 h" : `${horas} h`
  const dias = Math.floor(horas / 24)
  return dias === 1 ? "1 día" : `${dias} días`
}

// textoDeLectura — el estado de lectura tal cual, con su motivo COMPLETO cuando degradó.
// BR-3 + BR-4 juntas: si hubo una lectura exitosa vieja y AHORA no se puede leer, se dicen las
// dos cosas («leído hace 2 días · ahora sin acceso — <motivo>»). Nunca se afirma frescura que
// no hubo, y nunca se calla el motivo.
export function textoDeLectura(l: EstadoLectura, ahora: Date): string {
  if (l.tipo === "leido") {
    return l.cuando ? `leído hace ${haceCuanto(l.cuando, ahora)}` : "leído (sin fecha registrada)"
  }
  const rotulo = LECTURA_ROTULO[l.tipo]
  const conMotivo = l.motivo ? `${rotulo} — ${l.motivo}` : rotulo
  if (l.cuando) return `leído hace ${haceCuanto(l.cuando, ahora)} · ahora ${conMotivo}`
  return conMotivo
}

// esNavegable — spec §4.1.4: SOLO una fila leída navega a su catálogo. `sin-acceso`,
// `url-no-resuelve` y `no-leido` NO navegan: un catálogo vacío se leería como «este marketplace
// no tiene arneses», que es exactamente el pass fabricado que BR-4 mata.
export function esNavegable(m: MarketplaceConocido): boolean {
  return m.lectura.tipo === "leido"
}

// accionDeFila — qué botón ofrece la fila cuando no navega (`Leer catálogo` la primera vez,
// `Reintentar` cuando ya falló). Con lectura buena, la fila entera ES la acción.
export function accionDeFila(
  m: MarketplaceConocido,
): "ver-catalogo" | "leer-catalogo" | "reintentar" {
  switch (m.lectura.tipo) {
    case "leido":
      return "ver-catalogo"
    case "no-leido":
      return "leer-catalogo"
    default:
      return "reintentar"
  }
}

// ordenarMarketplaces — propios legibles → propios no legibles → referencia; dentro de cada
// grupo, por nombre ascendente (estable y determinista). No muta el array entrante.
export function ordenarMarketplaces(ms: MarketplaceConocido[]): MarketplaceConocido[] {
  const rango = (m: MarketplaceConocido): number => {
    if (m.clase !== "propio") return 2
    return esNavegable(m) ? 0 : 1
  }
  return [...ms].sort((a, b) => rango(a) - rango(b) || a.nombre.localeCompare(b.nombre))
}

// ── Catálogo: buscar y acotar (AG-D3), 100 % client-side ────────────────────────────────────

// filtrarEntradasCatalogo — buscar: substring case-insensitive sobre nombre/descripción/versión
// y el `source.crudo` (que es el dato con el que el operador reconoce un canal). Query vacía o
// solo-espacios ⇒ sin filtro (mismo convenio que `filtrarEntradas` del Portafolio).
export function filtrarEntradasCatalogo(es: EntradaCatalogo[], q: string): EntradaCatalogo[] {
  const query = q.trim().toLowerCase()
  if (query === "") return es
  return es.filter((e) => {
    const campos = [e.nombre, e.descripcion, e.version, e.source.crudo]
    return campos.some((c) => c?.toLowerCase().includes(query))
  })
}

// filtrarPorSituacion — acota por situación (afordancia DISTINTA de reagrupar). Set vacío ⇒ sin
// filtro. Multi-select: sobrevive toda fila cuya `situacion.tipo` esté en el set.
export function filtrarPorSituacion(
  es: EntradaCatalogo[],
  tipos: ReadonlySet<TipoSituacion>,
): EntradaCatalogo[] {
  if (tipos.size === 0) return es
  return es.filter((e) => tipos.has(e.situacion.tipo))
}

// ORDEN_SITUACION — orden canónico de las 6 ramas para el panel del filtro: enum cerrado ⇒
// orden fijo (más legible que «primera aparición», que haría bailar las chips entre catálogos).
const ORDEN_SITUACION: readonly TipoSituacion[] = [
  "no-lo-tengo",
  "al-hilo",
  "mi-copia-adelantada",
  "estante-adelantado",
  "instalaciones-en-deriva",
  "no-comparable",
]

// situacionesDisponibles — las situaciones REALMENTE presentes en las filas actuales, en orden
// canónico. Nunca un enum fijo de 6 chips donde 4 no matchean nada.
export function situacionesDisponibles(es: EntradaCatalogo[]): TipoSituacion[] {
  const presentes = new Set(es.map((e) => e.situacion.tipo))
  return ORDEN_SITUACION.filter((t) => presentes.has(t))
}
