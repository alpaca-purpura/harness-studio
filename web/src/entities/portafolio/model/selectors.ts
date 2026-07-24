// Selectores puros del Portafolio (S1-D6/D7): sin React, sin transporte — importables desde
// cualquier capa (FSD: entities/*/model). Unit-tested por el proyecto vitest `unit`
// (environment node, NO Storybook/browser) — ver selectors.test.ts.

import type { Candidato, EntradaPortafolio, IdentidadArnes, SaludPortafolio } from "./types"

// identificadorDe — cadena de identificación visible (S1-D26): `id` → `scope` → «(sin id)».
// Cuando no hay manifiesto, el scope ES el discriminador de la identidad provisional
// (RN-IDENT-2: ruta relativa al proyecto, o el remote del proyecto) — el backend ya lo manda
// en el wire; ocultarlo dejaba N tarjetas «(sin id)» indistinguibles (caso real luana-vitalia).
export function identificadorDe(identidad: IdentidadArnes): string {
  return identidad.id || identidad.scope || "(sin id)"
}

const GRUPO_RAIZ = "proyecto (raíz)"

// gruposCandidatosDe — lente monorepo del wizard (S1-D26): agrupa los candidatos de UN escaneo
// por la primera subcarpeta de `install_path` relativo a `proyecto_path`. Van al grupo
// «proyecto (raíz)»: la instalación del propio root, las que viven fuera del árbol del proyecto
// (cache CC, avisos sin dir físico) y las de dirs ocultos (`.claude/plugins/...` = instalación
// a nivel proyecto, no una subcarpeta del monorepo). Orden: raíz primero, resto por aparición
// (el orden del walker es estable). Puro, sin inferencia — solo reagrupa lo que el wire trae.
export function gruposCandidatosDe(cs: Candidato[]): { grupo: string; candidatos: Candidato[] }[] {
  const orden: string[] = []
  const porGrupo = new Map<string, Candidato[]>()
  for (const c of cs) {
    const grupo = grupoDeCandidato(c)
    let lista = porGrupo.get(grupo)
    if (!lista) {
      lista = []
      porGrupo.set(grupo, lista)
      orden.push(grupo)
    }
    lista.push(c)
  }
  orden.sort((a, b) => Number(b === GRUPO_RAIZ) - Number(a === GRUPO_RAIZ))
  return orden.map((grupo) => ({ grupo, candidatos: porGrupo.get(grupo) ?? [] }))
}

function grupoDeCandidato(c: Candidato): string {
  const raiz = c.instalacion.proyecto_path.replace(/\/+$/, "")
  const instalado = c.instalacion.install_path
  if (!raiz || !instalado || instalado === raiz || !instalado.startsWith(`${raiz}/`)) {
    return GRUPO_RAIZ
  }
  const primerSegmento = instalado.slice(raiz.length + 1).split("/")[0] ?? ""
  if (primerSegmento === "" || primerSegmento.startsWith(".")) return GRUPO_RAIZ
  return `${primerSegmento}/`
}

// saludDe — regla EXACTA de S1-D4 (definida + testeada, G9 «definir regla o quitar»): «no sé»
// ≠ «sano», jamás un verde fabricado.
//   atencion  ⟺ ∃ instalación con deriva "en-deriva" ∨ con aviso ∨ con discrepancias en su origen
//   ok        ⟺ ≥1 instalación y NINGUNA disparó atencion (todas al-hilo, sin aviso ni discrepancias)
//   sin-senal ⟺ el resto (todo deriva-no-evaluable mezclado sin señal de atención, o 0 instalaciones)
export function saludDe(e: EntradaPortafolio): SaludPortafolio {
  const instalaciones = e.instalaciones ?? []

  const necesitaAtencion = instalaciones.some((i) => {
    const discrepancias = i.origen.discrepancias?.length ?? 0
    return i.deriva === "en-deriva" || !!i.aviso || discrepancias > 0
  })
  if (necesitaAtencion) return "atencion"

  const todasAlHilo = instalaciones.length > 0 && instalaciones.every((i) => i.deriva === "al-hilo")
  if (todasAlHilo) return "ok"

  return "sin-senal"
}

