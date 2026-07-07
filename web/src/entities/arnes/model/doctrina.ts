// Diccionario DOCTRINAL de tooltips del inspector (RF-85..87, decisión #2 firmada del
// paquete inspector-drawer). Cementa UNA vez, en la entity (la dueña del dominio), el
// texto que el mockup v6 validó (mockup-drawer.html:309-360) — el widget solo consume.
// Son Maps y no objetos: las claves son títulos/valores de display (dato, no
// identificadores), fuera de la convención de naming del código.
//
// Fuente de cada definición (restricción del operador: nada redactado por fuera de la
// doctrina): METODOLOGIA §3 (contrato fusionado, 4 ejes) · §4 (honestidad, medido>estimado,
// gris≠verde) · §8.1 (arquetipo) · §8.2 (perfil T1-T3) · VISION Anatomía A1–A7 (caja/gate/
// Guardia/Base) · graph.l0.schema.json (banda · canal · origen) · knowledge/rules.md L2.6
// (alw derivado de paths:) · arch/contracts/nomenclatura-arnes.md (§4.1 fuente_path ·
// §4.5 no-reconocido).

// SEC_TIP — qué agrupa cada sección del drawer y su eje del contrato fusionado. La clave
// es el título EXACTO de la sección (RF-85).
export const SEC_TIP: ReadonlyMap<string, string> = new Map([
  [
    "Clasificación",
    "Dónde y cómo vive el nodo: banda/fase (geografía del mapa), canal (release), procedencia (honestidad del dato) y — si es caja — sus 3 ejes perpendiculares: clase ⊥ arquetipo ⊥ perfil (METODOLOGIA §8).",
  ],
  [
    "Fuente",
    "De dónde sale el nodo: el archivo real que el loader reconoció (fuente_path, nomenclatura §4.1) y quién lo puso en la instancia del arnés (origen — lo estampa el provisioner).",
  ],
  [
    "Activación",
    "Cuándo entra la regla en contexto: siempre (CLAUDE.md) o condicional (paths:). Derivado del archivo, no declarado (rules L2.6).",
  ],
  [
    "Rol",
    "Qué ES este nodo en la doctrina cuando no es caja de proceso — no tener contrato es su estado LEGAL, no un faltante.",
  ],
  [
    "Reconciliación",
    "Marcador honesto D-c (nomenclatura §4.5): el reconocedor no entendió el artefacto. Visible con warn, jamás descarte silencioso.",
  ],
  [
    "Intención",
    "Eje INTENCIÓN del contrato fusionado (METODOLOGIA §3): el porqué de la caja — qué promete.",
  ],
  [
    "Capabilities",
    "Eje INTENCIÓN: capacidades verificables — qué hace (what) y cómo se sabe que lo logró (success). Cada capability debe ser comprobable.",
  ],
  [
    "Necesita",
    "Eje CABLEADO: los insumos del contrato — qué artefactos requiere y de qué origen (usuario · caja: · base: · terceros: …).",
  ],
  [
    "Entrega",
    "Eje CABLEADO: los artefactos que produce. «Escritor único» = ningún otro componente escribe ese artefacto (A5).",
  ],
  [
    "Ruta",
    "Eje CABLEADO: a qué caja pasa el trabajo y bajo qué condición — orquestación determinista entre cajas.",
  ],
  [
    "Gate",
    "Eje ACEPTACIÓN (A4, «nada sin eval»): cómo se comprueba lo entregado. `none` es un HALLAZGO visible — gris ≠ verde (METODOLOGIA §4).",
  ],
  [
    "Handoff",
    "Eje ACEPTACIÓN: cuándo la caja se rinde y a quién escala (framed autonomy — autonomía acotada, no automatización ciega).",
  ],
  [
    "Viene de",
    "Eje CABLEADO, sentido inverso: quién invoca, lee o escribe sobre ESTE nodo. Derivado de los edges del grafo (invoca/lee/escribe), no declarado.",
  ],
  [
    "Hallazgos",
    "Honestidad del nodo (METODOLOGIA §4): hallazgos deterministas — gate:none (A4), no-reconocido (D-c) y checks rojos de conformance filtrados por nodo. Gris ≠ verde: el hueco se muestra, no se esconde.",
  ],
  [
    "Fuente del componente",
    "El artefacto real reconocido por el loader (autodocumentación como efecto): la fuente ES la verdad del componente. Lectura confinada al dir del arnés (S2).",
  ],
  [
    "Corridas donde actuó",
    "Sesiones REALES de Claude Code (JSONL de ~/.claude) en las que este nodo actuó — nada es flujo ideal: es lo que el usuario vivió. Esperan el indexer JSONL (telemetría de nacimiento, principio 9).",
  ],
])

