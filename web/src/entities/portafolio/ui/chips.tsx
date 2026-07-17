import { Glyph } from "@/shared/canvas"
import { cn } from "@/shared/lib/cn"
import type { EstadoDeriva, SaludPortafolio, TipoInstalacion } from "../model/types"

// Chips/dot de dominio del Portafolio (T3, plan §3 T3 · G1/G9): presentacionales, CERO
// transporte. Cada uno pinta SOLO lo que el wire trae — sin dato, el widget que lo consume
// (T4-T6) decide "desconocido"/disabled; estos componentes nunca fabrican un estado que no
// recibieron (BR-8). Clases via `data-`less modifier (el propio valor del wire, p.ej.
// `en-deriva`/`atencion`) para que los tests aserten CLASE, no color (G9 exige explícito).

// ── DerivaChip — el veredicto BR-4 tal cual (S1-D14: "al-hilo"/"en-deriva"/
// "deriva-no-evaluable" literal, sin parafrasear); `deriva_detalle` es tooltip nativo (G1: la
// deriva SOLO sale de datos reales, nunca de un contador inventado). ──
interface DerivaChipProps {
  estado: EstadoDeriva
  detalle?: string | undefined
}

export function DerivaChip({ estado, detalle }: DerivaChipProps) {
  return (
    <span className={cn("pf-chip", "pf-chip-deriva", estado)} title={detalle}>
      {estado}
    </span>
  )
}

// ── TipoInstalacionChip — la forma física tal cual (S0-D2): "materializada" /
// "referenciada-cc" / "proyecto-instalado", literal del wire (mismo espíritu que DerivaChip:
// ninguna copy amigable que pueda driftar del backend). ──
interface TipoInstalacionChipProps {
  tipo: TipoInstalacion | ""
}

export function TipoInstalacionChip({ tipo }: TipoInstalacionChipProps) {
  // "" = hallazgo sin forma física que clasificar (aviso declarado-sin-dir, C-P-14/S1-D26):
  // no hay tipo que pintar — nada, no un chip vacío (mismo criterio que AvisoChip).
  if (!tipo) return null
  return <span className={cn("pf-chip", "pf-chip-tipo", tipo)}>{tipo}</span>
}

// ── AvisoChip — SOLO existe si hay `aviso` (plan §3 T3): sin dato, no renderiza nada — el
// caller no necesita un `if` propio. El texto completo es el aviso (nunca truncado: BR-8, la
// honestidad no se recorta). ──
interface AvisoChipProps {
  aviso?: string | undefined
}

export function AvisoChip({ aviso }: AvisoChipProps) {
  if (!aviso) return null
  return (
    <span className="pf-chip pf-chip-aviso" title={aviso}>
      {aviso}
    </span>
  )
}

// ── DotSaludPortafolio — la regla S1-D4 pintada (G9): CADA rama con su propia clase/color;
// "sin-senal" es un hueco/muted que NUNCA comparte el color de "ok" ("no sé" ≠ "sano" — regla
// dura). `role="img"` + `aria-label`/`title` (a11y G8): el dot es decorativo-con-significado,
// el texto vive en el atributo, no en el DOM visible. ──
const SALUD_LABEL: Record<SaludPortafolio, string> = {
  ok: "ok",
  atencion: "atención",
  "sin-senal": "sin señal",
}

interface DotSaludPortafolioProps {
  salud: SaludPortafolio
}

export function DotSaludPortafolio({ salud }: DotSaludPortafolioProps) {
  const label = `salud: ${SALUD_LABEL[salud]}`
  return <span className={cn("pf-dot-salud", salud)} role="img" aria-label={label} title={label} />
}

// ── EmblemaInicial — inicial + color determinista sobre la paleta `--c-*` YA existente (kind
// tokens, base.tokens.json:165-211 — mismos 11 tokens que entities/arnes/model/kind.ts usa
// para las clases L0; acá se reutilizan como paleta NEUTRA de distinción visual, no como
// vocabulario de Clase). `colorDeterminista` es una función PURA aparte (unit-testeable sin
// DOM): hash simple (suma de code points) → índice — el MISMO string da SIEMPRE el mismo
// color (determinista, jamás random ni Math.random). No existía un helper hash→paleta
// reusable en entities/arnes (revisado antes de escribir este); T3 lo construye acá. ──
const PALETA_EMBLEMA: readonly string[] = [
  "var(--c-skill)",
  "var(--c-agent)",
  "var(--c-hook)",
  "var(--c-knowledge)",
  "var(--c-mcp)",
  "var(--c-rule)",
  "var(--c-command)",
  "var(--c-plugin)",
  "var(--c-settings)",
  "var(--c-output-style)",
  "var(--c-statusline)",
]

export function colorDeterminista(texto: string): string {
  let hash = 0
  for (let i = 0; i < texto.length; i++) {
    hash = (hash * 31 + texto.charCodeAt(i)) | 0
  }
  const indice = Math.abs(hash) % PALETA_EMBLEMA.length
  // PALETA_EMBLEMA es no-vacía (literal fijo arriba) — el índice siempre resuelve.
  return PALETA_EMBLEMA[indice] as string
}

interface EmblemaInicialProps {
  // La fuente del color+letra: normalmente `identidad.id` o `clave` — el caller elige (T4-T6).
  // Puede ser "" (identidad provisional real, G6 — ver fixture b): se pinta "?" honesto, jamás
  // un placeholder inventado.
  texto: string
  size?: number
}

export function EmblemaInicial({ texto, size = 22 }: EmblemaInicialProps) {
  const inicial = texto.trim().charAt(0).toUpperCase() || "?"
  return <Glyph color={colorDeterminista(texto)} char={inicial} shape="circle" size={size} />
}