// registriesDe — unión de entry.registries ∪ cada instalación.origen.registry, dedup, orden
// estable (primera aparición gana). Vacío ⇒ el caller pinta «desconocido» (G4/BR-3).
export function registriesDe(e: EntradaPortafolio): string[] {
  const vistos = new Set<string>()
  const out: string[] = []
  const agregar = (r: string | undefined) => {
    if (!r || vistos.has(r)) return
    vistos.add(r)
    out.push(r)
  }
  for (const r of e.registries ?? []) agregar(r)
  for (const inst of e.instalaciones ?? []) agregar(inst.origen.registry)
  return out
}

const SIN_EMPRESA = "sin empresa"

// agruparPorEmpresa — lente `empresa` (S1-D8): N:M, una entrada aparece en TANTOS grupos como
// empresas declare (nunca se elige una sola). El grupo «sin empresa» es SIEMPRE el último si
// alguna entrada no declara ninguna — jamás se infiere la empresa del path (G4).
export function agruparPorEmpresa(
  es: EntradaPortafolio[],
): { grupo: string; entradas: EntradaPortafolio[] }[] {
  const orden: string[] = []
  const porEmpresa = new Map<string, EntradaPortafolio[]>()
  const sinEmpresa: EntradaPortafolio[] = []

  for (const e of es) {
    const empresas = e.empresas ?? []
    if (empresas.length === 0) {
      sinEmpresa.push(e)
      continue
    }
    for (const emp of empresas) {
      let grupo = porEmpresa.get(emp)
      if (!grupo) {
        grupo = []
        porEmpresa.set(emp, grupo)
        orden.push(emp)
      }
      grupo.push(e)
    }
  }

  const grupos = orden.map((grupo) => ({ grupo, entradas: porEmpresa.get(grupo) ?? [] }))
  if (sinEmpresa.length > 0) grupos.push({ grupo: SIN_EMPRESA, entradas: sinEmpresa })
  return grupos
}

const SIN_PROYECTO = "sin proyecto instalado"

// agruparPorProyecto — lente `proyecto` (BACKLOG S1-D8, cerrada): N:M vía
// `instalaciones[].proyecto_path` — una entrada aparece en tantos grupos como proyectos tengan
// una instalación suya (mismo patrón N:M que agruparPorEmpresa). «sin proyecto instalado» =
// entradas con 0 instalaciones (solo canónico, nunca instalada en ningún proyecto escaneado).
export function agruparPorProyecto(
  es: EntradaPortafolio[],
): { grupo: string; entradas: EntradaPortafolio[] }[] {
  const orden: string[] = []
  const porProyecto = new Map<string, EntradaPortafolio[]>()
  const sinProyecto: EntradaPortafolio[] = []

  for (const e of es) {
    const proyectos = new Set(
      (e.instalaciones ?? []).map((i) => i.proyecto_path).filter((p): p is string => !!p),
    )
    if (proyectos.size === 0) {
      sinProyecto.push(e)
      continue
    }
    for (const p of proyectos) {
      let grupo = porProyecto.get(p)
      if (!grupo) {
        grupo = []
        porProyecto.set(p, grupo)
        orden.push(p)
      }
      grupo.push(e)
    }
  }

  const grupos = orden.map((grupo) => ({ grupo, entradas: porProyecto.get(grupo) ?? [] }))
  if (sinProyecto.length > 0) grupos.push({ grupo: SIN_PROYECTO, entradas: sinProyecto })
  return grupos
}

const SIN_MARKETPLACE = "origen desconocido"