// DEF_CAMPO — qué significa cada NOMBRE de campo (RF-86, primera mitad del tooltip).
export const DEF_CAMPO: ReadonlyMap<string, string> = new Map([
  [
    "banda",
    "Región fija de la geografía del mapa donde vive el nodo (eje ortogonal a clase y perfil).",
  ],
  [
    "fase",
    "Fase del proceso a la que pertenece — DATO que declara este arnés, no doctrina del producto (agnosticismo, VISION p3/p7).",
  ],
  [
    "estado",
    "La transición de la máquina de estados (spine) que esta caja POSEE — una caja = una transición (A2).",
  ],
  ["canal", "Canal de release del nodo en el train del kit (KIT-06)."],
  [
    "procedencia",
    "Honestidad del dato: de dónde salió lo que ves (METODOLOGIA §4 — «medido» manda sobre «estimado»).",
  ],
  ["arquetipo", "La FORMA del trabajo (METODOLOGIA §8.1) — autonomía ≠ automatización."],
  ["perfil", "CÓMO ejecuta su loop (METODOLOGIA §8.2, perfil de harness T1–T3)."],
  [
    "caja",
    "Si es caja de proceso: el frente de su fase, con contrato fusionado y una transición del spine (A1–A2).",
  ],
  [
    "fuente",
    "Ruta del artefacto real reconocido por el loader — la estampa la fábrica, no se escribe a mano (nomenclatura §4.1).",
  ],
  [
    "origen",
    "Quién puso el nodo en ESTA instancia del arnés — lo estampa el provisioner, no lo declara el autor (L0, HS-09 Fase D).",
  ],
  [
    "carga",
    "Cómo entra la regla en contexto — derivado del archivo real (CLAUDE.md vs paths:), rules L2.6.",
  ],
  ["tipo", "Tipo de eval del gate (A4)."],
])

// DEF_VALOR — qué significa el VALOR concreto de un campo (RF-86, segunda mitad).
export const DEF_VALOR: ReadonlyMap<string, ReadonlyMap<string, string>> = new Map<
  string,
  ReadonlyMap<string, string>
>([
  [
    "banda",
    new Map([
      [
        "guardia",
        "Guardia — banda superior: hooks transversales que corren en eventos del ciclo, cuidan sin bloquear.",
      ],
      ["fase", "Carriles del proceso — el nodo vive dentro de una fase del spine."],
      ["base", "Base — banda inferior: conocimiento, reglas y servicios que las skills leen (A6)."],
      ["libreria-expertos", "Librería de expertos — estante lateral, expertos bajo demanda."],
      ["meta-harness", "Meta-harness — configuración del propio arnés (permisos, estilos, salud)."],
      [
        "marcas-dormidas",
        "Marcas dormidas — instalado pero sin uso (0 corridas): candidato a podar o despertar.",
      ],
      ["terceros", "Terceros — capacidades externas al arnés."],
    ]),
  ],
  [
    "canal",
    new Map([
      ["beta", "beta — en prueba dentro del arnés; aún no es el estándar."],
      ["estable", "estable — en producción, versión vigente del release train."],
      ["propuesto", "propuesto — sugerido, aún NO aceptado por el operador."],
      ["deprecado", "deprecado — en salida; no usarlo en cableados nuevos."],
    ]),
  ],
  [
    "procedencia",
    new Map([
      ["medido", "medido — telemetría REAL; manda sobre cualquier estimación."],
      ["estimado", "estimado — proyección sin telemetría; se dibuja atenuado (gris ≠ verde)."],
      ["declarado", "declarado — lo afirmó el autor del componente; sin verificación automática."],
      ["inferido", "inferido — lo dedujo la fábrica a partir de otros datos."],
      ["no-declarado", "no-declarado — la AUSENCIA de dato hecha explícita, nunca silenciosa."],
    ]),
  ],
  [
    "arquetipo",
    new Map([
      [
        "pipeline",
        "pipeline — determinista/verificable: salida exacta, gate auto (automatización).",
      ],
      [
        "excepcion",
        "excepción — flujo feliz + casos raros: invariantes, gate en las ramas (autonomía acotada).",
      ],
      [
        "abierto",
        "abierto — generativo/relacional: acota el alcance, no los pasos (framed autonomy plena).",
      ],
      [
        "no-arnesar",
        "no-arnesar — juicio puro que no se descompone en partes contratables: queda FUERA del grafo, y saberlo es doctrina.",
      ],
    ]),
  ],
  [
    "perfil",
    new Map([
      ["T1", "T1 — una sola pasada, sin loop ni subagentes; modelo barato."],
      [
        "T2",
        "T2 — multi-paso con estado, a menudo interactivo; document-as-cache obligatorio salvo arquetipo abierto.",
      ],
      [
        "T3",
        "T3 — worker desatendido: el conductor Go posee el loop (spawn `claude -p`, repair-cap, blocked→handoff).",
      ],
    ]),
  ],
  [
    "caja",
    new Map([
      ["sí", "sí — caja de proceso: posee una transición del spine y responde por su contrato."],
      ["no", "no — componente de apoyo: sirve a una caja, no posee transición."],
    ]),
  ],
  [
    "origen",
    new Map([
      ["estandar", "estandar — proviene del kit genérico del rol."],
      [
        "del-puesto",
        "del-puesto — añadido/llenado en el onboarding del puesto (rol·proceso·empresa).",
      ],
    ]),
  ],
  [
    "carga",
    new Map([
      ["siempre en contexto", "siempre — vive en CLAUDE.md: entra en TODO contexto de la sesión."],
      [
        "condicional (paths:)",
        "condicional — carga solo cuando el trabajo toca sus paths: declarados.",
      ],
      [
        "desconocida",
        "desconocida — sin dato aún: el loader todavía no deriva la carga de esta regla.",
      ],
    ]),
  ],
  [
    "tipo",
    new Map([
      ["auto", "auto — eval automático: la fábrica comprueba sola la salida."],
      ["manual", "manual — decide un humano: el gate es una aprobación."],
      ["parcial", "parcial — auto + juicio humano."],
      ["none", "none — SIN eval formal: HALLAZGO visible a propósito (A4, gris ≠ verde)."],
    ]),
  ],
])

// PROP_NOTE — sufijo del tooltip cuando el valor viene del fixture PROPUESTA (spec §2.3).
export const PROP_NOTE =
  " · PROPUESTA: derivado de un fixture rotulado hasta que el dato real (provisioning/loader) lo estampe."

// tipDe — el tooltip completo de un campo k con valor v: definición del campo + significado
// del valor concreto; fase/estado (dato del arnés, no enum del producto) reciben la nota
// agnóstica; los valores «· PROPUESTA» añaden PROP_NOTE. "" cuando el campo no está en el
// diccionario (el widget entonces no subraya). Espejo de tipDe (mockup:353-360).
export function tipDe(k: string, v: string): string {
  const def = DEF_CAMPO.get(k)
  if (!def) return ""
  const base = v.replace(" · PROPUESTA", "")
  const vv = DEF_VALOR.get(k)?.get(base)
  let t = def
  if (vv) t += ` — ${vv}`
  else if (k === "fase") t += ` — «${base}» es una fase declarada por ESTE arnés.`
  else if (k === "estado") t += ` — esta caja transiciona «${base}».`
  if (v.includes("PROPUESTA")) t += PROP_NOTE
  return t
}