// agruparPorMarketplace — lente `marketplace` (BACKLOG S1-D8, cerrada): N:M vía
// `registriesDe(e)` (unión de `entry.registries` + cada `instalación.origen.registry`). «origen
// desconocido» = sin ningún registry resuelto (BR-3: nunca se infiere).
export function agruparPorMarketplace(
  es: EntradaPortafolio[],
): { grupo: string; entradas: EntradaPortafolio[] }[] {
  const orden: string[] = []
  const porRegistry = new Map<string, EntradaPortafolio[]>()
  const sinRegistry: EntradaPortafolio[] = []

  for (const e of es) {
    const registries = registriesDe(e)
    if (registries.length === 0) {
      sinRegistry.push(e)
      continue
    }
    for (const r of registries) {
      let grupo = porRegistry.get(r)
      if (!grupo) {
        grupo = []
        porRegistry.set(r, grupo)
        orden.push(r)
      }
      grupo.push(e)
    }
  }

  const grupos = orden.map((grupo) => ({ grupo, entradas: porRegistry.get(grupo) ?? [] }))
  if (sinRegistry.length > 0) grupos.push({ grupo: SIN_MARKETPLACE, entradas: sinRegistry })
  return grupos
}

// filtrarEntradas — buscar (S1-D8): substring case-insensitive sobre id/nombre/descripción.
// Query vacía o solo-espacios ⇒ sin filtro (todas las entradas).
export function filtrarEntradas(es: EntradaPortafolio[], q: string): EntradaPortafolio[] {
  const query = q.trim().toLowerCase()
  if (query === "") return es
  return es.filter((e) => {
    const campos = [e.identidad.id, e.nombre, e.descripcion]
    return campos.some((c) => c?.toLowerCase().includes(query))
  })
}

// filtrarPorSalud — filtro «Estado» de la toolbar (deuda viva S1-D8, cerrada 2026-07-24):
// afordancia DISTINTA de la lente — acota la lista, nunca la reagrupa. Set vacío ⇒ sin filtro
// (mismo convenio que filtrarEntradas con query vacía). Multi-select: cualquier entrada cuya
// saludDe() esté en el set sobrevive.
export function filtrarPorSalud(
  es: EntradaPortafolio[],
  saludes: ReadonlySet<SaludPortafolio>,
): EntradaPortafolio[] {
  if (saludes.size === 0) return es
  return es.filter((e) => saludes.has(saludDe(e)))
}

// filtrarPorMarketplace — filtro «Marketplace» de la toolbar: acota por intersección entre
// registriesDe(e) y el set elegido (mismo dato que agruparPorMarketplace, distinta afordancia).
// Set vacío ⇒ sin filtro.
export function filtrarPorMarketplace(
  es: EntradaPortafolio[],
  registries: ReadonlySet<string>,
): EntradaPortafolio[] {
  if (registries.size === 0) return es
  return es.filter((e) => registriesDe(e).some((r) => registries.has(r)))
}

// marketplacesDisponibles — valores DISTINTOS de registriesDe presentes en los datos actuales,
// orden de primera aparición (mismo criterio que agruparPorMarketplace). Alimenta las chips del
// panel del filtro Marketplace — nunca un enum fijo, `registry` es N-valuado y dinámico (BR-3).
export function marketplacesDisponibles(es: EntradaPortafolio[]): string[] {
  const vistos = new Set<string>()
  const out: string[] = []
  for (const e of es) {
    for (const r of registriesDe(e)) {
      if (!vistos.has(r)) {
        vistos.add(r)
        out.push(r)
      }
    }
  }
  return out
}

// idsColisionados — GAP-2/S1-D2 (el índice del Mapa keyea por `identidad.id` pelado, deuda
// viva): dos entradas del Portafolio con el mismo id se pisarían en silencio al observarlas.
// Devuelve SOLO los ids con ≥2 claves — mapa vacío ⇒ ninguna colisión.
export function idsColisionados(es: EntradaPortafolio[]): Map<string, string[]> {
  const porId = new Map<string, string[]>()
  for (const e of es) {
    const claves = porId.get(e.identidad.id) ?? []
    claves.push(e.clave)
    porId.set(e.identidad.id, claves)
  }

  const colisiones = new Map<string, string[]>()
  for (const [id, claves] of porId) {
    if (claves.length >= 2) colisiones.set(id, claves)
  }
  return colisiones
}
